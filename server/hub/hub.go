package hub

import (
	"errors"
	"math"
	"server/words"
)

type Dictionary struct {
	ActiveWord        string
	WordMap           map[string]words.WordEntry
	RandomWord        func() (words.WordEntry, error)
	CalculateDistance func(word string) float64
	IsValid           func(word string) bool
}

type GameHub struct {
	Dictionary Dictionary
}

func NewGameHub(dictionaryFiles []string) (*GameHub, error) {
	wordMap := words.ReadCSVFiles(dictionaryFiles)
	if len(wordMap) == 0 {
		return nil, errors.New("dictionary is empty")
	}

	gameHub := &GameHub{
		Dictionary: Dictionary{
			WordMap: wordMap,
			RandomWord: func() (words.WordEntry, error) {
				return words.RandomWordByAllowedPOSTypes(wordMap, words.RANDOM_WORD_ALLOWED_POS_TYPES)
			},
			IsValid: func(word string) bool {
				_, exists := wordMap[word]
				return exists
			},
		},
	}

	gameHub.Dictionary.CalculateDistance = func(word string) float64 {
		activeWordEntry, activeWordExists := wordMap[gameHub.Dictionary.ActiveWord]
		guessEntry, guessExists := wordMap[word]

		if !activeWordExists || !guessExists {
			return math.NaN()
		}

		return cosineDistance(activeWordEntry.WordVector, guessEntry.WordVector)
	}

	if err := gameHub.SetRandomActiveWord(); err != nil {
		return nil, err
	}

	return gameHub, nil
}

func (gh *GameHub) SetRandomActiveWord() error {
	if gh.Dictionary.RandomWord == nil {
		return errors.New("random word provider is not configured")
	}

	entry, err := gh.Dictionary.RandomWord()
	if err != nil {
		return err
	}

	gh.Dictionary.ActiveWord = entry.Word
	return nil
}

func (gh *GameHub) GetWordEntry(word string) (words.WordEntry, bool) {
	entry, exists := gh.Dictionary.WordMap[word]
	return entry, exists
}

func cosineDistance(vecA []float64, vecB []float64) float64 {
	if len(vecA) == 0 || len(vecB) == 0 || len(vecA) != len(vecB) {
		return math.NaN()
	}

	var dot float64
	var normA float64
	var normB float64

	for i := range vecA {
		dot += vecA[i] * vecB[i]
		normA += vecA[i] * vecA[i]
		normB += vecB[i] * vecB[i]
	}

	if normA == 0 || normB == 0 {
		return math.NaN()
	}

	cosineSimilarity := dot / (math.Sqrt(normA) * math.Sqrt(normB))
	return 1 - cosineSimilarity
}
