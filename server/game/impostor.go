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
	SHOW_WORD_DURATION    int = 8
	INTERMEDIATE_DURATION int = 5

	PhaseShowWord          ImpostorPhase = "show_word"
	PhaseInput             ImpostorPhase = "input"
	PhaseDiscussion        ImpostorPhase = "discussion"
	PhaseVote              ImpostorPhase = "vote"
	PhaseIntermediateCycle ImpostorPhase = "intermediate"
	PhaseResult            ImpostorPhase = "result"

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

// ImpostorPair holds the normal word and the pool of impostor candidates for a round.
type ImpostorPair struct {
	NormalWord         string
	ImpostorCandidates []string
}

type PlayerNode struct {
	self         uuid.UUID
	nextNode     *PlayerNode
	previousNode *PlayerNode
}

type ImpostorCycle struct {
	Submissions map[uuid.UUID]string     `json:"submissions"`
	Votes       map[uuid.UUID]*uuid.UUID `json:"votes"`
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

	// players maps player IDs to nodes in a circular doubly linked list that
	// represents turn order; eliminated players have a nil entry.
	players map[uuid.UUID]*PlayerNode

	// impostors is the set of player IDs assigned the impostor role.
	// Populated by pickImpostors at the start of Run.
	impostors map[uuid.UUID]struct{}

	// impostorWords maps each impostor's ID to their unique word drawn from the candidate pool.
	impostorWords map[uuid.UUID]string

	// wordPair holds the normal word and the full impostor candidate pool for this round.
	wordPair ImpostorPair

	// phase tracks the current lifecycle stage of the game.
	phase ImpostorPhase

	// cycles is an array containing all the submissions and votes for all cycles
	cycles []ImpostorCycle

	// cycleNumber is the index of the current cycle (0-based).
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
		GameBase: newGameBase(outputs, onDone),
		settings: settings,
		dict:     dict,
		players:  createPlayersCircularList(players),
		phase:    PhaseShowWord,
		cycles: []ImpostorCycle{{
			Submissions: make(map[uuid.UUID]string),
			Votes:       make(map[uuid.UUID]*uuid.UUID),
		}},
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

// getActivePlayers is a helper function that gets all players that have not been eliminated yet
// returns map[uuid.UUID]bool
func (g *ImpostorGame) getActivePlayers() map[uuid.UUID]bool {
	activePlayers := make(map[uuid.UUID]bool, len(g.players))
	for id, node := range g.players {
		if node != nil {
			activePlayers[id] = true
		}
	}

	return activePlayers
}

// doesWordExist is a helper function to check if a word has already been submited
// returns bool that is true if the word exists
func (g *ImpostorGame) doesWordExist(word string) bool {
	for _, cycle := range g.cycles {
		for _, sub_word := range cycle.Submissions {
			if sub_word == word {
				return true
			}
		}
	}

	return false
}

// getRandomPlayerNode picks a random node from the players map.
// Assumes g.players is non-empty.
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

// getPlayerWithMostVotes retrieves the player with the most votes
// returns uuid (nil if skipped or tie) and message if tied or skipped
func (g *ImpostorGame) getPlayerWithMostVotes() (*uuid.UUID, string) {
	voteCounts := make(map[uuid.UUID]int)
	skipCount := 0

	// Sum all the votes
	for _, votedFor := range g.cycles[g.cycleNumber].Votes {
		if votedFor == nil {
			skipCount++
		} else {
			voteCounts[*votedFor]++
		}
	}

	// Find the highest vote count among actual players
	maxVotes := 0
	var topCandidates []uuid.UUID

	for candidate, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
			topCandidates = []uuid.UUID{candidate} // Found a new max, reset the slice
		} else if count == maxVotes {
			topCandidates = append(topCandidates, candidate) // Tie for max, append to slice
		}
	}

	// Evaluate the results against skips and ties

	// Case A: Nobody voted at all or more skips than votes on a single player
	if maxVotes == 0 && skipCount == 0 || skipCount > maxVotes {
		return nil, "Majoritet röstade för att hoppa över."
	}

	// Case B: Tie between skipping and a player or multiple players with the same amount of votes
	if skipCount == maxVotes && maxVotes > 0 || len(topCandidates) > 1 {
		return nil, "Oavgjort mellan flera spelare eller hoppa över. Ingen blev eliminerad. Går vidare till nästa fas"
	}

	// Case C: A single player has the most votes
	if len(topCandidates) == 1 {
		return &topCandidates[0], ""
	}

	// Fallback safety (should rarely be hit)
	return nil, "Röstningen resulterade i inga elimineringar."
}

