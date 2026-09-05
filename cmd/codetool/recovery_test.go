package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOllToolCallFunc_UnmarshalsObjectArguments(t *testing.T) {
	var call ollToolCall
	if err := json.Unmarshal([]byte(`{"function":{"name":"view","arguments":{"file_path":"note.txt"}}}`), &call); err != nil {
		t.Fatal(err)
	}
	if call.Function.Arguments != `{"file_path":"note.txt"}` {
		t.Fatalf("arguments = %q", call.Function.Arguments)
	}
}

func TestParseToolCalls_RecoversFenced(t *testing.T) {
	tools := toolNames(flatTools())
	content := "```json\n{\n  \"name\": \"bash\",\n  \"arguments\": {\"command\": \"echo hi\"}\n}\n```"
	got, rec := parseToolCalls(content, nil, tools)
	if !rec || len(got) != 1 || got[0].Function.Name != "bash" {
		t.Fatalf("want recovered bash, got recovered=%v calls=%+v", rec, got)
	}
}

func TestParseToolCalls_RecoversAliasedName(t *testing.T) {
	tools := toolNames(flatTools())
	// 3b models sometimes invent names like "Run" -> bash
	content := `{"name":"Run","arguments":{"command":"ls -la"}}`
	got, rec := parseToolCalls(content, nil, tools)
	if !rec {
		t.Fatal("expected recovery")
	}
	if got[0].Function.Name != "bash" {
		t.Fatalf("expected alias Run->bash, got %s", got[0].Function.Name)
	}
}

func TestParseToolCalls_IgnoresInventedTool(t *testing.T) {
	tools := toolNames(flatTools())
	content := `{"name":"search_code","arguments":{"query":"foo"}}`
	got, rec := parseToolCalls(content, nil, tools)
	if rec || len(got) != 0 {
		t.Fatalf("invented tool must not execute: rec=%v got=%+v", rec, got)
	}
}

func TestExecTool_BashArrayCommand(t *testing.T) {
	// 3b emits commands as arrays; must be joined and run through the shell.
	out := execTool(context.Background(), "bash", `{"command":["echo","hello world"]}`, t.TempDir(), 0)
	if strings.TrimSpace(out) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", strings.TrimSpace(out))
	}
}

func TestExecTool_WriteAndView(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	r := execTool(ctx, "write", `{"path":"note.txt","content":"buy milk"}`, dir, 0)
	if r != "written" {
		t.Fatalf("write: %q", r)
	}
	v := execTool(ctx, "view", `{"path":"note.txt"}`, dir, 0)
	if v != "buy milk" {
		t.Fatalf("view: %q", v)
	}
}

func TestExecTool_ViewFilePathAlias(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := execTool(context.Background(), "view", `{"file_path":"note.txt"}`, dir, 0)
	if got != "hello" {
		t.Fatalf("view with file_path = %q", got)
	}
}

func TestParseToolCalls_RecoversTruncatedViewPath(t *testing.T) {
	content := `{"name":"view","arguments":{"file_path":"note.txt"`
	got, recovered := parseToolCalls(content, nil, toolNames(flatTools()))
	if !recovered || len(got) != 1 {
		t.Fatalf("recovered=%v calls=%+v", recovered, got)
	}
	if got[0].Function.Name != "view" || got[0].Function.Arguments != `{"path":"note.txt"}` {
		t.Fatalf("recovered call=%+v", got[0])
	}
}

func TestRecGlob_DoubleStar(t *testing.T) {
	root := t.TempDir()
	must := func(p string) {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, p)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, p), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must("cmd/a/main.go")
	must("cmd/b/x/main.go")
	must("cmd/a/util_test.go")
	must("cmd/keep.txt") // non-.go should not match

	got, err := recGlob(root, "cmd/**/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 .go files under cmd/, got %d: %v", len(got), got)
	}
}

func TestParseToolCalls_RecoversBareRegexValue(t *testing.T) {
	// qwen dumped {"name":"grep","arguments":{"path":"","pattern:"^.*$"}} —
	// the regex value is unquoted so braces never balance. matchBrace fails,
	// but the regex fallback must still recover the grep call.
	tools := toolNames(flatTools())
	content := `{"name":"grep","arguments":{"path":"","pattern:"^.*$"}}`
	got, rec := parseToolCalls(content, nil, tools)
	if !rec {
		t.Fatal("expected recovery=true for bare-regex-value dump")
	}
	if len(got) != 1 || got[0].Function.Name != "grep" {
		t.Fatalf("want recovered grep call, got %+v", got)
	}
}

func TestParseArgsTolerant_BareRegexValue(t *testing.T) {
	// The repaired arguments must let grep actually parse and search.
	got := parseArgsTolerant(`{"path":"","pattern:"^.*$"}`)
	if got == nil {
		t.Fatal("expected tolerant parse to succeed")
	}
	if got["pattern"] != "^.*$" {
		t.Fatalf("expected pattern=^.*$, got %#v", got["pattern"])
	}
}

func TestExecTool_GrepBareValueDefaultsToCwd(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\n// needle here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Malformed case: empty path defaults to cwd, bare regex value repaired.
	out := execTool(context.Background(), "grep", `{"path":"","pattern:"needle"}`, dir, 0)
	if !strings.Contains(out, "a.go") {
		t.Fatalf("grep should find needle in a.go, got %q", out)
	}
}

func TestNormalizeToolName(t *testing.T) {
	for in, want := range map[string]string{"Run": "bash", "LIST": "ls", "Read": "view", "Find": "glob", "bash": "bash"} {
		if got := normalizeToolName(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}
