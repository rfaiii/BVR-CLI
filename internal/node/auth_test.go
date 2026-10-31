package node

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestTokensEqualIgnoresTimingShape(t *testing.T) {
	require.True(t, tokensEqual("secret", "secret"))
	require.False(t, tokensEqual("secret", "secre"))
	require.False(t, tokensEqual("secret", "secret "))
	require.False(t, tokensEqual("", "x"))
}

func TestAuthorizedBearer(t *testing.T) {
	open := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/node/health", nil)
	require.True(t, authorizedBearer("", open))

	authed := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/node/health", nil)
	authed.Header.Set("Authorization", "Bearer secret")
	require.True(t, authorizedBearer("secret", authed))

	wrong := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/node/health", nil)
	wrong.Header.Set("Authorization", "Bearer nope")
	require.False(t, authorizedBearer("secret", wrong))
}

func TestExecPolicyAllowList(t *testing.T) {
	var policy ExecPolicy
	require.NoError(t, policy.Allows("uname"))
	require.NoError(t, policy.Allows("/usr/bin/uname"))
	require.Error(t, policy.Allows("printf"))
	require.Error(t, policy.Allows("bash"))

	policy.Extra = []string{"printf"}
	require.NoError(t, policy.Allows("printf"))

	policy.Unrestricted = true
	require.NoError(t, policy.Allows("bash"))
	require.Equal(t, []string{"*"}, policy.Names())
}

func TestOriginAllowed(t *testing.T) {
	same := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1:7777/v1/node/ws", nil)
	same.Host = "127.0.0.1:7777"
	same.Header.Set("Origin", "http://127.0.0.1:7777")
	require.True(t, originAllowed(same))

	native := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1:7777/v1/node/ws", nil)
	native.Host = "127.0.0.1:7777"
	require.True(t, originAllowed(native))

	evil := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1:7777/v1/node/ws", nil)
	evil.Host = "127.0.0.1:7777"
	evil.Header.Set("Origin", "https://evil.example")
	require.False(t, originAllowed(evil))
}

func TestWebSocketRejectsForeignOrigin(t *testing.T) {
	handler := NewAgentHandler(AgentOptions{Token: "secret"})
	server := httptest.NewServer(handler)
	defer server.Close()

	dialer := *websocket.DefaultDialer
	header := http.Header{}
	header.Set("Origin", "https://evil.example")
	header.Set("Authorization", "Bearer secret")
	_, resp, err := dialer.DialContext(t.Context(), wsURL(server.URL), header)
	require.Error(t, err)
	if resp != nil {
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
		_ = resp.Body.Close()
	}
}

func TestListenHostIsLoopback(t *testing.T) {
	require.True(t, ListenHostIsLoopback("127.0.0.1:7777"))
	require.True(t, ListenHostIsLoopback("localhost:7777"))
	require.False(t, ListenHostIsLoopback("0.0.0.0:7777"))
	require.False(t, ListenHostIsLoopback("192.168.1.20:7777"))
}

func wsURL(httpURL string) string {
	return "ws" + httpURL[len("http"):] + websocketPath
}
