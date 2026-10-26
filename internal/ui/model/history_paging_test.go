package model

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/richavery/bvr-cli/internal/history"
	"github.com/richavery/bvr-cli/internal/message"
	"github.com/richavery/bvr-cli/internal/session"
	"github.com/richavery/bvr-cli/internal/ui/chat"
)

// TestChatPrependMessagesKeepsIndexMap: pulling in an older page shifts every
// registered item index. Without the shift, MessageItem lookups would resolve
// to the wrong row and the next live message would be routed to a stale item.
func TestChatPrependMessagesKeepsIndexMap(t *testing.T) {
	t.Parallel()

	u := newTestUI()

	newer := chat.NewAssistantMessageItem(u.com.Styles, &message.Message{
		ID: "m-new", Role: message.Assistant,
		Parts: []message.ContentPart{message.TextContent{Text: "newer"}},
	})
	older := chat.NewAssistantMessageItem(u.com.Styles, &message.Message{
		ID: "m-old", Role: message.Assistant,
		Parts: []message.ContentPart{message.TextContent{Text: "older"}},
	})

	u.chat.SetMessages(newer)
	u.chat.PrependMessages(older)

	require.Equal(t, 2, u.chat.Len())
	require.Same(t, older, u.chat.list.ItemAt(0), "the older page must sit in front")
	require.Same(t, newer, u.chat.list.ItemAt(1))

	require.NotNil(t, u.chat.MessageItem("m-new"), "existing item lookup survived the shift")
	require.NotNil(t, u.chat.MessageItem("m-old"))
	require.Same(t, newer, u.chat.MessageItem("m-new"))
}

// TestPrependOlderMessagesLinksResultsAcrossPages: a tool call that arrived in
// the older page has its result in the newer page, so the item built for it
// must find that result. Tool results are collected per page, which on its own
// would render the call as still running.
func TestPrependOlderMessagesLinksResultsAcrossPages(t *testing.T) {
	t.Parallel()

	u := newTestUI()

	newerPage := []message.Message{{
		ID: "m-result", Role: message.Tool,
		Parts: []message.ContentPart{
			message.ToolResult{ToolCallID: "call-1", Name: "lookup", Content: "the answer"},
		},
	}}
	olderPage := []message.Message{{
		ID: "m-call", Role: message.Assistant,
		Parts: []message.ContentPart{
			message.ToolCall{ID: "call-1", Name: "lookup", Input: `{"q":"x"}`, Finished: true},
		},
	}}

	u.applyNewestHistory(u.prepareNewestPage(newerPage, 0))
	u.applyOlderHistory(u.prepareOlderPage(olderPage, u.historyToolResults))

	_, ok := u.historyToolResults["call-1"]
	require.True(t, ok, "results from the newer page must stay registered")

	// The call lives in the item built for the older page; the tool call ID is
	// that item's key.
	item, ok := u.chat.MessageItem("call-1").(chat.ToolMessageItem)
	require.True(t, ok, "the older page should have produced a tool item")
	require.Equal(t, "m-call", item.MessageID())
	require.Contains(t, ansi.Strip(item.RawRender(120)), "the answer",
		"the tool call must render with the result that arrived on a newer page")
}

// TestSessionOpenReadsOnePage: opening a session must read a page of history,
// not the whole session, and must do it in the command rather than in Update.
func TestSessionOpenReadsOnePage(t *testing.T) {
	t.Parallel()

	ws := &pagingWorkspace{
		countingWorkspace: &countingWorkspace{ready: true},
		page:              message.MessagePage{Messages: pageMessages(0, 3), Cursor: 97, HasMore: true},
	}
	m := newBusyUI(ws)

	loaded, ok := firstMsg(t, m.loadSession("s1")).(loadSessionMsg)
	require.True(t, ok, "loading a session must yield a loadSessionMsg")

	require.Equal(t, 1, ws.pageReads, "session open reads exactly one page")
	require.Zero(t, ws.fullReads, "session open must not read the whole history")
	require.Equal(t, message.DefaultPageLimit, ws.lastLimit)
	require.Zero(t, ws.lastBefore, "the first read starts at the newest message")

	require.Equal(t, int64(97), loaded.historyCursor)
	require.True(t, loaded.historyHasMore)
	require.Len(t, loaded.history.items, 3, "the page is converted to items off the Update goroutine")
}

// TestScrollToTopPullsOlderPageOnce: a session that still has history below
// what is on screen reads exactly one older page, and the in-flight flag stops
// a second read until the first one lands.
func TestScrollToTopPullsOlderPageOnce(t *testing.T) {
	ws := &pagingWorkspace{
		countingWorkspace: &countingWorkspace{ready: true},
		page:              message.MessagePage{Messages: pageMessages(3, 2), Cursor: 95, HasMore: false},
	}
	m := newBusyUI(ws)

	_, cmd := m.Update(loadSessionMsg{
		session:        &session.Session{ID: "s1"},
		history:        m.prepareNewestPage(pageMessages(0, 3), 0),
		historyCursor:  97,
		historyHasMore: true,
	})
	require.True(t, m.historyLoading, "the older page must be scheduled")

	// Update must not touch history itself, and a second frame while the read
	// is in flight must not schedule another one.
	_, second := m.Update(plainMsg{})
	require.Zero(t, ws.pageReads, "the read belongs to the command, not the Update goroutine")
	runCmds(m, second)
	require.Zero(t, ws.pageReads, "only one older page may be in flight")

	runCmds(m, cmd)
	require.Equal(t, 1, ws.pageReads)

	require.Equal(t, message.DefaultPageLimit, ws.lastLimit)
	require.Equal(t, int64(97), ws.lastBefore, "the read starts below the oldest shown message")
	require.Equal(t, int64(95), m.historyCursor, "the cursor advances to the new oldest message")
	require.False(t, m.historyHasMore)
	require.False(t, m.historyLoading)
	require.Equal(t, 5, m.chat.Len(), "the older page is prepended, not appended")
	require.Same(t, m.chat.MessageItem("msg-3"), m.chat.list.ItemAt(0))
	require.NotNil(t, m.chat.MessageItem("msg-0"), "newer items stay reachable after a prepend")
}

