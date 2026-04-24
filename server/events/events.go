package events

import "encoding/json"

type EventType string

// Go (server) -> Next.js (client)
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

// ParseEvent decodes a raw websocket message into an Event envelope.
func ParseEvent(message []byte) (Event, error) {
	var event Event
	err := json.Unmarshal(message, &event)
	return event, err
}

// DecodePayload decodes the event payload into the requested type.
func DecodePayload[T any](event Event) (T, error) {
	var out T
	err := json.Unmarshal(event.Payload, &out)
	return out, err
}

// EncodeEvent wraps payload in an Event envelope and encodes it to JSON.
func EncodeEvent(eventType EventType, payload any) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	updateEvent := Event{
		Type:    eventType,
		Payload: payloadBytes,
	}

	return json.Marshal(updateEvent)
}

// PrepareEvent wraps a given payload into an Event structure with the specified EventType.
// It returns the JSON-marshaled byte slice of the final event, ready to be sent over the socket.
func PrepareEvent(eventType EventType, payload any) []byte {
	finalMessage, _ := EncodeEvent(eventType, payload)
	return finalMessage
}
