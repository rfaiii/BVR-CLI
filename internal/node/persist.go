package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/richavery/bvr-cli/internal/config"
)

// devicesFile is the persisted view of the registry, written next to the
// global data file the provider cache uses.
const devicesFile = "nodes.json"

// DefaultDevicesPath is where the registry is saved between runs.
func DefaultDevicesPath() string {
	return filepath.Join(filepath.Dir(config.GlobalConfigData()), devicesFile)
}

// DeviceFromNode turns a node found on the local network into a registry
// entry. The ID published in the TXT record is the identity, so a host that
// returns under the same ID keeps the nickname the user gave it.
func DeviceFromNode(found Node) Device {
	return Device{
		ID:             firstNonEmpty(found.ID, found.Instance, found.HostPort()),
		Name:           found.DisplayName(),
		ConnectionType: "node",
		Address:        found.HostPort(),
		Status:         DeviceStatusOnline,
		LastSeen:       time.Now(),
		TransportURL:   found.WebSocketURL(),
	}
}

// RecordNodes adds every scanned node to the registry and returns the ones
// that carry a dial target. A node with no address is a browse answer with
// nothing to connect to, so recording it would only add a dead row.
func (m *Manager) RecordNodes(nodes []Node) []Device {
	recorded := make([]Device, 0, len(nodes))
	for _, found := range nodes {
		device := DeviceFromNode(found)
		if device.ID == "" || device.Address == "" {
			continue
		}
		// A scan rewrites the whole record, so carry over what the user named
		// this host; without this every scan resets a hand-picked nickname.
		device.Nickname = m.nicknameOf(device.ID)
		m.Upsert(device)
		recorded = append(recorded, device)
	}
	return recorded
}

func (m *Manager) nicknameOf(id string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.devices[id].Nickname
}

// Save writes the registry to path as JSON, replacing path atomically so an
// interrupted write cannot leave a half-file behind.
func (m *Manager) Save(path string) error {
	data, err := json.Marshal(m.Devices())
	if err != nil {
		return fmt.Errorf("marshal device registry: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create device registry directory: %w", err)
	}

	return writePrivateFile(path, data)
}

// writePrivateFile replaces path atomically. The mode is 0o600 because
// these files name LAN hosts or hold bearer tokens.
func writePrivateFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}

// Restore merges the registry saved at path into memory and reports how many
// entries were added. A saved record describes a host that has not been seen
// since this process started, so anything already registered wins and every
// restored entry is marked offline until a scan or probe confirms it. A missing
// file is not an error: it just means nothing has been discovered yet.
func (m *Manager) Restore(path string) (int, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read device registry: %w", err)
	}

	var saved []Device
	if err := json.Unmarshal(data, &saved); err != nil {
		return 0, fmt.Errorf("unmarshal device registry %s: %w", path, err)
	}

	added := 0
	m.mu.Lock()
	for _, device := range saved {
		if device.ID == "" {
			continue
		}
		if _, known := m.devices[device.ID]; known {
			continue
		}
		device.Status = DeviceStatusOffline
		m.devices[device.ID] = device
		added++
	}
	m.mu.Unlock()

	if added > 0 {
		m.emit()
	}
	return added, nil
}

// RecordScannedNodes registers the results of a scan and persists them, so the
// next run knows which hosts were reachable last time.
func RecordScannedNodes(nodes []Node) ([]Device, error) {
	recorded := defaultManager.RecordNodes(nodes)
	if len(recorded) == 0 {
		return recorded, nil
	}
	return recorded, defaultManager.Save(DefaultDevicesPath())
}

// LoadDeviceRegistry restores the persisted registry into the process-wide
// manager.
func LoadDeviceRegistry() (int, error) {
	return defaultManager.Restore(DefaultDevicesPath())
}
