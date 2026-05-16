package game

type ContextoBattleSettings struct {
	RoundDuration int `json:"round_duration"` // seconds per round
	WordType      int `json:"word_type"`
	Rounds        int `json:"rounds"`
}

func DefaultContextoBattleSettings() ContextoBattleSettings {
	return ContextoBattleSettings{
		RoundDuration: 120,
		WordType:      1, // 1 = Regular, 2 = Creative
		Rounds:        3,
	}
}
