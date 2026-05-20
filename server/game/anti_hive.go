package game

const defaultAntiHiveMaxDistance = 0.6

type AntiHiveSettings struct {
	InputDuration int     `json:"inputDuration"`
	MaxDistance   float64 `json:"maxDistance"` // semantic distance threshold
}

func DefaultAntiHiveSettings() AntiHiveSettings {
	return AntiHiveSettings{InputDuration: 2, MaxDistance: defaultAntiHiveMaxDistance}
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
