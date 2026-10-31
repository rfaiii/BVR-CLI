package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const credentialsFile = "node-credentials.json"

type credentialFile struct {
	Tokens map[string]string `json:"tokens"`
}

var (
	secretMu sync.Mutex
	secrets  = map[string]string{}
)

// DefaultCredentialsPath is the token file next to nodes.json. Tokens
// are not written into the device registry.
func DefaultCredentialsPath() string {
	return filepath.Join(filepath.Dir(DefaultDevicesPath()), credentialsFile)
}

// SaveToken records the bearer token for a device id.
func SaveToken(path, id, token string) error {
	if id == "" {
		return errors.New("missing device id")
	}
	current, err := loadTokenFile(path)
	if err != nil {
		return err
	}
	if current.Tokens == nil {
		current.Tokens = map[string]string{}
	}
	current.Tokens[id] = token
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	if err := writePrivateFile(path, data); err != nil {
		return fmt.Errorf("write node credentials: %w", err)
	}
	rememberToken(id, token)
	return nil
}

// LoadTokens reads the credentials file. A missing file is empty.
func LoadTokens(path string) (map[string]string, error) {
	file, err := loadTokenFile(path)
	if err != nil {
		return nil, err
	}
	if file.Tokens == nil {
		return map[string]string{}, nil
	}
	return file.Tokens, nil
}

func loadTokenFile(path string) (credentialFile, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return credentialFile{Tokens: map[string]string{}}, nil
	}
	if err != nil {
		return credentialFile{}, fmt.Errorf("read node credentials: %w", err)
	}
	var file credentialFile
	if err := json.Unmarshal(data, &file); err != nil {
		return credentialFile{}, fmt.Errorf("unmarshal node credentials: %w", err)
	}
	return file, nil
}

func rememberToken(id, token string) {
	secretMu.Lock()
	secrets[id] = token
	secretMu.Unlock()
}

// TokenFor returns the token remembered in this process.
func TokenFor(id string) (string, bool) {
	secretMu.Lock()
	defer secretMu.Unlock()
	token, ok := secrets[id]
	return token, ok
}

// AttachSavedTransports loads credentials and registers a transport
// for every saved device that has one. Call it when a process starts
// so a previous `node connect` is live in the TUI and the files CLI.
func AttachSavedTransports() error {
	tokens, err := LoadTokens(DefaultCredentialsPath())
	if err != nil {
		return err
	}
	secretMu.Lock()
	for id, token := range tokens {
		secrets[id] = token
	}
	secretMu.Unlock()
	for _, device := range Devices() {
		token, ok := tokens[device.ID]
		if !ok || device.TransportURL == "" {
			continue
		}
		if err := RegisterNodeTransport(device.ID, device.TransportURL, token); err != nil {
			return err
		}
	}
	return nil
}
