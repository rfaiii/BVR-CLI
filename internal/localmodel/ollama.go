package localmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"charm.land/catwalk/pkg/catwalk"
)

const DefaultOllamaURL = "http://127.0.0.1:11434"

var ErrOllamaUnavailable = errors.New("ollama is unavailable")

type Ollama struct {
	BaseURL  string
	Client   *http.Client
	LookPath func(string) (string, error)
	Command  func(context.Context, string, ...string) ([]byte, error)
}

type PullProgress struct {
	Status    string
	Digest    string
	Total     int64
	Completed int64
}

func NewOllama(baseURL string) *Ollama {
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_HOST")
	}
	if baseURL == "" {
		baseURL = DefaultOllamaURL
	}
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}
	return &Ollama{BaseURL: strings.TrimRight(baseURL, "/"), Client: &http.Client{Timeout: 10 * time.Second}, LookPath: exec.LookPath, Command: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).Output()
	}}
}

func (o *Ollama) Status(ctx context.Context) RuntimeStatus {
	status := RuntimeStatus{Status: StatusUnavailable}
	if o.LookPath != nil {
		_, err := o.LookPath("ollama")
		status.Installed = err == nil
	}
	// The Ollama API is the source of truth for installed models. A GUI-launched
	// BVR may not inherit the shell PATH, so a missing CLI must not hide a
	// healthy local daemon (and its /api/tags model list).
	if status.Installed && o.Command != nil {
		if version, err := o.Command(ctx, "ollama", "--version"); err == nil {
			status.Version = strings.TrimSpace(string(version))
		}
	}
	if err := o.Health(ctx); err != nil {
		if !status.Installed {
			return status
		}
		status.Status, status.Error = StatusOffline, err
		return status
	}
	status.Installed = true
	status.Running, status.Status = true, StatusOnline
	return status
}

func (o *Ollama) Health(ctx context.Context) error {
	var response struct {
		Models []any `json:"models"`
	}
	if err := o.request(ctx, http.MethodGet, "/api/tags", nil, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrOllamaUnavailable, err)
	}
	return nil
}

func (o *Ollama) ListModels(ctx context.Context) ([]Model, error) {
	var response struct {
		Models []struct {
			Name       string    `json:"name"`
			Model      string    `json:"model"`
			Size       int64     `json:"size"`
			ModifiedAt time.Time `json:"modified_at"`
			Details    struct {
				Family            string `json:"family"`
				ParameterSize     string `json:"parameter_size"`
				QuantizationLevel string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}
	if err := o.request(ctx, http.MethodGet, "/api/tags", nil, &response); err != nil {
		return nil, err
	}
	models := make([]Model, 0, len(response.Models))
	for _, item := range response.Models {
		name := item.Name
		if name == "" {
			name = item.Model
		}
		lower := strings.ToLower(name)
		coding := strings.Contains(lower, "coder") || strings.Contains(lower, "code") || strings.Contains(lower, "starcoder") || strings.Contains(lower, "codellama") || strings.Contains(lower, "qwen2.5-coder") || strings.Contains(lower, "deepseek-coder")
		models = append(models, Model{Name: name, DisplayName: name, Size: item.Size, ModifiedAt: item.ModifiedAt, Family: item.Details.Family, ParameterSize: item.Details.ParameterSize, Quantization: item.Details.QuantizationLevel, CodingCapable: coding})
	}
	return models, nil
}

func (o *Ollama) ShowModel(ctx context.Context, name string) (Model, error) {
	var response struct {
		ModelInfo map[string]any `json:"model_info"`
		Details   struct {
			Family            string `json:"family"`
			ParameterSize     string `json:"parameter_size"`
			QuantizationLevel string `json:"quantization_level"`
		} `json:"details"`
		Capabilities []string `json:"capabilities"`
	}
	if err := o.request(ctx, http.MethodPost, "/api/show", map[string]string{"model": name}, &response); err != nil {
		return Model{}, err
	}
	model := Model{Name: name, DisplayName: name, Family: response.Details.Family, ParameterSize: response.Details.ParameterSize, Quantization: response.Details.QuantizationLevel, Capabilities: response.Capabilities}
	model.ContextWindow = contextLength(response.ModelInfo)
	return model, nil
}

func (o *Ollama) Models(ctx context.Context) ([]Model, error) {
	models, err := o.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	for i := range models {
		metadata, showErr := o.ShowModel(ctx, models[i].Name)
		if showErr == nil {
			models[i].ContextWindow, models[i].Family, models[i].ParameterSize, models[i].Quantization, models[i].Capabilities = metadata.ContextWindow, metadata.Family, metadata.ParameterSize, metadata.Quantization, metadata.Capabilities
		}
	}
	return models, nil
}

// LoadModel asks Ollama to load a model into memory without generating a
// response. Ollama loads models lazily, so selecting a model alone does not
// necessarily start it. The keep-alive value keeps it warm for subsequent
// prompts while still allowing Ollama to reclaim memory later.
func (o *Ollama) LoadModel(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("ollama model name is required")
	}
	if err := o.Health(ctx); err != nil {
		if startErr := o.Start(ctx); startErr != nil {
			return startErr
		}
		deadline := time.NewTimer(30 * time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			if healthErr := o.Health(ctx); healthErr == nil {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-deadline.C:
				return fmt.Errorf("ollama did not become ready: %w", err)
			case <-ticker.C:
			}
		}
	}
	var response struct {
		Error string `json:"error"`
	}
	client := o.Client
	if client == nil {
		client = &http.Client{}
	} else {
		copy := *client
		copy.Timeout = 2 * time.Minute
		client = &copy
	}
	if err := o.requestWithClient(ctx, client, http.MethodPost, "/api/generate", map[string]any{
		"model":      name,
		"prompt":     "",
		"stream":     false,
		"keep_alive": "5m",
	}, &response); err != nil {
		return err
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return nil
}

func (o *Ollama) CatwalkModels(ctx context.Context) ([]catwalk.Model, error) {
	models, err := o.Models(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]catwalk.Model, 0, len(models))
	for _, model := range models {
		result = append(result, catwalk.Model{ID: model.Name, Name: model.DisplayName, ContextWindow: model.ContextWindow, DefaultMaxTokens: model.MaxTokens, SupportsImages: contains(model.Capabilities, "vision")})
	}
	return result, nil
}

func (o *Ollama) Pull(ctx context.Context, name string, progress func(PullProgress)) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("model name is required")
	}
	data, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/api/pull", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := o.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("ollama pull returned %s", response.Status)
	}
	decoder := json.NewDecoder(response.Body)
	for {
		var item struct {
			Status    string `json:"status"`
			Digest    string `json:"digest"`
			Total     int64  `json:"total"`
			Completed int64  `json:"completed"`
		}
		err := decoder.Decode(&item)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if progress != nil {
			progress(PullProgress{Status: item.Status, Digest: item.Digest, Total: item.Total, Completed: item.Completed})
		}
	}
}

