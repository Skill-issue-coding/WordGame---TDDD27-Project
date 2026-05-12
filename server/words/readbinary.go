package words

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
)

type binaryMeta struct {
	N    int `json:"n"`
	Dims int `json:"dims"`
}

// ReadBinaryFiles loads vocab.bin + vocab.json + meta.json produced by stage 6.
// Returns nil if the binary files are absent (caller falls back to CSV loader).
func ReadBinaryFiles() map[string]WordEntry {
	dir := baseFileDirectory()

	metaPath  := filepath.Join(dir, "meta.json")
	vocabPath := filepath.Join(dir, "vocab.json")
	binPath   := filepath.Join(dir, "vocab.bin")

	for _, p := range []string{metaPath, vocabPath, binPath} {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return nil
		}
	}

	// ── meta.json ────────────────────────────────────────────────────────────
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		log.Printf("words: could not read meta.json: %v", err)
		return nil
	}
	var meta binaryMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		log.Printf("words: could not parse meta.json: %v", err)
		return nil
	}
	if meta.N <= 0 || meta.Dims <= 0 {
		log.Printf("words: meta.json has invalid dimensions: %+v", meta)
		return nil
	}

	// ── vocab.json ───────────────────────────────────────────────────────────
	vocabBytes, err := os.ReadFile(vocabPath)
	if err != nil {
		log.Printf("words: could not read vocab.json: %v", err)
		return nil
	}
	var vocab []string
	if err := json.Unmarshal(vocabBytes, &vocab); err != nil {
		log.Printf("words: could not parse vocab.json: %v", err)
		return nil
	}
	if len(vocab) != meta.N {
		log.Printf("words: vocab.json has %d entries but meta.json says N=%d", len(vocab), meta.N)
		return nil
	}

	// ── vocab.bin ────────────────────────────────────────────────────────────
	f, err := os.Open(binPath)
	if err != nil {
		log.Printf("words: could not open vocab.bin: %v", err)
		return nil
	}
	defer f.Close()

	expectedBytes := int64(meta.N) * int64(meta.Dims) * 4
	info, err := f.Stat()
	if err != nil || info.Size() != expectedBytes {
		log.Printf("words: vocab.bin size mismatch (got %d, want %d)", info.Size(), expectedBytes)
		return nil
	}

	raw := make([]float32, meta.N*meta.Dims)
	if err := binary.Read(f, binary.LittleEndian, raw); err != nil && err != io.EOF {
		log.Printf("words: error reading vocab.bin: %v", err)
		return nil
	}

	// ── Build WordMap ─────────────────────────────────────────────────────────
	wordMap := make(map[string]WordEntry, meta.N)
	for i, word := range vocab {
		key := normalizeWordKey(word)
		if key == "" {
			continue
		}

		// Convert float32 slice to float64 (compatible with existing CosineDistance)
		row := raw[i*meta.Dims : (i+1)*meta.Dims]
		vec := make([]float64, meta.Dims)
		for j, v := range row {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				vec[j] = 0
			} else {
				vec[j] = float64(v)
			}
		}

		wordMap[key] = WordEntry{
			Word:       word,
			Type:       inferType(key, wordMap),
			WordVector: vec,
		}
	}

	log.Printf("words: loaded %d entries from binary files (dims=%d)", len(wordMap), meta.Dims)
	return wordMap
}

// inferType returns a generic type string; without per-word type metadata in
// the binary format we fall back to "general". Entities (proper nouns /
// named things) end up here as "general" which is fine for random-word
// selection — RANDOM_WORD_ALLOWED_POS_TYPES already includes "general".
func inferType(_ string, _ map[string]WordEntry) string {
	return "general"
}

// LoadTargets reads targets.json — the curated Contexto target word list
// produced by stage 7. Returns nil if the file is absent (callers fall back
// to SetRandomActiveWord over the full dictionary).
func LoadTargets() []string {
	p := filepath.Join(baseFileDirectory(), "targets.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var targets []string
	if err := json.Unmarshal(data, &targets); err != nil {
		log.Printf("words: could not parse targets.json: %v", err)
		return nil
	}
	log.Printf("words: loaded %d Contexto targets", len(targets))
	return targets
}

// ErrNoBinaryFiles is returned when the binary word files are absent or malformed.
var ErrNoBinaryFiles = fmt.Errorf("binary wordfiles not found")
