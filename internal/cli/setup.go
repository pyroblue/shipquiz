package cli

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shipquiz/shipquiz/internal/config"
	"github.com/shipquiz/shipquiz/internal/gitx"
	"github.com/shipquiz/shipquiz/internal/hook"
	"github.com/shipquiz/shipquiz/internal/keychain"
	"github.com/shipquiz/shipquiz/internal/llm"
	"github.com/shipquiz/shipquiz/internal/ui"
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

	ui.Hero(con.out)
	ui.Section(con.out, "Provider")
	ui.Choice(con.out, 1, "anthropic")
	ui.Choice(con.out, 2, "openai")
	ui.Choice(con.out, 3, "gemini")
	ui.Choice(con.out, 4, "xai")
	choice, err := con.prompt(ui.Ask(con.out, "Provider", "1"))
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
	case "4", "xai":
		cfg.Provider = "xai"
	default:
		fmt.Fprintln(con.out, "Use 1, 2, 3, or 4.")
		return 1
	}
	cfg.Model = config.DefaultModel(cfg.Provider)

	secret, err := con.secret("API key (hidden): ")
	if err != nil {
		return fail(err)
	}
	if secret == "" {
		ui.Hint(con.out, "An API key is required.")
		return 1
	}
	if err := keychain.Set(cfg.Provider, secret); err != nil {
		return fail(err)
	}

	base, err := con.prompt(ui.Ask(con.out, "Base URL", "blank for the provider default"))
	if err != nil {
		return fail(err)
	}
	cfg.BaseURL = base

	client, err := llm.New(cfg.Provider, cfg.BaseURL, secret)
	if err != nil {
		return fail(err)
	}
	model, err := pickModel(con, client, cfg.Provider, cfg.Model)
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
			fmt.Fprintln(con.out, ui.Fail(con.out, "Smoke test failed: ")+smokeErr.Error())
			ui.Hint(con.out, "The key is stored. Fix the model or key and run shipquiz setup again.")
		} else {
			fmt.Fprintln(con.out, ui.Pass(con.out, "Smoke test ok."))
		}
	}

	if err := config.Save(cfg); err != nil {
		return fail(err)
	}
	fmt.Fprintf(con.out, "%s  %s\n", ui.Pass(con.out, "Saved"), ui.Muted(con.out, mustConfigPath()+" ("+cfg.Model+")"))
	return offerHook(con)
}

func pickModel(con console, client llm.Client, provider, current string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	live, err := client.ListModels(ctx)
	cancel()
	if err != nil {
		fmt.Fprintln(con.out, ui.Fail(con.out, "Could not list models: ")+err.Error())
		live = nil
	}
	models := modelChoices(provider, live, current)
	if len(models) == 0 {
		entered, perr := con.prompt("Model: ")
		if perr != nil {
			return "", perr
		}
		return entered, nil
	}
	ui.Section(con.out, "Models")
	for i, model := range models {
		ui.Choice(con.out, i+1, model)
	}
	ui.Hint(con.out, "Pick a number, or type a model id.")
	entered, err := con.prompt(ui.Ask(con.out, "Model", "1"))
	if err != nil {
		return "", err
	}
	if entered == "" {
		return models[0], nil
	}
	if n, nerr := strconv.Atoi(entered); nerr == nil && n >= 1 && n <= len(models) {
		return models[n-1], nil
	}
	return entered, nil
}

var datedModelSuffix = regexp.MustCompile(`-(?:\d{8}|\d{4}-\d{2}-\d{2})$`)

// modelChoices is the numbered menu. The current model is first. Dated snapshots
// are hidden when the provider also offers the undated id.
func modelChoices(provider string, live []string, current string) []string {
	const limit = 15
	source := dropDatedAliases(live)
	if len(source) == 0 {
		source = suggestedModels(provider)
	}
	sort.SliceStable(source, func(i, j int) bool {
		return modelLess(source[i], source[j])
	})
	seen := map[string]struct{}{}
	out := make([]string, 0, limit)
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || len(out) >= limit {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	add(current)
	for _, id := range source {
		add(id)
	}
	return out
}

func suggestedModels(provider string) []string {
	switch provider {
	case "openai":
		return []string{config.DefaultModel("openai")}
	case "gemini":
		return []string{config.DefaultModel("gemini")}
	case "xai":
		return []string{config.DefaultModel("xai"), "grok-4.7"}
	default:
		return []string{config.DefaultModel("anthropic"), "claude-haiku-4-5"}
	}
}

func dropDatedAliases(ids []string) []string {
	set := map[string]struct{}{}
	for _, id := range ids {
		set[id] = struct{}{}
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if loc := datedModelSuffix.FindStringIndex(id); loc != nil {
			if _, ok := set[id[:loc[0]]]; ok {
				continue
			}
		}
		out = append(out, id)
	}
	return out
}

func modelLess(a, b string) bool {
	pa, pb := modelPriority(a), modelPriority(b)
	if pa != pb {
		return pa < pb
	}
	da := datedModelSuffix.MatchString(a)
	db := datedModelSuffix.MatchString(b)
	if da != db {
		return !da
	}
	return a < b
}

func modelPriority(id string) int {
	lower := strings.ToLower(id)
	switch {
	case strings.Contains(lower, "sonnet") || strings.Contains(lower, "gpt-4.1") || strings.Contains(lower, "gpt-5") || strings.Contains(lower, "gemini-2.5") || strings.Contains(lower, "grok-4.3") || strings.Contains(lower, "grok-4.7"):
		return 0
	case strings.Contains(lower, "haiku") || strings.Contains(lower, "opus") || strings.Contains(lower, "gpt-4") || strings.Contains(lower, "gemini") || strings.Contains(lower, "grok"):
		return 1
	default:
		return 2
	}
}

func offerHook(con console) int {
	if !gitx.Inside() {
		printNext(con)
		return 0
	}
	if hook.Installed() {
		fmt.Fprintln(con.out, ui.Pass(con.out, "Pre-push hook is already installed."))
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
		fmt.Fprintln(con.out, ui.Pass(con.out, "Pre-push hook installed."))
	}
	printNext(con)
	return 0
}

func printNext(con console) {
	fmt.Fprintln(con.out)
	fmt.Fprintln(con.out, ui.Pass(con.out, "Ready."))
	ui.Command(con.out, "shipquiz quiz", "quiz the current change")
	ui.Command(con.out, "shipquiz status", "show config, hook, and key")
	ui.Command(con.out, "git push", "runs the quiz when the hook is installed")
	fmt.Fprintln(con.out)
	ui.Hint(con.out, "Skip one push with SHIPQUIZ_SKIP=1 git push, or git push --no-verify.")
	ui.Hint(con.out, "If this shell cannot find shipquiz yet:")
	fmt.Fprintln(con.out, "  "+ui.Stamp(con.out, `export PATH="$HOME/.shipquiz/bin:$PATH"`))
}

func mustConfigPath() string {
	path, err := config.Path()
	if err != nil {
		return strings.TrimSpace(path)
	}
	return path
}
