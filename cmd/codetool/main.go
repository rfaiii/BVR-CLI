// codetool is a minimal, fantasy-free CLI coding agent for local Ollama models.
//
// It talks to Ollama's native /api/chat endpoint with a small, flat tool set
// (bash, view, edit, glob, grep, ls) and includes lenient recovery for models
// that dump a tool call as text instead of emitting tool_calls. Cline/Hermes
// skills are invoked by directly exec'ing the underlying CLI.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Ollama native chat request/response types (minimal subset).
type ollamaRequest struct {
	Model    string         `json:"model"`
	Stream   bool           `json:"stream"`
	Messages []ollMsg       `json:"messages"`
	Tools    []ollTool      `json:"tools,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
}

type ollMsg struct {
	Role      string        `json:"role"` // system|user|assistant|tool
	Content   string        `json:"content,omitempty"`
	Name      string        `json:"name,omitempty"`
	ToolCalls []ollToolCall `json:"tool_calls,omitempty"`
}

type ollTool struct {
	Type     string  `json:"type"` // "function"
	Function ollFunc `json:"function"`
}
type ollFunc struct {
	Desc       string         `json:"description"`
	Name       string         `json:"name"`
	Parameters map[string]any `json:"parameters"`
}
type ollToolCall struct {
	ID       string          `json:"id,omitempty"`
	Type     string          `json:"type,omitempty"`
	Function ollToolCallFunc `json:"function"`
}
type ollToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // Ollama sends a JSON string
}

// UnmarshalJSON accepts both Ollama's native object form and the string form
// emitted by OpenAI-compatible adapters. Small models and older Ollama builds
// commonly use either shape, so rejecting one here needlessly loses a valid
// tool call before the recovery logic can run.
func (f *ollToolCallFunc) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Name = raw.Name
	if len(raw.Arguments) == 0 || string(raw.Arguments) == "null" {
		f.Arguments = "{}"
		return nil
	}
	if raw.Arguments[0] == '"' {
		return json.Unmarshal(raw.Arguments, &f.Arguments)
	}
	f.Arguments = string(raw.Arguments)
	return nil
}

type ollamaResponse struct {
	Model   string `json:"model"`
	Message ollMsg `json:"message"`
	Done    bool   `json:"done"`
}

// streamContentBuffer keeps probable tool-call dumps off the terminal until
// the complete response is available. Ordinary prose is still printed as soon
// as it arrives; only content beginning with a JSON object or markdown fence
// is held back for classification.
type streamContentBuffer struct {
	pending   strings.Builder
	candidate bool
}

func (b *streamContentBuffer) Write(content string) {
	if content == "" {
		return
	}
	if b.candidate {
		b.pending.WriteString(content)
		return
	}
	trimmed := strings.TrimLeft(b.pending.String()+content, " \t\r\n")
	if trimmed == "{" || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "```") {
		b.candidate = true
		b.pending.WriteString(content)
		return
	}
	fmt.Print(content)
}

func (b *streamContentBuffer) Finish(isToolCall bool) {
	if b.pending.Len() == 0 {
		return
	}
	if !isToolCall {
		fmt.Print(b.pending.String())
	}
}

