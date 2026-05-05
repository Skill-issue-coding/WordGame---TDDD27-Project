package session

import (
	"log"
	"server/events"
	"server/game"
	"time"

	"github.com/google/uuid"
)

func NewLobby(id string) *GameLobby {
	lobby := &GameLobby{
		ID:           id,
		Clients:      make(map[*Client]bool),
		Broadcast:    make(chan []byte),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		SyncRequests: make(chan struct{}),
		Phase:        LobbyPhase,
		Users:        make(map[uuid.UUID]*UserProfile),
	}
	lobby.SetMode(ModeImpostor)
	return lobby
}

func (lobby *GameLobby) Run() {
	gameticker := time.NewTicker(1 * time.Second)
	defer gameticker.Stop()

	for {
		select {
		case client := <-lobby.Register:
			if lobby.Phase == GameStarted {
				client.SendError("Spelet har redan börjat, kan inte ansluta")
				continue
			}

			lobby.Clients[client] = true
			client.Lobby = lobby

			state := lobby.BuildLobbyState()

			for c := range lobby.Clients {
				_, exists := lobby.Users[c.UserId]
				if !exists {
					continue
				}

				if c == client {
					c.SendEvent(events.SyncGameStateEvent, SyncStatePayload{
						GameState: state,
						Message:   "Du gick med i spelet!",
					})
				} else {
					c.SendEvent(events.SyncGameStateEvent, SyncStatePayload{
						GameState: state,
					})
				}
			}

			client.SendEvent(events.JoinedLobbyEvent, nil)
			log.Printf("[Room %s] Player '%s' joined. Players in room: %d", lobby.ID, client.Username(), len(lobby.Clients))

		case client := <-lobby.Unregister:
			if _, exists := lobby.Clients[client]; !exists {
				continue
			}

			delete(lobby.Clients, client)
			delete(lobby.Users, client.UserId)

			log.Printf("[Room %s] Player '%s' left. Players remaining: %d", lobby.ID, client.Username(), len(lobby.Clients))

			if len(lobby.Clients) == 0 {
				log.Printf("[Room %s] Room is empty, closing.", lobby.ID)
				if client.Hub != nil {
					client.Hub.DeleteRoom(lobby.ID)
				}
				return
			}

			// Assign a new host if the host left
			if lobby.Host == client.UserId {
				for remaining := range lobby.Clients {
					lobby.Host = remaining.UserId
					break
				}
			}

			lobby.SyncStateToClients()

		case <-lobby.SyncRequests:
			lobby.SyncStateToClients()

		case <-gameticker.C:
			// Placeholder tick for game mode timers.
		}
	}
}

func (lobby *GameLobby) SyncStateToClients() {
	state := lobby.BuildLobbyState()
	for client := range lobby.Clients {
		_, exists := lobby.Users[client.UserId]
		if !exists {
			continue
		}
		client.SendEvent(events.SyncGameStateEvent, SyncStatePayload{
			GameState: state,
		})
	}
}

func (lobby *GameLobby) BuildLobbyState() LobbyState {
	return LobbyState{
		Code:     lobby.ID,
		Mode:     lobby.Mode,
		Phase:    lobby.Phase,
		Host:     lobby.Host,
		Users:    lobby.Users,
		Settings: lobby.ModeSettings(),
	}
}

func (lobby *GameLobby) ModeSettings() any {
	switch lobby.Mode {
	case ModeImpostor:
		return lobby.ImpostorSettings
	case ModeContextoBattle:
		return lobby.ContextoBattleSettings
	case ModeSynonymDuel:
		return lobby.SynonymDuelSettings
	case ModeAntiMatch:
		return lobby.AntiMatchSettings
	default:
		return nil
	}
}

func (lobby *GameLobby) SetMode(mode GameMode) {
	lobby.Mode = mode
	switch mode {
	case ModeImpostor:
		lobby.ImpostorSettings = game.DefaultImpostorSettings()
	case ModeContextoBattle:
		lobby.ContextoBattleSettings = game.DefaultContextoBattleSettings()
	case ModeSynonymDuel:
		lobby.SynonymDuelSettings = game.DefaultSynonymDuelSettings()
	case ModeAntiMatch:
		lobby.AntiMatchSettings = game.DefaultAntiHiveSettings()
	}
}
