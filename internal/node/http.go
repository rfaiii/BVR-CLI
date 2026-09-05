package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const protocolVersion = "1"

type HTTPTransport struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func NewHTTPTransport(baseURL, token string) *HTTPTransport {
	return &HTTPTransport{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, Client: &http.Client{Timeout: 30 * time.Second}}
}

func (t *HTTPTransport) request(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = strings.NewReader(string(data))
	}
	req, err := http.NewRequestWithContext(ctx, method, t.BaseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}
	client := t.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return fmt.Errorf("node HTTP %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(out)
}

func (t *HTTPTransport) Execute(ctx context.Context, request CommandRequest) (CommandResult, error) {
	var response executeResponse
	if err := t.request(ctx, http.MethodPost, "/v1/node/execute", executeRequest{CommandRequest: request}, &response); err != nil {
		return CommandResult{}, err
	}
	result := response.CommandResult
	if response.Error != "" {
		return result, errors.New(response.Error)
	}
	if result.ExitCode != 0 {
		return result, fmt.Errorf("remote command exited with code %d", result.ExitCode)
	}
	return result, nil
}

func (t *HTTPTransport) Ping(ctx context.Context) error {
	var response healthResponse
	if err := t.request(ctx, http.MethodGet, "/v1/node/health", nil, &response); err != nil {
		return err
	}
	if !response.OK {
		return errors.New("node reported unhealthy")
	}
	return nil
}

type executeRequest struct{ CommandRequest }
type executeResponse struct {
	CommandResult
	Error string `json:"error,omitempty"`
}
type healthResponse struct {
	OK       bool   `json:"ok"`
	Protocol string `json:"protocol"`
}

type NodeHTTPServer struct {
	Token     string
	Transport Transport
	MaxBody   int64
}

func (s *NodeHTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	s.registerHTTPRoutes(mux)
	return mux
}

func (s *NodeHTTPServer) registerHTTPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/node/health", s.handleHealth)
	mux.HandleFunc("GET /v1/node/capabilities", s.handleCapabilities)
	mux.HandleFunc("POST /v1/node/execute", s.handleExecute)
}

// NodeAgentHandler serves the request/response HTTP protocol and persistent
// WebSocket streaming protocol from one authenticated listener.
func NodeAgentHandler(token string, transport Transport) http.Handler {
	httpServer := &NodeHTTPServer{Token: token, Transport: transport}
	websocketServer := &NodeWebSocketServer{Token: token, Transport: transport}
	mux := http.NewServeMux()
	httpServer.registerHTTPRoutes(mux)
	mux.HandleFunc(websocketPath, websocketServer.handle)
	return mux
}

// ListenAndServe starts a NODE HTTP agent on addr. The caller owns shutdown
// when embedding Handler directly; this helper is intended for small agents.
func (s *NodeHTTPServer) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

func (s *NodeHTTPServer) authorized(r *http.Request) bool {
	return authorizedBearer(s.Token, r)
}

func authorizedBearer(token string, r *http.Request) bool {
	if token == "" {
		return true
	}
	return strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")) == token
}

func (s *NodeHTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, healthResponse{OK: true, Protocol: protocolVersion})
}

func (s *NodeHTTPServer) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]any{"protocol": protocolVersion, "commands": []string{"node", "npm"}})
}

func (s *NodeHTTPServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	maxBody := s.MaxBody
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	var request executeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(&request); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if request.Command == "" {
		http.Error(w, "missing command", http.StatusBadRequest)
		return
	}
	transport := s.Transport
	if transport == nil {
		transport = LocalTransport{}
	}
	result, err := transport.Execute(r.Context(), request.CommandRequest)
	response := executeResponse{CommandResult: result}
	if err != nil {
		response.Error = err.Error()
	}
	writeJSON(w, response)
}

// RegisterHTTPTransport attaches an authenticated HTTP transport to a device
// and records its endpoint for discovery/UI consumers.
func RegisterHTTPTransport(id, baseURL, token string) error {
	if strings.TrimSpace(baseURL) == "" {
		return errors.New("missing node transport URL")
	}
	RegisterTransport(id, NewHTTPTransport(baseURL, token))
	defaultManager.SetTransportURL(id, baseURL)
	return nil
}

// RegisterNodeTransport selects an authenticated transport from an endpoint
// URL. HTTP is request/response; WebSocket is persistent and streaming.
func RegisterNodeTransport(id, endpoint, token string) error {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid node endpoint: %q", endpoint)
	}
	switch parsed.Scheme {
	case "http", "https":
		RegisterTransport(id, NewHTTPTransport(strings.TrimRight(endpoint, "/"), token))
	case "ws", "wss":
		RegisterTransport(id, NewWebSocketTransport(strings.TrimRight(endpoint, "/"), token))
	default:
		return fmt.Errorf("unsupported node transport scheme %q", parsed.Scheme)
	}
	defaultManager.SetTransportURL(id, endpoint)
	return nil
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
