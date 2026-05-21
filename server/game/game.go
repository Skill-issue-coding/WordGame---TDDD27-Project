package game

import (
	"server/events"
	"time"

	"github.com/google/uuid"
)

// Game is the interface that all game modes must implement. A game is created
// by the lobby when the host starts the session, runs in its own goroutine,
// and is responsible for all game-phase logic and player communication.
type Game interface {
	// Run starts the game's main event loop in its own goroutine. It is
	// responsible for timers, phase transitions, and processing inputs until
	// the game ends or Stop is called.
	Run()

	// HandleInput delivers a player action to the game. It is called from the
	// lobby's Run goroutine, so implementations must forward the input to an
	// internal channel rather than processing it inline.
	HandleInput(GameInput)

	// Stop signals the game to clean up and exit its Run goroutine. It is
	// called by the lobby when the room is torn down or the game ends early.
	Stop()

	// StartTime returns the time at which Run was invoked. Zero if not yet started.
	StartTime() time.Time

	// EndTime returns the time at which the game finished. Zero if still running.
	EndTime() time.Time
}

// GameTimestamps can be embedded in any game struct to satisfy the StartTime
// and EndTime methods of the Game interface.
type GameTimestamps struct {
	startTime time.Time
	endTime   time.Time
}

func (t *GameTimestamps) StartTime() time.Time { return t.startTime }
func (t *GameTimestamps) EndTime() time.Time   { return t.endTime }

const SYNC_DELAY = 2 * time.Second

func (t *GameTimestamps) StartPhase(duration int) {
	now := time.Now()
	t.startTime = now
	t.endTime = now.Add((time.Duration(duration) * time.Second) + SYNC_DELAY)
}

// GameInput carries a single player action from the lobby to the active game.
// The lobby receives raw events from clients and wraps them here before
// forwarding to Game.HandleInput.
type GameInput struct {
	// ClientId identifies which player sent the action.
	ClientId uuid.UUID

	// Event is the raw parsed event from the client's WebSocket message.
	Event events.Event
}

// GameOutput carries an event from the game back to the lobby for delivery.
// A nil Target broadcasts to all players; a non-nil Target sends privately to
// the identified player only.
type GameOutput struct {
	Target  *uuid.UUID
	Type    events.EventType
	Payload any
}
