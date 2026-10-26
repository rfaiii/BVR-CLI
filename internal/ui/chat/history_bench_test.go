package chat

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richavery/bvr-cli/internal/db"
	"github.com/richavery/bvr-cli/internal/message"
	"github.com/richavery/bvr-cli/internal/ui/styles"
)

// The seed mirrors a real 2-hour agent session read out of .bvr/bvr.db
// (session 25bc4ec0): 595 rows, ~3 MB of parts JSON, and 85 % of those
// bytes sitting in tool_result blobs. The shape matters more than the size;
// it is the tool rows that dominate a history load.
const (
	seedAssistantBytes = 1580
	seedToolBytes      = 8886
	// Rows whose blob is far above the mean — the real session peaked at
	// 439 KB in a single tool result.
	seedHugeToolRows = 4
	seedHugeBytes    = 400_000
	// Nested agent-tool sessions, one per agentic_fetch/agent tool call.
	seedNestedSessions  = 12
	seedNestedTurns     = 15
	seedNestedToolBytes = 6000
)

var seedToolNames = []string{"bash", "edit", "view", "agentic_fetch"}

type seedShape struct {
	turns     int
	nested    int
	userEvery int
}

// seedHistory returns a message service backed by a throwaway SQLite file
// holding one long session plus its nested agent-tool sessions, and the ids
// of the nested sessions.
//
// BVR_BENCH_DATADIR and BVR_BENCH_SESSION swap in a real session instead,
// the same way BenchmarkResizeSession does.
func seedHistory(b *testing.B) (message.Service, string, []string, *db.Queries) {
	b.Helper()
	ctx := context.Background()

	if sessionID := os.Getenv("BVR_BENCH_SESSION"); sessionID != "" {
		dataDir := os.Getenv("BVR_BENCH_DATADIR")
		if dataDir == "" {
			dataDir = ".bvr"
		}
		// Fail rather than migrate a fresh database: a relative dataDir is
		// resolved against the package directory, so a typo silently produces
		// an empty session and microsecond numbers that look like a win.
		if _, err := os.Stat(filepath.Join(dataDir, "bvr.db")); err != nil {
			b.Fatalf("real-session bench: %v (run from the repo root or set an absolute BVR_BENCH_DATADIR)", err)
		}
		conn, err := db.Connect(ctx, dataDir, db.WithDataDirLock(false))
		if err != nil {
			b.Fatalf("connect: %v", err)
		}
		q := db.New(conn)
		svc := message.NewService(q)
		nested := realNestedSessions(ctx, b, svc, sessionID)
		if len(nested) == 0 {
			if page, err := svc.ListLatest(ctx, sessionID, 1); err != nil || len(page.Messages) == 0 {
				b.Fatalf("real-session bench: session %s has no history (%v)", sessionID, err)
			}
		}
		return svc, sessionID, nested, q
	}

	dataDir := filepath.Join(b.TempDir(), ".bvr")
	conn, err := db.Connect(ctx, dataDir, db.WithDataDirLock(false))
	if err != nil {
		b.Fatalf("connect: %v", err)
	}
	q := db.New(conn)
	svc := message.NewService(q)

	const main = "bench-session"
	if _, err := q.CreateSession(ctx, db.CreateSessionParams{ID: main, Title: "bench"}); err != nil {
		b.Fatalf("create session: %v", err)
	}
	seedSession(b, svc, main, seedShape{turns: 293, userEvery: 50})

	nested := make([]string, 0, seedNestedSessions)
	for i := range seedNestedSessions {
		id := fmt.Sprintf("%s$$call-%02d", main, i)
		if _, err := q.CreateSession(ctx, db.CreateSessionParams{
			ID:              id,
			ParentSessionID: sql.NullString{String: main, Valid: true},
			Title:           "bench nested",
		}); err != nil {
			b.Fatalf("create nested session: %v", err)
		}
		seedSession(b, svc, id, seedShape{turns: seedNestedTurns, userEvery: seedNestedTurns})
		nested = append(nested, id)
	}
	return svc, main, nested, q
}

