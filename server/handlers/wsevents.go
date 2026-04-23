package handlers

type EventType string

// Go (server) -> React (client)
const (
	GameCreatedEvent EventType = "lobby_created"
	JoinedGameEvent  EventType = "joined_lobby"
	ErrorEvent       EventType = "error"
	SuccessEvent     EventType = "success"
)

// Next.js (client) -> Go (server)
const (
	CreateGameEvent EventType = "create_lobby"
	JoinGameEvent   EventType = "join_lobby"
)