func (o *Ollama) Start(ctx context.Context) error {
	if o.LookPath == nil {
		return errors.New("ollama executable detection unavailable")
	}
	ollamaPath, err := o.LookPath("ollama")
	if err != nil {
		for _, candidate := range []string{
			"/Applications/Ollama.app/Contents/Resources/ollama",
			"/usr/local/bin/ollama",
			"/opt/homebrew/bin/ollama",
		} {
			if _, statErr := os.Stat(candidate); statErr == nil {
				ollamaPath = candidate
				break
			}
		}
		if ollamaPath == "" {
			return fmt.Errorf("ollama executable not found: %w", err)
		}
	}
	// The desktop application and system service commonly start Ollama for us.
	// Do not launch a second server (and do not attach a long-running process to
	// BVR's terminal) when the API is already healthy.
	if err := o.Health(ctx); err == nil {
		return nil
	}
	// Do not bind the daemon lifetime to the TUI command context. A successful
	// start must survive the short-lived readiness/warm-up operation.
	command := exec.Command(ollamaPath, "serve")
	command.Stdin = nil
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		return fmt.Errorf("start ollama: %w", err)
	}
	// Reap the detached daemon when it exits so it cannot become a zombie.
	go func() { _ = command.Wait() }()
	return nil
}

func (o *Ollama) request(ctx context.Context, method, endpoint string, body any, output any) error {
	client := o.Client
	if client == nil {
		client = http.DefaultClient
	}
	return o.requestWithClient(ctx, client, method, endpoint, body, output)
}

func (o *Ollama) requestWithClient(ctx context.Context, client *http.Client, method, endpoint string, body any, output any) error {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, o.BaseURL+path.Clean("/"+endpoint), reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("ollama returned %s", response.Status)
	}
	return json.NewDecoder(response.Body).Decode(output)
}

func contextLength(info map[string]any) int64 {
	for key, value := range info {
		if !strings.HasSuffix(key, ".context_length") {
			continue
		}
		switch number := value.(type) {
		case float64:
			return int64(number)
		case json.Number:
			parsed, _ := strconv.ParseInt(string(number), 10, 64)
			return parsed
		}
	}
	return 0
}
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
