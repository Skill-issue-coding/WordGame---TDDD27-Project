package game

import (
	"math/rand/v2"
	"server/events"
	"server/words"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ImpostorPhase string

const (
	SHOW_WORD_DURATION = 8 // seconds

	PhaseShowWord   ImpostorPhase = "show_word"
	PhaseInput      ImpostorPhase = "input"
	PhaseDiscussion ImpostorPhase = "discussion"
	PhaseVote       ImpostorPhase = "vote"
	PhaseResult     ImpostorPhase = "result"

	// Settings matching the client-side settings for Impostor game
	IMPOSTOR_COUNT_MIN               int = 1
	IMPOSTOR_COUNT_MAX               int = 4
	IMPOSTOR_INPUT_DURATION_MIN      int = 10
	IMPOSTOR_INPUT_DURATION_MAX      int = 60
	IMPOSTOR_DISCUSSION_DURATION_MIN int = 10
	IMPOSTOR_DISCUSSION_DURATION_MAX int = 60
	IMPOSTOR_VOTE_DURATION_MIN       int = 10
	IMPOSTOR_VOTE_DURATION_MAX       int = 60
)

type ImpostorSettings struct {
	InputDuration      int `json:"input_duration"`
	DiscussionDuration int `json:"discussion_duration"`
	ImpostorCount      int `json:"impostor_count"`
	VoteDuration       int `json:"vote_duration"`
}

// ImpostorPair holds the two words for a round: the word shown to normal
// players and the semantically similar word shown to impostors.
type ImpostorPair struct {
	NormalWord   string
	ImpostorWord string
}

// ImpostorGame implements the Impostor game mode. Normal players are assigned
// a secret word; impostors receive a semantically similar but distinct word.
// After the input phase players discuss and vote to eliminate the suspected
// impostor. Run must be started in its own goroutine via go game.Run().
type ImpostorGame struct {
	GameTimestamps

	// settings holds the host-configured durations and impostor count for
	// this game instance.
	settings ImpostorSettings

	// dict is the loaded word dictionary, used by pickImpostorPair on startup
	// to select the word pair for the round.
	dict *words.Dictionary

	// players is the ordered list of all player IDs participating in the game.
	// Order is stable throughout the game and used when building vote tallies.
	players []uuid.UUID

	// impostors is the set of player IDs assigned the impostor role.
	// Populated by pickImpostors at the start of Run.
	impostors map[uuid.UUID]struct{}

	// wordPair holds the two secret words chosen for this round.
	wordPair ImpostorPair

	// submissions maps each player ID to the word they submitted during
	// PhaseInput. Last write wins if a player submits more than once.
	submissions map[uuid.UUID]string

	// votes maps each voter's player ID to their chosen vote target.
	// A nil pointer value means the player cast a skip vote.
	votes map[uuid.UUID]*uuid.UUID

	// outputs is the channel shared with the lobby. The game writes GameOutput
	// values here; the lobby's Run goroutine drains them and delivers to clients.
	outputs chan GameOutput

	// onDone is called once when the Run goroutine exits, regardless of reason.
	// It signals the lobby to reset back to LobbyPhase.
	onDone func()

	// inputs is the channel through which the lobby forwards player actions to
	// the game's Run goroutine. Buffered to avoid blocking the lobby loop.
	inputs chan GameInput

	// stop is closed by Stop to signal the Run goroutine to exit cleanly.
	stop chan struct{}

	// once guards the close of stop so Stop is safe to call multiple times.
	once sync.Once

	// phase tracks the current lifecycle stage of the game.
	phase ImpostorPhase
}

// DefaultImpostorSettings returns the settings applied when a lobby first
// selects the Impostor mode. These are used as the baseline before the host
// makes any manual adjustments via UpdateSettingsRequestEvent.
func DefaultImpostorSettings() ImpostorSettings {
	return ImpostorSettings{
		InputDuration:      30,
		DiscussionDuration: 15,
		ImpostorCount:      1,
		VoteDuration:       30,
	}
}

// NewImpostorGame constructs an ImpostorGame ready to be started with Run.
// players must contain all participant IDs for this session; it is used to
// assign impostor roles and to build vote tallies.
// dict is the loaded word dictionary from which the word pair is drawn at
// game start; it must remain valid for the lifetime of the game.
// notify sends an event to a single player by UUID.
// broadcast sends an event to every player in the lobby.
// onDone is invoked once when the Run goroutine exits (for any reason) so the
// lobby can reset back to LobbyPhase.
func NewImpostorGame(
	settings ImpostorSettings,
	players []uuid.UUID,
	dict *words.Dictionary,
	outputs chan GameOutput,
	onDone func(),
) *ImpostorGame {
	return &ImpostorGame{
		settings:    settings,
		dict:        dict,
		players:     players,
		outputs:     outputs,
		onDone:      onDone,
		inputs:      make(chan GameInput, 16),
		stop:        make(chan struct{}),
		phase:       PhaseShowWord,
		submissions: make(map[uuid.UUID]string),
		votes:       make(map[uuid.UUID]*uuid.UUID),
	}
}

// pickImpostors randomly selects count player IDs from players to act as
// impostors. The original slice is not mutated — a copy is shuffled internally.
// Returns a set (map[uuid.UUID]bool) for O(1) role lookups during word
// assignment. If count >= len(players) all players become impostors; the
// caller is responsible for enforcing a sensible upper bound via settings.
func pickImpostors(players []uuid.UUID, count int) map[uuid.UUID]struct{} {
	shuffled := make([]uuid.UUID, len(players))
	copy(shuffled, players)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	impostors := make(map[uuid.UUID]struct{}, count)
	for i := 0; i < count && i < len(shuffled); i++ {
		impostors[shuffled[i]] = struct{}{}
	}
	return impostors
}

// pickImpostorPair selects a random target from dict.Targets whose
// ImpostorCandidates list is non-empty, then randomly picks one candidate as
// the impostor word. ImpostorCandidates are pre-validated by the preprocessing
// pipeline (stage 9) to be semantically close but clearly distinct from the
// target, ensuring the impostor's word is plausible but not identical.
// Returns (ImpostorPair, true) on success, or (ImpostorPair{}, false) if no
// eligible target exists — for example when preprocessing has not been run.
func pickImpostorPair(dict *words.Dictionary) (ImpostorPair, bool) {
	eligible := make([]words.Target, 0, len(dict.Targets))
	for _, t := range dict.Targets {
		if len(t.ImpostorCandidates) > 0 {
			eligible = append(eligible, t)
		}
	}
	if len(eligible) == 0 {
		return ImpostorPair{}, false
	}
	target := eligible[rand.IntN(len(eligible))]
	impostorWord := target.ImpostorCandidates[rand.IntN(len(target.ImpostorCandidates))]
	return ImpostorPair{NormalWord: target.Word, ImpostorWord: impostorWord}, true
}

// Run starts the game's main event loop and must be called in its own goroutine.
// On startup it picks a word pair via pickImpostorPair and assigns impostor
// roles via pickImpostors, then privately delivers ImpostorWordAssignedEvent
// to each player containing their word, role, and the phase timestamps
// (ShownUntil, InputEndsAt) so the client can render countdown timers.
// The loop then steps through phases driven by a one-second ticker:
//
//	PhaseShowWord → PhaseInput → PhaseDiscussion → PhaseVote → PhaseResult
//
// If no eligible word pair exists in the dictionary, Run exits immediately and
// calls onDone, leaving players on the game-started screen. onDone is always
// called on exit to signal the lobby to reset back to LobbyPhase.
func (g *ImpostorGame) Run() {
	g.StartPhase(SHOW_WORD_DURATION)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer g.onDone()

	pair, ok := pickImpostorPair(g.dict)
	if !ok {
		return
	}
	g.wordPair = pair
	g.impostors = pickImpostors(g.players, g.settings.ImpostorCount)

	// shownUntil = end of ShowWord phase; inputEndsAt = end of the following Input phase.
	shownUntil := g.endTime
	inputEndsAt := shownUntil.Add(SYNC_DELAY + time.Duration(g.settings.InputDuration)*time.Second)

	for _, playerID := range g.players {
		word := g.wordPair.NormalWord
		role := events.RoleNormal
		_, isImpostor := g.impostors[playerID]
		if isImpostor {
			word = g.wordPair.ImpostorWord
			role = events.RoleImpostor
		}
		id := playerID
		g.outputs <- GameOutput{
			Target: &id,
			Type:   events.ImpostorWordAssignedEvent,
			Payload: events.ImpostorWordAssignedPayload{
				Word:          word,
				Role:          role,
				ImpostorCount: g.settings.ImpostorCount,
				ShownUntil:    shownUntil,
				InputEndsAt:   inputEndsAt,
			},
		}
	}

	for {
		select {
		case <-g.stop:
			return
		case input := <-g.inputs:
			g.processInput(input)
		case <-ticker.C:
			if !time.Now().Before(g.endTime) {
				g.advancePhase()
			}
		}
	}
}

// advancePhase transitions the game to the next phase once the current phase's
// deadline has passed. It is called exclusively from the Run ticker goroutine,
// so all field access is implicitly single-threaded. Each transition starts the
// new phase timer via StartPhase and broadcasts the appropriate event:
//   - PhaseShowWord → PhaseInput: no broadcast; clients already hold InputEndsAt.
//   - PhaseInput → PhaseDiscussion: broadcasts ImpostorDiscussionStartedEvent
//     with the full submission list and the discussion deadline.
//   - PhaseDiscussion → PhaseVote: broadcasts ImpostorVoteStartedEvent with the
//     candidate list and the vote deadline.
//   - PhaseVote → PhaseResult: calls broadcastResult then stops the game.
func (g *ImpostorGame) advancePhase() {
	switch g.phase {
	case PhaseShowWord:
		g.phase = PhaseInput
		g.StartPhase(g.settings.InputDuration)
	case PhaseInput:
		g.phase = PhaseDiscussion
		g.StartPhase(g.settings.DiscussionDuration)
		subs := make([]events.ImpostorWordSubmission, 0, len(g.submissions))
		for id, word := range g.submissions {
			subs = append(subs, events.ImpostorWordSubmission{UserId: id, Word: word})
		}
		g.outputs <- GameOutput{
			Type: events.ImpostorDiscussionStartedEvent,
			Payload: events.ImpostorDiscussionStartedPayload{
				Submissions:      subs,
				DiscussionEndsAt: g.endTime,
			},
		}
	case PhaseDiscussion:
		g.phase = PhaseVote
		g.StartPhase(g.settings.VoteDuration)
		candidates := make([]uuid.UUID, len(g.players))
		copy(candidates, g.players)
		g.outputs <- GameOutput{
			Type: events.ImpostorVoteStartedEvent,
			Payload: events.ImpostorVoteStartedPayload{
				Candidates: candidates,
				VoteEndsAt: g.endTime,
			},
		}
	case PhaseVote:
		g.phase = PhaseResult
		g.broadcastResult()
		g.Stop()
	case PhaseResult:
		g.Stop()
	}
}

// broadcastResult tallies the votes collected during PhaseVote and broadcasts
// ImpostorRoundResultEvent to all players. The eliminated player is the one
// with a strict majority of votes (most votes with no tie). A tie or zero total
// votes results in no elimination — eliminated is nil in the payload. The full
// impostor list and both secret words (NormalWord and ImpostorWord) are always
// revealed so players can judge how close the pair was.
func (g *ImpostorGame) broadcastResult() {
	votersByCandidate := make(map[uuid.UUID][]uuid.UUID, len(g.players))
	for _, id := range g.players {
		votersByCandidate[id] = []uuid.UUID{}
	}
	for voter, target := range g.votes {
		if target == nil {
			continue
		}
		votersByCandidate[*target] = append(votersByCandidate[*target], voter)
	}

	// Find the eliminated player: most votes, no tie, at least one vote.
	var eliminated *uuid.UUID
	maxVotes := 0
	for id, voters := range votersByCandidate {
		if len(voters) > maxVotes {
			maxVotes = len(voters)
			id := id
			eliminated = &id
		}
	}
	tieCount := 0
	for _, voters := range votersByCandidate {
		if len(voters) == maxVotes {
			tieCount++
		}
	}
	if maxVotes == 0 || tieCount > 1 {
		eliminated = nil
	}

	var wasImpostor bool
	if eliminated != nil {
		_, wasImpostor = g.impostors[*eliminated]
	}

	impostors := make([]uuid.UUID, 0, len(g.impostors))
	for id := range g.impostors {
		impostors = append(impostors, id)
	}

	tallies := make([]events.ImpostorVoteTally, 0, len(g.players))
	for _, id := range g.players {
		tallies = append(tallies, events.ImpostorVoteTally{
			UserId: id,
			Votes:  votersByCandidate[id],
		})
	}

	g.outputs <- GameOutput{
		Type: events.ImpostorRoundResultEvent,
		Payload: events.ImpostorRoundResultPayload{
			Eliminated:   eliminated,
			WasImpostor:  wasImpostor,
			Impostors:    impostors,
			Tallies:      tallies,
			NormalWord:   g.wordPair.NormalWord,
			ImpostorWord: g.wordPair.ImpostorWord,
		},
	}
}

// processInput handles a single player action forwarded from HandleInput.
// It is called exclusively from the Run goroutine, so field access is safe
// without additional locking. Behaviour is phase-dependent:
//   - PhaseInput: accepts GameSubmitWordRequestEvent. Records the player's word
//     in submissions; a second submission from the same player overwrites the
//     first (last write wins).
//   - PhaseVote: accepts GameSubmitVoteRequestEvent. Records the player's vote
//     target in votes; nil target means the player chose to skip. Last write
//     wins if a player votes more than once before the deadline.
//
// Inputs received during any other phase, or with unexpected event types, are
// silently dropped.
func (g *ImpostorGame) processInput(input GameInput) {
	switch g.phase {
	case PhaseInput:
		if input.Event.Type != events.GameSubmitWordRequestEvent {
			return
		}
		payload, err := events.DecodePayload[events.GameSubmitWordPayload](input.Event)
		if err != nil || payload.Word == "" {
			return
		}
		g.submissions[input.ClientId] = payload.Word
	case PhaseVote:
		if input.Event.Type != events.GameSubmitVoteRequestEvent {
			return
		}
		payload, err := events.DecodePayload[events.GameSubmitVotePayload](input.Event)
		if err != nil {
			return
		}
		g.votes[input.ClientId] = payload.Target
	}
}

// HandleInput is called from the lobby's Run goroutine — forwards to internal channel.
func (g *ImpostorGame) HandleInput(input GameInput) {
	g.inputs <- input
}

// Stop signals the game goroutine to exit. Safe to call multiple times.
func (g *ImpostorGame) Stop() {
	g.once.Do(func() { close(g.stop) })
}
