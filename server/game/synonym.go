package game

type SynonymDuelSettings struct {
	RoundDuration int `json:"roundDuration"`
	Rounds        int `json:"rounds"`
}

func DefaultSynonymDuelSettings() SynonymDuelSettings {
	return SynonymDuelSettings{RoundDuration: 3, Rounds: 5}
}
