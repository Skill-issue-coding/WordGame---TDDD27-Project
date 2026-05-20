package game

import "server/words"

type ImpostorSettings struct {
	InputDuration      int `json:"inputDuration"`      // seconds to submit word
	DiscussionDuration int `json:"discussionDuration"` // seconds for discussion/voting
	ImpostorCount      int `json:"impostorCount"`
}

func DefaultImpostorSettings() ImpostorSettings {
	return ImpostorSettings{
		InputDuration:      2,
		DiscussionDuration: 15,
		ImpostorCount:      1,
	}
}

type ImpostorPair struct {
	NormalWord   words.WordEntry
	ImpostorWord words.WordEntry
}

func GenerateImpostorPair(dictionary *words.Dictionary) (ImpostorPair, error) {
	normal, impostor, err := dictionary.RandomImpostorPairFromTargets()
	if err != nil {
		return ImpostorPair{}, err
	}

	return ImpostorPair{
		NormalWord:   normal,
		ImpostorWord: impostor,
	}, nil
}
