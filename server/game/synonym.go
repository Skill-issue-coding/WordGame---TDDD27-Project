package game

const ( // Settings matching the client-side settings for Synonym game
	SYNONYM_WORD_TYPE_MIN      int = 1
	SYNONYM_WORD_TYPE_MAX      int = 2
	SYNONYM_ROUND_DURATION_MIN int = 10
	SYNONYM_ROUND_DURATION_MAX int = 60
	SYNONYM_ROUNDS_MIN         int = 1
	SYNONYM_ROUNDS_MAX         int = 5
)

type SynonymDuelSettings struct {
	RoundDuration int `json:"round_duration"`
	Rounds        int `json:"rounds"`
	WordType      int `json:"word_type"`
}

func DefaultSynonymDuelSettings() SynonymDuelSettings {
	return SynonymDuelSettings{
		RoundDuration: 20,
		Rounds:        3,
		WordType:      1,
	}
}
