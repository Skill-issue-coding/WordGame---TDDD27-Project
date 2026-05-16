package game

type AntiHiveSettings struct {
	InputDuration int     `json:"input_duration"`
	MaxDistance   float64 `json:"max_distance"` // semantic distance threshold
	Rounds        int     `json:"rounds"`
}

func DefaultAntiHiveSettings() AntiHiveSettings {
	return AntiHiveSettings{
		InputDuration: 20, 
		MaxDistance:   0.5,
		Rounds:        3,
	}
}
