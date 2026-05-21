package game

import (
	"server/events"
	"sync"
	"time"

	"github.com/google/uuid"
)

const defaultAntiHiveMaxDistance = 0.5

const (
	// Settings matching the client-side settings for Anti-match game
	ANTIMATCH_ROUND_DURATION_MIN = 10
	ANTIMATCH_ROUND_DURATION_MAX = 60
	ANTIMATCH_ROUNDS_MIN         = 1
	ANTIMATCH_ROUNDS_MAX         = 5
	ANTIMATCH_DISTANCE_MIN       = 0.1
	ANTIMATCH_DISTANCE_MAX       = 1.0
)

type AntiMatchSettings struct {
	InputDuration int     `json:"input_duration"`
	MaxDistance   float64 `json:"max_distance"` // semantic distance threshold
	Rounds        int     `json:"rounds"`
}

type AntiMatchPhase string

const (
	AntiMatchPhaseInput  AntiMatchPhase = "input"
	AntiMatchPhaseResult AntiMatchPhase = "result"
)

type AntiMatchGame struct {
	GameTimestamps
	settings  AntiMatchSettings
	notify    func(uuid.UUID, events.EventType, any) // send to one player
	broadcast func(events.EventType, any)            // send to all players
	onDone    func()                                 // called when game ends, resets lobby phase
	inputs    chan GameInput
	stop      chan struct{}
	once      sync.Once     // prevents double-close of stop
	phase     AntiMatchPhase
	timer     int
	round     int
}

func DefaultAntiMatchSettings() AntiMatchSettings {
	return AntiMatchSettings{
		InputDuration: 20,
		MaxDistance:   defaultAntiHiveMaxDistance,
		Rounds:        3,
	}
}

func NewAntimatchGame(
	settings AntiMatchSettings,
	notify func(uuid.UUID, events.EventType, any),
	broadcast func(events.EventType, any),
	onDone func(),
) *AntiMatchGame {
	return &AntiMatchGame{
		settings:  settings,
		notify:    notify,
		broadcast: broadcast,
		onDone:    onDone,
		inputs:    make(chan GameInput, 16),
		stop:      make(chan struct{}),
		phase:     AntiMatchPhaseInput,
		timer:     settings.InputDuration,
	}
}

// AntiHiveThresholdFor returns the per-target distance threshold when the
// target was enriched by stage 9, otherwise the default global threshold.
// Pass the AntiHiveThreshold field from the active words.Target.
func AntiHiveThresholdFor(perTargetThreshold float64) float64 {
	if perTargetThreshold > 0 {
		return perTargetThreshold
	}
	return defaultAntiHiveMaxDistance
}

// Run starts the AntiMatch game loop. It must be called in its own goroutine.
func (g *AntiMatchGame) Run() {
	g.startTime = time.Now()
	defer func() { g.endTime = time.Now() }()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer g.onDone()

	// TODO: pick target word, notify players

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

func (g *AntiMatchGame) advancePhase() {
	switch g.phase {
	case AntiMatchPhaseInput:
		g.phase = AntiMatchPhaseResult
		// TODO: score submissions, g.broadcast(events.ResultEvent, ...)
		g.round++
		if g.round >= g.settings.Rounds {
			g.Stop()
			return
		}
		g.phase = AntiMatchPhaseInput
		g.timer = g.settings.InputDuration
		// TODO: start next round
	case AntiMatchPhaseResult:
		g.Stop()
	}
}

func (g *AntiMatchGame) processInput(input GameInput) {
	// TODO: switch on g.phase and input.Event.Type
}

// HandleInput is called from the lobby's Run goroutine — forwards to internal channel.
func (g *AntiMatchGame) HandleInput(input GameInput) {
	g.inputs <- input
}

// Stop signals the game goroutine to exit. Safe to call multiple times.
func (g *AntiMatchGame) Stop() {
	g.once.Do(func() { close(g.stop) })
}