// TestNoOlderPageWhenHistoryComplete: a short session must never ask for more.
func TestNoOlderPageWhenHistoryComplete(t *testing.T) {
	ws := &pagingWorkspace{
		countingWorkspace: &countingWorkspace{ready: true},
		page:              message.MessagePage{Messages: pageMessages(0, 2)},
	}
	m := newBusyUI(ws)

	_, cmd := m.Update(loadSessionMsg{
		session: &session.Session{ID: "s1"},
		history: m.prepareNewestPage(pageMessages(0, 2), 0),
	})
	runCmds(m, cmd)

	require.Zero(t, ws.pageReads)
	require.Zero(t, ws.fullReads)
	require.False(t, m.historyLoading)
	require.Equal(t, 2, m.chat.Len())
}

// TestUpdateDoesNotBuildHistory pins the split that keeps a session switch from
// stalling the paint: by the time a page lands in Update it is already converted
// into chat items, so Update only moves them into the list.
func TestUpdateDoesNotBuildHistory(t *testing.T) {
	ws := &countingWorkspace{ready: true}
	m := newBusyUI(ws)

	history := m.prepareNewestPage(pageMessages(0, 3), 0)
	require.Len(t, history.items, 3, "items must exist before the page reaches Update")
	builtDuring := ws.historyReads

	_, cmd := m.Update(loadSessionMsg{
		session: &session.Session{ID: "s1"},
		history: history,
	})
	runCmds(m, cmd)

	require.Equal(t, builtDuring, ws.historyReads,
		"Update must not read history when a page lands: nested agent sessions are the command's job")
	require.Equal(t, 3, m.chat.Len())
}

// pagingWorkspace serves one fixed history page and counts how it was read.
// It embeds the busy-state stub so Update can run end to end.
type pagingWorkspace struct {
	*countingWorkspace

	page message.MessagePage

	fullReads int
	pageReads int

	lastBefore int64
	lastLimit  int
}

func (w *pagingWorkspace) ListMessages(context.Context, string) ([]message.Message, error) {
	w.fullReads++
	return nil, nil
}

func (w *pagingWorkspace) ListMessagesPage(
	_ context.Context, _ string, limit int, before int64,
) (message.MessagePage, error) {
	w.pageReads++
	w.lastLimit = limit
	w.lastBefore = before
	return w.page, nil
}

func (w *pagingWorkspace) GetSession(context.Context, string) (session.Session, error) {
	return session.Session{ID: "s1"}, nil
}

func (w *pagingWorkspace) SetCurrentSession(context.Context, string) error { return nil }

func (w *pagingWorkspace) ListSessionHistory(context.Context, string) ([]history.File, error) {
	return nil, nil
}

func (w *pagingWorkspace) FileTrackerListReadFiles(context.Context, string) ([]string, error) {
	return nil, nil
}

// firstMsg runs a command the way the Bubble Tea runtime would and returns the
// first message it produces, unwrapping a batch.
func firstMsg(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	require.NotNil(t, cmd, "expected a command")

	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if m := c(); m != nil {
				return m
			}
		}
		return nil
	}
	return msg
}

// pageMessages builds n messages with ids msg-offset .. msg-(offset+n-1).
func pageMessages(offset, n int) []message.Message {
	msgs := make([]message.Message, 0, n)
	for i := range n {
		msgs = append(msgs, message.Message{
			ID:   fmt.Sprintf("msg-%d", offset+i),
			Role: message.User,
			Parts: []message.ContentPart{
				message.TextContent{Text: fmt.Sprintf("prompt %d", offset+i)},
			},
			CreatedAt: int64(1000 + offset + i),
		})
	}
	return msgs
}

// benchPage builds n messages with bodies sized like real turns, so the two
// halves of a session open are measured against plausible content.
func benchPage(n int) []message.Message {
	body := strings.Repeat("rendered history line\n", 20)
	msgs := make([]message.Message, 0, n)
	for i := range n {
		msgs = append(msgs, message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			Role:      message.User,
			Parts:     []message.ContentPart{message.TextContent{Text: fmt.Sprintf("prompt %d: %s", i, body)}},
			CreatedAt: int64(1000 + i),
		})
	}
	return msgs
}

// BenchmarkSessionOpenStages splits a session open into the part that runs on a
// command goroutine (prepare: render every message into items and read nested
// agent sessions) and the part still left between frames (apply: put the
// prepared items into the chat list).
func BenchmarkSessionOpenStages(b *testing.B) {
	ws := &countingWorkspace{ready: true}
	m := newBusyUI(ws)
	msgs := benchPage(message.DefaultPageLimit)

	b.Run("prepare-off-thread", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = m.prepareNewestPage(msgs, 0)
		}
	})

	history := m.prepareNewestPage(msgs, 0)
	b.Run("apply-in-update", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = m.applyNewestHistory(history)
		}
	})
}
