package words

type Dictionary struct {
	ActiveWord  string
	WordMap     map[string]WordEntry
	// Targets is the curated Contexto target list from targets.json.
	// Nil means fall back to random selection from the full WordMap.
	Targets     []string
}

type RelatedWord struct {
	Word       string
	Similarity float64
}

type WordEntry struct {
	Word       string
	Type       string
	WordVector []float64
	Popularity float64
	Sitelinks  float64
	Score      float64
	IsSeed     bool
}
