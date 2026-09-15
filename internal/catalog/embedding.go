package catalog

import (
	"encoding/json"
	"errors"
	"math"
)

// TextEmbedder produces a dense vector for a piece of text. In production this
// calls the Ollama /api/embed endpoint with the nomic-embed-text model. In
// tests it can be stubbed to avoid network calls.
type TextEmbedder func(text string) ([]float64, error)

// CosineSimilarity returns the cosine similarity between two vectors of equal
// length. It returns 0 if either vector has zero magnitude.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	magA := math.Sqrt(normA)
	magB := math.Sqrt(normB)
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (magA * magB)
}

// DecodeEmbedding reads the embedding_json column from product_embeddings and
// returns the vector.
func DecodeEmbedding(data string) ([]float64, error) {
	var vec []float64
	if err := json.Unmarshal([]byte(data), &vec); err != nil {
		return nil, errors.New("invalid embedding JSON")
	}
	return vec, nil
}

// EncodeEmbedding serialises a vector to JSON for storage.
func EncodeEmbedding(vec []float64) (string, error) {
	data, err := json.Marshal(vec)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
