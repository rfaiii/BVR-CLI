package node

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func discoveredNode(id, address string) Node {
	return Node{
		Instance:  "studio",
		Addresses: []net.IP{net.ParseIP(address)},
		Port:      7799,
		Protocol:  ProtocolVersion,
		ID:        id,
		Name:      "Studio Mac",
	}
}

func TestDeviceFromNode(t *testing.T) {
	device := DeviceFromNode(discoveredNode("studio-mac", "172.20.20.20"))

	require.Equal(t, "studio-mac", device.ID, "the TXT id is the registry identity")
	require.Equal(t, "Studio Mac", device.Name)
	require.Equal(t, "node", device.ConnectionLabel())
	require.Equal(t, "172.20.20.20:7799", device.Address)
	require.Equal(t, "ws://172.20.20.20:7799/v1/node/ws", device.TransportURL)
	require.Equal(t, DeviceStatusOnline, device.Status)
	require.False(t, device.LastSeen.IsZero())
}

func TestRecordNodesSkipsUnaddressableNodes(t *testing.T) {
	m := NewManager()

	// A browse answer without an A record has nothing to dial.
	recorded := m.RecordNodes([]Node{{ID: "ghost", Instance: "ghost"}, discoveredNode("studio", "10.0.0.5")})

	require.Len(t, recorded, 1)
	require.Equal(t, []Device{{
		ID: "studio", Name: "Studio Mac", ConnectionType: "node",
		Address: "10.0.0.5:7799", Status: DeviceStatusOnline,
		TransportURL: "ws://10.0.0.5:7799/v1/node/ws",
	}}, withoutLastSeen(recorded))
	require.Empty(t, m.Devices()[0].Nickname)
}

func TestRecordNodesKeepsNicknameAcrossScans(t *testing.T) {
	m := NewManager()
	m.RecordNodes([]Node{discoveredNode("studio", "10.0.0.5")})
	m.SetNickname("studio", "Workstation")

	// Re-scanning rewrites the record; the name the user chose must survive it.
	m.RecordNodes([]Node{discoveredNode("studio", "10.0.0.5")})

	require.Equal(t, "Workstation", m.Devices()[0].Nickname)
}

func TestSaveRestoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "nodes.json")
	source := NewManager()
	source.RecordNodes([]Node{discoveredNode("studio", "10.0.0.5")})
	seen := source.Devices()[0].LastSeen

	require.NoError(t, source.Save(path))

	target := NewManager()
	added, err := target.Restore(path)
	require.NoError(t, err)
	require.Equal(t, 1, added)

	restored := target.Devices()[0]
	require.Equal(t, "studio", restored.ID)
	require.Equal(t, "10.0.0.5:7799", restored.Address)
	require.Equal(t, "ws://10.0.0.5:7799/v1/node/ws", restored.TransportURL)
	require.Equal(t, seen.Unix(), restored.LastSeen.Unix(), "when the host was last seen is the point of persisting it")
	// Nothing has confirmed this host since the process started.
	require.Equal(t, DeviceStatusOffline, restored.Status)

	// The file lists LAN hosts, so it must not be readable by other accounts.
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	require.NoFileExists(t, path+".tmp")
}

func TestRestoreKeepsLiveEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nodes.json")
	saved := NewManager()
	saved.RecordNodes([]Node{discoveredNode("studio", "10.0.0.5")})
	require.NoError(t, saved.Save(path))

	live := NewManager()
	live.RecordNodes([]Node{discoveredNode("studio", "192.168.1.9")})

	added, err := live.Restore(path)
	require.NoError(t, err)
	require.Zero(t, added, "a host seen in this run is not replaced by an older record")
	require.Equal(t, "192.168.1.9:7799", live.Devices()[0].Address)
	require.Equal(t, DeviceStatusOnline, live.Devices()[0].Status)
}

func TestRestoreWithoutAFileIsNotAnError(t *testing.T) {
	added, err := NewManager().Restore(filepath.Join(t.TempDir(), "nodes.json"))
	require.NoError(t, err)
	require.Zero(t, added)
}

func TestRestoreRejectsACorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nodes.json")
	require.NoError(t, os.WriteFile(path, []byte(`[{"ID":`), 0o600))

	_, err := NewManager().Restore(path)
	require.ErrorContains(t, err, "unmarshal device registry")
}

func TestRestoreIgnoresAnonymousRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nodes.json")
	data, err := json.Marshal([]Device{{ID: "ok", Address: "10.0.0.5:7799"}, {Address: "orphan"}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	m := NewManager()
	added, err := m.Restore(path)
	require.NoError(t, err)
	require.Equal(t, 1, added)
	require.Len(t, m.Devices(), 1)
}

// withoutLastSeen zeroes timestamps that are only asserted to exist, so the
// rest of a record can be compared exactly.
func withoutLastSeen(devices []Device) []Device {
	out := make([]Device, len(devices))
	for i, device := range devices {
		device.LastSeen = time.Time{}
		out[i] = device
	}
	return out
}
