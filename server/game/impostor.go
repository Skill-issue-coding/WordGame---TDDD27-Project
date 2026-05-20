package game

import "server/words"

type ImpostorSettings struct {
	InputDuration      int `json:"input_duration"`      // seconds to submit word
	DiscussionDuration int `json:"discussion_duration"` // seconds for discussion
	ImpostorCount      int `json:"impostor_count"`      // amount of impostors
	VoteDuration       int `json:"vote_duration"`       // seconds for voting
}

func DefaultImpostorSettings() ImpostorSettings {
	return ImpostorSettings{
		InputDuration:      30,
		DiscussionDuration: 15,
		ImpostorCount:      1,
		VoteDuration:       30,
	}
}

type ImpostorPair struct {
	NormalWord   words.WordEntry
	ImpostorWord words.WordEntry
}
