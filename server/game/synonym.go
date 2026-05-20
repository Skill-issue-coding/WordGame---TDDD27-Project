package game

type SynonymDuelSettings struct {
	RoundDuration int `json:"round_duration"`
	Rounds        int `json:"rounds"`
	WordType      int `json:"word_type"`
}

func DefaultSynonymDuelSettings() SynonymDuelSettings {
	return SynonymDuelSettings{
		RoundDuration: 20,
		Rounds:        3,
		WordType:      1, // 1 = Regular, 2 = Creative
	}
}
