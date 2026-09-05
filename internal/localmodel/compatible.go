package localmodel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CompatibleRuntime struct {
	Name    string
	BaseURL string
	Client  *http.Client
}

func NewLMStudio(baseURL string) *CompatibleRuntime {
	return newCompatible("LM Studio", baseURL, "http://127.0.0.1:1234/v1")
}
func NewLlamaCPP(baseURL string) *CompatibleRuntime {
	return newCompatible("llama.cpp", baseURL, "http://127.0.0.1:8080/v1")
}

func newCompatible(name, baseURL, fallback string) *CompatibleRuntime {
	if baseURL == "" {
		baseURL = fallback
	}
	return &CompatibleRuntime{Name: name, BaseURL: strings.TrimRight(baseURL, "/"), Client: &http.Client{Timeout: 10 * time.Second}}
}

func (r *CompatibleRuntime) Status(ctx context.Context) RuntimeStatus {
	if _, err := r.ListModels(ctx); err != nil {
		return RuntimeStatus{Status: StatusOffline, Error: err}
	}
	return RuntimeStatus{Installed: true, Running: true, Status: StatusOnline}
}

func (r *CompatibleRuntime) ListModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s unavailable: %w", r.Name, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%s returned %s", r.Name, response.Status)
	}
	var payload struct {
		Data []struct {
			ID      string `json:"id"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	models := make([]Model, 0, len(payload.Data))
	for _, item := range payload.Data {
		models = append(models, Model{Runtime: r.Name, Name: item.ID, DisplayName: item.ID, Family: item.OwnedBy})
	}
	return models, nil
}
