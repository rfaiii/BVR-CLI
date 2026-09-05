package license

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richavery/bvr-cli/internal/config"
)

const AuthFilename = "auth.json"

type AuthData struct {
	LicenseKey string `json:"license_key"`
}

// SaveLicense saves the license key to the global auth.json file.
func SaveLicense(key string) error {
	configDir := config.GlobalConfig()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	authPath := filepath.Join(configDir, AuthFilename)
	data := AuthData{LicenseKey: key}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	if err := os.WriteFile(authPath, b, 0o600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// LoadLicense reads the license key from the global auth.json file.
func LoadLicense() string {
	configDir := config.GlobalConfig()
	authPath := filepath.Join(configDir, AuthFilename)

	b, err := os.ReadFile(authPath)
	if err != nil {
		return ""
	}

	var data AuthData
	if err := json.Unmarshal(b, &data); err != nil {
		return ""
	}

	return data.LicenseKey
}

// ValidateLicense checks the license key against the licensing API.
func ValidateLicense(key string) error {
	// TODO: Integrate actual API calls to Lemon Squeezy/Paddle/Keygen.
	// For now, simple format validation logic.
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("license key cannot be empty")
	}
	if !strings.HasPrefix(strings.ToUpper(key), "BVR-") {
		return fmt.Errorf("invalid license key format, should start with BVR-")
	}

	// Mocking successful network validation
	return nil
}
