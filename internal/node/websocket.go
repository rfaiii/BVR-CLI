package node

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const websocketPath = "/v1/node/ws"

type WSFrame struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Stream  string          `json:"stream,omitempty"`
	Data    string          `json:"data,omitempty"`
	Request *CommandRequest `json:"request,omitempty"`
	Result  *CommandResult  `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type WebSocketTransport struct {
	URL       string
	Token     string
	Dialer    *websocket.Dialer
	KeepAlive time.Duration
	mu        sync.Mutex
	conn      *websocket.Conn
}

func NewWebSocketTransport(url, token string) *WebSocketTransport {
	return &WebSocketTransport{URL: strings.TrimRight(url, "/"), Token: token, Dialer: websocket.DefaultDialer, KeepAlive: 30 * time.Second}
}

func (t *WebSocketTransport) connect(ctx context.Context) (*websocket.Conn, error) {
	if t.conn != nil {
		return t.conn, nil
	}
	dialer := t.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	header := http.Header{}
	if t.Token != "" {
		header.Set("Authorization", "Bearer "+t.Token)
	}
	conn, _, err := dialer.DialContext(ctx, t.URL+websocketPath, header)
	if err != nil {
		return nil, err
	}
	t.conn = conn
	return conn, nil
}

func (t *WebSocketTransport) closeLocked() {
	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}
}

func (t *WebSocketTransport) do(ctx context.Context, frame WSFrame) (CommandResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for attempt := 0; attempt < 2; attempt++ {
		conn, err := t.connect(ctx)
		if err != nil {
			return CommandResult{}, err
		}
		if deadline, ok := ctx.Deadline(); ok {
			_ = conn.SetWriteDeadline(deadline)
			_ = conn.SetReadDeadline(deadline)
		}
		if err := conn.WriteJSON(frame); err != nil {
			t.closeLocked()
			continue
		}
		var result CommandResult
		for {
			var response WSFrame
			if err := conn.ReadJSON(&response); err != nil {
				t.closeLocked()
				break
			}
			switch response.Type {
			case "stdout":
				result.Stdout += response.Data
			case "stderr":
				result.Stderr += response.Data
			case "result":
				if response.Result != nil {
					result.Stdout += response.Result.Stdout
					result.Stderr += response.Result.Stderr
					result.ExitCode = response.Result.ExitCode
				}
				if response.Error != "" {
					return result, errors.New(response.Error)
				}
				return result, nil
			case "pong":
				if frame.Type == "ping" {
					return result, nil
				}
			case "error":
				return result, errors.New(response.Error)
			}
		}
	}
	return CommandResult{}, errors.New("websocket connection closed")
}

func (t *WebSocketTransport) Execute(ctx context.Context, request CommandRequest) (CommandResult, error) {
	return t.do(ctx, WSFrame{ID: fmt.Sprintf("%d", time.Now().UnixNano()), Type: "execute", Request: &request})
}

func (t *WebSocketTransport) Ping(ctx context.Context) error {
	_, err := t.do(ctx, WSFrame{ID: fmt.Sprintf("%d", time.Now().UnixNano()), Type: "ping"})
	return err
}

func (t *WebSocketTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return nil
	}
	err := t.conn.Close()
	t.conn = nil
	return err
}

type NodeWebSocketServer struct {
	Token     string
	Transport Transport
	Upgrader  websocket.Upgrader
}

func (s *NodeWebSocketServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(websocketPath, s.handle)
	return mux
}

func (s *NodeWebSocketServer) authorized(r *http.Request) bool {
	return authorizedBearer(s.Token, r)
}

func (s *NodeWebSocketServer) handle(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	upgrader := s.Upgrader
	if upgrader.CheckOrigin == nil {
		upgrader.CheckOrigin = func(*http.Request) bool { return true }
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		var request WSFrame
		if err := conn.ReadJSON(&request); err != nil {
			return
		}
		switch request.Type {
		case "ping":
			_ = conn.WriteJSON(WSFrame{ID: request.ID, Type: "pong"})
		case "execute":
			s.execute(conn, request)
		default:
			_ = conn.WriteJSON(WSFrame{ID: request.ID, Type: "error", Error: "unknown websocket request"})
		}
	}
}

func (s *NodeWebSocketServer) execute(conn *websocket.Conn, frame WSFrame) {
	if frame.Request == nil || frame.Request.Command == "" {
		_ = conn.WriteJSON(WSFrame{ID: frame.ID, Type: "error", Error: "missing command"})
		return
	}
	transport := s.Transport
	if transport == nil {
		transport = LocalTransport{}
	}
	if streamer, ok := transport.(StreamTransport); ok {
		result, err := streamer.StreamExecute(context.Background(), *frame.Request, func(stream, data string) error {
			return conn.WriteJSON(WSFrame{ID: frame.ID, Type: stream, Stream: stream, Data: data})
		})
		response := WSFrame{ID: frame.ID, Type: "result", Result: &result}
		if err != nil {
			response.Error = err.Error()
		}
		_ = conn.WriteJSON(response)
		return
	}
	result, err := transport.Execute(context.Background(), *frame.Request)
	response := WSFrame{ID: frame.ID, Type: "result", Result: &result}
	if err != nil {
		response.Error = err.Error()
	}
	_ = conn.WriteJSON(response)
}

type StreamTransport interface {
	Transport
	StreamExecute(context.Context, CommandRequest, func(string, string) error) (CommandResult, error)
}

func (LocalTransport) StreamExecute(ctx context.Context, request CommandRequest, emit func(string, string) error) (CommandResult, error) {
	if request.Command == "" {
		return CommandResult{}, errors.New("missing command")
	}
	cmd := exec.CommandContext(ctx, request.Command, request.Args...)
	cmd.Dir = request.WorkingDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return CommandResult{}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return CommandResult{}, err
	}
	if err := cmd.Start(); err != nil {
		return CommandResult{}, err
	}
	var wg sync.WaitGroup
	var streamErr error
	var mu sync.Mutex
	copyStream := func(name string, reader interface{ Read([]byte) (int, error) }) {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			mu.Lock()
			if streamErr == nil {
				streamErr = emit(name, scanner.Text()+"\n")
			}
			mu.Unlock()
			if streamErr != nil {
				return
			}
		}
	}
	wg.Add(2)
	go copyStream("stdout", stdout)
	go copyStream("stderr", stderr)
	wg.Wait()
	err = cmd.Wait()
	result := CommandResult{}
	if err != nil {
		result.ExitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		if streamErr != nil {
			return result, streamErr
		}
		return result, err
	}
	return result, streamErr
}
