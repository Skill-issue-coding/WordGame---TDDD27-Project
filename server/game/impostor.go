package game

import (
	"server/events"
	"server/words"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ImpostorPhase string

const (
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
	InputDuration      int `json:"input_duration"`      // seconds to submit word
	DiscussionDuration int `json:"discussion_duration"` // seconds for discussion
	ImpostorCount      int `json:"impostor_count"`      // amount of impostors
	VoteDuration       int `json:"vote_duration"`       // seconds for voting
}

func DefaultImpostorSettings() ImpostorSettings {
	return ImpostorSettings{
		InputDuration:      30,
		DiscussionDuration: 15,
		ImpostorCount:      1,
		VoteDuration:       30,
	}
}

type ImpostorPair struct {
	NormalWord   words.WordEntry
	ImpostorWord words.WordEntry
}

type ImpostorGame struct {
	settings  ImpostorSettings
	notify    func(uuid.UUID, events.EventType, any) // send to one player
	broadcast func(events.EventType, any)            // send to all players
	onDone    func()                                 // called when game ends, resets lobby phase
	inputs    chan GameInput
	stop      chan struct{}
	once      sync.Once // prevents double-close of stop
	phase     ImpostorPhase
	timer     int
}

func NewImpostorGame(
	settings ImpostorSettings,
	notify func(uuid.UUID, events.EventType, any),
	broadcast func(events.EventType, any),
	onDone func(),
) *ImpostorGame {
	return &ImpostorGame{
		settings:  settings,
		notify:    notify,
		broadcast: broadcast,
		onDone:    onDone,
		inputs:    make(chan GameInput, 16),
		stop:      make(chan struct{}),
		phase:     PhaseInput,
		timer:     settings.InputDuration,
	}
}

func (g *ImpostorGame) Run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer g.onDone() // notify lobby when game exits for any reason

	// TODO: pick words, assign impostor, notify players

	for {
		select {
		case <-g.stop:
			return
		case input := <-g.inputs:
			g.processInput(input)
		case <-ticker.C:
			g.timer--
			if g.timer <= 0 {
				g.advancePhase()
			}
		}
	}
}

func (g *ImpostorGame) advancePhase() {
	switch g.phase {
	case PhaseInput:
		g.phase = PhaseDiscussion
		g.timer = g.settings.DiscussionDuration
		// TODO: g.broadcast(events.DiscussionStartedEvent, ...)
	case PhaseDiscussion:
		g.phase = PhaseVote
		g.timer = g.settings.VoteDuration
		// TODO: g.broadcast(events.VoteStartedEvent, ...)
	case PhaseVote:
		g.phase = PhaseResult
		// TODO: tally votes, g.broadcast(events.ResultEvent, ...)
	case PhaseResult:
		g.Stop()
	}
}

func (g *ImpostorGame) processInput(input GameInput) {
	// TODO: switch on g.phase and input.Event.Type
}

// HandleInput is called from the lobby's Run goroutine — forwards to internal channel.
func (g *ImpostorGame) HandleInput(input GameInput) {
	g.inputs <- input
}

// Stop signals the game goroutine to exit. Safe to call multiple times.
func (g *ImpostorGame) Stop() {
	g.once.Do(func() { close(g.stop) })
}
