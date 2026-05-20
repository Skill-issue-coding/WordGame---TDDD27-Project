package words

import (
	"fmt"
	"strings"
)

func normalizeWordKey(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}

// inferType returns a generic type string; without per-word type metadata in
// the binary format we fall back to "general". Entities (proper nouns /
// named things) end up here as "general" which is fine for random-word
// selection — RANDOM_WORD_ALLOWED_POS_TYPES already includes "general".
func inferType(_ string, _ map[string]WordEntry) string {
	return "general"
}

// ErrNoBinaryFiles is returned when the binary word files are absent or malformed.
var ErrNoBinaryFiles = fmt.Errorf("binary wordfiles not found")
