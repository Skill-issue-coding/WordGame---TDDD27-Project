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
// default settings. The lobby's Run goroutine is NOT started here — the
// caller (GameHub.CreateUniqueRoom) is responsible for calling go lobby.Run().
func NewLobby(id string) *GameLobby {
	lobby := &GameLobby{
		ID:           id,
		Clients:      make(map[*Client]bool),
		Broadcast:    make(chan []byte),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		ChatMessages: make(chan ChatMessage),
		SyncRequests: make(chan struct{}, 8),
		Phase:        LobbyPhase,
		Users:        make(map[uuid.UUID]*UserProfile),
	}
	lobby.SetMode(ModeImpostor)
	return lobby
}

// Run is the lobby's main event loop. It must be started in its own goroutine
// and is the only place where lobby state is mutated, making all field access
// implicitly single-threaded and safe without additional locking.
//
// The loop handles four cases:
//   - Register: a new client joins the lobby.
//   - Unregister: a client leaves or disconnects.
//   - SyncRequests: an external signal (e.g. profile update) requesting a broadcast.
//   - gameticker: a 1-second tick reserved for future game-phase timer logic.
//
// Run exits when the last player leaves, at which point it calls DeleteRoom
// on the hub to clean up the lobby entry.
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

			// Build state once and broadcast to all clients. The joining client
			// receives an extra Message field to confirm they successfully joined.
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

			// Notify the client's frontend to navigate to the lobby view.
			client.SendEvent(events.JoinedLobbyEvent, nil)
			log.Printf("[Room %s] Player '%s' joined. Players in room: %d", lobby.ID, client.Username(), len(lobby.Clients))

		case client := <-lobby.Unregister:
			if _, exists := lobby.Clients[client]; !exists {
				continue
			}

			delete(lobby.Clients, client)
			delete(lobby.Users, client.UserId)

			log.Printf("[Room %s] Player '%s' left. Players remaining: %d", lobby.ID, client.Username(), len(lobby.Clients))

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

		// SyncRequests is a lightweight signal used by events such as
		// update_user that mutate shared profile data outside the lobby
		// goroutine. Receiving here ensures the sync happens inside Run,
		// where all reads of lobby state are safe.
		case <-lobby.SyncRequests:
			lobby.SyncStateToClients()

		case message := <-lobby.ChatMessages:
			for client := range lobby.Clients {
				client.SendEvent(events.SendChatMessageEvent, message)
			}

		case <-gameticker.C:
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
// state, ready to be serialised and sent to clients. The Settings field is
// populated from the currently active mode's settings struct via ModeSettings.
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
// Only the struct matching lobby.Mode is meaningful; the others hold defaults.
// The returned value is embedded in LobbyState.Settings and serialised as a
// concrete typed JSON object for the frontend.
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
// to the mode's defaults. Any previously customised settings for that mode
// are discarded. Should only be called from within the lobby's Run goroutine.
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
