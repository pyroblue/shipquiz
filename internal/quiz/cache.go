package quiz

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry is a cached quiz. Passed is set after a successful grade so a
// retried push of the same diff does not ask again.
type Entry struct {
	Generated Generated `json:"generated"`
	Passed    bool      `json:"passed"`
	Score     float64   `json:"score"`
	At        time.Time `json:"at"`
}

// Key identifies a generated quiz for this diff, model, and style.
func Key(base, head, diff, model, style string) string {
	sum := sha256.Sum256([]byte(base + "\n" + head + "\n" + diff + "\n" + model + "\n" + style))
	return hex.EncodeToString(sum[:])
}

// LoadEntry reads a cache entry. The boolean is false when the key is absent.
func LoadEntry(key string) (Entry, bool, error) {
	path, err := cachePath(key)
	if err != nil {
		return Entry{}, false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, false, nil
		}
		return Entry{}, false, err
	}
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return Entry{}, false, fmt.Errorf("read cache: %w", err)
	}
	return entry, true, nil
}

// StoreEntry writes a cache entry.
func StoreEntry(key string, entry Entry) error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, key+".json")
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func cacheDir() (string, error) {
	if dir := os.Getenv("SHIPQUIZ_CACHE_DIR"); dir != "" {
		return dir, nil
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "shipquiz"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "shipquiz"), nil
}

func cachePath(key string) (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, key+".json"), nil
}
