package localmodel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOllamaStatusAndModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/tags":
			json.NewEncoder(w).Encode(map[string]any{"models": []any{map[string]any{"name": "qwen:latest", "size": 123, "details": map[string]string{"family": "qwen2"}}}})
		case "/api/show":
			json.NewEncoder(w).Encode(map[string]any{"model_info": map[string]any{"qwen2.context_length": float64(32768)}, "details": map[string]string{"family": "qwen2"}, "capabilities": []string{"completion"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	ollama := NewOllama(server.URL)
	ollama.LookPath = func(string) (string, error) { return "/usr/bin/ollama", nil }
	ollama.Command = func(context.Context, string, ...string) ([]byte, error) { return []byte("ollama version 0.1"), nil }
	status := ollama.Status(context.Background())
	require.Equal(t, StatusOnline, status.Status)
	require.Equal(t, "ollama version 0.1", status.Version)
	models, err := ollama.Models(context.Background())
	require.NoError(t, err)
	require.Len(t, models, 1)
	require.Equal(t, int64(32768), models[0].ContextWindow)
}

func TestOllamaUnavailable(t *testing.T) {
	ollama := NewOllama("http://127.0.0.1:1")
	ollama.LookPath = func(string) (string, error) { return "/usr/bin/ollama", nil }
	status := ollama.Status(context.Background())
	require.Equal(t, StatusOffline, status.Status)
	require.Error(t, status.Error)
}

func TestOllamaExecutableUnavailable(t *testing.T) {
	ollama := NewOllama("http://127.0.0.1:1")
	ollama.LookPath = func(string) (string, error) { return "", ErrOllamaUnavailable }
	status := ollama.Status(context.Background())
	require.Equal(t, StatusUnavailable, status.Status)
	require.False(t, status.Installed)
}

func TestOllamaAPIIsDiscoverableWithoutCLIOnPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/tags", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"qwen3:4b"},{"name":"llama3.2:latest"}]}`))
	}))
	defer server.Close()

	ollama := NewOllama(server.URL)
	ollama.LookPath = func(string) (string, error) { return "", ErrOllamaUnavailable }
	status := ollama.Status(context.Background())
	require.Equal(t, StatusOnline, status.Status)
	require.True(t, status.Installed)

	models, err := ollama.ListModels(context.Background())
	require.NoError(t, err)
	require.Len(t, models, 2)
	require.Equal(t, "qwen3:4b", models[0].Name)
	require.Equal(t, "llama3.2:latest", models[1].Name)
}

func TestOllamaLoadModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[]}`))
		case "/api/generate":
			var request map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			require.Equal(t, "qwen3:4b", request["model"])
			require.Equal(t, false, request["stream"])
			require.Equal(t, "5m", request["keep_alive"])
			_, _ = w.Write([]byte(`{"response":""}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ollama := NewOllama(server.URL)
	require.NoError(t, ollama.LoadModel(context.Background(), "qwen3:4b"))
}

func TestOllamaPullStreamsProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/pull", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"pulling","digest":"sha256:abc","total":100,"completed":40}` + "\n" + `{"status":"success"}` + "\n"))
	}))
	defer server.Close()
	ollama := NewOllama(server.URL)
	var progress []PullProgress
	require.NoError(t, ollama.Pull(context.Background(), "qwen", func(value PullProgress) { progress = append(progress, value) }))
	require.Len(t, progress, 2)
	require.Equal(t, int64(40), progress[0].Completed)
}
