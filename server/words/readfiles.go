package words

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const DEFAULT_WORD_FILES_DIRECTORY string = "wordfiles/"

var VECTOR_FILES = []string{"celebrities_vectors.csv", "companies_vectors.csv", "culture_vectors.csv", "games_vectors.csv"}

var TERMINAL_TEST_VECTOR_FILES = []string{"celebrities_vectors.csv", "companies_vectors.csv", "culture_vectors.csv", "games_vectors.csv"}

var RANDOM_WORD_ALLOWED_POS_TYPES = []string{
	"NOUN",
	"noun",
	"noun-en",
	"noun-ett",
	"noun-en/-ett",
	"proper_noun",
	"celebrity",
	"company",
	"character",
	"game",
	"media",
	"geography",
	"general",
}

var IMPOSTOR_PRIMARY_TYPES = []string{
	"celebrity",
	"company",
	"character",
	"game",
	"media",
	"geography",
}

type vectorColumn struct {
	Index int
	Pos   int
}

func normalizeWordKey(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}

func StringToFloatSlice(vectorStr string) ([]float32, error) {
	strValues := strings.Fields(vectorStr)
	vector := make([]float32, len(strValues))

	for i, strVal := range strValues {
		floatVal, err := strconv.ParseFloat(strVal, 32)
		if err != nil {
			return nil, fmt.Errorf("error parsing value '%s' at index %d: %w \n", strVal, i, err)
		}
		vector[i] = float32(floatVal)
	}

	return vector, nil
}

func parseOptionalFloat(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseOptionalBool(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "true" || value == "1" || value == "yes"
}

func firstNonEmpty(record []string, columns map[string]int, keys ...string) string {
	for _, key := range keys {
		pos, ok := columns[strings.ToLower(key)]
		if !ok || pos >= len(record) {
			continue
		}
		value := strings.TrimSpace(record[pos])
		if value != "" {
			return value
		}
	}
	return ""
}

func parseVectorColumns(record []string, columns map[string]int) []float32 {
	indexed := make([]vectorColumn, 0)
	for name, pos := range columns {
		if !strings.HasPrefix(name, "v") {
			continue
		}
		dimension, err := strconv.Atoi(strings.TrimPrefix(name, "v"))
		if err != nil {
			continue
		}
		indexed = append(indexed, vectorColumn{Index: dimension, Pos: pos})
	}

	if len(indexed) == 0 {
		return nil
	}

	sort.Slice(indexed, func(i int, j int) bool {
		return indexed[i].Index < indexed[j].Index
	})

	vector := make([]float32, len(indexed))
	for i, col := range indexed {
		if col.Pos >= len(record) {
			return nil
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(record[col.Pos]), 32)
		if err != nil {
			return nil
		}
		vector[i] = float32(parsed)
	}

	return vector
}

func readCSVFile(wordMap *map[string]WordEntry, filename string) {
	fullPath := filepath.Join(DEFAULT_WORD_FILES_DIRECTORY, filename)
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Skipping missing or unreadable file %s: %v", fullPath, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		log.Printf("Skipping malformed csv headers for %s: %v", fullPath, err)
		return
	}

	columns := make(map[string]int, len(headers))
	for i, header := range headers {
		columns[strings.ToLower(strings.TrimSpace(header))] = i
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Skipping malformed csv row in %s: %v", fullPath, err)
			continue
		}

		word := firstNonEmpty(record, columns, "word", "cleanname", "name")
		if word == "" {
			continue
		}

		wordType := firstNonEmpty(record, columns, "category", "type", "pos")
		if wordType == "" {
			wordType = "unknown"
		}

		vector := parseVectorColumns(record, columns)
		if vector == nil {
			vectorString := firstNonEmpty(record, columns, "vector")
			if vectorString != "" {
				parsed, parseErr := StringToFloatSlice(vectorString)
				if parseErr == nil {
					vector = parsed
				}
			}
		}

		if vector == nil && len(record) >= 3 {
			parsed, parseErr := StringToFloatSlice(record[2])
			if parseErr == nil {
				vector = parsed
			}
		}

		if len(vector) == 0 {
			log.Printf("Skipping row without vector in %s for word '%s'", filename, word)
			continue
		}

		entry := WordEntry{
			Word:       word,
			Type:       wordType,
			WordVector: vector,
		}

		key := normalizeWordKey(entry.Word)
		if key == "" {
			continue
		}

		(*wordMap)[key] = entry
	}
}

func discoverVectorFiles() []string {
	entries, err := os.ReadDir(DEFAULT_WORD_FILES_DIRECTORY)
	if err != nil {
		log.Printf("Could not list vector files in %s: %v", DEFAULT_WORD_FILES_DIRECTORY, err)
		return nil
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, "_vectors.csv") {
			files = append(files, name)
		}
	}

	sort.Strings(files)
	return files
}

func ReadAllCSVFiles() map[string]WordEntry {
	files := discoverVectorFiles()
	if len(files) == 0 {
		files = VECTOR_FILES
	}
	return ReadCSVFiles(files)
}

func ReadCSVFiles(filepaths []string) map[string]WordEntry {
	var wordMap map[string]WordEntry = make(map[string]WordEntry)

	for _, filepath := range filepaths {
		readCSVFile(&wordMap, filepath)
	}

	return wordMap
}
