package session

import (
	"log"
	"server/events"
	"server/game"
	"time"

	"github.com/google/uuid"
)

// NewLobby creates and returns a new GameLobby with the given room code.
// All channels are initialised and the mode is set to ModeImpostor with
// default settings. The caller is responsible for starting the lobby.Run() goroutine.
func NewLobby(id string) *GameLobby {
	lobby := &GameLobby{
		ID:                    id,
		Clients:               make(map[*Client]bool),
		Broadcast:             make(chan []byte),
		Register:              make(chan *Client),
		Unregister:            make(chan *Client),
		ModeUpdateRequests:    make(chan GameMode),
		SettingUpdateRequests: make(chan UpdateSettingPayload),
		ChatMessages:          make(chan ChatMessage),
		SyncRequests:          make(chan struct{}, 8),
		Phase:                 LobbyPhase,
		Users:                 make(map[uuid.UUID]*UserProfile),
	}
	lobby.SetMode(ModeImpostor)
	return lobby
}

// Run is the lobby's main event loop. It must be started in its own goroutine
// and is the only place where lobby state is mutated, making all field access
// implicitly single-threaded and safe without additional locking.
func (lobby *GameLobby) Run() {
	gameTicker := time.NewTicker(1 * time.Second)
	defer gameTicker.Stop()

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
				if _, exists := lobby.Users[c.UserId]; !exists {
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

			client.SendEvent(events.LeftLobbyEvent, nil)

			// If the lobby is now empty, shut it down.
			if len(lobby.Clients) == 0 {
				log.Printf("[Room %s] Room is empty, closing.", lobby.ID)
				if client.Hub != nil {
					client.Hub.DeleteRoom(lobby.ID)
				}
				return
			}

			// If the host left, promote an arbitrary remaining player.
			if lobby.Host == client.UserId {
				for remaining := range lobby.Clients {
					lobby.Host = remaining.UserId
					break
				}
			}

			lobby.SyncStateToClients()

		case <-lobby.SyncRequests:
			lobby.SyncStateToClients()

		case mode := <-lobby.ModeUpdateRequests:
			lobby.SetMode(mode)
			lobby.SyncStateToClients()

		case update := <-lobby.SettingUpdateRequests:
			lobby.ApplySetting(update.Key, update.Value)
			lobby.SyncStateToClients()

		case message := <-lobby.ChatMessages:
			for client := range lobby.Clients {
				client.SendEvent(events.SendChatMessageEvent, message)
			}

		case <-gameTicker.C:
			// Reserved for game-phase countdown timers and round management.
		}
	}
}

// SyncStateToClients broadcasts the current LobbyState to every client in the
// lobby. It should only be called from within the lobby's Run goroutine.
func (lobby *GameLobby) SyncStateToClients() {
	state := lobby.BuildLobbyState()
	for client := range lobby.Clients {
		if _, exists := lobby.Users[client.UserId]; !exists {
			continue
		}
		client.SendEvent(events.SyncGameStateEvent, SyncStatePayload{
			GameState: state,
		})
	}
}

// BuildLobbyState assembles a point-in-time snapshot of the lobby's shared
// state, ready to be serialised and sent to clients.
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

// ModeSettings returns the settings struct for the currently active game mode.
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

// SetMode switches the lobby to the given game mode and resets its settings
// to the mode's defaults.
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

// ApplySetting updates a specific setting for the currently active game mode
// based on the provided key and value.
func (lobby *GameLobby) ApplySetting(key string, value float64) {
	switch lobby.Mode {
	case ModeImpostor:
		switch key {
		case "input_duration":
			lobby.ImpostorSettings.InputDuration = int(value)
		case "discussion_duration":
			lobby.ImpostorSettings.DiscussionDuration = int(value)
		case "impostor_count":
			lobby.ImpostorSettings.ImpostorCount = int(value)
		}
	case ModeContextoBattle:
		if key == "round_duration" {
			lobby.ContextoBattleSettings.RoundDuration = int(value)
		}
	case ModeSynonymDuel:
		switch key {
		case "round_duration":
			lobby.SynonymDuelSettings.RoundDuration = int(value)
		case "rounds":
			lobby.SynonymDuelSettings.Rounds = int(value)
		}
	case ModeAntiMatch:
		switch key {
		case "input_duration":
			lobby.AntiMatchSettings.InputDuration = int(value)
		case "max_distance":
			lobby.AntiMatchSettings.MaxDistance = value
		}
	}
}