func main() {
	var (
		model   = flag.String("model", "qwen2.5-coder:7b", "Ollama model to use")
		baseURL = flag.String("url", defaultOllamaURL(), "Ollama base URL")
		cwd     = flag.String("cwd", ".", "working directory")
		skill   = flag.String("skill", "", "delegate to a skill CLI instead of a local model (cline|hermes)")
		prompt  = flag.String("p", "", "one-shot prompt (otherwise reads stdin)")
		stream  = flag.Bool("stream", true, "stream model tokens as they arrive")
		timeout = flag.Duration("timeout", 30*time.Second, "default bash tool timeout")
		budget  = flag.Duration("budget", 120*time.Second, "overall agent time budget")
	)
	flag.Parse()

	if *skill != "" {
		runSkill(*skill, strings.Join(flag.Args(), " "), *cwd)
		return
	}

	if err := runAgent(*baseURL, *model, *cwd, *timeout, *budget, *stream, *prompt, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func defaultOllamaURL() string {
	if v := strings.TrimSpace(os.Getenv("OLLAMA_HOST")); v != "" {
		if strings.Contains(v, "://") {
			return strings.TrimRight(v, "/") + "/api/chat"
		}
		return "http://" + strings.TrimRight(v, "/") + "/api/chat"
	}
	return "http://127.0.0.1:11434/api/chat"
}

func runAgent(baseURL, modelName, cwd string, timeout, budget time.Duration, stream bool, prompt string, extra []string) error {
	if flag.NArg() == 0 && prompt == "" {
		return errors.New("no prompt given (pass -p, or args after flags, or pipe via stdin)")
	}
	userPrompt := prompt
	if userPrompt == "" {
		userPrompt = strings.Join(extra, " ")
		if userPrompt == "" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			userPrompt = strings.TrimSpace(string(data))
		}
	}
	if userPrompt == "" {
		return errors.New("empty prompt")
	}

	tools := flatTools()
	msgs := []ollMsg{
		{Role: "system", Content: "You are a concise CLI coding agent. Make edits and run commands to complete the task. Call exactly one tool when needed, then use the returned tool result as authoritative. Do not claim a file is missing when the tool result contains its contents. Otherwise give a final answer in one or two sentences."},
		{Role: "user", Content: userPrompt},
	}

	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	for i := 0; i < 4; i++ {
		turnCtx, turnCancel := turnContext(ctx, 45*time.Second)
		if os.Getenv("CODETOOL_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[ollama] turn %d/%d model=%s stream=%t\n", i+1, 4, modelName, stream)
		}
		var resp *ollamaResponse
		var err error
		if stream {
			resp, err = chatStream(turnCtx, baseURL, ollamaRequest{
				Model: modelName, Stream: true, Messages: msgs, Tools: tools,
				Options: map[string]any{"num_predict": 256},
			})
		} else {
			resp, err = chat(turnCtx, baseURL, ollamaRequest{
				Model: modelName, Stream: false, Messages: msgs, Tools: tools,
				Options: map[string]any{"num_predict": 256},
			})
		}
		turnCancel()
		if err != nil {
			return fmt.Errorf("ollama: %w", err)
		}
		asst := resp.Message
		if !stream {
			fmt.Print(asst.Content)
		}

		calls, recovered := parseToolCalls(asst.Content, asst.ToolCalls, toolNames(tools))
		if recovered {
			fmt.Fprintln(os.Stderr, "[tool-call recovered from model text]")
		}
		if len(calls) == 0 {
			if asst.Content != "" {
				fmt.Println()
			}
			return nil
		}
		fmt.Println()
		msgs = append(msgs, asst)
		for _, c := range calls {
			out := execTool(ctx, c.Function.Name, c.Function.Arguments, cwd, timeout)
			msgs = append(msgs, ollMsg{
				Role: "tool", Name: c.Function.Name, Content: fmt.Sprintf("Tool %s result:\n%s", c.Function.Name, out),
			})
			fmt.Fprintf(os.Stderr, "[%s] %s\n", c.Function.Name, truncate(out, 500))
		}
	}
	return nil
}

func turnContext(parent context.Context, max time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := parent.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < max {
			max = remaining
		}
	}
	return context.WithTimeout(parent, max)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}

