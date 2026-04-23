package game

import (
	"server/session"
	"time"

	"github.com/google/uuid"
)

// BaseState holds the data EVERY game mode shares
type BaseState struct {
	Mode          string                        `json:"mode"`
	Phase         string                        `json:"phase"` // e.g., "starting", "playing", "voting", "ended"
	TimeRemaining float64                       `json:"timeRemaining"`
	Players       map[uuid.UUID]*session.Player `json:"players"` // List of Player IDs
}

// Subject to change
type ImpostorState struct {
	BaseState
	ImpostorID   string            `json:"impostorId"`
	SecretWord   string            `json:"secretWord"`
	PlayerInputs map[string]string `json:"playerInputs"`
}
type RoyaleState struct {
	BaseState
	TargetWord      string             `json:"targetWord"`
	PlayerDistances map[string]float64 `json:"playerDistances"`
	Eliminated      []string           `json:"eliminated"`
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
