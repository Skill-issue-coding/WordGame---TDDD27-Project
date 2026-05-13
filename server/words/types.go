package words

// Target is one entry from targets.json — a word with its category.
type Target struct {
	Word string `json:"word"`
	Type string `json:"type"`
}

type Dictionary struct {
	ActiveWord string
	WordMap    map[string]WordEntry
	// Targets is the curated Contexto target list from targets.json.
	// Nil means fall back to random selection from the full WordMap.
	Targets []Target
}

type RelatedWord struct {
	Word       string
	Similarity float64
}

type WordEntry struct {
	Word        string
	Type        string
	WordVector  []float32 // passage vector — used when this word is the target
	QueryVector []float32 // query vector — used when this word is guessed; falls back to WordVector if nil
	Popularity  float64
	Sitelinks   float64
	Score       float64
	IsSeed      bool
}
