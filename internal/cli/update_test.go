package cli

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestVersionNewer(t *testing.T) {
	cases := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v0.1.10", "0.1.9", true},
		{"v0.2.0", "0.1.9", true},
		{"v0.1.0", "0.1.0", false},
		{"0.1.0", "v0.1.1", false},
		{"v1", "0.9.9", true},
		{"nope", "0.1.0", false},
		{"", "0.1.0", false},
	}
	for _, tc := range cases {
		if got := versionNewer(tc.latest, tc.current); got != tc.want {
			t.Fatalf("versionNewer(%q, %q) = %v", tc.latest, tc.current, got)
		}
	}
}

func TestConsiderUpdateUsesFreshCache(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_NO_UPDATE", "")
	t.Setenv("CI", "")
	t.Setenv("SHIPQUIZ_JUST_UPDATED", "")
	save := fetchReleaseFn
	t.Cleanup(func() { fetchReleaseFn = save })
	fetchReleaseFn = func(context.Context) (githubRelease, error) {
		t.Fatal("fetched despite a fresh cache")
		return githubRelease{}, nil
	}
	if err := saveUpdateState(updateState{CheckedAt: time.Now(), Latest: "v9.0.0", OK: true}); err != nil {
		t.Fatal(err)
	}
	tag, ask := considerUpdate("quiz")
	if !ask || tag != "v9.0.0" {
		t.Fatalf("tag=%s ask=%v", tag, ask)
	}
}

func TestConsiderUpdateSkipsDeclinedRelease(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_NO_UPDATE", "")
	t.Setenv("CI", "")
	if err := saveUpdateState(updateState{CheckedAt: time.Now(), Latest: "v9.0.0", Declined: "v9.0.0", OK: true}); err != nil {
		t.Fatal(err)
	}
	if _, ask := considerUpdate("quiz"); ask {
		t.Fatal("asked about a declined release")
	}
}

func TestConsiderUpdateRetriesAfterFailedLookup(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_NO_UPDATE", "")
	t.Setenv("CI", "")
	save := fetchReleaseFn
	t.Cleanup(func() { fetchReleaseFn = save })
	fetchReleaseFn = func(context.Context) (githubRelease, error) {
		t.Fatal("retried inside the backoff")
		return githubRelease{}, nil
	}
	if err := saveUpdateState(updateState{CheckedAt: time.Now(), OK: false}); err != nil {
		t.Fatal(err)
	}
	if _, ask := considerUpdate("quiz"); ask {
		t.Fatal("asked after a failed lookup")
	}
}

func TestConsiderUpdateFetchesWhenCacheIsStale(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_NO_UPDATE", "")
	t.Setenv("CI", "")
	save := fetchReleaseFn
	t.Cleanup(func() { fetchReleaseFn = save })
	fetchReleaseFn = func(context.Context) (githubRelease, error) {
		return githubRelease{Tag: "v1.2.3"}, nil
	}
	if err := saveUpdateState(updateState{CheckedAt: time.Now().Add(-48 * time.Hour), Latest: "v0.0.1", OK: true}); err != nil {
		t.Fatal(err)
	}
	tag, ask := considerUpdate("quiz")
	if !ask || tag != "v1.2.3" {
		t.Fatalf("tag=%s ask=%v", tag, ask)
	}
}

func TestConsiderUpdateRecordsMissingRelease(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_NO_UPDATE", "")
	t.Setenv("CI", "")
	save := fetchReleaseFn
	t.Cleanup(func() { fetchReleaseFn = save })
	fetchReleaseFn = func(context.Context) (githubRelease, error) {
		return githubRelease{}, errNoRelease
	}
	if _, ask := considerUpdate("status"); ask {
		t.Fatal("asked when no release exists")
	}
	state := loadUpdateState()
	if !state.OK || state.Latest != "" {
		t.Fatalf("state = %#v", state)
	}
}

func TestConsiderUpdateSkipsQuietCommands(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_NO_UPDATE", "")
	t.Setenv("CI", "")
	t.Setenv("SHIPQUIZ_JUST_UPDATED", "")
	if err := saveUpdateState(updateState{CheckedAt: time.Now(), Latest: "v9.0.0", OK: true}); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"hook-run", "action-run", "version", "update"} {
		if _, ask := considerUpdate(command); ask {
			t.Fatalf("asked during %s", command)
		}
	}
	t.Setenv("CI", "true")
	if _, ask := considerUpdate("quiz"); ask {
		t.Fatal("asked in CI")
	}
	t.Setenv("CI", "")
	t.Setenv("SHIPQUIZ_NO_UPDATE", "1")
	if _, ask := considerUpdate("quiz"); ask {
		t.Fatal("asked when updates are disabled")
	}
}

func TestSkipUpdateCheckClearsRerunMarker(t *testing.T) {
	t.Setenv("SHIPQUIZ_NO_UPDATE", "")
	t.Setenv("CI", "")
	t.Setenv("SHIPQUIZ_JUST_UPDATED", "1")
	if !skipUpdateCheck() {
		t.Fatal("expected the rerun marker to skip the check")
	}
	if os.Getenv("SHIPQUIZ_JUST_UPDATED") != "" {
		t.Fatal("rerun marker still set")
	}
}
