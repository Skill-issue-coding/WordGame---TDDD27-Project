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
	SHOW_WORD_DURATION int = 8

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

type PlayerNode struct {
	self         uuid.UUID
	nextNode     *PlayerNode
	previousNode *PlayerNode
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

	// players is a map that returns a playerlistentry, playerentries is a circular
	// doubly linked list that keeps track of the playing order of all players
	players map[uuid.UUID]*PlayerNode

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

	// startingPlayer keeps track which player started
	startingPlayer *PlayerNode

	// currentPlayer keeps track of which players turn it is during the input phase
	currentPlayer *PlayerNode
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
		players:     createPlayersCircularList(players),
		phase:       PhaseShowWord,
		submissions: make(map[uuid.UUID][]string),
		votes:       make(map[uuid.UUID][]*uuid.UUID),
		cycleNumber: 0,
	}
}

// createPlayersCircularList builds a circular doubly linked list of players.
// players is the ordered list of all player IDs used to define the play order.
// Returns a map from player ID to its node in the list for O(1) lookup.
func createPlayersCircularList(players []uuid.UUID) map[uuid.UUID]*PlayerNode {
	playerEntriesMap := make(map[uuid.UUID]*PlayerNode, len(players))
	if len(players) == 0 {
		return playerEntriesMap
	}

	entries := make([]*PlayerNode, len(players))
	for idx, player := range players {
		entries[idx] = &PlayerNode{self: player}
	}

	for idx, player := range players {
		nextIdx := idx + 1
		if nextIdx >= len(players) {
			nextIdx = 0
		}
		previousIdx := idx - 1
		if previousIdx < 0 {
			previousIdx = len(players) - 1
		}

		entries[idx].nextNode = entries[nextIdx]
		entries[idx].previousNode = entries[previousIdx]
		playerEntriesMap[player] = entries[idx]
	}

	return playerEntriesMap
}

// getRandomPlayerNode picks a random node from the players map.
// Returns nil if the map is empty.
func (g *ImpostorGame) getRandomPlayerNode() *PlayerNode {
	target := rand.IntN(len(g.players))
	i := 0
	for _, node := range g.players {
		if i == target {
			return node
		}
		i++
	}
	return nil
}

// eliminatePlayer is a function that eliminates a player from the game
// it removes the player from the circular list and the entry in the hashmap
// now points to a nil pointer
func (g *ImpostorGame) eliminatePlayer(player uuid.UUID) {
	playerNode := g.players[player]
	if playerNode == nil {
		return
	}

	if g.startingPlayer == playerNode {
		g.startingPlayer = playerNode.nextNode
	}

	playerNode.previousNode.nextNode = playerNode.nextNode
	playerNode.nextNode.previousNode = playerNode.previousNode
	g.players[player] = nil
}

// getNormalAndImposterCount is a function that loops through the circular list
// and counts the active normal and impostor players
// returns normalPlayerCount, impostorPlayerCount
func (g *ImpostorGame) getNormalAndImpostorCount() (int, int) {
	currentNode := g.startingPlayer

	var normalPlayerCount int = 0
	var impostorPlayerCount int = 0

	for {
		if _, exists := g.impostors[currentNode.nextNode.self]; exists {
			impostorPlayerCount++
		} else {
			normalPlayerCount++
		}

		if currentNode.nextNode == g.startingPlayer {
			return normalPlayerCount, impostorPlayerCount
		}
		currentNode = currentNode.nextNode
	}
}

// pickImpostors randomly assigns ImpostorCount players from g.players as impostors.
// Returns a set for O(1) role lookups. The caller is responsible for ensuring
// ImpostorCount < len(g.players) to avoid infinite recursion.
func (g *ImpostorGame) pickImpostors() map[uuid.UUID]struct{} {
	return g.pickImpostorsRecursive(make(map[uuid.UUID]struct{}), g.settings.ImpostorCount)
}

func (g *ImpostorGame) pickImpostorsRecursive(impostors map[uuid.UUID]struct{}, remaining int) map[uuid.UUID]struct{} {
	if remaining == 0 {
		return impostors
	}
	node := g.getRandomPlayerNode()
	if _, exists := impostors[node.self]; exists {
		return g.pickImpostorsRecursive(impostors, remaining)
	}
	impostors[node.self] = struct{}{}
	return g.pickImpostorsRecursive(impostors, remaining-1)
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

	g.startingPlayer = g.getRandomPlayerNode()

	pair, ok := pickImpostorPair(g.dict)
	if !ok {
		return
	}
	g.wordPair = pair
	g.impostors = g.pickImpostors()

	g.StartPhase(SHOW_WORD_DURATION)

	g.sendInitialGameState()

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
//   - PhaseShowWord → PhaseInput: broadcasts GameNewPhaseEvent with InputEndsAt.
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
		if g.currentPlayer == g.startingPlayer {
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
			return
		}

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
		if payload.Target != nil {
			valid := false
			for _, p := range g.players {
				if p == *payload.Target {
					valid = true
					break
				}
			}
			if !valid {
				return
			}
		}
		v := g.votes[input.ClientId]
		for int(g.cycleNumber) >= len(v) {
			v = append(v, nil)
		}
		v[g.cycleNumber] = payload.Target
		g.votes[input.ClientId] = v
	}
}

func (g *ImpostorGame) sendInitialGameState() {
	activePlayers := make(map[uuid.UUID]bool, len(g.players))
	for id, node := range g.players {
		if node != nil {
			activePlayers[id] = true
		}
	}

	for playerId := range g.players {
		word := g.wordPair.NormalWord
		role := ImpostorRoleNormal
		if _, exists := g.impostors[playerId]; exists {
			word = g.wordPair.ImpostorWord
			role = ImpostorRoleImpostor
		}

		state := ImpostorClientGameState{
			Role:                   role,
			Word:                   word,
			ActivePlayers:          activePlayers,
			PreviousSubmittedWords: g.submissions,
		}
		id := playerId
		g.Send(&id, events.ImpostorNewRoundEvent, state)
	}
}
