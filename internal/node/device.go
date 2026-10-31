// Package node contains the device registry and discovery primitives used by
// the BVR NODE connection UI and future remote transports.
package node

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"slices"
	"sort"
	"sync"
	"time"
)

type DeviceStatus string

const (
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusError   DeviceStatus = "error"
)

// Device is one row in the NODE registry. JSON uses snake_case. Older
// nodes.json files used the Go field names (ID, TransportURL, …); UnmarshalJSON
// accepts both so a file written before the tags still loads.
type Device struct {
	ID             string       `json:"id"`
	Name           string       `json:"name,omitempty"`
	Nickname       string       `json:"nickname,omitempty"`
	ConnectionType string       `json:"connection_type,omitempty"`
	Address        string       `json:"address,omitempty"`
	Status         DeviceStatus `json:"status,omitempty"`
	LastSeen       time.Time    `json:"last_seen,omitempty"`
	TransportURL   string       `json:"transport_url,omitempty"`
}

// UnmarshalJSON reads a current nodes.json object or a legacy one that
// stored exported Go names.
func (d *Device) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var out Device
	if err := firstJSON(raw, &out.ID, "id", "ID"); err != nil {
		return err
	}
	if err := firstJSON(raw, &out.Name, "name", "Name"); err != nil {
		return err
	}
	if err := firstJSON(raw, &out.Nickname, "nickname", "Nickname"); err != nil {
		return err
	}
	if err := firstJSON(raw, &out.ConnectionType, "connection_type", "ConnectionType"); err != nil {
		return err
	}
	if err := firstJSON(raw, &out.Address, "address", "Address"); err != nil {
		return err
	}
	if err := firstJSON(raw, &out.Status, "status", "Status"); err != nil {
		return err
	}
	if err := firstJSON(raw, &out.LastSeen, "last_seen", "LastSeen"); err != nil {
		return err
	}
	if err := firstJSON(raw, &out.TransportURL, "transport_url", "TransportURL"); err != nil {
		return err
	}
	*d = out
	return nil
}

func firstJSON(raw map[string]json.RawMessage, dest any, keys ...string) error {
	for _, key := range keys {
		val, ok := raw[key]
		if !ok || string(val) == "null" {
			continue
		}
		return json.Unmarshal(val, dest)
	}
	return nil
}

func (d Device) DisplayName() string {
	if d.Nickname != "" {
		return d.Nickname
	}
	if d.Name != "" {
		return d.Name
	}
	return d.ID
}

func (d Device) ConnectionLabel() string {
	if d.ConnectionType == "" {
		return "local"
	}
	return d.ConnectionType
}

type Manager struct {
	mu       sync.RWMutex
	devices  map[string]Device
	onChange func([]Device)
}

func NewManager() *Manager { return &Manager{devices: make(map[string]Device)} }

func (m *Manager) OnChange(fn func([]Device)) {
	m.mu.Lock()
	m.onChange = fn
	m.mu.Unlock()
}

func (m *Manager) Devices() []Device {
	m.mu.RLock()
	devices := make([]Device, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	m.mu.RUnlock()
	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })
	return devices
}

func (m *Manager) emit() {
	m.mu.RLock()
	fn := m.onChange
	m.mu.RUnlock()
	if fn != nil {
		fn(m.Devices())
	}
}

func (m *Manager) Upsert(device Device) {
	if device.ID == "" {
		device.ID = fmt.Sprintf("%s-%s", device.Name, device.Address)
	}
	if device.LastSeen.IsZero() && device.Status == DeviceStatusOnline {
		device.LastSeen = time.Now()
	}
	m.mu.Lock()
	m.devices[device.ID] = device
	m.mu.Unlock()
	m.emit()
}

func (m *Manager) SetStatus(id string, status DeviceStatus) {
	m.mu.Lock()
	if device, ok := m.devices[id]; ok {
		device.Status = status
		if status == DeviceStatusOnline {
			device.LastSeen = time.Now()
		}
		m.devices[id] = device
	}
	m.mu.Unlock()
	m.emit()
}

func (m *Manager) SetNickname(id, nickname string) {
	m.mu.Lock()
	if device, ok := m.devices[id]; ok {
		// Two statements, not one tuple: in `device.Nickname, m.devices[id] =
		// nickname, device` the right-hand `device` is evaluated before the
		// nickname assignment, so the map would keep the unrenamed copy.
		device.Nickname = nickname
		m.devices[id] = device
	}
	m.mu.Unlock()
	m.emit()
}

func (m *Manager) SetTransportURL(id, endpoint string) {
	m.mu.Lock()
	if device, ok := m.devices[id]; ok {
		device.TransportURL = endpoint
		m.devices[id] = device
	}
	m.mu.Unlock()
	m.emit()
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	delete(m.devices, id)
	m.mu.Unlock()
	m.emit()
}

func (m *Manager) EnsureDefault() {
	if len(m.Devices()) != 0 {
		return
	}
	m.Upsert(Device{ID: "local", Name: "local", Nickname: "Local Device", ConnectionType: "local", Address: "127.0.0.1", Status: DeviceStatusOnline})
}

// Probe is a network endpoint to test during discovery.
type Probe struct{ ID, Name, Address, ConnectionType string }

type ProbeFunc func(Probe) bool

func Discover(probes []Probe, probe ProbeFunc) []Device {
	devices := make([]Device, 0, len(probes))
	for _, candidate := range probes {
		if probe == nil || !probe(candidate) {
			continue
		}
		devices = append(devices, Device{ID: candidate.ID, Name: candidate.Name, Address: candidate.Address, ConnectionType: candidate.ConnectionType, Status: DeviceStatusOnline, LastSeen: time.Now()})
	}
	return devices
}

func TCPProbe(timeout time.Duration) ProbeFunc {
	if timeout <= 0 {
		timeout = 250 * time.Millisecond
	}
	return func(candidate Probe) bool {
		conn, err := net.DialTimeout("tcp", candidate.Address, timeout) //nolint:noctx // the probe API is synchronous and carries its own timeout; there is no caller context to thread
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}
}

var defaultManager = NewManager()

func Devices() []Device                              { return defaultManager.Devices() }
func UpsertDevice(device Device)                     { defaultManager.Upsert(device) }
func SetDeviceStatus(id string, status DeviceStatus) { defaultManager.SetStatus(id, status) }
func SetDeviceNickname(id, nickname string)          { defaultManager.SetNickname(id, nickname) }
func SetDeviceTransportURL(id, endpoint string)      { defaultManager.SetTransportURL(id, endpoint) }
func RemoveDevice(id string)                         { defaultManager.Remove(id) }
func EnsureDefaultDevice()                           { defaultManager.EnsureDefault() }

// DiscoverDevices probes NODEs already saved in this process. It used to
// scan a fixed list of local dev-server ports, which never spoke the NODE
// protocol. Refresh in the settings dialog calls ProbeSavedNodes directly
// so it can time out; this wrapper remains for older callers.
func DiscoverDevices() {
	ProbeSavedNodes(context.Background())
}

// DefaultProbes is exposed for callers that want to run discovery on their
// own scheduler without using the process-wide registry.
func DefaultProbes() []Probe {
	probes := []Probe{}
	for _, device := range Devices() {
		probes = append(probes, Probe{ID: device.ID, Name: device.Name, Address: device.Address, ConnectionType: device.ConnectionType})
	}
	return slices.Clone(probes)
}
