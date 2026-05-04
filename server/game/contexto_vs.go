package game

type ContextoBattleSettings struct {
	RoundDuration int `json:"roundDuration"` // seconds per round
}

func DefaultContextoBattleSettings() ContextoBattleSettings {
	return ContextoBattleSettings{RoundDuration: 30}
}
