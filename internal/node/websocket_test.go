package node

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebSocketTransportStreamsAndPings(t *testing.T) {
	httpServer := httptest.NewServer((&NodeWebSocketServer{Token: "secret"}).Handler())
	defer httpServer.Close()
	parsed, err := url.Parse(httpServer.URL)
	require.NoError(t, err)
	transport := NewWebSocketTransport("ws://"+parsed.Host, "secret")
	defer transport.Close()
	result, err := transport.Execute(context.Background(), CommandRequest{Command: "printf", Args: []string{"hello"}})
	require.NoError(t, err)
	require.Equal(t, "hello", strings.TrimSpace(result.Stdout))
	require.NoError(t, transport.Ping(context.Background()))
}

func TestWebSocketTransportRejectsBadToken(t *testing.T) {
	httpServer := httptest.NewServer((&NodeWebSocketServer{Token: "secret"}).Handler())
	defer httpServer.Close()
	parsed, err := url.Parse(httpServer.URL)
	require.NoError(t, err)
	err = NewWebSocketTransport("ws://"+parsed.Host, "wrong").Ping(context.Background())
	require.Error(t, err)
}
