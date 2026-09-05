package node

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPTransportExecutesAndAuthenticates(t *testing.T) {
	server := httptest.NewServer((&NodeHTTPServer{Token: "secret", Transport: fakeHTTPTransport{}}).Handler())
	defer server.Close()
	transport := NewHTTPTransport(server.URL, "secret")
	result, err := transport.Execute(context.Background(), CommandRequest{Command: "node"})
	require.NoError(t, err)
	require.Equal(t, "remote output", result.Stdout)
	bad := NewHTTPTransport(server.URL, "wrong")
	_, err = bad.Execute(context.Background(), CommandRequest{Command: "node"})
	require.Error(t, err)
}

func TestHTTPTransportPing(t *testing.T) {
	server := httptest.NewServer((&NodeHTTPServer{Token: "secret"}).Handler())
	defer server.Close()
	require.NoError(t, NewHTTPTransport(server.URL, "secret").Ping(context.Background()))
}

func TestHTTPServerRejectsOversizedRequest(t *testing.T) {
	server := httptest.NewServer((&NodeHTTPServer{MaxBody: 8}).Handler())
	defer server.Close()
	request := NewHTTPTransport(server.URL, "")
	_, err := request.Execute(context.Background(), CommandRequest{Command: "this request is too large"})
	require.Error(t, err)
}

func TestNodeAgentHandlerServesHTTPAndWebSocket(t *testing.T) {
	server := httptest.NewServer(NodeAgentHandler("secret", fakeHTTPTransport{}))
	defer server.Close()
	transport := NewHTTPTransport(server.URL, "secret")
	require.NoError(t, transport.Ping(context.Background()))
	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)
	ws := NewWebSocketTransport("ws://"+parsed.Host, "secret")
	defer ws.Close()
	result, err := ws.Execute(context.Background(), CommandRequest{Command: "node"})
	require.NoError(t, err)
	require.Equal(t, "remote output", result.Stdout)
}

type fakeHTTPTransport struct{}

func (fakeHTTPTransport) Execute(context.Context, CommandRequest) (CommandResult, error) {
	return CommandResult{Stdout: "remote output"}, nil
}
func (fakeHTTPTransport) Ping(context.Context) error { return nil }
