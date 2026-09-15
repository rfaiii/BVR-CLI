package imagelookup

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OllamaClient talks to the local Ollama daemon. It is intentionally separate
// from [internal/localmodel.Ollama] (which focuses on model management) because
// the image-lookup flow needs chat-with-vision and embedding helpers.
type OllamaClient struct {
	BaseURL     string
	VisionModel string
	EmbedModel  string
	HTTPClient *http.Client
}

// NewOllamaClient creates a client pointed at the given Ollama base URL.
func NewOllamaClient(baseURL, visionModel, embedModel string) *OllamaClient {
	return &OllamaClient{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		VisionModel: visionModel,
		EmbedModel:  embedModel,
		HTTPClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}
}

type ollamaChatRequest struct {
	Model    string           `json:"model"`
	Stream   bool             `json:"stream"`
	Format   json.RawMessage  `json:"format,omitempty"`
	Messages []ollamaMessage  `json:"messages"`
	Options  ollamaOptions    `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumCtx      int     `json:"num_ctx,omitempty"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
}

// jsonSchemaForAnalysis returns a minimal JSON-schema object that constrains
// the model output to the ImageAnalysis shape. Ollama honours this via the
// "format" field so the response is schema-validated by the model runtime.
func jsonSchemaForAnalysis() json.RawMessage {
	props := map[string]any{}
	for name := range analysisFieldNames() {
		props[name] = map[string]any{
			"type": analysisFieldType(name),
		}
	}
	b, _ := json.Marshal(map[string]any{
		"type":       "object",
		"properties": props,
	})
	return b
}

// analysisFieldNames returns the set of top-level JSON field names in
// ImageAnalysis with a JSON tag. It is used to build the schema above.
func analysisFieldNames() map[string]struct{} {
	return map[string]struct{}{
		"image_quality": {}, "item_category": {}, "probable_brand": {},
		"visible_text": {}, "logos_and_marks": {}, "silhouette": {},
		"materials": {}, "colors": {}, "patterns": {}, "hardware": {},
		"construction": {}, "dimensions_estimate": {}, "distinguishing_cues": {},
		"damage_or_wear": {}, "unknowns": {}, "next_photos_needed": {},
		"authenticity_note": {},
	}
}

func analysisFieldType(name string) string {
	switch name {
	case "image_quality", "authenticity_note":
		return "string"
	default:
		return "array"
	}
}

// AnalyzeImage sends the normalized image at imagePath to the configured
// vision model and returns the parsed ImageAnalysis. The model is instructed
// to return JSON matching the analysis schema only.
func (c *OllamaClient) AnalyzeImage(ctx context.Context, imagePath string) (ImageAnalysis, error) {
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return ImageAnalysis{}, fmt.Errorf("read normalized image: %w", err)
	}

	requestBody, err := json.Marshal(ollamaChatRequest{
		Model:  c.VisionModel,
		Stream: false,
		Format: jsonSchemaForAnalysis(),
		Messages: []ollamaMessage{
			{
				Role:    "user",
				Content: VisionPrompt,
				Images: []string{
					base64.StdEncoding.EncodeToString(imageBytes),
				},
			},
		},
		Options: ollamaOptions{
			Temperature: 0.1,
			NumCtx:      8192,
		},
	})
	if err != nil {
		return ImageAnalysis{}, fmt.Errorf("encode Ollama request: %w", err)
	}

	resp, err := c.doChatRequest(ctx, requestBody)
	if err != nil {
		return ImageAnalysis{}, err
	}

	var analysis ImageAnalysis
	if err := json.Unmarshal([]byte(resp.Message.Content), &analysis); err != nil {
		return ImageAnalysis{}, fmt.Errorf("model returned invalid analysis JSON: %w", err)
	}

	analysis.AuthenticityNote = "Image-based identification is not authentication."

	return analysis, nil
}

// doChatRequest sends a raw chat request and returns the decoded response.
func (c *OllamaClient) doChatRequest(ctx context.Context, body []byte) (ollamaChatResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return ollamaChatResponse{}, fmt.Errorf("create Ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return ollamaChatResponse{}, fmt.Errorf("call Ollama: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(httpResp.Body, 16<<10))
		return ollamaChatResponse{}, fmt.Errorf("Ollama returned %s: %s", httpResp.Status, strings.TrimSpace(string(respBody)))
	}

	var payload ollamaChatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&payload); err != nil {
		return ollamaChatResponse{}, fmt.Errorf("decode Ollama response: %w", err)
	}
	return payload, nil
}

// EmbedText calls the Ollama embedding endpoint with the configured embed
// model (e.g. nomic-embed-text) and returns the resulting vector.
func (c *OllamaClient) EmbedText(ctx context.Context, text string) ([]float64, error) {
	if c.EmbedModel == "" {
		return nil, fmt.Errorf("embed model is not configured")
	}

	requestBody, err := json.Marshal(map[string]any{
		"model": c.EmbedModel,
		"input": text,
	})
	if err != nil {
		return nil, fmt.Errorf("encode embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/embed", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embed API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<10))
		return nil, fmt.Errorf("embed API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Embedding  []float64    `json:"embedding"`
		Embeddings []float64    `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}

	if len(payload.Embedding) > 0 {
		return payload.Embedding, nil
	}
	if len(payload.Embeddings) > 0 {
		return payload.Embeddings, nil
	}
	return nil, fmt.Errorf("no embedding returned")
}

// LoadModel asks Ollama to load the vision model into memory so subsequent
// inference is fast. This mirrors [internal/localmodel.Ollama.LoadModel] but
// keeps the imagelookup package self-contained.
func (c *OllamaClient) LoadModel(ctx context.Context, name string) error {
	requestBody, err := json.Marshal(map[string]any{
		"model":      name,
		"prompt":     "",
		"stream":     false,
		"keep_alive": "5m",
	})
	if err != nil {
		return fmt.Errorf("encode load request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("create load request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("call generate API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<10))
		return fmt.Errorf("generate API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode load response: %w", err)
	}
	if payload.Error != "" {
		return fmt.Errorf("load model: %s", payload.Error)
	}
	return nil
}

// Health checks whether the Ollama daemon is reachable.
func (c *OllamaClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama is unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned %s", resp.Status)
	}
	return nil
}
