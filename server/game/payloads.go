package game

import (
	"github.com/google/uuid"
)

type GamePhasePayload struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

// =============================================================================
// Impostor mode payload types
// =============================================================================

// ImpostorRole identifies whether a player is a normal player or an impostor.
type ImpostorGameRoles string

const (
	ImpostorRoleNormal   ImpostorGameRoles = "normal"
	ImpostorRoleImpostor ImpostorGameRoles = "impostor"
)

// =============================================================================
// Server -> Client
// =============================================================================

type ImpostorClientGameState struct {
	Role                   ImpostorGameRoles      `json:"role"`
	Word                   string                 `json:"word"`
	ActivePlayers          map[uuid.UUID]bool     `json:"active_players"`
	PreviousSubmittedWords map[uuid.UUID][]string `json:"previous_submitted_words,omitempty"`
}

// =============================================================================
// Client → Server payload types
// =============================================================================

// GameSubmitWordPayload is the payload for GameSubmitWordRequestEvent.
type GameSubmitWordPayload struct {
	Word string `json:"word"`
}

// GameSubmitGuessPayload is the payload for GameSubmitGuessRequestEvent.
type GameSubmitGuessPayload struct {
	Word string `json:"word"`
}

// GameSubmitVotePayload is the payload for GameSubmitVoteRequestEvent.
// A nil Target indicates a skip vote (no elimination this round).
type GameSubmitVotePayload struct {
	Target *uuid.UUID `json:"target"`
}
