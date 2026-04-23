package game

import (
	"server/lobby"
	"time"

	"github.com/google/uuid"
)

type BaseGameState struct {
	Host *lobby.Client
}

type GameMode interface {
	// Called When The Host Starts The Game
	Start()

	// Called when a player types a word and hits enter
	HandleInput(playerID uuid.UUID, word string)

	// Called every second (or tick) to handle time limits (e.g., 2 or 3 seconds)
	Tick(dt time.Duration)

	// Returns the current state of the game to be broadcasted to the React frontend
	GetState() interface{}
}
