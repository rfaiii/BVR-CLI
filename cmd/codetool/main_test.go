package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestChatStream_AccumulatesAndCapturesToolCalls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(
			`{"model":"m","message":{"role":"assistant","content":"Hel"},"done":false}` + "\n" +
				`{"model":"m","message":{"role":"assistant","content":"lo"},"done":false}` + "\n" +
				`{"model":"m","message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"bash","arguments":"{\"command\":\"ls\"}"}}]},"done":true}` + "\n",
		))
	}))
	defer srv.Close()

	old := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = old }()

	got, err := chatStream(context.Background(), srv.URL+"/api/chat", ollamaRequest{Model: "m", Stream: true})
	if err != nil {
		t.Fatal(err)
	}
	if got.Message.Content != "Hello" {
		t.Fatalf("expected accumulated content 'Hello', got %q", got.Message.Content)
	}
	if len(got.Message.ToolCalls) != 1 || got.Message.ToolCalls[0].Function.Name != "bash" {
		t.Fatalf("expected captured bash tool call, got %+v", got.Message.ToolCalls)
	}
	if !got.Done {
		t.Fatal("expected done=true")
	}
}

func TestChatStream_RejectsIncompleteResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(`{"model":"m","message":{"role":"assistant","content":"partial"},"done":false}` + "\n"))
	}))
	defer srv.Close()

	old := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = old }()

	if _, err := chatStream(context.Background(), srv.URL+"/api/chat", ollamaRequest{Model: "m", Stream: true}); err == nil {
		t.Fatal("expected incomplete stream error")
	}
}

func TestStreamContentBuffer_HoldsToolCallCandidate(t *testing.T) {
	var b streamContentBuffer
	b.Write("```json\n")
	b.Write(`{"name":"view"}`)
	if !b.candidate {
		t.Fatal("expected markdown JSON to be buffered as a candidate")
	}
	if got := b.pending.String(); got != "```json\n{\"name\":\"view\"}" {
		t.Fatalf("buffered content = %q", got)
	}
}

func TestStreamContentBuffer_DoesNotBufferProse(t *testing.T) {
	var b streamContentBuffer
	b.Write("The answer is")
	if b.candidate || b.pending.Len() != 0 {
		t.Fatal("ordinary prose should not be buffered")
	}
}

func TestParseToolCalls_Reported(t *testing.T) {
	tools := toolNames(flatTools())
	in := []ollToolCall{{Function: ollToolCallFunc{Name: "bash", Arguments: `{"command":"echo hi"}`}}}
	got, rec := parseToolCalls("", in, tools)
	if rec {
		t.Fatal("should not flag recovery when provider supplied tool_calls")
	}
	if len(got) != 1 || got[0].Function.Name != "bash" {
		t.Fatalf("want 1 bash call, got %+v", got)
	}
}

func TestParseToolCalls_RecoversFromText(t *testing.T) {
	tools := toolNames(flatTools())
	content := `{"name":"bash","arguments":"{\"command\":\"echo hi\"}"}`
	got, rec := parseToolCalls(content, nil, tools)
	if !rec {
		t.Fatal("expected recovery=true")
	}
	if len(got) != 1 || got[0].Function.Name != "bash" {
		t.Fatalf("want recovered bash call, got %+v", got)
	}
	if got[0].Function.Arguments != `{"command":"echo hi"}` {
		t.Fatalf("bad arguments: %s", got[0].Function.Arguments)
	}
}

func TestParseToolCalls_NoRecoveryForPlainText(t *testing.T) {
	tools := toolNames(flatTools())
	content := "Here is what this project does:\n\n it is a CLI coding agent."
	got, rec := parseToolCalls(content, nil, tools)
	if rec {
		t.Fatal("plain text must not be recovered as a tool call")
	}
	if len(got) != 0 {
		t.Fatalf("expected no calls, got %+v", got)
	}
}

func TestParseToolCalls_IgnoresUnknownToolName(t *testing.T) {
	tools := toolNames(flatTools())
	// qwen-style invented tool name must NOT be executed.
	content := `{"name":"search_code","arguments":"{\"query\":\"foo\"}"}`
	got, rec := parseToolCalls(content, nil, tools)
	if rec {
		t.Fatal("invented tool name should not be recovered")
	}
	if len(got) != 0 {
		t.Fatalf("expected no calls for unknown tool, got %+v", got)
	}
}

func TestExecTool_EditAndGlob(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	f := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(f, []byte("hello world hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := execTool(ctx, "edit", `{"path":"note.txt","old":"hello","new":"goodbye"}`, dir, 0)
	if out != "edited" {
		t.Fatalf("edit out: %q", out)
	}
	b, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "goodbye world goodbye" {
		t.Fatalf("unexpected file content: %q", string(b))
	}
	g := execTool(ctx, "glob", `{"pattern":"*.txt"}`, dir, 0)
	if g != f {
		t.Fatalf("glob expected %s, got %s", f, g)
	}
}
