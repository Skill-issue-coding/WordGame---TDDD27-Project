package events

type CreateLobbyPayload struct {
	Username string `json:"username"`
	// TODO: Add gamesettings?
}