// getNormalAndImpostorCount is a function that loops through the circular list
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
// ImpostorCount <= len(g.players) to avoid excessive retries.
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
func pickImpostorPair(dict *words.Dictionary, impostorCount int) (ImpostorPair, bool) {
	eligible := make([]words.Target, 0, len(dict.Targets))
	for _, t := range dict.Targets {
		if len(t.ImpostorCandidates) >= impostorCount {
			eligible = append(eligible, t)
		}
	}
	if len(eligible) == 0 {
		return ImpostorPair{}, false
	}
	target := eligible[rand.IntN(len(eligible))]
	return ImpostorPair{NormalWord: target.Word, ImpostorCandidates: target.ImpostorCandidates}, true
}

// assignImpostorWords picks a unique word from the candidate pool for each impostor.
// pickImpostorPair guarantees len(candidates) >= impostorCount, so every impostor
// gets a distinct word.
func (g *ImpostorGame) assignImpostorWords() map[uuid.UUID]string {
	result := make(map[uuid.UUID]string, len(g.impostors))
	candidates := make([]string, len(g.wordPair.ImpostorCandidates))
	copy(candidates, g.wordPair.ImpostorCandidates)
	rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	i := 0
	for id := range g.impostors {
		result[id] = candidates[i]
		i++
	}
	return result
}

// Run starts the game's main event loop and must be called in its own goroutine.
// On startup it picks a word pair via pickImpostorPair and assigns impostor
// roles via pickImpostors, then privately delivers ImpostorNewRoundEvent
// to each player containing their word, role, and phase timestamps so the
// client can render countdown timers.
// The loop then steps through phases driven by a one-second ticker:
//
//	PhaseShowWord → PhaseInput → PhaseDiscussion → PhaseVote → PhaseIntermediateCycle
//	(repeats PhaseInput for the next cycle until the game ends)
//
// If no eligible word pair exists in the dictionary, Run exits immediately and
// calls onDone, leaving players on the game-started screen. onDone is always
// called on exit to signal the lobby to reset back to LobbyPhase.
func (g *ImpostorGame) Run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer g.onDone()

	g.startingPlayer = g.getRandomPlayerNode()
	g.currentPlayer = g.startingPlayer

	pair, ok := pickImpostorPair(g.dict, g.settings.ImpostorCount)
	if !ok {
		return
	}
	g.wordPair = pair
	g.impostors = g.pickImpostors()
	g.impostorWords = g.assignImpostorWords()

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
// new phase timer via StartPhase and broadcasts the appropriate events:
//   - PhaseShowWord → PhaseInput: broadcasts ImpostorNewPhaseEvent.
//   - PhaseInput → PhaseDiscussion: handled by advanceInputPlayer; broadcasts
//     ImpostorNewPhaseEvent when the cycle's input is complete.
//   - PhaseDiscussion → PhaseVote: broadcasts ImpostorNewPhaseEvent.
//   - PhaseVote → PhaseIntermediateCycle: broadcasts ImpostorVoteResultEvent.
//   - PhaseIntermediateCycle → PhaseInput: broadcasts ImpostorNewCycleEvent and
//     ImpostorNewPhaseEvent, or stops the game if it has ended.
func (g *ImpostorGame) advancePhase() {
	switch g.phase {
	case PhaseShowWord:
		g.phase = PhaseInput
		g.StartPhase(g.settings.InputDuration)
		g.sendGamePhaseUpdate()

	case PhaseInput:
		g.advanceInputPlayer()

	case PhaseDiscussion:
		g.phase = PhaseVote
		g.StartPhase(g.settings.VoteDuration)
		g.sendGamePhaseUpdate()

	case PhaseVote:
		votedOutPlayer, message := g.getPlayerWithMostVotes()
		if votedOutPlayer != nil {
			g.eliminatePlayer(*votedOutPlayer)
		}
		g.phase = PhaseIntermediateCycle
		g.StartPhase(INTERMEDIATE_DURATION)
		g.Broadcast(events.ImpostorVoteResultEvent, ImpostorVoteResultPayload{VotedOut: votedOutPlayer, Message: message})

	case PhaseIntermediateCycle:
		gameOver, impostorsWon := g.getCycleResult()
		if gameOver {
			g.getGameResults(impostorsWon)
			g.Stop()
		} else {
			g.cycleNumber++
			g.cycles = append(g.cycles, ImpostorCycle{
				Submissions: make(map[uuid.UUID]string),
				Votes:       make(map[uuid.UUID]*uuid.UUID),
			})
			g.currentPlayer = g.startingPlayer
			g.phase = PhaseInput
			g.StartPhase(g.settings.InputDuration)
			g.sendGameCycleState()
			g.sendGamePhaseUpdate()
		}
	}
}

