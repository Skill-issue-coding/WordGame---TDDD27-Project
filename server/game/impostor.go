package game

import "server/words"

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
