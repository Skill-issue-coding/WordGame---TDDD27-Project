package game

import (
	"server/events"
	"sync"
	"time"

	"github.com/google/uuid"
)

const SYNC_DELAY = 2 * time.Second

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

// GameBase can be embedded in any game struct to satisfy HandleInput, Stop,
// StartTime, and EndTime from the Game interface, leaving only Run to implement.
// It also provides Broadcast and Send helpers to cut down on output boilerplate.
type GameBase struct {
	startTime     time.Time
	endTime       time.Time
	outputs       chan GameOutput
	onDone        func()
	inputs        chan GameInput
	stop          chan struct{}
	once          sync.Once
	skipPhaseVote chan struct{}
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

func newGameBase(outputs chan GameOutput, onDone func()) GameBase {
	return GameBase{
		outputs:       outputs,
		onDone:        onDone,
		inputs:        make(chan GameInput, 16),
		stop:          make(chan struct{}),
		skipPhaseVote: make(chan struct{}),
	}
}

func (b *GameBase) StartTime() time.Time { return b.startTime }
func (b *GameBase) EndTime() time.Time   { return b.endTime }

// Resets startTime and endTime when a new phase starts
func (b *GameBase) StartPhase(duration int) {
	now := time.Now()
	b.startTime = now
	b.endTime = now.Add((time.Duration(duration) * time.Second) + SYNC_DELAY)
}

// HandleInput forwards the input to the game's internal channel.
// Safe to call from any goroutine; blocks if the input channel is full.
func (b *GameBase) HandleInput(input GameInput) {
	b.inputs <- input
}

// Stop signals the Run goroutine to exit. Safe to call multiple times.
func (b *GameBase) Stop() {
	b.once.Do(func() { close(b.stop) })
}

// Broadcast sends an event to all players in the lobby.
func (b *GameBase) Broadcast(eventType events.EventType, payload any) {
	b.outputs <- GameOutput{Type: eventType, Payload: payload}
}

// Send sends an event to a single player identified by target.
func (b *GameBase) Send(target *uuid.UUID, eventType events.EventType, payload any) {
	b.outputs <- GameOutput{Target: target, Type: eventType, Payload: payload}
}

// Broadcasts the current phase timers to all players in the lobby
func (b *GameBase) SendPhaseTimes() {
	b.Broadcast(events.GameNewPhaseEvent, NewGamePhasePayload{StartTime: b.startTime.UnixMilli(), EndTime: b.endTime.UnixMilli()})
}
