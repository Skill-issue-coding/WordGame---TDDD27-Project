package events

// =============================================================================
// Server → Client: Impostor mode
// =============================================================================

const (
	// ImpostorWordAssignedEvent is sent privately to each player at the start of
	// the Impostor game. It delivers the player's secret word and their role.
	// Payload: ImpostorWordAssignedPayload
	ImpostorNewRoundEvent EventType = "impostor_new_round"

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

	// GameNewPhaseEvent  is broadcast once a game is in a new phase
	// Payload: StartTime int64, EndTime int64
	GameNewPhaseEvent EventType = "new_game_phase"
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
