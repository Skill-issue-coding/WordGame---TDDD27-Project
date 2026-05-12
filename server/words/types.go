package words

type Dictionary struct {
	ActiveWord string
	WordMap    map[string]WordEntry
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
