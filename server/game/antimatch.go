package game

import "time"

const defaultAntiMatchMaxDistance = 0.5

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
	GameBase
	settings AntiMatchSettings
	phase    AntiMatchPhase
	timer    int
	round    int
}

func DefaultAntiMatchSettings() AntiMatchSettings {
	return AntiMatchSettings{
		InputDuration: 20,
		MaxDistance:   defaultAntiMatchMaxDistance,
		Rounds:        3,
	}
}

func NewAntimatchGame(
	settings AntiMatchSettings,
	outputs chan GameOutput,
	onDone func(),
) *AntiMatchGame {
	return &AntiMatchGame{
		GameBase: newGameBase(outputs, onDone),
		settings: settings,
		phase:    AntiMatchPhaseInput,
		timer:    settings.InputDuration,
	}
}

// AntiMatchThresholdFor returns the per-target distance threshold when the
// target was enriched by stage 9, otherwise the default global threshold.
// Pass the AntiHiveThreshold field from the active words.Target.
func AntiMatchThresholdFor(perTargetThreshold float64) float64 {
	if perTargetThreshold > 0 {
		return perTargetThreshold
	}
	return defaultAntiMatchMaxDistance
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

