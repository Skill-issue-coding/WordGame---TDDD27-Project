package game

import (
	"github.com/google/uuid"
)

type NewGamePhasePayload struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

// =============================================================================
// Impostor mode payload types
// =============================================================================

// ImpostorRole identifies whether a player is a normal player or an impostor.
type ImpostorRole string

const (
	RoleNormal   ImpostorRole = "normal"
	RoleImpostor ImpostorRole = "impostor"
)

// =============================================================================
// Server -> Client
// =============================================================================

// ImpostorWordAssignedPayload is the payload for ImpostorWordAssignedEvent.
// Sent privately to each player at game start, during the word-reveal phase.
// ShownUntil is when the reveal phase ends and input opens.
// InputEndsAt is when the input phase closes.
type ImpostorBaseClientStatePayload struct {
	Role                   ImpostorRole         `json:"role"`
	ClientWord             string               `json:"word"`
	PreviousSubmittedWords map[uuid.UUID]string `json:"previous_submitted_words,omitempty"`
}

// ImpostorDiscussionStartedPayload is broadcast when the input phase ends.
type ImpostorDiscussionStartedPayload struct {
	Submissions map[uuid.UUID]string `json:"submissions"`
}

// ImpostorVoteStartedPayload is broadcast when the discussion phase ends.
type ImpostorVoteStartedPayload struct {
	Candidates []uuid.UUID `json:"candidates"`
}

// ImpostorRoundResultPayload is the payload for ImpostorRoundResultEvent.
// Eliminated is nil when the vote was a tie or everyone skipped. NormalWord and
// ImpostorWord are always revealed so players can judge how close the pair was.
// GameOver signals whether the game has ended; ImpostorsWin indicates who won.
type ImpostorRoundResultPayload struct {
	Eliminated   *uuid.UUID                `json:"eliminated,omitempty"`
	WasImpostor  bool                      `json:"was_impostor"`
	Impostors    []uuid.UUID               `json:"impostors"`
	VoteResults  map[uuid.UUID][]uuid.UUID `json:"vote_results"`
	NormalWord   string                    `json:"normal_word"`
	ImpostorWord string                    `json:"impostor_word"`
	GameOver     bool                      `json:"game_over"`
	ImpostorsWin bool                      `json:"impostors_win,omitempty"`
}

// =============================================================================
// Contexto Battle mode payload types
// =============================================================================

// ContextoGuessResultPayload is the payload for ContextoGuessResultEvent.
// Sent privately after each guess. Rank 1 means the guess is the target word.
type ContextoGuessResultPayload struct {
	Word     string  `json:"word"`
	Distance float64 `json:"distance"`
	Rank     int     `json:"rank,omitempty"`
}

// =============================================================================
// Shared multi-mode payload types
// =============================================================================

// PlayerRoundSubmission captures a single player's submission within a round.
// Duplicate is set by Anti-Match when two or more players submit the same word.
type PlayerRoundSubmission struct {
	UserId    uuid.UUID `json:"user_id"`
	Word      string    `json:"word"`
	Distance  float64   `json:"distance"`
	Points    int       `json:"points"`
	Duplicate bool      `json:"duplicate,omitempty"`
}

// GameRoundStartedPayload is the payload for GameRoundStartedEvent.
//
//   - TotalRounds is 0 for Synonym Duel Battle Royale (unknown number of rounds).
//   - TargetWord is empty for Contexto Battle (target is hidden until round end).
//   - Players lists the remaining participants; used by Synonym Duel to track
//     who has been eliminated.
type GameRoundStartedPayload struct {
	Round       int         `json:"round"`
	TotalRounds int         `json:"total_rounds,omitempty"`
	TargetWord  string      `json:"target_word,omitempty"`
	EndsAt      int64       `json:"ends_at"`
	Players     []uuid.UUID `json:"players,omitempty"`
}

// GameRoundResultPayload is the payload for GameRoundResultEvent.
//
//   - Eliminated is set only for Synonym Duel (the player farthest from target).
//   - RoundWinner is the player with the closest (or most unique closest) word.
//   - TargetWord is always revealed here, even for Contexto Battle.
type GameRoundResultPayload struct {
	Submissions []PlayerRoundSubmission `json:"submissions"`
	Eliminated  *uuid.UUID              `json:"eliminated,omitempty"`
	RoundWinner *uuid.UUID              `json:"round_winner,omitempty"`
	TargetWord  string                  `json:"target_word"`
}

// PlayerFinalScore captures a player's end-of-game rank and total score.
type PlayerFinalScore struct {
	UserId uuid.UUID `json:"user_id"`
	Score  int       `json:"score"`
	Rank   int       `json:"rank"`
}

// GameResultPayload is the payload for GameResultEvent.
type GameResultPayload struct {
	Winner uuid.UUID          `json:"winner"`
	Scores []PlayerFinalScore `json:"scores"`
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
