package session

import (
	"log"
	"server/events"
	"server/game"
	"server/util"
	"time"

	"github.com/google/uuid"
)

const (
	// Settings matching the client-side settings for Impostor game
	IPOSTOR_COUNT_MIN                = 1
	IPOSTOR_COUNT_MAX                = 4
	IMPOSTOR_INPUT_DURATION_MIN      = 10
	IMPOSTOR_INPUT_DURATION_MAX      = 60
	IMPOSTOR_DISCUSSION_DURATION_MIN = 10
	IMPOSTOR_DISCUSSION_DURATION_MAX = 60
	IMPOSTOR_VOTE_DURATION_MIN       = 10
	IMPOSTOR_VOTE_DURATION_MAX       = 60

	// Settings matching the client-side settings for Contexto game
	CONTEXTO_WORD_TYPE_MIN      = 1
	CONTEXTO_WORD_TYPE_MAX      = 2
	CONTEXTO_ROUND_DURATION_MIN = 60
	CONTEXTO_ROUND_DURATION_MAX = 600
	CONTEXTO_ROUNDS_MIN         = 1
	CONTEXTO_ROUNDS_MAX         = 5

	// Settings matching the client-side settings for Synonym game
	SYNONYM_WORD_TYPE_MIN      = 1
	SYNONYM_WORD_TYPE_MAX      = 2
	SYNONYM_ROUND_DURATION_MIN = 10
	SYNONYM_ROUND_DURATION_MAX = 60
	SYNONYM_ROUNDS_MIN         = 1
	SYNONYM_ROUNDS_MAX         = 5

	// Settings matching the client-side settings for Anti-match game
	ANTIMATCH_ROUND_DURATION_MIN = 10
	ANTIMATCH_ROUND_DURATION_MAX = 60
	ANTIMATCH_ROUNDS_MIN         = 1
	ANTIMATCH_ROUNDS_MAX         = 5
	ANTIMATCH_DISTANCE_MIN       = 0.1
	ANTIMATCH_DISTANCE_MAX       = 1.0
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
		PlayerInputs:          make(chan PlayerInputRequest),
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

		case input := <-lobby.PlayerInputs:
			if lobby.Phase != GameStarted || lobby.Mode != ModeImpostor {
				continue
			}

			// Initialize the state if it hasn't been already
			if lobby.ImpostorState == nil {
				lobby.ImpostorState = &ImpostorGameState{
					SubmittedPlayers: []uuid.UUID{},
				}
			}

			// Check if the user has already submitted to avoid duplicates
			alreadySubmitted := false
			for _, id := range lobby.ImpostorState.SubmittedPlayers {
				if id == input.UserID {
					alreadySubmitted = true
					break
				}
			}

			if !alreadySubmitted {
				// Record the submission
				lobby.ImpostorState.SubmittedPlayers = append(lobby.ImpostorState.SubmittedPlayers, input.UserID)

				// TODO: You will also want to store the actual `input.Word` somewhere!
				// e.g., lobby.ImpostorState.PlayerClues[input.UserID] = input.Word

				// 2. Broadcast the updated state to all clients
				lobby.SyncStateToClients()

				// Optional: Check if ALL players have submitted.
				// If they have, you can automatically advance the phase to "discuss".
				// if len(lobby.ImpostorState.SubmittedPlayers) == len(lobby.Clients) { ... }
			}
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
	state := LobbyState{
		Code:     lobby.ID,
		Mode:     lobby.Mode,
		Phase:    lobby.Phase,
		Host:     lobby.Host,
		Users:    lobby.Users,
		Settings: lobby.ModeSettings(),
	}

	if lobby.Mode == ModeImpostor && lobby.ImpostorState != nil {
		state.GameState = lobby.ImpostorState
	}

	return state
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

func (lobby *GameLobby) StartLobby(client *Client) {
	// TODO: Add host, setting and other checks
	for client := range lobby.Clients {
		client.SendEvent(events.GameStartedEvent, nil)
	}
}

// ApplySetting updates a specific setting for the currently active game mode
// based on the provided key and value.
func (lobby *GameLobby) ApplySetting(key string, value float64) {
	switch lobby.Mode {
	case ModeImpostor:
		switch key {
		case "input_duration":
			lobby.ImpostorSettings.InputDuration = util.ClampInt(value, IMPOSTOR_INPUT_DURATION_MIN, IMPOSTOR_INPUT_DURATION_MAX)
		case "discussion_duration":
			lobby.ImpostorSettings.DiscussionDuration = util.ClampInt(value, IMPOSTOR_DISCUSSION_DURATION_MIN, IMPOSTOR_DISCUSSION_DURATION_MAX)
		case "impostor_count":
			lobby.ImpostorSettings.ImpostorCount = util.ClampInt(value, IPOSTOR_COUNT_MIN, IPOSTOR_COUNT_MIN)
		case "vote_duration":
			lobby.ImpostorSettings.VoteDuration = util.ClampInt(value, IMPOSTOR_VOTE_DURATION_MIN, IMPOSTOR_DISCUSSION_DURATION_MAX)
		}
	case ModeContextoBattle:
		switch key {
		case "round_duration":
			lobby.ContextoBattleSettings.RoundDuration = util.ClampInt(value, CONTEXTO_ROUND_DURATION_MIN, CONTEXTO_ROUND_DURATION_MAX)
		case "rounds":
			lobby.ContextoBattleSettings.Rounds = util.ClampInt(value, CONTEXTO_ROUNDS_MIN, CONTEXTO_ROUNDS_MAX)
		case "word_type":
			lobby.ContextoBattleSettings.WordType = util.ClampInt(value, CONTEXTO_WORD_TYPE_MIN, CONTEXTO_WORD_TYPE_MAX)
		}
	case ModeSynonymDuel:
		switch key {
		case "round_duration":
			lobby.SynonymDuelSettings.RoundDuration = util.ClampInt(value, SYNONYM_ROUND_DURATION_MIN, SYNONYM_ROUNDS_MAX)
		case "rounds":
			lobby.SynonymDuelSettings.Rounds = util.ClampInt(value, SYNONYM_ROUNDS_MIN, SYNONYM_ROUNDS_MAX)
		case "word_type":
			lobby.SynonymDuelSettings.WordType = util.ClampInt(value, SYNONYM_WORD_TYPE_MIN, SYNONYM_WORD_TYPE_MAX)
		}
	case ModeAntiMatch:
		switch key {
		case "input_duration":
			lobby.AntiMatchSettings.InputDuration = util.ClampInt(value, ANTIMATCH_ROUND_DURATION_MIN, ANTIMATCH_ROUND_DURATION_MAX)
		case "max_distance":
			lobby.AntiMatchSettings.MaxDistance = util.ClampFloat(value, ANTIMATCH_DISTANCE_MIN, ANTIMATCH_DISTANCE_MAX)
		case "rounds":
			lobby.AntiMatchSettings.Rounds = util.ClampInt(value, SYNONYM_ROUNDS_MIN, SYNONYM_ROUNDS_MAX)
		}
	}
}
