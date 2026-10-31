package node

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfineRejectsEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("nope"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(root, "sub"), 0o755))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "sub", "link")))
	require.NoError(t, os.Symlink(filepath.Join(root, "sub"), filepath.Join(root, "inside-link")))

	resolvedRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	ok, err := confine(root, "sub/note.txt")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(resolvedRoot, "sub", "note.txt"), ok)

	_, err = confine(root, "../secret.txt")
	require.ErrorIs(t, err, errPathEscape)
	_, err = confine(root, "/etc/passwd")
	require.ErrorIs(t, err, errPathEscape)
	_, err = confine(root, "sub/link/secret.txt")
	require.ErrorIs(t, err, errSymlink)
	_, err = confine(root, "inside-link/note.txt")
	require.ErrorIs(t, err, errSymlink)
}

func TestFileAPIConflictAndConfinement(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("nope"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "escape")))

	share, err := NewFileShare([]string{root})
	require.NoError(t, err)
	t.Cleanup(share.Close)
	projects := share.Projects()
	require.Len(t, projects, 1)
	id := projects[0].ID

	handler := NewAgentHandler(AgentOptions{Token: "secret", Share: share, NodeID: "main"})
	server := httptest.NewServer(handler)
	defer server.Close()
	client := &FilesClient{
		HTTP:         NewHTTPTransport(server.URL, "secret"),
		WebSocketURL: "ws://" + server.Listener.Addr().String() + websocketPath,
		Token:        "secret",
		Project:      id,
	}
	ctx := context.Background()

	created, err := client.Write(ctx, "notes/a.txt", []byte("one"), "")
	require.NoError(t, err)
	require.Equal(t, etagBytes([]byte("one")), created.ETag)
	onDisk, err := os.ReadFile(filepath.Join(root, "notes", "a.txt"))
	require.NoError(t, err)
	require.Equal(t, "one", string(onDisk))

	_, err = client.Write(ctx, "notes/a.txt", []byte("two"), "stale")
	require.True(t, IsConflict(err), "err=%v", err)
	body, err := os.ReadFile(filepath.Join(root, "notes", "a.txt"))
	require.NoError(t, err)
	require.Equal(t, "one", string(body), "a stale etag must not clobber the file")

	updated, err := client.Write(ctx, "notes/a.txt", []byte("two"), created.ETag)
	require.NoError(t, err)
	require.Equal(t, etagBytes([]byte("two")), updated.ETag)

	_, err = client.Read(ctx, "../secret.txt")
	require.Error(t, err)
	_, err = client.Read(ctx, "escape")
	require.Error(t, err)
	_, err = client.Write(ctx, "escape", []byte("pwn"), "")
	require.Error(t, err)
	_, err = os.ReadFile(filepath.Join(outside, "secret.txt"))
	require.NoError(t, err)
	require.Equal(t, "nope", string(mustRead(t, filepath.Join(outside, "secret.txt"))))

	list, err := client.List(ctx, ".", 1)
	require.NoError(t, err)
	require.NotEmpty(t, list.Entries)

	events, err := client.Watch(ctx, "notes")
	require.NoError(t, err)
	_, err = client.Put(ctx, "notes/b.txt", []byte("watch-me"))
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		select {
		case ev := <-events:
			return ev.Path == "notes/b.txt" && (ev.Op == "create" || ev.Op == "write")
		default:
			return false
		}
	}, 3*time.Second, 20*time.Millisecond)
}

func TestExecuteAllowListOnAgent(t *testing.T) {
	server := httptest.NewServer(NewAgentHandler(AgentOptions{Token: "secret"}))
	defer server.Close()
	blocked := NewHTTPTransport(server.URL, "secret")
	_, err := blocked.Execute(context.Background(), CommandRequest{Command: "printf", Args: []string{"hi"}})
	require.Error(t, err)
	var remote *RemoteError
	require.ErrorAs(t, err, &remote)
	require.Equal(t, http.StatusForbidden, remote.Status)

	open := httptest.NewServer(NewAgentHandler(AgentOptions{
		Token: "secret",
		Exec:  ExecPolicy{Unrestricted: true},
	}))
	defer open.Close()
	result, err := NewHTTPTransport(open.URL, "secret").Execute(context.Background(), CommandRequest{Command: "printf", Args: []string{"hi"}})
	require.NoError(t, err)
	require.Equal(t, "hi", result.Stdout)
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}
