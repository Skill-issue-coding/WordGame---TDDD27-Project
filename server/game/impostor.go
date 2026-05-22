package game

import (
	"math/rand/v2"
	"server/events"
	"server/words"
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
	GameBase

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
	submissions map[uuid.UUID][]string

	// votes maps each voter's player ID to their chosen vote target.
	// A nil pointer value means the player cast a skip vote.
	votes map[uuid.UUID][]*uuid.UUID

	// phase tracks the current lifecycle stage of the game.
	phase ImpostorPhase

	// cycleNumber keeps track of how many cycles the game has completed
	cycleNumber int8
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
		GameBase:    newGameBase(outputs, onDone),
		settings:    settings,
		dict:        dict,
		players:     players,
		phase:       PhaseShowWord,
		submissions: make(map[uuid.UUID][]string),
		votes:       make(map[uuid.UUID][]*uuid.UUID),
		cycleNumber: 0,
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
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer g.onDone()

	pair, ok := pickImpostorPair(g.dict)
	if !ok {
		return
	}
	g.wordPair = pair
	g.impostors = pickImpostors(g.players, g.settings.ImpostorCount)
	g.StartPhase(SHOW_WORD_DURATION)
	g.sendBaseState()

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
		subs := make(map[uuid.UUID]string)
		for playerId, words := range g.submissions {
			word := ""
			if int(g.cycleNumber) < len(words) {
				word = words[g.cycleNumber]
			}
			subs[playerId] = word
		}
		g.Broadcast(events.ImpostorDiscussionStartedEvent, ImpostorDiscussionStartedPayload{Submissions: subs})
		g.SendPhaseTimes()

	case PhaseDiscussion:
		g.phase = PhaseVote
		g.StartPhase(g.settings.VoteDuration)
		candidates := make([]uuid.UUID, len(g.players))
		copy(candidates, g.players)
		g.Broadcast(events.ImpostorVoteStartedEvent, ImpostorVoteStartedPayload{Candidates: candidates})
		g.SendPhaseTimes()

	case PhaseVote:
		over, _ := g.broadcastResult()
		if over {
			g.Stop()
		} else {
			g.cycleNumber++
			g.phase = PhaseShowWord
			g.StartPhase(SHOW_WORD_DURATION)
			g.sendBaseState()
		}
	}
}

// broadcastResult tallies the votes collected during PhaseVote and broadcasts
// ImpostorRoundResultEvent to all players. The eliminated player is the one
// with a strict majority of votes (most votes with no tie). A tie or zero total
// votes results in no elimination — eliminated is nil in the payload. The full
// impostor list and both secret words (NormalWord and ImpostorWord) are always
// revealed so players can judge how close the pair was.
// broadcastResult tallies votes for the current cycle, removes the eliminated
// player from active state, checks the win condition, and broadcasts the result.
// Returns (gameOver, impostorsWin) so the caller can decide whether to start a
// new cycle or stop the game.
func (g *ImpostorGame) broadcastResult() (gameOver bool, impostorsWin bool) {
	votersByCandidate := make(map[uuid.UUID][]uuid.UUID, len(g.players))
	for _, id := range g.players {
		votersByCandidate[id] = []uuid.UUID{}
	}
	for voter, cycleVotes := range g.votes {
		if int(g.cycleNumber) >= len(cycleVotes) {
			continue
		}
		target := cycleVotes[g.cycleNumber]
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

	if eliminated != nil {
		delete(g.impostors, *eliminated)
		g.players = removePlayer(g.players, *eliminated)
	}

	over, impostWin := g.isGameOver()

	impostors := make([]uuid.UUID, 0, len(g.impostors))
	for id := range g.impostors {
		impostors = append(impostors, id)
	}

	g.Broadcast(events.ImpostorRoundResultEvent, ImpostorRoundResultPayload{
		Eliminated:   eliminated,
		WasImpostor:  wasImpostor,
		Impostors:    impostors,
		VoteResults:  votersByCandidate,
		NormalWord:   g.wordPair.NormalWord,
		ImpostorWord: g.wordPair.ImpostorWord,
		GameOver:     over,
		ImpostorsWin: impostWin,
	})

	return over, impostWin
}

// isGameOver returns whether the game should end and, if so, who won.
// Normal players win when no impostors remain; impostors win when their count
// reaches or exceeds the number of remaining normal players.
func (g *ImpostorGame) isGameOver() (over bool, impostorsWin bool) {
	nImpostors := len(g.impostors)
	if nImpostors == 0 {
		return true, false
	}
	nNormal := len(g.players) - nImpostors
	if nImpostors >= nNormal {
		return true, true
	}
	return false, false
}

// removePlayer returns a new slice with id removed. Order is not preserved.
func removePlayer(players []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for i, p := range players {
		if p == id {
			players[i] = players[len(players)-1]
			return players[:len(players)-1]
		}
	}
	return players
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
		payload, err := events.DecodePayload[GameSubmitWordPayload](input.Event)
		if err != nil || payload.Word == "" {
			return
		}
		sub := g.submissions[input.ClientId]
		for int(g.cycleNumber) >= len(sub) {
			sub = append(sub, "")
		}
		sub[g.cycleNumber] = payload.Word
		g.submissions[input.ClientId] = sub
	case PhaseVote:
		if input.Event.Type != events.GameSubmitVoteRequestEvent {
			return
		}
		payload, err := events.DecodePayload[GameSubmitVotePayload](input.Event)
		if err != nil {
			return
		}
		v := g.votes[input.ClientId]
		for int(g.cycleNumber) >= len(v) {
			v = append(v, nil)
		}
		v[g.cycleNumber] = payload.Target
		g.votes[input.ClientId] = v
	}
}

// Sends a base state to clients individually also calls SendPhaseTimes() from game.go
func (g *ImpostorGame) sendBaseState() {
	var previousRoundSubmissions map[uuid.UUID]string = make(map[uuid.UUID]string)
	if g.cycleNumber > 0 {
		for playerId, wordArr := range g.submissions {
			previousRoundSubmissions[playerId] = wordArr[(g.cycleNumber - 1)]
		}
	}
	for _, playerID := range g.players {
		word := g.wordPair.NormalWord
		role := RoleNormal
		_, isImpostor := g.impostors[playerID]
		if isImpostor {
			word = g.wordPair.ImpostorWord
			role = RoleImpostor
		}

		g.Send(&playerID, events.ImpostorNewRoundEvent, ImpostorBaseClientStatePayload{
			Role:                   role,
			ClientWord:             word,
			PreviousSubmittedWords: previousRoundSubmissions,
		})
	}
	g.SendPhaseTimes()
}