func chat(ctx context.Context, baseURL string, req ollamaRequest) (*ollamaResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := httpNewRequest(ctx, "POST", baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if os.Getenv("CODETOOL_DEBUG") != "" {
		fmt.Fprintln(os.Stderr, "[debug] raw ollama response:")
		fmt.Fprintln(os.Stderr, string(raw))
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %s: %s", resp.Status, truncate(string(raw), 1000))
	}
	var r ollamaResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// chatStream issues a streaming /api/chat request. Each NDJSON line carries a
// content delta; content is printed to stdout as it arrives and the deltas are
// accumulated (along with any tool_calls) into a single response for the agent
// loop.
func chatStream(ctx context.Context, baseURL string, req ollamaRequest) (*ollamaResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := httpNewRequest(ctx, "POST", baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %s: %s", resp.Status, truncate(string(raw), 1000))
	}
	dec := json.NewDecoder(resp.Body)
	var full ollamaResponse
	seen := false
	content := &streamContentBuffer{}
	for {
		var part ollamaResponse
		if err := dec.Decode(&part); err != nil {
			if errors.Is(err, io.EOF) && seen && full.Done {
				break
			}
			return nil, fmt.Errorf("ollama stream ended before done=true: %w", err)
		}
		seen = true
		full.Model = part.Model
		full.Message.Role = part.Message.Role
		full.Message.Content += part.Message.Content
		if len(part.Message.ToolCalls) > 0 {
			full.Message.ToolCalls = part.Message.ToolCalls
		}
		if part.Message.Content != "" {
			content.Write(part.Message.Content)
		}
		full.Done = part.Done
		if part.Done {
			break
		}
	}
	if !seen || !full.Done {
		return nil, errors.New("ollama stream returned no completed response")
	}
	calls, _ := parseToolCalls(full.Message.Content, full.Message.ToolCalls, toolNames(flatTools()))
	content.Finish(len(calls) > 0)
	return &full, nil
}

// parseToolCalls returns the tool calls parsed by the provider, OR — as a
// lenient recovery for small models that dump tool calls as text instead of
// emitting a tool_calls array — one or more recognized tool calls extracted
// from `content`. `recovered` is true when the latter path was used.
func parseToolCalls(content string, reported []ollToolCall, known map[string]bool) ([]ollToolCall, bool) {
	if len(reported) > 0 {
		return normalizeToolCalls(reported), false
	}
	content = strings.TrimSpace(content)
	if !strings.Contains(content, "{") {
		return nil, false
	}
	calls := extractToolCallsFromText(content, known)
	return calls, len(calls) > 0
}

// extractToolCallsFromText scans free text for every brace-balanced JSON object
// and recovers those that look exactly like a tool call ({name, arguments}) for
// a known tool. This is the lenient recovery that lets small local models that
// "text-dump" tool calls (including inside ```json fences or multiple in one
// turn) still drive the agent loop. `arguments` may be a JSON string
// (OpenAI/Ollama shape) or a bare object/array (some models inline it); both
// are normalized to a JSON-string form for execTool.
func extractToolCallsFromText(content string, known map[string]bool) []ollToolCall {
	content = stripJSONFences(content)
	var out []ollToolCall
	for {
		content = strings.TrimLeft(content, " \t\r\n")
		idx := strings.IndexByte(content, '{')
		if idx < 0 {
			return out
		}
		content = content[idx:]
		end, ok := matchBrace(content)
		if !ok {
			// Malformed JSON: braces never balance (e.g. an unquoted regex
			// value like "pattern:"^.*$" swallows the closing braces). Don't
			// give up — the regex fallback in decodeToolCall can still pull
			// out a recognized {name, arguments} pair from the raw text.
			if tc := decodeToolCall(content, known); tc != nil {
				out = append(out, *tc)
			}
			return out
		}
		if tc := decodeToolCall(content[:end], known); tc != nil {
			out = append(out, *tc)
		}
		content = content[end:]
	}
}

// nameRe / argRe extract a tool call from malformed JSON that small models
// produce (e.g. bare identifiers instead of quoted strings). nameRe is also
// tolerant of the bare-word form qwen3 emits ("name": glob).
var (
	nameRe    = regexp.MustCompile(`"name"\s*:\s*"?([A-Za-z_][A-Za-z0-9_.\-/]*)"?`)
	argObjRe  = regexp.MustCompile(`"arguments"\s*:\s*(\{.*\})`)
	pathArgRe = regexp.MustCompile(`"(?:file_path|path)"\s*:\s*"([^"\r\n]*)`)
)

// decodeToolCall parses a single {...} object as a {name, arguments} tool call.
// It first tries strict JSON; if that fails (common for small models that emit
// bare identifiers), it falls back to regex extraction of name + arguments.
// Returns nil unless the object has a recognized tool name and an arguments
// field, which keeps prose JSON examples from being mis-executed.
func decodeToolCall(obj string, known map[string]bool) *ollToolCall {
	// Fast path: well-formed JSON.
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(obj), &m); err == nil {
		if _, ok := m["name"]; ok && m["arguments"] != nil {
			return decodeFromMap(m, known)
		}
	}
	// Lenient fallback for malformed/bare-word JSON.
	if mm := nameRe.FindStringSubmatch(obj); mm != nil {
		name := normalizeToolName(mm[1])
		if !known[name] {
			return nil
		}
		argStr := "{}"
		if am := argObjRe.FindStringSubmatch(obj); am != nil {
			argStr = strings.TrimSpace(am[1])
		} else if pm := pathArgRe.FindStringSubmatch(obj); pm != nil {
			// Recover a truncated view call while the model is still producing
			// its JSON. This is intentionally limited to a path argument; never
			// execute a partially recovered shell command or file mutation.
			argBytes, _ := json.Marshal(map[string]string{"path": pm[1]})
			argStr = string(argBytes)
		}
		return &ollToolCall{Function: ollToolCallFunc{Name: name, Arguments: argStr}}
	}
	return nil
}

func decodeFromMap(m map[string]json.RawMessage, known map[string]bool) *ollToolCall {
	var rawName string
	if err := json.Unmarshal(m["name"], &rawName); err != nil {
		return nil
	}
	name := normalizeToolName(rawName)
	if !known[name] {
		return nil
	}
	args := m["arguments"]
	argStr := ""
	if len(args) > 0 && args[0] == '"' {
		_ = json.Unmarshal(args, &argStr)
	} else {
		argStr = strings.TrimSpace(string(args))
	}
	if argStr == "" {
		argStr = "{}"
	}
	return &ollToolCall{Function: ollToolCallFunc{Name: name, Arguments: argStr}}
}

// matchBrace finds the index just past the closing '}' matching the opening '{'
// at s[0], respecting string literals and backslash escapes.
func matchBrace(s string) (int, bool) {
	depth := 0
	inStr := false
	esc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1, true
			}
		}
	}
	return 0, false
}

