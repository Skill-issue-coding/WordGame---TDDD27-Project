package session

import "encoding/json"

type CreatedLobbyResponsePayload struct {
	User      UserProfile `json:"user"`
	GameState LobbyState  `json:"game_state"`
	Message   string      `json:"message"`
}

type SyncStatePayload struct {
	GameState LobbyState  `json:"gamestate"`
	You       UserProfile `json:"you"`
	Message   string      `json:"message,omitempty"`
}

type JoinLobbyPayload struct {
	Code string `json:"code"`
}

// Host changes the mode from the lobby
type ChangeModePayload struct {
	Mode GameMode `json:"mode"`
}

type UpdateSettingPayload struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

type ConnectedToHubPayload struct {
	User UserProfile `json:"user"`
}
