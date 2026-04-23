package words

type Dictionary struct {
	ActiveWord string
	WordMap    map[string]WordEntry
}

type WordEntry struct {
	Word       string
	Type       string
	WordVector []float64
}
