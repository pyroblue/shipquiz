package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/shipquiz/shipquiz/internal/config"
	"github.com/shipquiz/shipquiz/internal/ui"
	"github.com/shipquiz/shipquiz/internal/version"
)

const releaseRepo = "pyroblue/shipquiz"

const (
	updateInterval = 24 * time.Hour
	updateRetry    = time.Hour
	updateLookup   = 2 * time.Second
)

var errNoRelease = errors.New("no release published")

type githubRelease struct {
	Tag string
	URL string
}

type updateState struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
	Declined  string    `json:"declined"`
	OK        bool      `json:"ok"`
}

var (
	updateNow      = time.Now
	fetchReleaseFn = fetchRelease
)

// offerUpdate asks before replacing this binary when a newer release exists.
// A yes replaces the binary and runs the same command again.
func offerUpdate(command string) {
	if os.Getenv("SHIPQUIZ_JUST_UPDATED") == "1" {
		_ = os.Unsetenv("SHIPQUIZ_JUST_UPDATED")
		return
	}
	con := openConsole()
	defer con.close()
	if con.file == nil || !term.IsTerminal(int(con.file.Fd())) {
		return
	}
	tag, ask := considerUpdate(command)
	if !ask {
		return
	}
	yes, err := con.confirm(fmt.Sprintf("shipquiz %s is available (you have %s). Update?", tag, version.Current), true)
	if err != nil || !yes {
		if err == nil {
			state := loadUpdateState()
			state.Declined = tag
			_ = saveUpdateState(state)
		}
		return
	}
	fmt.Fprintln(con.out, ui.Muted(con.out, "updating…"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	rel, err := fetchRelease(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: update failed: %v\n", err)
		return
	}
	if err := installRelease(ctx, rel); err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: update failed: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "shipquiz: updated to %s\n", rel.Tag)
	if err := rerun(); err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: run the command again to use %s\n", rel.Tag)
	}
}

// considerUpdate reports the newer release to offer, when this command should ask.
func considerUpdate(command string) (string, bool) {
	if !interactiveCommand(command) || skipUpdateCheck() {
		return "", false
	}
	state := loadUpdateState()
	age := updateNow().Sub(state.CheckedAt)
	if !state.CheckedAt.IsZero() && state.OK && age < updateInterval {
		return pendingUpdate(state)
	}
	if !state.CheckedAt.IsZero() && !state.OK && age < updateRetry {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), updateLookup)
	defer cancel()
	rel, err := fetchReleaseFn(ctx)
	if errors.Is(err, errNoRelease) {
		state.CheckedAt = updateNow()
		state.Latest = ""
		state.OK = true
		_ = saveUpdateState(state)
		return "", false
	}
	if err != nil {
		state.CheckedAt = updateNow()
		state.OK = false
		_ = saveUpdateState(state)
		return "", false
	}
	state.CheckedAt = updateNow()
	state.Latest = rel.Tag
	state.OK = true
	_ = saveUpdateState(state)
	return pendingUpdate(state)
}

// interactiveCommand is a command a person runs in a terminal.
func interactiveCommand(name string) bool {
	switch name {
	case "setup", "install", "uninstall", "quiz", "status", "config", "github-action":
		return true
	default:
		return false
	}
}

// skipUpdateCheck is set for CI and for the command that just replaced itself.
func skipUpdateCheck() bool {
	if os.Getenv("SHIPQUIZ_NO_UPDATE") != "" || os.Getenv("CI") != "" {
		return true
	}
	if os.Getenv("SHIPQUIZ_JUST_UPDATED") == "1" {
		_ = os.Unsetenv("SHIPQUIZ_JUST_UPDATED")
		return true
	}
	return false
}

// pendingUpdate is the cached release when it is newer than this binary and was not declined.
func pendingUpdate(state updateState) (string, bool) {
	if !versionNewer(state.Latest, version.Current) || state.Declined == state.Latest {
		return "", false
	}
	return state.Latest, true
}

// versionNewer reports whether latest is a newer dotted version than current.
func versionNewer(latest, current string) bool {
	lMajor, lMinor, lPatch, okL := parseVersion(latest)
	cMajor, cMinor, cPatch, okC := parseVersion(current)
	if !okL || !okC {
		return false
	}
	if lMajor != cMajor {
		return lMajor > cMajor
	}
	if lMinor != cMinor {
		return lMinor > cMinor
	}
	return lPatch > cPatch
}

// parseVersion reads a major.minor.patch tag. A leading v is ignored.
func parseVersion(raw string) (major, minor, patch int, ok bool) {
	raw = strings.TrimPrefix(strings.TrimSpace(raw), "v")
	if raw == "" {
		return 0, 0, 0, false
	}
	parts := strings.Split(raw, ".")
	if len(parts) > 3 {
		return 0, 0, 0, false
	}
	nums := []int{0, 0, 0}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return 0, 0, 0, false
		}
		nums[i] = n
	}
	return nums[0], nums[1], nums[2], true
}

// fetchRelease reads the latest GitHub release for this OS and architecture.
func fetchRelease(ctx context.Context) (githubRelease, error) {
	asset := fmt.Sprintf("shipquiz_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", releaseRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("accept", "application/vnd.github+json")
	req.Header.Set("user-agent", "shipquiz/"+version.Current)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return githubRelease{}, errNoRelease
	}
	if res.StatusCode >= 300 {
		return githubRelease{}, fmt.Errorf("release lookup returned %d", res.StatusCode)
	}
	var payload struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return githubRelease{}, err
	}
	for _, item := range payload.Assets {
		if item.Name == asset {
			return githubRelease{Tag: payload.TagName, URL: item.URL}, nil
		}
	}
	return githubRelease{}, fmt.Errorf("release %s has no %s", payload.TagName, asset)
}

// rerun replaces this process with the binary that was just installed.
func rerun() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if err := os.Setenv("SHIPQUIZ_JUST_UPDATED", "1"); err != nil {
		return err
	}
	return syscall.Exec(exe, append([]string{exe}, os.Args[1:]...), os.Environ())
}

func updateStatePath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update.json"), nil
}

func loadUpdateState() updateState {
	path, err := updateStatePath()
	if err != nil {
		return updateState{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return updateState{}
	}
	var state updateState
	if err := json.Unmarshal(data, &state); err != nil {
		return updateState{}
	}
	return state
}

func saveUpdateState(state updateState) error {
	path, err := updateStatePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
