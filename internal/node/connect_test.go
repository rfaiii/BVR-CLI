package node

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectSavesDeviceWithoutToken(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hi"), 0o644))
	share, err := NewFileShare([]string{root})
	require.NoError(t, err)
	t.Cleanup(share.Close)

	server := httptest.NewServer(NewAgentHandler(AgentOptions{
		Token: "secret", Share: share, NodeID: "studio", NodeName: "Studio",
	}))
	defer server.Close()

	dataDir := t.TempDir()
	t.Setenv("BVR_GLOBAL_DATA", dataDir)
	resetDefaultManager(t)

	device, err := Connect(context.Background(), server.Listener.Addr().String(), "secret", "Desk")
	require.NoError(t, err)
	require.Equal(t, "studio", device.ID)
	require.Equal(t, "Desk", device.Nickname)
	require.Equal(t, DeviceStatusOnline, device.Status)

	saved, err := os.ReadFile(DefaultDevicesPath())
	require.NoError(t, err)
	require.NotContains(t, string(saved), "secret")
	require.Contains(t, string(saved), `"id"`)
	require.NotContains(t, string(saved), `"ID"`)

	creds, err := os.ReadFile(DefaultCredentialsPath())
	require.NoError(t, err)
	require.Contains(t, string(creds), "secret")
	if runtime.GOOS != "windows" {
		info, err := os.Stat(DefaultCredentialsPath())
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	projects, err := NewHTTPTransport(mustHTTPBase(t, device.TransportURL), "secret").ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)

	client := &FilesClient{
		HTTP: NewHTTPTransport(mustHTTPBase(t, device.TransportURL), "secret"), Project: projects[0].ID,
	}
	body, err := client.Read(context.Background(), "hello.txt")
	require.NoError(t, err)
	require.Equal(t, "hi", body.Content)
}

func TestLegacyNodesJSONStillLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nodes.json")
	raw := []byte(`[{"ID":"studio","Name":"Studio","Nickname":"Desk","ConnectionType":"node","Address":"10.0.0.5:7799","Status":"online","LastSeen":"2026-01-02T03:04:05Z","TransportURL":"ws://10.0.0.5:7799/v1/node/ws"}]`)
	require.NoError(t, os.WriteFile(path, raw, 0o600))

	manager := NewManager()
	added, err := manager.Restore(path)
	require.NoError(t, err)
	require.Equal(t, 1, added)
	device := manager.Devices()[0]
	require.Equal(t, "studio", device.ID)
	require.Equal(t, "Desk", device.Nickname)
	require.Equal(t, "10.0.0.5:7799", device.Address)
	require.Equal(t, "ws://10.0.0.5:7799/v1/node/ws", device.TransportURL)
	require.Equal(t, DeviceStatusOffline, device.Status)

	encoded, err := json.Marshal(device)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"id":"studio"`)
	require.Contains(t, string(encoded), `"connection_type":"node"`)
	require.NotContains(t, string(encoded), `"ID"`)
}

func TestStartupBannerAlwaysNamesTheListener(t *testing.T) {
	banner := StartupInfo{Addr: "127.0.0.1:7777"}.Banner()
	require.Contains(t, banner, "NODE agent listening on 127.0.0.1:7777")
	require.Contains(t, banner, "uname")
	require.Contains(t, banner, "No projects shared")

	open := StartupInfo{Addr: "0.0.0.0:7777", Exec: ExecPolicy{Unrestricted: true}}.Banner()
	require.Contains(t, open, "unrestricted")
}

func TestParseDialDefaultsPort(t *testing.T) {
	httpBase, wsURL, hostport, err := parseDial("127.0.0.1")
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:7777", hostport)
	require.Equal(t, "http://127.0.0.1:7777", httpBase)
	require.Equal(t, "ws://127.0.0.1:7777/v1/node/ws", wsURL)

	_, wsURL, _, err = parseDial("ws://192.168.1.20:7777/v1/node/ws")
	require.NoError(t, err)
	require.Equal(t, "ws://192.168.1.20:7777/v1/node/ws", wsURL)
}

func resetDefaultManager(t *testing.T) {
	t.Helper()
	defaultManager = NewManager()
	defaultTransportManager = NewTransportManager(defaultManager)
	secretMu.Lock()
	secrets = map[string]string{}
	secretMu.Unlock()
	t.Cleanup(func() {
		defaultManager = NewManager()
		defaultTransportManager = NewTransportManager(defaultManager)
		secretMu.Lock()
		secrets = map[string]string{}
		secretMu.Unlock()
	})
}

func mustHTTPBase(t *testing.T, transportURL string) string {
	t.Helper()
	base, err := HTTPBase(transportURL)
	require.NoError(t, err)
	return base
}
