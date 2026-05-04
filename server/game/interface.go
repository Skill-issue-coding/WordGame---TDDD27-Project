package game

import (
	"time"

	"github.com/google/uuid"
)

type GameMode interface {
	// Called When The Host Starts The Game
	Start()

	// Called when a player types a word and hits enter
	HandleInput(playerID uuid.UUID, word string)

	// Called every second (or tick) to handle time limits (e.g., 2 or 3 seconds)
	Tick(dt time.Duration)

	// Returns the full authoritative state (server internal use).
	GetStateInternal() interface{}

	// Returns a client-facing view of state for one specific player.
	GetStateForClient(viewerID uuid.UUID) interface{}
}
