// Package events defines the WebSocket event protocol between the Go server
// and the Next.js client. All messages are JSON objects with a "type" string
// and a "payload" object.
//
// Direction conventions:
//   - Server → Client: events the backend emits (ConnectedEvent, SyncGameStateEvent, …)
//   - Client → Server: events the frontend sends (CreateLobbyEvent, JoinLobbyEvent, …)
package events

import "encoding/json"

// EventType is the discriminator field used to route incoming and outgoing
// WebSocket messages. Values are stable strings shared between Go and TypeScript.
type EventType string

// Server → Client event types.
const (
	// ConnectedEvent is sent once immediately after a client connects to the hub.
	// Payload: ConnectedToHubPayload — contains the server-generated user profile.
	ConnectedEvent EventType = "connected_to_hub"

	// JoinedLobbyEvent is sent to the joining client after they are registered
	// into a lobby. The frontend uses this to navigate to the lobby view.
	// Payload: nil.
	JoinedLobbyEvent EventType = "joined_lobby"

	// SyncGameStateEvent is the primary state-sync event. It is broadcast to all
	// lobby clients whenever any shared state changes (player joined/left, profile
	// updated, settings changed, phase changed).
	// Payload: SyncStatePayload.
	SyncGameStateEvent EventType = "sync_gamestate"

	// ErrorEvent delivers a human-readable Swedish error message to the client.
	// Payload: { "message": string }
	ErrorEvent EventType = "error"

	// SuccessEvent delivers a human-readable Swedish success message to the client.
	// Payload: { "message": string }
	SuccessEvent EventType = "success"
)

// Client → Server event types.
const (
	// CreateLobbyEvent requests creation of a new lobby. The client is
	// automatically registered as host. Payload: nil.
	CreateLobbyEvent EventType = "create_lobby"

	// JoinLobbyEvent requests joining an existing lobby by room code.
	// Payload: JoinLobbyPayload.
	JoinLobbyEvent EventType = "join_lobby"

	// UpdateUserEvent requests a profile update (username and/or background).
	// The server mutates the client's shared UserProfile pointer and triggers
	// a SyncGameStateEvent broadcast to the lobby.
	// Payload: UpdateUserPayload.
	UpdateUserEvent EventType = "update_user"
)

// Event is the wire envelope for all WebSocket messages. Both directions use
// this structure: a type discriminator and a raw JSON payload that is decoded
// separately based on the type.
type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ParseEvent decodes a raw WebSocket message into an Event envelope.
// The Payload field is left as raw JSON for subsequent typed decoding
// via DecodePayload.
func ParseEvent(message []byte) (Event, error) {
	var event Event
	err := json.Unmarshal(message, &event)
	return event, err
}

// DecodePayload decodes the raw JSON payload of an event into the requested
// concrete type T. Use after ParseEvent when the event type is known.
//
// Example:
//
//	payload, err := events.DecodePayload[JoinLobbyPayload](event)
func DecodePayload[T any](event Event) (T, error) {
	var out T
	err := json.Unmarshal(event.Payload, &out)
	return out, err
}

// EncodeEvent serialises the given payload and wraps it in an Event envelope
// with the specified type. Returns the final JSON bytes or an error.
func EncodeEvent(eventType EventType, payload any) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Event{Type: eventType, Payload: payloadBytes})
}

// PrepareEvent is a convenience wrapper around EncodeEvent that returns the
// JSON-encoded event bytes ready to be written to a WebSocket connection.
func PrepareEvent(eventType EventType, payload any) ([]byte, error) {
	return EncodeEvent(eventType, payload)
}
