package util

import (
	"fmt"
	"math"
	"math/rand"
)

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
