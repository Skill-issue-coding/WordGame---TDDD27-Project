package words

import (
	"errors"
	"math"
	"math/rand/v2"
	"server/util"
	"strings"
)

func InitializeDictionary() (Dictionary, error) {
	wordMap := ReadAllCSVFiles()
	if len(wordMap) == 0 {
		return Dictionary{}, errors.New("dictionary is empty")
	}

	return Dictionary{
		WordMap: wordMap,
	}, nil
}

func (dictionary *Dictionary) CalculateDistance(word string) float64 {
	activeWordEntry, activeWordExists := dictionary.WordMap[dictionary.ActiveWord]
	guessEntry, guessExists := dictionary.WordMap[word]

	if !activeWordExists || !guessExists {
		return math.NaN()
	}

	return util.CosineDistance(activeWordEntry.WordVector, guessEntry.WordVector)
}

func (dictionary *Dictionary) IsValid(word string) bool {
	_, exists := dictionary.WordMap[word]
	return exists
}

func (dictionary *Dictionary) SetRandomActiveWord() error {
	entry, err := dictionary.RandomWord()
	if err != nil {
		return err
	}

	dictionary.ActiveWord = entry.Word
	return nil
}

func (dictionary *Dictionary) RandomWord() (WordEntry, error) {
	if len(dictionary.WordMap) == 0 {
		return WordEntry{}, errors.New("word map is empty")
	}

	targetIndex := rand.IntN(len(dictionary.WordMap))
	currentIndex := 0

	for _, entry := range dictionary.WordMap {
		if currentIndex == targetIndex {
			return entry, nil
		}
		currentIndex++
	}

	return WordEntry{}, errors.New("failed to select random word")
}

func (dictionary *Dictionary) RandomWordByAllowedPOSTypes(wordMap map[string]WordEntry, allowedTypes []string) (WordEntry, error) {
	if len(wordMap) == 0 {
		return WordEntry{}, errors.New("word map is empty")
	}

	if len(allowedTypes) == 0 {
		return WordEntry{}, errors.New("allowed POS types are empty")
	}

	allowedTypeSet := make(map[string]struct{}, len(allowedTypes))
	for _, posType := range allowedTypes {
		allowedTypeSet[strings.TrimSpace(posType)] = struct{}{}
	}

	filteredEntries := make([]WordEntry, 0)
	for _, entry := range wordMap {
		if _, ok := allowedTypeSet[strings.TrimSpace(entry.Type)]; ok {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	if len(filteredEntries) == 0 {
		return WordEntry{}, errors.New("no words matched allowed POS types")
	}

	return filteredEntries[rand.IntN(len(filteredEntries))], nil
}
