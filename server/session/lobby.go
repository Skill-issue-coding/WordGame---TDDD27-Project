package session

import (
	"log"
	"server/events"
	"server/game"
	"server/util"

	"github.com/google/uuid"
)

type GameSetting string

const (
	INPUT_DURATION      GameSetting = "input_duration"
	DISCUSSION_DURATION GameSetting = "discussion_duration"
	IMPOSTOR_COUNT      GameSetting = "impostor_count"
	VOTE_DURATION       GameSetting = "vote_duration"
	ROUND_DURATION      GameSetting = "round_duration"
	NUMBER_OF_ROUNDS    GameSetting = "rounds"
	WORD_TYPE           GameSetting = "word_type"
	MAX_DISTANCE        GameSetting = "max_distance"
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
		StartGameRequests:     make(chan *Client),
		GameInputs:            make(chan game.GameInput, 16),
		GameDone:              make(chan struct{}, 1),
	}
	lobby.SetMode(ModeImpostor)
	return lobby
}

// Run is the lobby's main event loop. It must be started in its own goroutine
// and is the only place where lobby state is mutated, making all field access
// implicitly single-threaded and safe without additional locking.
func (lobby *GameLobby) Run() {
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

		case client := <-lobby.StartGameRequests:
			if client.UserId != lobby.Host {
				client.SendError("Endast hosten kan starta spelet.")
				continue
			}

			if lobby.Phase == GameStarted {
				client.SendError("Spelet har redan startat.")
				continue
			}

			// Snapshot send functions — game goroutine must not read lobby.Clients directly.
			sendFns := make(map[uuid.UUID]func(events.EventType, any), len(lobby.Clients))
			for c := range lobby.Clients {
				fn := c.SendEvent // capture by value
				sendFns[c.UserId] = fn
			}

			notify := func(id uuid.UUID, t events.EventType, p any) {
				if fn, ok := sendFns[id]; ok {
					fn(t, p)
				}
			}
			broadcast := func(t events.EventType, p any) {
				for _, fn := range sendFns {
					fn(t, p)
				}
			}
			onDone := func() { lobby.GameDone <- struct{}{} }

			players := make([]uuid.UUID, 0, len(lobby.Users))
			for id := range lobby.Users {
				players = append(players, id)
			}

			switch lobby.Mode {
			case ModeImpostor:
				lobby.CurrentGame = game.NewImpostorGame(lobby.ImpostorSettings, players, &client.Hub.Dictionary, notify, broadcast, onDone)
			case ModeAntiMatch:
				lobby.CurrentGame = game.NewAntimatchGame(lobby.AntiMatchSettings, notify, broadcast, onDone)
			}
			lobby.Phase = GameStarted
			go lobby.CurrentGame.Run()

		case <-lobby.GameDone:
			lobby.CurrentGame = nil
			lobby.Phase = LobbyPhase
			lobby.SyncStateToClients()
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
		lobby.AntiMatchSettings = game.DefaultAntiMatchSettings()
	}
}

// ApplySetting updates a specific setting for the currently active game mode
// based on the provided key and value.
func (lobby *GameLobby) ApplySetting(key GameSetting, value float64) {
	switch lobby.Mode {
	case ModeImpostor:
		switch key {
		case INPUT_DURATION:
			lobby.ImpostorSettings.InputDuration = util.ClampInt(value, game.IMPOSTOR_INPUT_DURATION_MIN, game.IMPOSTOR_INPUT_DURATION_MAX)
		case DISCUSSION_DURATION:
			lobby.ImpostorSettings.DiscussionDuration = util.ClampInt(value, game.IMPOSTOR_DISCUSSION_DURATION_MIN, game.IMPOSTOR_DISCUSSION_DURATION_MAX)
		case IMPOSTOR_COUNT:
			lobby.ImpostorSettings.ImpostorCount = util.ClampInt(value, game.IMPOSTOR_COUNT_MIN, game.IMPOSTOR_COUNT_MIN)
		case VOTE_DURATION:
			lobby.ImpostorSettings.VoteDuration = util.ClampInt(value, game.IMPOSTOR_VOTE_DURATION_MIN, game.IMPOSTOR_DISCUSSION_DURATION_MAX)
		}
	case ModeContextoBattle:
		switch key {
		case ROUND_DURATION:
			lobby.ContextoBattleSettings.RoundDuration = util.ClampInt(value, game.CONTEXTO_ROUND_DURATION_MIN, game.CONTEXTO_ROUND_DURATION_MAX)
		case NUMBER_OF_ROUNDS:
			lobby.ContextoBattleSettings.Rounds = util.ClampInt(value, game.CONTEXTO_ROUNDS_MIN, game.CONTEXTO_ROUNDS_MAX)
		case WORD_TYPE:
			lobby.ContextoBattleSettings.WordType = util.ClampInt(value, game.CONTEXTO_WORD_TYPE_MIN, game.CONTEXTO_WORD_TYPE_MAX)
		}
	case ModeSynonymDuel:
		switch key {
		case ROUND_DURATION:
			lobby.SynonymDuelSettings.RoundDuration = util.ClampInt(value, game.SYNONYM_ROUND_DURATION_MIN, game.SYNONYM_ROUNDS_MAX)
		case NUMBER_OF_ROUNDS:
			lobby.SynonymDuelSettings.Rounds = util.ClampInt(value, game.SYNONYM_ROUNDS_MIN, game.SYNONYM_ROUNDS_MAX)
		case WORD_TYPE:
			lobby.SynonymDuelSettings.WordType = util.ClampInt(value, game.SYNONYM_WORD_TYPE_MIN, game.SYNONYM_WORD_TYPE_MAX)
		}
	case ModeAntiMatch:
		switch key {
		case INPUT_DURATION:
			lobby.AntiMatchSettings.InputDuration = util.ClampInt(value, game.ANTIMATCH_ROUND_DURATION_MIN, game.ANTIMATCH_ROUND_DURATION_MAX)
		case MAX_DISTANCE:
			lobby.AntiMatchSettings.MaxDistance = util.ClampFloat(value, game.ANTIMATCH_DISTANCE_MIN, game.ANTIMATCH_DISTANCE_MAX)
		case NUMBER_OF_ROUNDS:
			lobby.AntiMatchSettings.Rounds = util.ClampInt(value, game.SYNONYM_ROUNDS_MIN, game.SYNONYM_ROUNDS_MAX)
		}
	}
}