// normalizeToolCalls keeps only well-formed name+arguments pairs.
func normalizeToolCalls(in []ollToolCall) []ollToolCall {
	out := make([]ollToolCall, 0, len(in))
	for _, t := range in {
		if t.Function.Name == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

func toolNames(tools []ollTool) map[string]bool {
	m := make(map[string]bool, len(tools))
	for _, t := range tools {
		m[t.Function.Name] = true
	}
	return m
}

// toolAliases maps common small-model inventions to real tool names so that
// weak models that rename tools (e.g. call bash "Run") still execute.
var toolAliases = map[string]string{
	"run":        "bash",
	"execute":    "bash",
	"shell":      "bash",
	"terminal":   "bash",
	"list_files": "ls",
	"list":       "ls",
	"ls_dir":     "ls",
	"cat":        "view",
	"read":       "view",
	"read_file":  "view",
	"write":      "write",
	"create":     "write",
	"find":       "glob",
	"grep":       "grep",
	"search":     "grep",
}

func normalizeToolName(name string) string {
	if n, ok := toolAliases[strings.ToLower(name)]; ok {
		return n
	}
	return name
}

// stripJSONFences removes a surrounding markdown code fence (with optional
// "json" language tag) so a fenced tool-call dump can be recovered.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx > 0 {
			s = strings.TrimSpace(s[idx+1:])
		}
		if strings.HasSuffix(s, "```") {
			s = strings.TrimSpace(s[:len(s)-len("```")])
		}
	} else if strings.HasPrefix(s, "```json") {
		if idx := strings.Index(s, "\n"); idx > 0 {
			s = strings.TrimSpace(s[idx+1:])
		}
		if strings.HasSuffix(s, "```") {
			s = strings.TrimSpace(s[:len(s)-len("```")])
		}
	}
	return s
}

