package words

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
)

const BASE_FILE_DIRECTORY string = "words/"

var VECTOR_FILES = []string{"celebrities_vectors.csv", "companies_vectors.csv", "kelly_vectors.csv", "korp_vectors.csv", "maktbarometern_vectors.csv"}

var TERMINAL_TEST_VECTOR_FILES = []string{"celebrities_vectors.csv", "companies_vectors.csv", "kelly_vectors.csv", "maktbarometern_vectors.csv"}

var RANDOM_WORD_ALLOWED_POS_TYPES = []string{"NOUN", "noun", "noun-en", "noun-ett", "noun-en/-ett", "proper_noun"}

type WordEntry struct {
	Word       string
	Type       string
	WordVector []float64
}

func StringToFloatSlice(vectorStr string) ([]float64, error) {
	strValues := strings.Fields(vectorStr)
	vector := make([]float64, len(strValues))

	for i, strVal := range strValues {
		floatVal, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing value '%s' at index %d: %w \n", strVal, i, err)
		}
		vector[i] = floatVal
	}

	return vector, nil
}

func readCSVFile(wordMap *map[string]WordEntry, filepath string) {
	file, err := os.Open(BASE_FILE_DIRECTORY + filepath)

	if err != nil {
		log.Fatalf("Unable to read input file %s, %v \n", filepath, err)
		return
	}
	defer file.Close()

	r := csv.NewReader(file)

	if _, err := r.Read(); err != nil {
		log.Fatal(err)
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		vector, err := StringToFloatSlice(record[2])
		if err != nil {
			log.Printf("Error parsing wordvector for: %s, %s \n", filepath, record[0])
			continue
		}

		var entry WordEntry = WordEntry{Word: record[0], Type: record[1], WordVector: vector}

		if _, exists := (*wordMap)[entry.Word]; exists {
			continue
		}

		(*wordMap)[entry.Word] = entry
	}
}

func ReadAllCSVFiles() map[string]WordEntry {
	return ReadCSVFiles(VECTOR_FILES)
}

func ReadCSVFiles(filepaths []string) map[string]WordEntry {
	var wordMap map[string]WordEntry = make(map[string]WordEntry)

	for _, filepath := range filepaths {
		readCSVFile(&wordMap, filepath)
	}

	return wordMap
}

func RandomWord(wordMap map[string]WordEntry) (WordEntry, error) {
	if len(wordMap) == 0 {
		return WordEntry{}, errors.New("word map is empty")
	}

	targetIndex := rand.IntN(len(wordMap))
	currentIndex := 0

	for _, entry := range wordMap {
		if currentIndex == targetIndex {
			return entry, nil
		}
		currentIndex++
	}

	return WordEntry{}, errors.New("failed to select random word")
}

func RandomWordByAllowedPOSTypes(wordMap map[string]WordEntry, allowedTypes []string) (WordEntry, error) {
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
