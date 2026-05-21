package events

import (
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Server → Client: Impostor mode
// =============================================================================

const (
	// ImpostorWordAssignedEvent is sent privately to each player at the start of
	// the Impostor game. It delivers the player's secret word and their role.
	// Payload: ImpostorWordAssignedPayload
	ImpostorWordAssignedEvent EventType = "impostor_word_assigned"

	// ImpostorDiscussionStartedEvent is broadcast when the input phase ends.
	// It reveals every player's submitted word for the discussion phase.
	// Payload: ImpostorDiscussionStartedPayload
	ImpostorDiscussionStartedEvent EventType = "impostor_discussion_started"

	// ImpostorVoteStartedEvent is broadcast when the discussion phase ends.
	// It opens the voting phase and lists eligible vote targets.
	// Payload: ImpostorVoteStartedPayload
	ImpostorVoteStartedEvent EventType = "impostor_vote_started"

	// ImpostorRoundResultEvent is broadcast when the vote phase ends.
	// It reveals who was eliminated, their role, and all impostors.
	// Payload: ImpostorRoundResultPayload
	ImpostorRoundResultEvent EventType = "impostor_round_result"
)

// =============================================================================
// Server → Client: Contexto Battle mode
// =============================================================================

const (
	// ContextoGuessResultEvent is sent privately to the guessing player after
	// each submission. It reports the word's semantic distance to the hidden target.
	// Payload: ContextoGuessResultPayload
	ContextoGuessResultEvent EventType = "contexto_guess_result"
)

// =============================================================================
// Server → Client: shared across multiple modes
// =============================================================================

const (
	// GameRoundStartedEvent is broadcast at the start of each round for
	// Anti-Match, Synonym Duel, and Contexto Battle modes.
	// Payload: GameRoundStartedPayload
	GameRoundStartedEvent EventType = "game_round_started"

	// GameRoundResultEvent is broadcast when a round ends for Anti-Match,
	// Synonym Duel, and Contexto Battle modes.
	// Payload: GameRoundResultPayload
	GameRoundResultEvent EventType = "game_round_result"

	// GameResultEvent is broadcast when the game is completely over, for all modes.
	// Payload: GameResultPayload
	GameResultEvent EventType = "game_result"
)

// =============================================================================
// Client → Server: game input events
// =============================================================================

const (
	// GameSubmitWordRequestEvent is sent by a player to submit a word during
	// the input phase of Impostor, Anti-Match, and Synonym Duel modes.
	// Payload: GameSubmitWordPayload
	GameSubmitWordRequestEvent EventType = "game_submit_word"

	// GameSubmitGuessRequestEvent is sent by a player to submit a guess in
	// Contexto Battle mode. Players may submit multiple guesses per round.
	// Payload: GameSubmitGuessPayload
	GameSubmitGuessRequestEvent EventType = "game_submit_guess"

	// GameSubmitVoteRequestEvent is sent by a player to cast a vote during the
	// Impostor vote phase. A nil target indicates a skip vote.
	// Payload: GameSubmitVotePayload
	GameSubmitVoteRequestEvent EventType = "game_submit_vote"
)

// =============================================================================
// Impostor mode payload types
// =============================================================================

// ImpostorRole identifies whether a player is a normal player or an impostor.
type ImpostorRole string

const (
	RoleNormal   ImpostorRole = "normal"
	RoleImpostor ImpostorRole = "impostor"
)

// ImpostorWordAssignedPayload is the payload for ImpostorWordAssignedEvent.
// Sent privately to each player at game start, during the word-reveal phase.
// ShownUntil is when the reveal phase ends and input opens.
// InputEndsAt is when the input phase closes.
type ImpostorWordAssignedPayload struct {
	Word          string       `json:"word"`
	Role          ImpostorRole `json:"role"`
	ImpostorCount int          `json:"impostor_count"`
	ShownUntil    time.Time    `json:"shown_until"`
	InputEndsAt   time.Time    `json:"input_ends_at"`
}

// ImpostorWordSubmission pairs a player with the word they submitted during
// the input phase. Used inside ImpostorDiscussionStartedPayload.
type ImpostorWordSubmission struct {
	UserId uuid.UUID `json:"user_id"`
	Word   string    `json:"word"`
}

// ImpostorDiscussionStartedPayload is the payload for ImpostorDiscussionStartedEvent.
type ImpostorDiscussionStartedPayload struct {
	Submissions      []ImpostorWordSubmission `json:"submissions"`
	DiscussionEndsAt time.Time                `json:"discussion_ends_at"`
}

// ImpostorVoteStartedPayload is the payload for ImpostorVoteStartedEvent.
type ImpostorVoteStartedPayload struct {
	Candidates []uuid.UUID `json:"candidates"` // players eligible to be voted out
	VoteEndsAt time.Time   `json:"vote_ends_at"`
}

// ImpostorVoteTally records which players voted for a given candidate.
type ImpostorVoteTally struct {
	UserId uuid.UUID   `json:"user_id"`
	Votes  []uuid.UUID `json:"votes"` // IDs of players who voted for this candidate
}

// ImpostorRoundResultPayload is the payload for ImpostorRoundResultEvent.
// Eliminated is nil when the skip vote won. NormalWord and ImpostorWord are
// revealed so players can see how close the two words were.
type ImpostorRoundResultPayload struct {
	Eliminated   *uuid.UUID          `json:"eliminated,omitempty"`
	WasImpostor  bool                `json:"was_impostor"`
	Impostors    []uuid.UUID         `json:"impostors"`
	Tallies      []ImpostorVoteTally `json:"tallies"`
	NormalWord   string              `json:"normal_word"`
	ImpostorWord string              `json:"impostor_word"`
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
	EndsAt      time.Time   `json:"ends_at"`
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
