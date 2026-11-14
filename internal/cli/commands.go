package cli

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/shipquiz/shipquiz/internal/config"
	"github.com/shipquiz/shipquiz/internal/gitx"
	"github.com/shipquiz/shipquiz/internal/hook"
	"github.com/shipquiz/shipquiz/internal/keychain"
	"github.com/shipquiz/shipquiz/internal/llm"
	"github.com/shipquiz/shipquiz/internal/quiz"
	"github.com/shipquiz/shipquiz/internal/version"
)

const releaseRepo = "pyroblue/shipquiz"

func cmdInstall() int {
	if !gitx.Inside() {
		fmt.Fprintln(os.Stderr, "shipquiz: run this inside a git repository")
		return 1
	}
	if err := hook.Install(); err != nil {
		return fail(err)
	}
	fmt.Println("shipquiz: pre-push hook installed")
	fmt.Println("Skip once with SHIPQUIZ_SKIP=1 git push, or git push --no-verify.")
	return 0
}

func cmdUninstall() int {
	if !gitx.Inside() {
		fmt.Fprintln(os.Stderr, "shipquiz: run this inside a git repository")
		return 1
	}
	if err := hook.Uninstall(); err != nil {
		return fail(err)
	}
	fmt.Println("shipquiz: pre-push hook removed")
	return 0
}

func cmdStatus() int {
	root, _ := gitx.Root()
	cfg, err := config.LoadForRepo(root)
	if err != nil {
		return fail(err)
	}
	path, _ := config.Path()
	fmt.Printf("version:  %s\n", version.Current)
	fmt.Printf("provider: %s\n", cfg.Provider)
	fmt.Printf("model:    %s\n", cfg.Model)
	if cfg.BaseURL != "" {
		fmt.Printf("base url: %s\n", cfg.BaseURL)
	}
	fmt.Printf("config:   %s\n", path)
	fmt.Printf("timeout:  %ds\n", cfg.Quiz.TimeoutSeconds)
	fmt.Printf("pass:     average >= %.2f and no zeros\n", cfg.Quiz.PassingAverage)
	if root != "" {
		fmt.Printf("repo:     %s\n", root)
		if hook.Installed() {
			fmt.Println("hook:     installed")
		} else {
			fmt.Println("hook:     not installed")
		}
	} else {
		fmt.Println("repo:     (not a git repository)")
	}
	fmt.Print("api key:  ")
	apiKey, err := keychain.Get(cfg.Provider)
	if err != nil || apiKey == "" {
		fmt.Println("missing")
		return 0
	}
	client, err := llm.New(cfg.Provider, cfg.BaseURL, apiKey)
	if err != nil {
		fmt.Printf("present (%v)\n", err)
		return 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	_, err = client.ListModels(ctx)
	cancel()
	if err != nil {
		fmt.Printf("present, provider rejected it (%v)\n", err)
		return 0
	}
	fmt.Println("valid")
	return 0
}

func cmdConfig(args []string) int {
	if len(args) == 0 {
		root, _ := gitx.Root()
		cfg, err := config.LoadForRepo(root)
		if err != nil {
			return fail(err)
		}
		if err := toml.NewEncoder(os.Stdout).Encode(cfg); err != nil {
			return fail(err)
		}
		return 0
	}
	if len(args) == 3 && args[0] == "set" {
		cfg, err := config.Load()
		if err != nil {
			return fail(err)
		}
		if err := config.Set(&cfg, args[1], args[2]); err != nil {
			return fail(err)
		}
		if err := config.Save(cfg); err != nil {
			return fail(err)
		}
		fmt.Printf("shipquiz: set %s\n", args[1])
		return 0
	}
	fmt.Fprintln(os.Stderr, "usage: shipquiz config [set <key> <value>]")
	return 1
}

func cmdUpdate() int {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	asset := fmt.Sprintf("shipquiz_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", releaseRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fail(err)
	}
	req.Header.Set("accept", "application/vnd.github+json")
	req.Header.Set("user-agent", "shipquiz/"+version.Current)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fail(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		fmt.Fprintln(os.Stderr, "shipquiz: no release published yet. Build from source:")
		fmt.Fprintln(os.Stderr, "  go build -o ~/.shipquiz/bin/shipquiz ./cmd/shipquiz")
		return 1
	}
	if res.StatusCode >= 300 {
		return fail(fmt.Errorf("release lookup returned %d", res.StatusCode))
	}
	var payload struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return fail(err)
	}
	download := ""
	for _, item := range payload.Assets {
		if item.Name == asset {
			download = item.URL
			break
		}
	}
	if download == "" {
		return fail(fmt.Errorf("release %s has no %s", payload.TagName, asset))
	}
	bin, err := downloadBinary(ctx, download)
	if err != nil {
		return fail(err)
	}
	if err := replaceExecutable(bin); err != nil {
		return fail(err)
	}
	fmt.Printf("shipquiz: updated to %s\n", payload.TagName)
	return 0
}

func downloadBinary(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", "shipquiz/"+version.Current)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("download returned %d", res.StatusCode)
	}
	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if filepath.Base(hdr.Name) != "shipquiz" || hdr.Typeflag != tar.TypeReg {
			continue
		}
		return io.ReadAll(io.LimitReader(tr, 100<<20))
	}
	return nil, fmt.Errorf("archive did not contain shipquiz")
}

func replaceExecutable(data []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	tmp := exe + ".new"
	if err := os.WriteFile(tmp, data, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func cmdQuiz(args []string) int {
	fs := flag.NewFlagSet("quiz", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	cached := fs.Bool("cached", false, "use cached questions only")
	again := fs.Bool("again", false, "ask again even if this diff already passed")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if !gitx.Inside() {
		fmt.Fprintln(os.Stderr, "shipquiz: run this inside a git repository")
		return 1
	}
	fmt.Fprintln(os.Stderr, "shipquiz: analyzing diff…")
	diff, err := diffForQuiz()
	if err != nil {
		return fail(err)
	}
	return runQuiz(diff, false, *cached, *again)
}

func cmdHookRun() int {
	if !gitx.Inside() {
		fmt.Fprintln(os.Stderr, "shipquiz: run this inside a git repository")
		return 1
	}
	fmt.Fprintln(os.Stderr, "shipquiz: analyzing diff…")
	var diffErr error
	var diff quiz.Diff
	if isTerminal(os.Stdin) {
		diff, diffErr = diffForQuiz()
	} else {
		raw := readAll(os.Stdin)
		refs := parsePush(raw)
		if len(strings.TrimSpace(raw)) == 0 || len(refs) == 0 {
			fmt.Fprintln(os.Stderr, "shipquiz: nothing to push")
			return 0
		}
		diff, diffErr = pushDiff(raw)
	}
	if diffErr != nil {
		return fail(diffErr)
	}
	return runQuiz(diff, true, false, false)
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
