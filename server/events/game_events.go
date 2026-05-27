package events

// =============================================================================
// Server → Client: Impostor mode
// =============================================================================

const (
	// ImpostorNewRoundEvent is sent privately to each player at the start of
	// the Impostor game. It delivers the player's secret word, role, and timers.
	// Payload: ImpostorClientGameStatePayload
	ImpostorNewRoundEvent EventType = "impostor_new_round"

	// ImpostorNewPhaseEvent is sent to all players when the phase or current turn advances.
	// Payload: ImpostorGamePhaseUpdatePayload
	ImpostorNewPhaseEvent EventType = "impostor_new_phase"

	// ImpostorVoteResultEvent is sent when a vote phase is over
	// Payload: ImpostorVoteResultPayload
	ImpostorVoteResultEvent EventType = "impostor_vote_result"

	// ImpostorNewCycleEvent is sent once a cycle is completed.
	// Payload: ImpostorGameCycleUpdatePayload
	ImpostorNewCycleEvent EventType = "impostor_new_cycle"

	// ImpostorVoteUpdateEvent is broadcast immediately when any player casts a
	// vote, so clients can show live vote statistics during the vote phase.
	// Payload: ImpostorVoteUpdatePayload
	ImpostorVoteUpdateEvent EventType = "impostor_vote_update"

	// ImpostorResultEvent is sent when an impostor game is finished.
	// TODO: Payload: GameResultPayload
	ImpostorResultEvent EventType = "impostor_game_result"
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

	// GameNewPhaseEvent is broadcast once a game is in a new phase.
	// Payload: GamePhasePayload
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
