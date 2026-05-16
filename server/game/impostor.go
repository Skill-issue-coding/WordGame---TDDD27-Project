package game

import "server/words"

type ImpostorSettings struct {
	InputDuration      int `json:"input_duration"`      // seconds to submit word
	DiscussionDuration int `json:"discussion_duration"` // seconds for discussion/voting
	ImpostorCount      int `json:"impostor_count"`
	VoteDuration       int `json:"vote_duration"`
}

func DefaultImpostorSettings() ImpostorSettings {
	return ImpostorSettings{
		InputDuration:      2,
		DiscussionDuration: 15,
		ImpostorCount:      1,
		VoteDuration:       30,
	}
}

type ImpostorPair struct {
	NormalWord   words.WordEntry
	ImpostorWord words.WordEntry
}

func GenerateImpostorPair(dictionary *words.Dictionary) (ImpostorPair, error) {
	normal, impostor, err := dictionary.RandomRelatedPair(words.IMPOSTOR_PRIMARY_TYPES)
	if err != nil {
		return ImpostorPair{}, err
	}

	return ImpostorPair{
		NormalWord:   normal,
		ImpostorWord: impostor,
	}, nil
}
