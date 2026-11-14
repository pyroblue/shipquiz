// Package keychain stores provider API keys in the macOS Keychain,
// with a mode-0600 file fallback for Linux and for SHIPQUIZ_KEY_STORE=file.
package keychain

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shipquiz/shipquiz/internal/config"
)

// Set stores the API key for provider.
func Set(provider, secret string) error {
	provider = strings.TrimSpace(provider)
	secret = strings.TrimSpace(secret)
	if provider == "" || secret == "" {
		return fmt.Errorf("provider and API key are required")
	}
	if backend() == "file" {
		return setFile(provider, secret)
	}
	user := account()
	cmd := exec.Command("security", "add-generic-password",
		"-U",
		"-s", service(provider),
		"-a", user,
		"-w", secret,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychain save failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// Get returns the API key. SHIPQUIZ_API_KEY wins so CI can supply a secret
// without writing it to disk.
func Get(provider string) (string, error) {
	if key := strings.TrimSpace(os.Getenv("SHIPQUIZ_API_KEY")); key != "" {
		return key, nil
	}
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return "", fmt.Errorf("provider is required")
	}
	if backend() == "file" {
		return getFile(provider)
	}
	key, err := getKeychain(provider)
	if err == nil && key != "" {
		return key, nil
	}
	if fileKey, ferr := getFile(provider); ferr == nil && fileKey != "" {
		return fileKey, nil
	}
	if err != nil {
		return "", fmt.Errorf("no API key for %s in the keychain", provider)
	}
	return "", fmt.Errorf("no API key for %s", provider)
}

// Stored reports whether a key is available without returning the secret.
func Stored(provider string) bool {
	key, err := Get(provider)
	return err == nil && key != ""
}

func service(provider string) string {
	return "shipquiz:" + provider
}

func account() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	return "shipquiz"
}

func backend() string {
	switch os.Getenv("SHIPQUIZ_KEY_STORE") {
	case "file":
		return "file"
	case "keychain":
		return "keychain"
	}
	if runtime.GOOS == "darwin" {
		return "keychain"
	}
	return "file"
}

func getKeychain(provider string) (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", service(provider), "-a", account(), "-w")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func secretsPath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "secrets.json"), nil
}

func loadSecrets() (map[string]string, error) {
	path, err := secretsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]string{}, nil
	}
	var secrets map[string]string
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, fmt.Errorf("read secrets: %w", err)
	}
	if secrets == nil {
		secrets = map[string]string{}
	}
	return secrets, nil
}

func getFile(provider string) (string, error) {
	secrets, err := loadSecrets()
	if err != nil {
		return "", err
	}
	key := strings.TrimSpace(secrets[provider])
	if key == "" {
		return "", fmt.Errorf("no API key for %s", provider)
	}
	return key, nil
}

func setFile(provider, secret string) error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	secrets, err := loadSecrets()
	if err != nil {
		return err
	}
	secrets[provider] = secret
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path, err := secretsPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
