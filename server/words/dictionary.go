package words

import (
	"errors"
	"strings"
)

func InitializeDictionary() (Dictionary, error) {
	wordMap := ReadBinaryFiles()
	if len(wordMap) == 0 {
		return Dictionary{}, errors.New("dictionary is empty")
	}

	return Dictionary{
		WordMap:  wordMap,
		Targets:  LoadTargets(),
		LemmaMap: LoadLemmaMap(),
	}, nil
}

// Resolve maps a surface form to its canonical lemma key using LemmaMap.
// "bilar" → "bil", "röda" → "röd". Falls back to the lowercased input
// when the word is not in the map (entities, already-canonical lemmas, etc.).
func (dictionary *Dictionary) Resolve(word string) string {
	key := strings.ToLower(strings.TrimSpace(word))
	if lemma, ok := dictionary.LemmaMap[key]; ok {
		return lemma
	}
	return key
}

func (dictionary *Dictionary) Lookup(word string) (WordEntry, bool) {
	entry, exists := dictionary.WordMap[dictionary.Resolve(word)]
	return entry, exists
}
