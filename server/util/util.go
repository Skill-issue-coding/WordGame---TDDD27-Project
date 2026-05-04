package util

import (
	"fmt"
	"math"
	"math/rand"
)

var palette = []string{
	"#8b5cf6",
	"#ec4899",
	"#3b82f6",
	"#10b981",
	"#f59e0b",
	"#ef4444",
	"#06b6d4",
	"#a855f7",
}

var adjectives = []string{
	"Snabb", "Trög", "Hungrig", "Sömnig", "Arg",
	"Glad glad", "Mystisk", "Knasig", "Luden", "Blöt",
	"Stolt", "Förvirrad", "Listig", "Modig", "Klumpig",
}

var nouns = []string{
	"Älg", "Kråka", "Igelkott", "Valross", "Lama",
	"Bäver", "Flamingo", "Pingvin", "Vessla", "Mård",
	"Lobster", "Axolotl", "Mufflon", "Tapir", "Manet",
}

func GenerateGameCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	segment := func() string {
		b := make([]byte, 4)
		for i := range b {
			b[i] = chars[rand.Intn(len(chars))]
		}
		return string(b)
	}

	return fmt.Sprintf("%s-%s", segment(), segment())
}

func CosineDistance(vecA []float64, vecB []float64) float64 {
	if len(vecA) == 0 || len(vecB) == 0 || len(vecA) != len(vecB) {
		return math.NaN()
	}

	var dot float64
	var normA float64
	var normB float64

	for i := range vecA {
		dot += vecA[i] * vecB[i]
		normA += vecA[i] * vecA[i]
		normB += vecB[i] * vecB[i]
	}

	if normA == 0 || normB == 0 {
		return math.NaN()
	}

	cosineSimilarity := dot / (math.Sqrt(normA) * math.Sqrt(normB))
	return 1 - cosineSimilarity
}

func GenerateUsername() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	number := rand.Intn(100)
	return fmt.Sprintf("%s%s%d", adj, noun, number)
}

func GenerateBackgroundColor() string {
	randomIndex := rand.Intn(len(palette))
	return palette[randomIndex]
}