// globRegex compiles a glob pattern (with ** support) into a regex matching
// path-relative file paths. ** matches any number of path segments.
func globRegex(pat string) (*regexp.Regexp, error) {
	b := &strings.Builder{}
	b.WriteByte('^')
	i := 0
	for i < len(pat) {
		switch {
		case pat[i] == '*' && i+1 < len(pat) && pat[i+1] == '*':
			if i+2 < len(pat) && pat[i+2] == '/' {
				b.WriteString("(?:.*/)?")
				i += 3
			} else {
				b.WriteString(".*")
				i += 2
			}
		case pat[i] == '*':
			b.WriteString("[^/]*")
			i++
		case pat[i] == '?':
			b.WriteString("[^/]")
			i++
		default:
			b.WriteString(regexp.QuoteMeta(string(pat[i])))
			i++
		}
	}
	b.WriteByte('$')
	return regexp.Compile(b.String())
}

// recGlob supports ** (matches any number of path segments) on top of standard
// single-star/single-segment globs, matching path-relative file paths under root.
func recGlob(root, pattern string) ([]string, error) {
	pattern = strings.TrimPrefix(strings.TrimPrefix(pattern, "./"), "/")
	if pattern == "" || pattern == "." {
		return nil, nil
	}
	re, err := globRegex(pattern)
	if err != nil {
		return nil, err
	}
	if os.Getenv("CODETOOL_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[debug recGlob] regex=%s\n", re.String())
	}
	var matches []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Never skip the root itself: a "." or hidden root would otherwise
			// abort the walk via SkipDir before descending at all.
			if path != root && isExcludedDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return nil
		}
		if re.MatchString(filepath.ToSlash(rel)) {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}

// isExcludedDir reports directories that should never be traversed by local
// tools (VCS internals, package caches, hidden build dirs, etc.).
func isExcludedDir(name string) bool {
	switch name {
	case ".git", ".bvr", "node_modules", "vendor", ".venv", ".cache":
		return true
	}
	return strings.HasPrefix(name, ".")
}

// extractJSONObject does a best-effort extraction of the first {...} object.
func extractJSONObject(s string) map[string]any {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return nil
	}
	// lenient: try to balance braces
	depth := 0
	var end int
	for i, r := range s {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
		if end != 0 {
			break
		}
	}
	if end == 0 {
		return nil
	}
	var m map[string]any
	_ = json.Unmarshal([]byte(s[:end]), &m)
	return m
}

func flatTools() []ollTool {
	obj := func(props map[string]any, required ...string) map[string]any {
		return map[string]any{"type": "object", "properties": props, "required": required}
	}
	strProp := func(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }
	intProp := func(desc string) map[string]any { return map[string]any{"type": "integer", "description": desc} }

	return []ollTool{
		{Type: "function", Function: ollFunc{Name: "bash", Desc: "Run a shell command in the cwd.", Parameters: obj(map[string]any{
			"command": strProp("The shell command to run."),
			"timeout": intProp("Max seconds to run."),
		}, "command")}},
		{Type: "function", Function: ollFunc{Name: "view", Desc: "Read a file's full contents.", Parameters: obj(map[string]any{
			"path":      strProp("File to read."),
			"file_path": strProp("Alias for path; file to read."),
		}, "path")}},
		{Type: "function", Function: ollFunc{Name: "edit", Desc: "Replace old with new in a file.", Parameters: obj(map[string]any{
			"path":  strProp("File to edit."),
			"old":   strProp("Exact text to replace."),
			"new":   strProp("Replacement text."),
			"count": intProp("Maximum replacements (optional)."),
		}, "path", "old", "new")}},
		{Type: "function", Function: ollFunc{Name: "glob", Desc: "Match files by glob pattern.", Parameters: obj(map[string]any{
			"pattern": strProp("Glob pattern, e.g. **/*.go."),
		}, "pattern")}},
		{Type: "function", Function: ollFunc{Name: "grep", Desc: "Search file contents with a regexp.", Parameters: obj(map[string]any{
			"pattern": strProp("Regexp pattern."),
			"path":    strProp("Root path to search."),
		}, "pattern")}},
		{Type: "function", Function: ollFunc{Name: "ls", Desc: "List a directory.", Parameters: obj(map[string]any{
			"path": strProp("Directory to list."),
		}, "path")}},
		{Type: "function", Function: ollFunc{Name: "write", Desc: "Create or overwrite a text file.", Parameters: obj(map[string]any{
			"path":    strProp("File path to write."),
			"content": strProp("Full file content to write."),
		}, "path", "content")}},
	}
}

