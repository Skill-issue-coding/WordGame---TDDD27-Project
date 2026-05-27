package game

const (
	// Settings matching the client-side settings for Contexto game
	CONTEXTO_WORD_TYPE_MIN      int = 1
	CONTEXTO_WORD_TYPE_MAX      int = 2
	CONTEXTO_ROUND_DURATION_MIN int = 60
	CONTEXTO_ROUND_DURATION_MAX int = 600
	CONTEXTO_ROUNDS_MIN         int = 1
	CONTEXTO_ROUNDS_MAX         int = 5
)

type ContextoBattleSettings struct {
	RoundDuration int `json:"round_duration"` // seconds per round
	WordType      int `json:"word_type"`
	Rounds        int `json:"rounds"`
}

type ContextoGame struct {
	GameBase

	settings ContextoBattleSettings
}

func DefaultContextoBattleSettings() ContextoBattleSettings {
	return ContextoBattleSettings{
		RoundDuration: 120,
		WordType:      1, // 1 = Regular, 2 = Creative
		Rounds:        3,
	}
}