// advanceInputPlayer moves to the next player's turn. If the full cycle
// is complete (we've wrapped back to startingPlayer), it advances to discussion.
func (g *ImpostorGame) advanceInputPlayer() {
	next := g.currentPlayer.nextNode
	if next == g.startingPlayer {
		// Full cycle done — move to discussion
		g.phase = PhaseDiscussion
		g.StartPhase(g.settings.DiscussionDuration)
	} else {
		g.currentPlayer = next
		g.StartPhase(g.settings.InputDuration)
	}
	g.sendGamePhaseUpdate()
}

// processInput handles a single player action forwarded from HandleInput.
// It is called exclusively from the Run goroutine, so field access is safe
// without additional locking. Behaviour is phase-dependent:
//   - PhaseInput: accepts GameSubmitWordRequestEvent from the current player
//     only. Records the submission and advances to the next player or phase.
//   - PhaseVote: accepts GameSubmitVoteRequestEvent. Records the player's vote
//     target in votes; nil target means the player chose to skip. Last write
//     wins if a player votes more than once before the deadline.
//
// Inputs received during any other phase, or with unexpected event types, are
// silently dropped.
func (g *ImpostorGame) processInput(input GameInput) {
	switch g.phase {
	case PhaseInput:
		if input.Event.Type != events.GameSubmitWordRequestEvent || input.ClientId != g.currentPlayer.self {
			return
		}

		payload, err := events.DecodePayload[GameSubmitWordPayload](input.Event)
		if err != nil || payload.Word == "" {
			return
		}

		if g.doesWordExist(payload.Word) {
			g.Send(&g.currentPlayer.self, events.ErrorEvent, map[string]string{"message": "Order har redan skickats in tidigare."})
			return
		}

		g.cycles[g.cycleNumber].Submissions[input.ClientId] = payload.Word
		g.advanceInputPlayer()

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
			for id, node := range g.players {
				if node != nil && id == *payload.Target {
					valid = true
					break
				}
			}
			if !valid {
				return
			}
		}
		g.cycles[g.cycleNumber].Votes[input.ClientId] = payload.Target
		g.Broadcast(events.ImpostorVoteUpdateEvent, ImpostorVoteUpdatePayload{
			Votes: g.cycles[g.cycleNumber].Votes,
		})
	}
}

// getCycleResult is called once a cycle is completed
// returns gameOver, impostorsWon
func (g *ImpostorGame) getCycleResult() (bool, bool) {
	normalCount, impostorCount := g.getNormalAndImpostorCount()

	if impostorCount >= normalCount {
		return true, true
	} else if impostorCount == 0 {
		return true, false
	}

	return false, false
}

func (g *ImpostorGame) getGameResults(impostorsWon bool) {

}

func (g *ImpostorGame) sendInitialGameState() {
	for playerId := range g.players {
		word := g.wordPair.NormalWord
		role := ImpostorRoleNormal
		if _, exists := g.impostors[playerId]; exists {
			word = g.impostorWords[playerId]
			role = ImpostorRoleImpostor
		}

		state := ImpostorClientGameStatePayload{
			Role:             role,
			Word:             word,
			ActivePlayers:    g.getActivePlayers(),
			GamePhasePayload: GamePhasePayload{StartTime: g.startTime.UnixMilli(), EndTime: g.endTime.UnixMilli()},
		}
		g.Send(&playerId, events.ImpostorNewRoundEvent, state)
	}
}

func (g *ImpostorGame) sendGameCycleState() {
	state := ImpostorGameCycleUpdatePayload{Cycles: g.cycles, ActivePlayers: g.getActivePlayers()}
	g.Broadcast(events.ImpostorNewCycleEvent, state)
}

func (g *ImpostorGame) sendGamePhaseUpdate() {
	phaseTimers := GamePhasePayload{StartTime: g.startTime.UnixMilli(), EndTime: g.endTime.UnixMilli()}
	state := ImpostorGamePhaseUpdatePayload{
		GamePhasePayload: phaseTimers,
		WordsCycle:       g.cycles[g.cycleNumber].Submissions,
		VotesCycle:       g.cycles[g.cycleNumber].Votes,
		CurrentPlayer:    g.currentPlayer.self,
		Phase:            g.phase,
	}
	g.Broadcast(events.ImpostorNewPhaseEvent, state)
}
