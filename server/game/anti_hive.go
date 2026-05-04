package game

type AntiHiveSettings struct {
	InputDuration int     `json:"inputDuration"`
	MaxDistance   float64 `json:"maxDistance"` // semantic distance threshold
}

func DefaultAntiHiveSettings() AntiHiveSettings {
	return AntiHiveSettings{InputDuration: 2, MaxDistance: 0.6}
}
