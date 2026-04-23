package words

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const BASE_FILE_DIRECTORY string = "wordfiles/"

var VECTOR_FILES = []string{"celebrities_vectors.csv", "companies_vectors.csv", "kelly_vectors.csv", "korp_vectors.csv", "maktbarometern_vectors.csv"}

var TERMINAL_TEST_VECTOR_FILES = []string{"celebrities_vectors.csv", "companies_vectors.csv", "kelly_vectors.csv", "maktbarometern_vectors.csv"}

var RANDOM_WORD_ALLOWED_POS_TYPES = []string{"NOUN", "noun", "noun-en", "noun-ett", "noun-en/-ett", "proper_noun"}

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
