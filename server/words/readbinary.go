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
	N    int  `json:"n"`
	Dims int  `json:"dims"`
	Dual bool `json:"dual"`
}

// ReadBinaryFiles loads vocab.bin + vocab.json + meta.json produced by stage 6.
// If vocab_query.bin is present (dual=true in meta.json), query vectors are loaded
// into WordEntry.QueryVector for asymmetric passage/query similarity.
// Returns nil if the required binary files are absent (caller falls back to CSV loader).
func ReadBinaryFiles() map[string]WordEntry {
	dir := baseFileDirectory()

	metaPath     := filepath.Join(dir, "meta.json")
	vocabPath    := filepath.Join(dir, "vocab.json")
	binPath      := filepath.Join(dir, "vocab.bin")
	queryBinPath := filepath.Join(dir, "vocab_query.bin")

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

	// ── Load vocab_query.bin (optional, dual-vector mode) ────────────────────
	var rawQuery []float32
	if meta.Dual {
		if info, err := os.Stat(queryBinPath); err == nil && info.Size() == expectedBytes {
			qf, err := os.Open(queryBinPath)
			if err != nil {
				log.Printf("words: could not open vocab_query.bin: %v", err)
			} else {
				defer qf.Close()
				rawQuery = make([]float32, meta.N*meta.Dims)
				if err := binary.Read(qf, binary.LittleEndian, rawQuery); err != nil && err != io.EOF {
					log.Printf("words: error reading vocab_query.bin: %v — falling back to passage vectors", err)
					rawQuery = nil
				}
			}
		} else {
			log.Printf("words: meta.dual=true but vocab_query.bin absent or wrong size — falling back to passage vectors")
		}
	}

	// ── Load sources.json for category metadata (optional) ───────────────────
	var sources []string
	if data, err := os.ReadFile(filepath.Join(dir, "sources.json")); err == nil {
		_ = json.Unmarshal(data, &sources)
	}

	// ── Build WordMap ─────────────────────────────────────────────────────────
	wordMap := make(map[string]WordEntry, meta.N)
	for i, word := range vocab {
		key := normalizeWordKey(word)
		if key == "" {
			continue
		}

		wordType := "general"
		if i < len(sources) && sources[i] != "" {
			wordType = sources[i]
		}

		// Slice directly into the raw backing array — no per-word allocation.
		// Sanitize any NaN/Inf in-place before storing.
		row := raw[i*meta.Dims : (i+1)*meta.Dims]
		for j, v := range row {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				row[j] = 0
			}
		}

		entry := WordEntry{
			Word:       word,
			Type:       wordType,
			WordVector: row,
		}

		if rawQuery != nil {
			qrow := rawQuery[i*meta.Dims : (i+1)*meta.Dims]
			for j, v := range qrow {
				if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
					qrow[j] = 0
				}
			}
			entry.QueryVector = qrow
		}

		wordMap[key] = entry
	}

	dualNote := ""
	if rawQuery != nil {
		dualNote = " (dual passage+query vectors)"
	}
	log.Printf("words: loaded %d entries from binary files (dims=%d)%s", len(wordMap), meta.Dims, dualNote)
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
// Accepts both the new format ([{"word":"…","type":"…"}]) and the legacy
// format (["word",…]) — legacy entries get type "general".
func LoadTargets() []Target {
	p := filepath.Join(baseFileDirectory(), "targets.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}

	// Try new format first.
	var targets []Target
	if err := json.Unmarshal(data, &targets); err == nil && len(targets) > 0 {
		log.Printf("words: loaded %d Contexto targets", len(targets))
		return targets
	}

	// Fall back to legacy flat string list.
	var legacy []string
	if err := json.Unmarshal(data, &legacy); err != nil {
		log.Printf("words: could not parse targets.json: %v", err)
		return nil
	}
	targets = make([]Target, len(legacy))
	for i, w := range legacy {
		targets[i] = Target{Word: w, Type: "general"}
	}
	log.Printf("words: loaded %d Contexto targets (legacy format)", len(targets))
	return targets
}

// LoadLemmaMap reads lemma_map.json produced by stage 5/6.
// It maps lowercased surface forms to their canonical lemma, e.g. "bilar" → "bil".
// Returns nil if the file is absent (callers treat nil as no-op).
func LoadLemmaMap() map[string]string {
	p := filepath.Join(baseFileDirectory(), "lemma_map.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		log.Printf("words: could not parse lemma_map.json: %v", err)
		return nil
	}
	log.Printf("words: loaded %d lemma mappings", len(m))
	return m
}

// ErrNoBinaryFiles is returned when the binary word files are absent or malformed.
var ErrNoBinaryFiles = fmt.Errorf("binary wordfiles not found")
