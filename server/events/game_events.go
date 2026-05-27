package events

// =============================================================================
// Server → Client: Impostor mode — events unique to this mode
// =============================================================================

const (
	// ImpostorVoteResultEvent is sent when a vote phase is over.
	// Payload: ImpostorVoteResultPayload
	ImpostorVoteResultEvent EventType = "impostor_vote_result"

	// ImpostorNewCycleEvent is sent once a cycle is completed.
	// Payload: ImpostorGameCycleUpdatePayload
	ImpostorNewCycleEvent EventType = "impostor_new_cycle"

	// ImpostorVoteUpdateEvent is broadcast immediately when any player casts a
	// vote, so clients can show live vote statistics during the vote phase.
	// Payload: ImpostorVoteUpdatePayload
	ImpostorVoteUpdateEvent EventType = "impostor_vote_update"
)

// =============================================================================
// Server → Client: canonical lifecycle events — used by all game modes.
// Each mode sends these with its own payload shape; the client uses
// lobbyState.mode to know how to parse the payload.
// =============================================================================

const (
	// GameRoundStartedEvent is sent at game/round start. For Impostor it is
	// delivered privately per-player; for other modes it is broadcast.
	// Payload shape varies by mode (see each mode's payloads file).
	GameRoundStartedEvent EventType = "game_round_started"

	// GameRoundResultEvent is broadcast when a round ends with per-player scores.
	// Payload shape varies by mode.
	GameRoundResultEvent EventType = "game_round_result"

	// GameResultEvent is broadcast when the game is completely over.
	// Payload shape varies by mode.
	GameResultEvent EventType = "game_result"

	// GameNewPhaseEvent is broadcast when the active phase changes.
	// Payload shape varies by mode.
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
