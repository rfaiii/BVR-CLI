package imagelookup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func newMockOllama(t *testing.T, handler http.HandlerFunc) *OllamaClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewOllamaClient(srv.URL, "qwen2.5vl:3b", "nomic-embed-text")
}

func TestAnalyzeImage_Success(t *testing.T) {
	t.Parallel()
	analysis := ImageAnalysis{
		ImageQuality:       "good",
		ItemCategory:       Observation{Value: "handbag", Confidence: ConfidenceHigh, Evidence: "tote shape"},
		ProbableBrand:      Observation{Value: "Louis Vuitton", Confidence: ConfidenceMedium, Evidence: "LV monogram"},
		DistinguishingCues: []Observation{{Value: "side laces", Confidence: ConfidenceHigh, Evidence: "visible"}},
		AuthenticityNote:   "Image-based identification is not authentication.",
	}
	analysisJSON, _ := json.Marshal(analysis)

	client := newMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/chat", r.URL.Path)
		var req ollamaChatRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, "qwen2.5vl:3b", req.Model)
		require.Len(t, req.Messages, 1)
		require.NotEmpty(t, req.Messages[0].Images)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: ollamaMessage{
				Role:    "assistant",
				Content: string(analysisJSON),
			},
		})
	})

	result, err := client.AnalyzeImage(context.Background(), writeJPEGTestImage(t, 200, 200))
	require.NoError(t, err)
	require.Equal(t, "good", result.ImageQuality)
	require.Equal(t, "handbag", result.ItemCategory.Value)
	require.Equal(t, ConfidenceHigh, result.ItemCategory.Confidence)
	require.Equal(t, "Louis Vuitton", result.ProbableBrand.Value)
	require.Len(t, result.DistinguishingCues, 1)
	require.Equal(t, "Image-based identification is not authentication.", result.AuthenticityNote)
}

func TestAnalyzeImage_ErrorStatus(t *testing.T) {
	t.Parallel()
	client := newMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("model not found"))
	})
	_, err := client.AnalyzeImage(context.Background(), writeJPEGTestImage(t, 200, 200))
	require.Error(t, err)
	require.Contains(t, err.Error(), "500")
}

func TestEmbedText_Success(t *testing.T) {
	t.Parallel()
	client := newMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/embed", r.URL.Path)
		var req map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, "nomic-embed-text", req["model"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"embedding": []float64{0.1, 0.2, 0.3, 0.4},
		})
	})

	vec, err := client.EmbedText(context.Background(), "handbag; brand: Louis Vuitton")
	require.NoError(t, err)
	require.Equal(t, []float64{0.1, 0.2, 0.3, 0.4}, vec)
}

func TestEmbedText_NoModelConfigured(t *testing.T) {
	t.Parallel()
	client := NewOllamaClient("http://localhost", "", "")
	_, err := client.EmbedText(context.Background(), "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "embed model is not configured")
}

func TestHealth_Success(t *testing.T) {
	t.Parallel()
	client := newMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/tags", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"models": []any{}})
	})
	require.NoError(t, client.Health(context.Background()))
}

func TestHealth_Error(t *testing.T) {
	t.Parallel()
	client := NewOllamaClient("http://127.0.0.1:1", "m", "e")
	err := client.Health(context.Background())
	require.Error(t, err)
}

func TestLoadModel_Success(t *testing.T) {
	t.Parallel()
	client := newMockOllama(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/generate", r.URL.Path)
		var req map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, "qwen2.5vl:3b", req["model"])
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": "", "done": true})
	})
	require.NoError(t, client.LoadModel(context.Background(), "qwen2.5vl:3b"))
}
