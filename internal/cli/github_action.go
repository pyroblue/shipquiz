package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shipquiz/shipquiz/internal/action"
	"github.com/shipquiz/shipquiz/internal/config"
	"github.com/shipquiz/shipquiz/internal/gitx"
	"github.com/shipquiz/shipquiz/internal/keychain"
	"github.com/shipquiz/shipquiz/internal/llm"
)

const llmSecret = "SHIPQUIZ_LLM_API_KEY"

var errBadProvider = errors.New("use 1, 2, 3, or 4")

func cmdGitHubAction(args []string) int {
	fs := flag.NewFlagSet("github-action", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	printOnly := fs.Bool("print", false, "print the workflow without asking")
	_ = fs.Bool("write", false, "ask for the model and key, then write the workflow")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		return fail(err)
	}
	if *printOnly {
		fmt.Print(action.Workflow(cfg.Provider, cfg.Model, cfg.BaseURL))
		return 0
	}

	con := openConsole()
	defer con.close()

	provider, err := askProvider(con, cfg.Provider)
	if errors.Is(err, errBadProvider) {
		fmt.Fprintln(con.out, "Use 1, 2, 3, or 4.")
		return 1
	}
	if err != nil {
		return fail(err)
	}
	baseURL, err := askBaseURL(con, cfg, provider)
	if err != nil {
		return fail(err)
	}
	secret, err := askSecret(con, provider)
	if err != nil {
		return fail(err)
	}
	client, err := llm.New(provider, baseURL, secret)
	if err != nil {
		return fail(err)
	}
	current := config.DefaultModel(provider)
	if provider == cfg.Provider && cfg.Model != "" {
		current = cfg.Model
	}
	model, err := pickModel(con, client, provider, current)
	if err != nil {
		return fail(err)
	}
	smoke, err := con.confirm("Send a one-token smoke test?", true)
	if err != nil {
		return fail(err)
	}
	if smoke {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		_, smokeErr := client.Complete(ctx, llm.Request{
			Model:     model,
			System:    "Reply with the single word ok.",
			User:      "ok",
			MaxTokens: 16,
		})
		cancel()
		if smokeErr != nil {
			fmt.Fprintf(con.out, "Smoke test failed: %v\n", smokeErr)
			fmt.Fprintln(con.out, "The workflow will still be written.")
		} else {
			fmt.Fprintln(con.out, "Smoke test ok.")
		}
	}

	cfg.Provider = provider
	cfg.Model = model
	cfg.BaseURL = baseURL
	if err := config.Save(cfg); err != nil {
		return fail(err)
	}
	fmt.Fprintf(con.out, "Saved %s (%s, %s).\n", mustConfigPath(), provider, model)

	wrote := false
	path := ""
	root, rootErr := gitx.Root()
	if rootErr != nil {
		fmt.Fprintln(con.out, "\nWorkflow")
		fmt.Fprint(con.out, action.Workflow(provider, model, baseURL))
	} else {
		var err error
		path, wrote, err = writeWorkflow(con, root, provider, model, baseURL)
		if err != nil {
			return fail(err)
		}
	}
	return printSecretSteps(con, path, wrote, secret)
}

func askProvider(con console, current string) (string, error) {
	if current == "" {
		current = "anthropic"
	}
	def := "1"
	switch current {
	case "openai":
		def = "2"
	case "gemini":
		def = "3"
	case "xai":
		def = "4"
	}
	fmt.Fprintln(con.out, "\nProvider")
	fmt.Fprintln(con.out, "  1  anthropic")
	fmt.Fprintln(con.out, "  2  openai")
	fmt.Fprintln(con.out, "  3  gemini")
	fmt.Fprintln(con.out, "  4  xai")
	choice, err := con.prompt("Provider [" + def + "]: ")
	if err != nil {
		return "", err
	}
	if choice == "" {
		choice = def
	}
	switch choice {
	case "1", "anthropic":
		return "anthropic", nil
	case "2", "openai":
		return "openai", nil
	case "3", "gemini":
		return "gemini", nil
	case "4", "xai":
		return "xai", nil
	default:
		return "", errBadProvider
	}
}

