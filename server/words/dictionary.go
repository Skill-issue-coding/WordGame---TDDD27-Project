package words

import (
	"errors"
	"log"
	"strings"
)

func InitializeDictionary() (Dictionary, error) {
	wordMap := ReadBinaryFiles()
	if len(wordMap) == 0 {
		log.Printf("words: binary files absent or invalid, falling back to CSV loader")
		wordMap = ReadAllCSVFiles()
	}
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
	if dictionary != nil && dictionary.LemmaMap != nil {
		if lemma, ok := dictionary.LemmaMap[key]; ok {
			return lemma
		}
	}
	return key
}

func (dictionary *Dictionary) Lookup(word string) (WordEntry, bool) {
	if dictionary == nil || len(dictionary.WordMap) == 0 {
		return WordEntry{}, false
	}
	entry, exists := dictionary.WordMap[dictionary.Resolve(word)]
	return entry, exists
}
