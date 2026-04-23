package session

// Next.js (client) -> Go (server)
type CreateLobbyPayload struct {
	Username string `json:"username"`
	// TODO: Add gamesettings?
}

// Go (server) -> Next.js (client)
type CreatedLobbyResponsePayload struct {
	GameState BaseState `json:"gamestate"`
	User      User      `json:"user"`
	Message   string    `json:"message"`
}
