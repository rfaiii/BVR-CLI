package chat

import (
	"context"
	"testing"

	"github.com/richavery/bvr-cli/internal/db"

	"github.com/richavery/bvr-cli/internal/message"
)

// BenchmarkListCostSplit separates the two things that make a history read
// expensive: the SQLite round-trip itself (locking, page reads, row copy)
// and the parts JSON decode. It runs the same read through the raw sqlc
// query and through the message service, on sessions of different sizes.
//
//	go test ./internal/ui/chat/ -run x -bench BenchmarkListCostSplit -benchtime 30x
func BenchmarkListCostSplit(b *testing.B) {
	svc, sessionID, nested, q := seedHistory(b)
	ctx := context.Background()

	blobs := totalPartsBytes(ctx, b, q, sessionID)
	b.Logf("session %s: %d bytes of parts", sessionID, blobs)

	// A session with a single row isolates the fixed per-query cost.
	const tiny = "bench-tiny"
	if _, err := q.CreateSession(ctx, db.CreateSessionParams{ID: tiny, Title: "tiny"}); err != nil {
		b.Fatalf("create tiny session: %v", err)
	}
	if _, err := svc.Create(ctx, tiny, message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: "hello"}},
	}); err != nil {
		b.Fatalf("create tiny message: %v", err)
	}

	b.Run("rows-only/main", func(b *testing.B) {
		b.SetBytes(int64(blobs))
		for b.Loop() {
			rows, err := q.ListMessagesBySession(ctx, sessionID)
			if err != nil {
				b.Fatalf("list: %v", err)
			}
			if n := len(rows); n == 0 {
				b.Fatal("no rows")
			}
		}
	})

	b.Run("service/main", func(b *testing.B) {
		b.SetBytes(int64(blobs))
		for b.Loop() {
			if _, err := svc.List(ctx, sessionID); err != nil {
				b.Fatalf("list: %v", err)
			}
		}
	})

	b.Run("rows-only/tiny", func(b *testing.B) {
		for b.Loop() {
			if _, err := q.ListMessagesBySession(ctx, tiny); err != nil {
				b.Fatalf("list: %v", err)
			}
		}
	})

	b.Run("service/tiny", func(b *testing.B) {
		for b.Loop() {
			if _, err := svc.List(ctx, tiny); err != nil {
				b.Fatalf("list: %v", err)
			}
		}
	})

	b.Run("rows-only/nested-one", func(b *testing.B) {
		b.SetBytes(int64(totalPartsBytes(ctx, b, q, nested[0])))
		for b.Loop() {
			if _, err := q.ListMessagesBySession(ctx, nested[0]); err != nil {
				b.Fatalf("list: %v", err)
			}
		}
	})

	b.Run("service/nested-one", func(b *testing.B) {
		for b.Loop() {
			if _, err := svc.List(ctx, nested[0]); err != nil {
				b.Fatalf("list: %v", err)
			}
		}
	})
}

// totalPartsBytes sums the parts column so benchmarks can report MB/s.
func totalPartsBytes(ctx context.Context, b *testing.B, q *db.Queries, sessionID string) int64 {
	b.Helper()
	rows, err := q.ListMessagesBySession(ctx, sessionID)
	if err != nil {
		b.Fatalf("list rows: %v", err)
	}
	var total int64
	for _, row := range rows {
		total += int64(len(row.Parts))
	}
	return total
}
