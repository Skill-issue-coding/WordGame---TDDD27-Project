package words

import (
	"errors"
	"log"
	"math"
	"math/rand/v2"
	"server/util"
	"sort"
	"strings"
)

const (
	minRelatedSimilarity = 0.45
	maxRelatedSimilarity = 0.92
	relatedTargetMean    = 0.72
	relatedTargetSigma   = 0.08
	topicTypeSuffix      = "_topic"
)

func InitializeDictionary() (Dictionary, error) {
	// Prefer the compact binary format produced by stage 6; fall back to CSVs.
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
// "bilar" → "bil", "röda" → "röd".  Falls back to the lowercased input when
// the word is not in the map (entities, already-canonical lemmas, etc.).
func (dictionary *Dictionary) Resolve(word string) string {
	key := strings.ToLower(strings.TrimSpace(word))
	if dictionary != nil && dictionary.LemmaMap != nil {
		if lemma, ok := dictionary.LemmaMap[key]; ok {
			return lemma
		}
	}
	return key
}

func (dictionary *Dictionary) lookup(word string) (WordEntry, bool) {
	if dictionary == nil || len(dictionary.WordMap) == 0 {
		return WordEntry{}, false
	}
	entry, exists := dictionary.WordMap[dictionary.Resolve(word)]
	return entry, exists
}

func (dictionary *Dictionary) CalculateDistance(word string) float64 {
	activeWordEntry, activeWordExists := dictionary.lookup(dictionary.ActiveWord)
	guessEntry, guessExists := dictionary.lookup(word)

	if !activeWordExists || !guessExists {
		return math.NaN()
	}

	// Use asymmetric passage(target) · query(guess) when available.
	guessVec := guessEntry.QueryVector
	if len(guessVec) == 0 {
		guessVec = guessEntry.WordVector
	}
	return util.CosineDistance(activeWordEntry.WordVector, guessVec)
}

func (dictionary *Dictionary) IsValid(word string) bool {
	_, exists := dictionary.lookup(word)
	return exists
}

// SetRandomContextoTarget picks a random word from the curated target list.
// Falls back to SetRandomActiveWord if no target list is loaded.
func (dictionary *Dictionary) SetRandomContextoTarget() error {
	return dictionary.setContextoTarget(dictionary.Targets)
}

// SetRandomContextoTargetByType picks a random target whose type matches
// category (case-insensitive). Falls back to SetRandomContextoTarget if no
// matching targets are found.
func (dictionary *Dictionary) SetRandomContextoTargetByType(category string) error {
	cat := strings.ToLower(strings.TrimSpace(category))
	filtered := make([]Target, 0)
	for _, t := range dictionary.Targets {
		if strings.ToLower(strings.TrimSpace(t.Type)) == cat {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) == 0 {
		return dictionary.SetRandomContextoTarget()
	}
	return dictionary.setContextoTarget(filtered)
}

func (dictionary *Dictionary) setContextoTarget(pool []Target) error {
	if len(pool) > 0 {
		t := pool[rand.IntN(len(pool))]
		if _, ok := dictionary.lookup(t.Word); ok {
			dictionary.ActiveWord = t.Word
			return nil
		}
	}
	return dictionary.SetRandomActiveWord()
}

func (dictionary *Dictionary) SetRandomActiveWord() error {
	entry, err := dictionary.RandomWordByPopularity(RANDOM_WORD_ALLOWED_POS_TYPES)
	if err != nil {
		entry, err = dictionary.RandomWord()
		if err != nil {
			return err
		}
	}

	dictionary.ActiveWord = entry.Word
	return nil
}

func (dictionary *Dictionary) RandomWord() (WordEntry, error) {
	if len(dictionary.WordMap) == 0 {
		return WordEntry{}, errors.New("word map is empty")
	}

	keys := make([]string, 0, len(dictionary.WordMap))
	for key := range dictionary.WordMap {
		keys = append(keys, key)
	}

	entry := dictionary.WordMap[keys[rand.IntN(len(keys))]]
	return entry, nil
}

func normalizeAllowedTypes(allowedTypes []string) map[string]struct{} {
	if len(allowedTypes) == 0 {
		return nil
	}

	set := make(map[string]struct{}, len(allowedTypes))
	for _, posType := range allowedTypes {
		normalized := strings.ToLower(strings.TrimSpace(posType))
		if normalized != "" {
			set[normalized] = struct{}{}
		}
	}

	return set
}

func (dictionary *Dictionary) entriesByAllowedTypes(allowedTypes []string) []WordEntry {
	allowedTypeSet := normalizeAllowedTypes(allowedTypes)
	entries := make([]WordEntry, 0, len(dictionary.WordMap))

	for _, entry := range dictionary.WordMap {
		if allowedTypeSet != nil {
			entryType := strings.ToLower(strings.TrimSpace(entry.Type))
			if _, ok := allowedTypeSet[entryType]; !ok {
				continue
			}
		}
		entries = append(entries, entry)
	}

	return entries
}

func (dictionary *Dictionary) RandomWordByAllowedPOSTypes(wordMap map[string]WordEntry, allowedTypes []string) (WordEntry, error) {
	if len(wordMap) == 0 {
		return WordEntry{}, errors.New("word map is empty")
	}

	if len(allowedTypes) == 0 {
		return WordEntry{}, errors.New("allowed POS types are empty")
	}

	allowedTypeSet := normalizeAllowedTypes(allowedTypes)
	filteredEntries := make([]WordEntry, 0)
	for _, entry := range wordMap {
		if _, ok := allowedTypeSet[strings.ToLower(strings.TrimSpace(entry.Type))]; ok {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	if len(filteredEntries) == 0 {
		return WordEntry{}, errors.New("no words matched allowed POS types")
	}

	return filteredEntries[rand.IntN(len(filteredEntries))], nil
}

func sampleNormal() float64 {
	u1 := rand.Float64()
	if u1 <= 0 {
		u1 = 1e-9
	}
	u2 := rand.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

func clamp(value float64, low float64, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func comparableWord(value string) string {
	value = strings.ToLower(value)
	builder := strings.Builder{}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == 'å' || r == 'ä' || r == 'ö' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func topicTypeFor(primaryType string) string {
	normalized := strings.ToLower(strings.TrimSpace(primaryType))
	if normalized == "" {
		return ""
	}
	if strings.HasSuffix(normalized, topicTypeSuffix) {
		return normalized
	}
	return normalized + topicTypeSuffix
}

func weightedChoice(entries []WordEntry, weights []float64) (WordEntry, error) {
	if len(entries) == 0 || len(entries) != len(weights) {
		return WordEntry{}, errors.New("invalid weighted choice input")
	}

	var total float64
	for _, weight := range weights {
		total += weight
	}
	if total <= 0 {
		return WordEntry{}, errors.New("weights sum to zero")
	}

	threshold := rand.Float64() * total
	var cumulative float64
	for i, weight := range weights {
		cumulative += weight
		if threshold <= cumulative {
			return entries[i], nil
		}
	}

	return entries[len(entries)-1], nil
}

func (dictionary *Dictionary) RandomWordByPopularity(allowedTypes []string) (WordEntry, error) {
	entries := dictionary.entriesByAllowedTypes(allowedTypes)
	if len(entries) == 0 {
		return WordEntry{}, errors.New("no words matched allowed types")
	}

	seedEntries := make([]WordEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsSeed {
			seedEntries = append(seedEntries, entry)
		}
	}
	if len(seedEntries) >= 20 {
		entries = seedEntries
	}

	if len(entries) == 1 {
		return entries[0], nil
	}

	popularities := make([]float64, len(entries))
	var sum float64
	for i, entry := range entries {
		popularity := entry.Popularity
		if popularity <= 0 {
			popularity = 1
		}
		popularities[i] = popularity
		sum += popularity
	}

	mean := sum / float64(len(popularities))
	var variance float64
	for _, popularity := range popularities {
		delta := popularity - mean
		variance += delta * delta
	}
	std := math.Sqrt(variance / float64(len(popularities)))
	if std <= 0 {
		return entries[rand.IntN(len(entries))], nil
	}

	targetZ := clamp(sampleNormal(), -2.0, 2.0)
	weights := make([]float64, len(entries))
	for i, popularity := range popularities {
		z := (popularity - mean) / std
		distributionWeight := math.Exp(-0.5 * math.Pow((z-targetZ)/0.75, 2))
		popularityFloor := 0.25 + math.Min(1.0, popularity/(mean+std))
		weights[i] = distributionWeight * popularityFloor
	}

	return weightedChoice(entries, weights)
}

func (dictionary *Dictionary) ClosestWords(word string, limit int, allowedTypes []string) ([]RelatedWord, error) {
	if limit <= 0 {
		return []RelatedWord{}, nil
	}

	sourceEntry, exists := dictionary.lookup(word)
	if !exists {
		return nil, errors.New("word not found in dictionary")
	}

	allowedTypeSet := normalizeAllowedTypes(allowedTypes)
	related := make([]RelatedWord, 0, limit)
	sourceKey := normalizeWordKey(sourceEntry.Word)

	for key, candidate := range dictionary.WordMap {
		if key == sourceKey {
			continue
		}

		if allowedTypeSet != nil {
			candidateType := strings.ToLower(strings.TrimSpace(candidate.Type))
			if _, ok := allowedTypeSet[candidateType]; !ok {
				continue
			}
		}

		distance := util.CosineDistance(sourceEntry.WordVector, candidate.WordVector)
		if math.IsNaN(distance) {
			continue
		}
		similarity := 1 - distance
		if math.IsNaN(similarity) || math.IsInf(similarity, 0) {
			continue
		}

		related = append(related, RelatedWord{
			Word:       candidate.Word,
			Similarity: similarity,
		})
	}

	sort.Slice(related, func(i int, j int) bool {
		if related[i].Similarity == related[j].Similarity {
			return related[i].Word < related[j].Word
		}
		return related[i].Similarity > related[j].Similarity
	})

	if len(related) > limit {
		return related[:limit], nil
	}
	return related, nil
}

func (dictionary *Dictionary) RandomRelatedPair(allowedTypes []string) (WordEntry, WordEntry, error) {
	for attempts := 0; attempts < 8; attempts++ {
		primary, err := dictionary.RandomWordByPopularity(allowedTypes)
		if err != nil {
			return WordEntry{}, WordEntry{}, err
		}

		primaryType := strings.ToLower(strings.TrimSpace(primary.Type))
		relatedTypeSet := map[string]struct{}{}
		if primaryType != "" {
			relatedTypeSet[primaryType] = struct{}{}
			if topicType := topicTypeFor(primaryType); topicType != "" {
				relatedTypeSet[topicType] = struct{}{}
			}
		}

		relatedAllowedTypes := make([]string, 0, len(relatedTypeSet))
		for typeName := range relatedTypeSet {
			relatedAllowedTypes = append(relatedAllowedTypes, typeName)
		}

		relatedCandidates, err := dictionary.ClosestWords(primary.Word, 60, relatedAllowedTypes)
		if err != nil {
			continue
		}

		filteredWords := make([]WordEntry, 0)
		filteredWeights := make([]float64, 0)
		for _, candidate := range relatedCandidates {
			if candidate.Similarity < minRelatedSimilarity || candidate.Similarity > maxRelatedSimilarity {
				continue
			}

			entry, exists := dictionary.lookup(candidate.Word)
			if !exists {
				continue
			}
			candidateType := strings.ToLower(strings.TrimSpace(entry.Type))
			if _, ok := relatedTypeSet[candidateType]; !ok {
				continue
			}

			minPopularity := 20.0
			if strings.HasSuffix(candidateType, topicTypeSuffix) {
				minPopularity = 5.0
			}
			if !entry.IsSeed && entry.Popularity < minPopularity {
				continue
			}
			primaryLower := strings.ToLower(primary.Word)
			relatedLower := strings.ToLower(entry.Word)
			if strings.Contains(primaryLower, relatedLower) || strings.Contains(relatedLower, primaryLower) {
				continue
			}
			if comparableWord(primary.Word) == comparableWord(entry.Word) {
				continue
			}

			similarityWeight := math.Exp(-0.5 * math.Pow((candidate.Similarity-relatedTargetMean)/relatedTargetSigma, 2))
			popularityWeight := 0.4 + math.Min(1.0, math.Max(0.0, entry.Popularity/(primary.Popularity+1)))
			weight := similarityWeight * popularityWeight
			if weight <= 0 {
				continue
			}

			filteredWords = append(filteredWords, entry)
			filteredWeights = append(filteredWeights, weight)
		}

		if len(filteredWords) == 0 {
			continue
		}

		related, err := weightedChoice(filteredWords, filteredWeights)
		if err != nil {
			continue
		}

		return primary, related, nil
	}

	return WordEntry{}, WordEntry{}, errors.New("failed to find related word pair")
}

// RandomImpostorPairFromTargets picks a normal/impostor word pair using the
// precomputed impostor_candidates list in targets.json (stage 9). Falls back
// to RandomRelatedPair when no enriched targets are available.
func (dictionary *Dictionary) RandomImpostorPairFromTargets() (WordEntry, WordEntry, error) {
	eligible := make([]Target, 0, len(dictionary.Targets))
	for _, t := range dictionary.Targets {
		if len(t.ImpostorCandidates) >= 3 {
			eligible = append(eligible, t)
		}
	}

	if len(eligible) == 0 {
		log.Printf("words: no enriched impostor targets — falling back to RandomRelatedPair")
		return dictionary.RandomRelatedPair(IMPOSTOR_PRIMARY_TYPES)
	}

	// Shuffle candidates so we don't always try them in pipeline order.
	perm := rand.Perm(len(eligible))
	for _, pi := range perm {
		t := eligible[pi]
		normalEntry, ok := dictionary.lookup(t.Word)
		if !ok {
			continue
		}

		// Try impostor candidates in random order.
		candPerm := rand.Perm(len(t.ImpostorCandidates))
		for _, ci := range candPerm {
			impostorEntry, ok := dictionary.lookup(t.ImpostorCandidates[ci])
			if !ok {
				continue
			}
			return normalEntry, impostorEntry, nil
		}
	}

	log.Printf("words: all enriched impostor targets failed lookup — falling back to RandomRelatedPair")
	return dictionary.RandomRelatedPair(IMPOSTOR_PRIMARY_TYPES)
}