// bareColonRe matches the malformed "key:"value form a small model sometimes
// emits (the colon lands inside the key's quotes), e.g. "pattern:"^.*$". It is
// repaired to the well-formed "key":"value" by inserting a quote after the
// colon. It only matches when the colon is INSIDE the quoted key, so normal
// "key":value output is left untouched.
var bareColonRe = regexp.MustCompile(`"([A-Za-z_][A-Za-z0-9_]*):"`)

// parseArgsTolerant best-effort parses a tool's argument JSON when strict
// json.Unmarshal fails. It first applies a light repair for the bare-value
// colon malformation, then a naive per-key regex as a final fallback. Returns
// nil only when nothing usable can be extracted.
func parseArgsTolerant(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// Light repair: "key:"^.*$" -> "key":"^.*$".
	if fixed := bareColonRe.ReplaceAllString(raw, `"$1":"`); fixed != raw {
		var m map[string]any
		if err := json.Unmarshal([]byte(fixed), &m); err == nil {
			return m
		}
	}
	// Final fallback: pull out "key":"value" pairs with a tolerant regex.
	m := map[string]any{}
	re := regexp.MustCompile(`"([A-Za-z_][A-Za-z0-9_]*)":\s*("(?:[^"\\]|\\.)*"|\[[^\]]*\]|[^,}]+)`)
	for _, mm := range re.FindAllStringSubmatch(raw, -1) {
		key, val := mm[1], strings.TrimSpace(mm[2])
		if strings.HasPrefix(val, `"`) {
			var s string
			if json.Unmarshal([]byte(val), &s) == nil {
				m[key] = s
			} else {
				m[key] = strings.Trim(val, `"`)
			}
		} else {
			m[key] = val
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func execTool(ctx context.Context, name, args, cwd string, timeout time.Duration) string {
	var p map[string]any
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		// Small models mangle quoting (e.g. "pattern:"^.*$" leaves a regex
		// value bare). Try a light repair, then a tolerant fallback before
		// giving up so the tool still runs.
		p = parseArgsTolerant(args)
	}
	if p == nil {
		return "invalid tool arguments"
	}
	normalizeToolArgs(p)
	name = normalizeToolName(name)
	switch name {
	case "bash":
		to := timeout
		if t, ok := p["timeout"].(float64); ok && t > 0 {
			to = time.Duration(t) * time.Second
		}
		if to <= 0 {
			to = 30 * time.Second
		}
		var cmd string
		switch c := p["command"].(type) {
		case string:
			cmd = c
		case []any:
			parts := make([]string, len(c))
			for i, v := range c {
				parts[i] = fmt.Sprint(v)
			}
			cmd = strings.Join(parts, " ")
		default:
			cmd = fmt.Sprint(p["command"])
		}
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			return "bash: empty command"
		}
		cctx, cancel := context.WithTimeout(ctx, to)
		defer cancel()
		out, err := exec.CommandContext(cctx, "sh", "-c", cmd).CombinedOutput()
		return string(out) + errMsg(err)
	case "view":
		b, err := os.ReadFile(fileJoin(cwd, fmt.Sprint(p["path"])))
		if err != nil {
			return fmt.Sprintf("view error: %v", err)
		}
		return string(b)
	case "edit":
		fp := fileJoin(cwd, fmt.Sprint(p["path"]))
		old := fmt.Sprint(p["old"])
		new := fmt.Sprint(p["new"])
		b, err := os.ReadFile(fp)
		if err != nil {
			return fmt.Sprintf("edit error: %v", err)
		}
		limit := 0
		if c, ok := p["count"].(float64); ok && c > 0 {
			limit = int(c)
		}
		if limit == 0 {
			b = bytes.Replace(b, []byte(old), []byte(new), -1)
		} else {
			b = bytes.Replace(b, []byte(old), []byte(new), limit)
		}
		if err := os.WriteFile(fp, b, 0o644); err != nil {
			return fmt.Sprintf("edit write error: %v", err)
		}
		return "edited"
	case "grep":
		re, err := regexp.Compile(fmt.Sprint(p["pattern"]))
		if err != nil {
			return fmt.Sprintf("grep compile error: %v", err)
		}
		var hits []string
		root := fileJoin(cwd, fmt.Sprint(p["path"]))
		if root == "" {
			root = cwd
		}
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if isExcludedDir(info.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			for _, line := range strings.Split(string(b), "\n") {
				if re.MatchString(line) {
					hits = append(hits, path+": "+line)
				}
			}
			return nil
		})
		return strings.Join(hits, "\n")
	case "glob":
		pattern := fmt.Sprint(p["pattern"])
		matches, err := recGlob(cwd, pattern)
		if os.Getenv("CODETOOL_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[debug glob] pattern=%q root=%q matches=%d err=%v", pattern, cwd, len(matches), err)
			for _, m := range matches {
				fmt.Fprintf(os.Stderr, "\n  %s", m)
			}
			fmt.Fprintln(os.Stderr)
		}
		if err != nil {
			return fmt.Sprintf("glob error: %v", err)
		}
		return strings.Join(matches, "\n")
	case "ls":
		entries, err := os.ReadDir(fileJoin(cwd, fmt.Sprint(p["path"])))
		if err != nil {
			return fmt.Sprintf("ls error: %v", err)
		}
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return strings.Join(names, "\n")
	case "write":
		fp := fileJoin(cwd, fmt.Sprint(p["path"]))
		if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
			return fmt.Sprintf("write error: %v", err)
		}
		if err := os.WriteFile(fp, []byte(fmt.Sprint(p["content"])), 0o644); err != nil {
			return fmt.Sprintf("write error: %v", err)
		}
		return "written"
	default:
		return fmt.Sprintf("unknown tool: %s", name)
	}
}

