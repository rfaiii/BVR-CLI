package message

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// walkPages reads a whole session through the paging API, newest page first,
// and returns the messages oldest first.
func walkPages(t *testing.T, svc Service, sessionID string, pageSize int) []Message {
	t.Helper()

	var (
		all    []Message
		cursor int64
	)
	for page := 0; ; page++ {
		require.Less(t, page, 1000, "paging must terminate")

		var (
			p   MessagePage
			err error
		)
		if cursor == 0 {
			p, err = svc.ListLatest(t.Context(), sessionID, pageSize)
		} else {
			p, err = svc.ListOlder(t.Context(), sessionID, cursor, pageSize)
		}
		require.NoError(t, err)

		if cursor != 0 {
			require.NotEmpty(t, p.Messages,
				"ListOlder returned an empty page while HasMore was set (page %d)", page)
		}
		require.LessOrEqual(t, len(p.Messages), pageSize)

		// Pages arrive oldest first and are read newest first, so each new
		// page belongs in front of everything collected so far.
		all = append(p.Messages, all...)
		if p.Cursor == 0 {
			require.Empty(t, p.Messages, "a non-empty page always carries a cursor")
		}
		cursor = p.Cursor

		if !p.HasMore {
			return all
		}
	}
}

// TestListPagesWalkSessionExactly: paging a session to its start must return
// every message exactly once, in the order the rows were written.
//
// This is the property a created_at cursor cannot offer. Messages created in
// the same second tie, and a busy session has roughly one collision per two
// rows, so paging over created_at drops or repeats whichever row the sort
// happens to pick first.
func TestListPagesWalkSessionExactly(t *testing.T) {
	svc, sessionID := newTestService(t)

	const total = 250
	want := make([]string, 0, total)
	for i := range total {
		msg, err := svc.Create(t.Context(), sessionID, CreateMessageParams{
			Role:  User,
			Parts: []ContentPart{TextContent{Text: fmt.Sprintf("prompt %d", i)}},
		})
		require.NoError(t, err)
		want = append(want, msg.ID)
	}

	got := make([]string, 0, total)
	for _, msg := range walkPages(t, svc, sessionID, DefaultPageLimit) {
		got = append(got, msg.ID)
	}
	require.Equal(t, want, got)
}

// TestListPagesSizeAndHasMore: a full page reports more history, a short one
// does not, and only the newest page may be short.
func TestListPagesSizeAndHasMore(t *testing.T) {
	svc, sessionID := newTestService(t)

	for i := range 230 {
		_, err := svc.Create(t.Context(), sessionID, CreateMessageParams{
			Role:  User,
			Parts: []ContentPart{TextContent{Text: fmt.Sprintf("prompt %d", i)}},
		})
		require.NoError(t, err)
	}

	page, err := svc.ListLatest(t.Context(), sessionID, DefaultPageLimit)
	require.NoError(t, err)
	require.Len(t, page.Messages, DefaultPageLimit)
	require.True(t, page.HasMore)
	require.NotZero(t, page.Cursor)

	page, err = svc.ListOlder(t.Context(), sessionID, page.Cursor, DefaultPageLimit)
	require.NoError(t, err)
	require.Len(t, page.Messages, DefaultPageLimit)
	require.True(t, page.HasMore)

	// The oldest page holds the remaining 30 and ends the walk.
	page, err = svc.ListOlder(t.Context(), sessionID, page.Cursor, DefaultPageLimit)
	require.NoError(t, err)
	require.Len(t, page.Messages, 30)
	require.False(t, page.HasMore)
}

// TestListOlderIsStableAcrossInserts: the cursor is a sequence number, not an
// offset, so messages written between two page reads cannot shift the older
// page. An OFFSET-based page would slide by however many rows landed in
// between and show the same message twice.
func TestListOlderIsStableAcrossInserts(t *testing.T) {
	svc, sessionID := newTestService(t)

	for i := range 200 {
		_, err := svc.Create(t.Context(), sessionID, CreateMessageParams{
			Role:  User,
			Parts: []ContentPart{TextContent{Text: fmt.Sprintf("prompt %d", i)}},
		})
		require.NoError(t, err)
	}

	first, err := svc.ListLatest(t.Context(), sessionID, DefaultPageLimit)
	require.NoError(t, err)
	require.Len(t, first.Messages, DefaultPageLimit)

	// A run starts on another client while the reader is scrolling up.
	for i := range 20 {
		_, err := svc.Create(t.Context(), sessionID, CreateMessageParams{
			Role:  User,
			Parts: []ContentPart{TextContent{Text: fmt.Sprintf("appended %d", i)}},
		})
		require.NoError(t, err)
	}

	second, err := svc.ListOlder(t.Context(), sessionID, first.Cursor, DefaultPageLimit)
	require.NoError(t, err)
	require.Len(t, second.Messages, DefaultPageLimit)

	// The newest message of the older page must sit immediately before the
	// oldest message of the page it came from: no gap, no overlap.
	require.Equal(t,
		fmt.Sprintf("prompt %d", 200-DefaultPageLimit-1),
		textOf(t, second.Messages[len(second.Messages)-1]),
	)
	require.Equal(t,
		fmt.Sprintf("prompt %d", 200-DefaultPageLimit-2),
		textOf(t, second.Messages[len(second.Messages)-2]),
	)
}

// TestListLatestEmptySession: a session with no messages yields an empty page
// that ends the walk rather than a cursor that would page forever.
func TestListLatestEmptySession(t *testing.T) {
	svc, sessionID := newTestService(t)

	page, err := svc.ListLatest(t.Context(), sessionID, DefaultPageLimit)
	require.NoError(t, err)
	require.Empty(t, page.Messages)
	require.Zero(t, page.Cursor)
	require.False(t, page.HasMore)
}

// TestListPagesMatchFullReadOrder: the paged walk agrees with List about which
// messages exist. List still orders by created_at, so only the set is compared.
func TestListPagesMatchFullReadOrder(t *testing.T) {
	svc, sessionID := newTestService(t)

	for i := range 137 {
		_, err := svc.Create(t.Context(), sessionID, CreateMessageParams{
			Role:  User,
			Parts: []ContentPart{TextContent{Text: fmt.Sprintf("prompt %d", i)}},
		})
		require.NoError(t, err)
	}

	full, err := svc.List(t.Context(), sessionID)
	require.NoError(t, err)

	paged := walkPages(t, svc, sessionID, 40)
	require.Len(t, paged, len(full))

	ids := func(msgs []Message) map[string]bool {
		set := make(map[string]bool, len(msgs))
		for _, m := range msgs {
			require.False(t, set[m.ID], "message %s returned by two pages", m.ID)
			set[m.ID] = true
		}
		return set
	}
	require.Equal(t, ids(full), ids(paged))
}

func textOf(t *testing.T, m Message) string {
	t.Helper()
	text, ok := m.Parts[0].(TextContent)
	require.True(t, ok)
	return text.Text
}