func askBaseURL(con console, cfg config.Config, provider string) (string, error) {
	current := ""
	if provider == cfg.Provider {
		current = cfg.BaseURL
	}
	label := "Base URL (blank for the provider default): "
	if current != "" {
		label = "Base URL [" + current + "]: "
	}
	entered, err := con.prompt(label)
	if err != nil {
		return "", err
	}
	if entered == "" {
		return current, nil
	}
	return entered, nil
}

func askSecret(con console, provider string) (string, error) {
	if keychain.Stored(provider) {
		use, err := con.confirm("Use the saved API key for "+provider+"?", true)
		if err != nil {
			return "", err
		}
		if use {
			return keychain.Get(provider)
		}
	}
	secret, err := con.secret("LLM API key (hidden): ")
	if err != nil {
		return "", err
	}
	if secret == "" {
		return "", fmt.Errorf("an API key is required")
	}
	if err := keychain.Set(provider, secret); err != nil {
		return "", err
	}
	return secret, nil
}

func writeWorkflow(con console, root, provider, model, baseURL string) (string, bool, error) {
	path := filepath.Join(root, ".github", "workflows", "shipquiz.yml")
	if _, err := os.Stat(path); err == nil {
		replace, err := con.confirm("Replace "+path+"?", false)
		if err != nil {
			return path, false, err
		}
		if !replace {
			return path, false, nil
		}
	} else if !os.IsNotExist(err) {
		return path, false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return path, false, err
	}
	body := action.Workflow(provider, model, baseURL)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return path, false, err
	}
	fmt.Fprintf(con.out, "Wrote %s\n", path)
	return path, true, nil
}

func printSecretSteps(con console, path string, wrote bool, secret string) int {
	repo := githubSlug(originURL())
	fmt.Fprintln(con.out, "\nRepository secret")
	fmt.Fprintf(con.out, "  %s\n", llmSecret)
	fmt.Fprintln(con.out, "  Value: the LLM API key you just confirmed")
	if wrote {
		fmt.Fprintf(con.out, "\nThe workflow is %s. The key is not in that file.\n", path)
	}
	fmt.Fprintln(con.out, "Require the shipquiz check in branch protection if a failed quiz should block merging.")
	fmt.Fprintln(con.out, "People reply to the quiz comment in their own words.")

	if repo == "" {
		fmt.Fprintf(con.out, "\n  gh secret set %s\n", llmSecret)
		return 0
	}
	setNow, err := con.confirm("Set "+llmSecret+" on "+repo+" now?", true)
	if err != nil {
		return fail(err)
	}
	if !setNow {
		fmt.Fprintf(con.out, "\n  gh secret set %s --repo %s\n", llmSecret, repo)
		return 0
	}
	if err := setGitHubSecret(repo, secret); err != nil {
		fmt.Fprintf(con.out, "Could not set the secret: %v\n", err)
		fmt.Fprintf(con.out, "  gh secret set %s --repo %s\n", llmSecret, repo)
		return 0
	}
	fmt.Fprintf(con.out, "Set %s on %s.\n", llmSecret, repo)
	return 0
}

func setGitHubSecret(repo, value string) error {
	cmd := exec.Command("gh", "secret", "set", llmSecret, "--repo", repo)
	cmd.Stdin = strings.NewReader(value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func originURL() string {
	out, err := gitx.Output("remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// githubSlug returns owner/name from a GitHub remote URL.
func githubSlug(remote string) string {
	s := strings.TrimSpace(remote)
	s = strings.TrimSuffix(s, ".git")
	if i := strings.Index(s, "github.com:"); i >= 0 {
		return s[i+len("github.com:"):]
	}
	if i := strings.Index(s, "github.com/"); i >= 0 {
		return s[i+len("github.com/"):]
	}
	return ""
}