// normalizeToolArgs bridges argument names used by common coding-agent
// prompts to codetool's deliberately small canonical schema. In particular,
// qwen-coder frequently emits file_path even when the advertised field is
// path; executing the call is preferable to echoing it back as failed JSON.
func normalizeToolArgs(args map[string]any) {
	if _, ok := args["path"]; !ok {
		if path, ok := args["file_path"]; ok {
			args["path"] = path
		}
	}
}

func fileJoin(cwd, p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~/"))
		}
	}
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(cwd, p)
}

func errMsg(err error) string {
	if err == nil {
		return ""
	}
	stderr := ""
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		stderr = "\n" + string(ee.Stderr)
	}
	return fmt.Sprintf("\n(exit: %v%s)", err, stderr)
}

// runSkill delegates to an external coding agent CLI (cline/hermes) by name.
func runSkill(name, task, cwd string) {
	bin := name
	args := []string{}
	switch name {
	case "cline":
		args = []string{"--yolo", task}
	case "hermes":
		args = []string{"--yolo", task}
	case "crush":
		args = []string{task}
	default:
		fmt.Fprintf(os.Stderr, "unknown skill %q (use cline|hermes)\n", name)
		os.Exit(2)
	}
	cmd := exec.CommandContext(context.Background(), bin, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "skill error:", err)
		os.Exit(1)
	}
}

// ---- tiny HTTP/transport shims to keep this self-contained and testable ----

var httpClient = defaultHTTPClient()

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Minute}
}

func httpNewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// strconv is used for int->string only in truncation guards above.
var _ = strconv.Itoa
