package localmodel

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompatibleRuntimeListsModelsAndStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/models", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "qwen", "owned_by": "local"}}})
	}))
	defer server.Close()
	runtime := NewLMStudio(server.URL)
	status := runtime.Status(context.Background())
	require.Equal(t, StatusOnline, status.Status)
	models, err := runtime.ListModels(context.Background())
	require.NoError(t, err)
	require.Equal(t, "LM Studio", models[0].Runtime)
}
