package events

import "encoding/json"

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

type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// PrepareEvent wraps a given payload into an Event structure with the specified EventType.
// It returns the JSON-marshaled byte slice of the final event, ready to be sent over the socket.
func PrepareEvent(eventType EventType, payload any) []byte {
	payloadBytes, _ := json.Marshal(payload)

	updateEvent := Event{
		Type:    eventType,
		Payload: payloadBytes,
	}

	finalMessage, _ := json.Marshal(updateEvent)

	return finalMessage
}
