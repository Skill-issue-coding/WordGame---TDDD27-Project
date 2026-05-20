// Package util provides stateless helper functions used across the server.
package util

import (
	"fmt"
	"math"
	"math/rand"
)

// palette is the set of background colors a player can be assigned.
// Kept in sync with the frontend palette in GameContextProvider.
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
	"Glad", "Mystisk", "Knasig", "Luden", "Blöt",
	"Stolt", "Förvirrad", "Listig", "Modig", "Klumpig",
}

var nouns = []string{
	"Älg", "Kråka", "Igelkott", "Valross", "Lama",
	"Bäver", "Flamingo", "Pingvin", "Vessla", "Mård",
	"Lobster", "Axolotl", "Mufflon", "Tapir", "Manet",
}

// GenerateGameCode returns a random room code in the format "xxxx-xxxx",
// where each segment is four alphanumeric characters. Used as the human-readable
// lobby identifier players share with friends.
func GenerateGameCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

	segment := func() string {
		b := make([]byte, 4)
		for i := range b {
			b[i] = chars[rand.Intn(len(chars))]
		}
		return string(b)
	}

	return fmt.Sprintf("%s-%s", segment(), segment())
}

// CosineDistance computes the cosine distance between two equal-length float64
// vectors. Returns a value in [0, 2] where 0 means identical direction and 2
// means opposite direction. Returns math.NaN() if either vector is empty,
// they differ in length, or either has zero magnitude.
//
// This is the core similarity primitive used by all game modes to compare
// Swedish fastText word vectors.
func CosineDistance(vecA []float32, vecB []float32) float64 {
	if len(vecA) == 0 || len(vecB) == 0 || len(vecA) != len(vecB) {
		return math.NaN()
	}

	var dot, normA, normB float64

	for i := range vecA {
		a, b := float64(vecA[i]), float64(vecB[i])
		dot += a * b
		normA += a * a
		normB += b * b
	}

	if normA == 0 || normB == 0 {
		return math.NaN()
	}

	cosineSimilarity := dot / (math.Sqrt(normA) * math.Sqrt(normB))
	return 1 - cosineSimilarity
}

// GenerateUsername returns a random Swedish display name in the format
// "<Adjective><Noun><Number>", e.g. "KnasigFlamingo42".
// Called once per WebSocket connection in the HTTP upgrade handler.
func GenerateUsername() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	return fmt.Sprintf("%s%s%d", adj, noun, rand.Intn(100))
}

// GenerateBackgroundColor returns a random color hex string from the shared
// player palette. Called once per WebSocket connection alongside GenerateUsername.
func GenerateBackgroundColor() string {
	return palette[rand.Intn(len(palette))]
}

// Helper functions to ensure client-provided values remain within safe bounds
// Example, used in validating client inputs and applying gamesettings on serverside
func ClampInt(val float64, min int, max int) int {
	v := int(val)
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func ClampFloat(val float64, min float64, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}