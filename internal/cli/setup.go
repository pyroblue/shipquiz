package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shipquiz/shipquiz/internal/config"
	"github.com/shipquiz/shipquiz/internal/gitx"
	"github.com/shipquiz/shipquiz/internal/hook"
	"github.com/shipquiz/shipquiz/internal/keychain"
	"github.com/shipquiz/shipquiz/internal/llm"
)

func cmdSetup() int {
	con := openConsole()
	defer con.close()

	cfg, err := config.Load()
	if err != nil {
		return fail(err)
	}
	if _, statErr := os.Stat(mustConfigPath()); statErr == nil {
		again, err := con.confirm("shipquiz is already configured. Reconfigure?", false)
		if err != nil {
			return fail(err)
		}
		if !again {
			return offerHook(con)
		}
	}

	fmt.Fprintln(con.out, "\nProvider")
	fmt.Fprintln(con.out, "  1  anthropic")
	fmt.Fprintln(con.out, "  2  openai")
	fmt.Fprintln(con.out, "  3  gemini")
	choice, err := con.prompt("Provider [1]: ")
	if err != nil {
		return fail(err)
	}
	switch choice {
	case "", "1", "anthropic":
		cfg.Provider = "anthropic"
	case "2", "openai":
		cfg.Provider = "openai"
	case "3", "gemini":
		cfg.Provider = "gemini"
	default:
		fmt.Fprintln(con.out, "Use 1, 2, or 3.")
		return 1
	}
	cfg.Model = config.DefaultModel(cfg.Provider)

	secret, err := con.secret("API key (hidden): ")
	if err != nil {
		return fail(err)
	}
	if secret == "" {
		fmt.Fprintln(con.out, "An API key is required.")
		return 1
	}
	if err := keychain.Set(cfg.Provider, secret); err != nil {
		return fail(err)
	}

	base, err := con.prompt("Base URL (blank for the provider default): ")
	if err != nil {
		return fail(err)
	}
	cfg.BaseURL = base

	client, err := llm.New(cfg.Provider, cfg.BaseURL, secret)
	if err != nil {
		return fail(err)
	}
	model, err := pickModel(con, client, cfg.Model)
	if err != nil {
		return fail(err)
	}
	cfg.Model = model

	smoke, err := con.confirm("Send a one-token smoke test?", true)
	if err != nil {
		return fail(err)
	}
	if smoke {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		_, smokeErr := client.Complete(ctx, llm.Request{
			Model:     cfg.Model,
			System:    "Reply with the single word ok.",
			User:      "ok",
			MaxTokens: 16,
		})
		cancel()
		if smokeErr != nil {
			fmt.Fprintf(con.out, "Smoke test failed: %v\n", smokeErr)
			fmt.Fprintln(con.out, "The key is stored. Fix the model or key and run shipquiz setup again.")
		} else {
			fmt.Fprintln(con.out, "Smoke test ok.")
		}
	}

	if err := config.Save(cfg); err != nil {
		return fail(err)
	}
	fmt.Fprintf(con.out, "Saved %s (%s).\n", mustConfigPath(), cfg.Model)
	return offerHook(con)
}

func pickModel(con console, client llm.Client, current string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	models, err := client.ListModels(ctx)
	cancel()
	if err != nil {
		fmt.Fprintf(con.out, "Could not list models: %v\n", err)
		entered, perr := con.prompt(fmt.Sprintf("Model [%s]: ", current))
		if perr != nil {
			return "", perr
		}
		if entered == "" {
			return current, nil
		}
		return entered, nil
	}
	if len(models) > 15 {
		models = models[:15]
	}
	fmt.Fprintln(con.out, "\nModels")
	selected := 1
	for i, model := range models {
		fmt.Fprintf(con.out, "  %d  %s\n", i+1, model)
		if model == current {
			selected = i + 1
		}
	}
	entered, err := con.prompt(fmt.Sprintf("Model [%d]: ", selected))
	if err != nil {
		return "", err
	}
	if entered == "" {
		return models[selected-1], nil
	}
	if n, nerr := strconv.Atoi(entered); nerr == nil && n >= 1 && n <= len(models) {
		return models[n-1], nil
	}
	return entered, nil
}

func offerHook(con console) int {
	if !gitx.Inside() {
		printNext(con)
		return 0
	}
	if hook.Installed() {
		fmt.Fprintln(con.out, "Pre-push hook is already installed.")
		printNext(con)
		return 0
	}
	yes, err := con.confirm("Install the pre-push hook in this repo?", true)
	if err != nil {
		return fail(err)
	}
	if yes {
		if err := hook.Install(); err != nil {
			return fail(err)
		}
		fmt.Fprintln(con.out, "Pre-push hook installed.")
	}
	printNext(con)
	return 0
}

func printNext(con console) {
	fmt.Fprintln(con.out, `
Ready.
  shipquiz quiz     quiz the current change
  shipquiz status   show config, hook, and key
  git push          runs the quiz when the hook is installed

Skip one push with SHIPQUIZ_SKIP=1 git push, or git push --no-verify.
If this shell cannot find shipquiz yet:
  export PATH="$HOME/.shipquiz/bin:$PATH"`)
}

func mustConfigPath() string {
	path, err := config.Path()
	if err != nil {
		return strings.TrimSpace(path)
	}
	return path
}