// seedSession writes user turns interleaved with assistant/tool pairs, so
// that tool result N references the tool call in assistant message N.
func seedSession(b *testing.B, svc message.Service, sessionID string, shape seedShape) {
	b.Helper()
	ctx := context.Background()
	isNested := strings.Contains(sessionID, "$$")

	huge := map[int]bool{}
	for i := range seedHugeToolRows {
		huge[i*shape.turns/seedHugeToolRows] = true
	}

	toolCallID := func(n int) string { return fmt.Sprintf("%s-call-%d", sessionID, n) }

	for turn := range shape.turns {
		if shape.userEvery > 0 && turn%shape.userEvery == 0 {
			if _, err := svc.Create(ctx, sessionID, message.CreateMessageParams{
				Role:  message.User,
				Parts: []message.ContentPart{message.TextContent{Text: strings.Repeat("prompt ", 30)}},
			}); err != nil {
				b.Fatalf("create user message: %v", err)
			}
		}

		name := seedToolNames[turn%len(seedToolNames)]
		if _, err := svc.Create(ctx, sessionID, message.CreateMessageParams{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.ReasoningContent{Thinking: strings.Repeat("thinking ", 30)},
				message.TextContent{Text: strings.Repeat("response ", 20)},
				message.ToolCall{
					ID:       toolCallID(turn),
					Name:     name,
					Input:    `{"command":"go test ./..."}`,
					Finished: true,
				},
			},
			Model: "claude",
		}); err != nil {
			b.Fatalf("create assistant message: %v", err)
		}

		size := seedToolBytes
		if isNested {
			size = seedNestedToolBytes
		}
		if huge[turn] && !isNested {
			size = seedHugeBytes
		}
		if _, err := svc.Create(ctx, sessionID, message.CreateMessageParams{
			Role: message.Tool,
			Parts: []message.ContentPart{
				message.ToolResult{
					ToolCallID: toolCallID(turn),
					Name:       name,
					Content:    strings.Repeat("x", size),
				},
			},
		}); err != nil {
			b.Fatalf("create tool message: %v", err)
		}
	}
}

func realNestedSessions(ctx context.Context, b *testing.B, svc message.Service, sessionID string) []string {
	b.Helper()
	msgs, err := svc.List(ctx, sessionID)
	if err != nil {
		b.Fatalf("list: %v", err)
	}
	var nested []string
	for _, m := range msgs {
		for _, part := range m.Parts {
			tc, ok := part.(message.ToolCall)
			if !ok || (tc.Name != "agent" && tc.Name != "agentic_fetch") {
				continue
			}
			nested = append(nested, sessionID+"$$"+tc.ID)
		}
	}
	return nested
}

func asPointers(msgs []message.Message) []*message.Message {
	ptrs := make([]*message.Message, len(msgs))
	for i := range msgs {
		ptrs[i] = &msgs[i]
	}
	return ptrs
}

// BenchmarkSessionOpen measures what it costs to reopen a long session, split
// into the stages a session switch really runs: the unbounded DB read and
// parts decode, the item build, the per-nested-agent List calls that
// loadNestedToolCalls makes one at a time, and the prompt-history refetch.
//
//	go test ./internal/ui/chat/ -run x -bench BenchmarkSessionOpen -benchtime 20x
func BenchmarkSessionOpen(b *testing.B) {
	svc, sessionID, nested, _ := seedHistory(b)
	ctx := context.Background()
	sty := styles.CharmtonePantera()

	b.Run("list", func(b *testing.B) {
		for b.Loop() {
			msgs, err := svc.List(ctx, sessionID)
			if err != nil {
				b.Fatalf("list: %v", err)
			}
			b.SetBytes(int64(len(msgs)))
		}
	})

	b.Run("build-items", func(b *testing.B) {
		msgs, err := svc.List(ctx, sessionID)
		if err != nil {
			b.Fatalf("list: %v", err)
		}
		ptrs := asPointers(msgs)
		for b.Loop() {
			toolResults := BuildToolResultMap(ptrs)
			var items []MessageItem
			for _, m := range ptrs {
				items = append(items, ExtractMessageItems(&sty, m, toolResults)...)
			}
			b.SetBytes(int64(len(items)))
		}
	})

	b.Run("nested-fetch", func(b *testing.B) {
		for b.Loop() {
			var total int
			for _, id := range nested {
				msgs, err := svc.List(ctx, id)
				if err != nil {
					b.Fatalf("list nested: %v", err)
				}
				total += len(msgs)
			}
			b.SetBytes(int64(total))
		}
	})

	// What a session open costs now: one page of history rather than the whole
	// session, plus the items the chat builds from it.
	b.Run("paged-open", func(b *testing.B) {
		for b.Loop() {
			page, err := svc.ListLatest(ctx, sessionID, message.DefaultPageLimit)
			if err != nil {
				b.Fatalf("list latest: %v", err)
			}
			ptrs := asPointers(page.Messages)
			toolResults := BuildToolResultMap(ptrs)
			var items []MessageItem
			for _, m := range ptrs {
				items = append(items, ExtractMessageItems(&sty, m, toolResults)...)
			}
			b.SetBytes(int64(len(items)))
		}
	})

	b.Run("prompt-history", func(b *testing.B) {
		for b.Loop() {
			msgs, err := svc.ListUserMessages(ctx, sessionID)
			if err != nil {
				b.Fatalf("list user messages: %v", err)
			}
			b.SetBytes(int64(len(msgs)))
		}
	})
}
