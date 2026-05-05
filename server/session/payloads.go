package session

// Go (server) -> Next.js (client)
type SyncStatePayload struct {
	GameState LobbyState `json:"lobbystate"`
	Message   string     `json:"message,omitempty"`
}

type ConnectedToHubPayload struct {
	User UserProfile `json:"user"`
}

// Next.js (client) -> Go (server)
type JoinLobbyPayload struct {
	Code string `json:"code"`
}

type UpdateUserPayload struct {
	Username   string `json:"username,omitempty"`
	Background string `json:"background,omitempty"`
}
