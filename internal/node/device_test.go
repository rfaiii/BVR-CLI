package node

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManagerSnapshotsAreSortedAndIndependent(t *testing.T) {
	m := NewManager()
	m.Upsert(Device{ID: "z", Name: "Z", Status: DeviceStatusOnline})
	m.Upsert(Device{ID: "a", Name: "A", Status: DeviceStatusOffline})
	devices := m.Devices()
	require.Equal(t, []string{"a", "z"}, []string{devices[0].ID, devices[1].ID})
	devices[0].Name = "changed"
	require.Equal(t, "A", m.Devices()[0].Name)
}

func TestDiscoverUsesInjectedProbe(t *testing.T) {
	probes := []Probe{{ID: "one", Address: "one:1"}, {ID: "two", Address: "two:2"}}
	got := Discover(probes, func(probe Probe) bool { return probe.ID == "two" })
	require.Len(t, got, 1)
	require.Equal(t, "two", got[0].ID)
	require.Equal(t, DeviceStatusOnline, got[0].Status)
}

func TestSetStatusUpdatesLastSeen(t *testing.T) {
	m := NewManager()
	m.Upsert(Device{ID: "one", Status: DeviceStatusOffline})
	before := time.Now()
	m.SetStatus("one", DeviceStatusOnline)
	device := m.Devices()[0]
	require.Equal(t, DeviceStatusOnline, device.Status)
	require.False(t, device.LastSeen.Before(before))
}

func TestSetNicknamePersists(t *testing.T) {
	m := NewManager()
	m.Upsert(Device{ID: "one", Name: "One"})
	m.SetNickname("one", "Workstation")

	require.Equal(t, "Workstation", m.Devices()[0].Nickname)
	require.Equal(t, "Workstation", m.Devices()[0].DisplayName())

	// Renaming an unknown id must not create a device.
	m.SetNickname("missing", "Ghost")
	require.Len(t, m.Devices(), 1)
}
