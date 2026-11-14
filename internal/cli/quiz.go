package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/shipquiz/shipquiz/internal/config"
	"github.com/shipquiz/shipquiz/internal/gitx"
	"github.com/shipquiz/shipquiz/internal/keychain"
	"github.com/shipquiz/shipquiz/internal/llm"
	"github.com/shipquiz/shipquiz/internal/quiz"
	"github.com/shipquiz/shipquiz/internal/ui"
)

func runQuiz(diff quiz.Diff, hook, cachedOnly, again bool) int {
	root, _ := gitx.Root()
	cfg, err := config.LoadForRepo(root)
	if err != nil {
		return fail(err)
	}
	if hook && !cfg.Hooks.PrePush {
		fmt.Fprintln(os.Stderr, "shipquiz: skipped (hooks.pre_push is false)")
		return 0
	}
	if diff.Empty() {
		fmt.Fprintln(os.Stderr, "shipquiz: nothing to quiz")
		return 0
	}
	if quiz.ShouldSkip(diff.Files, cfg.SkipPaths) {
		fmt.Fprintln(os.Stderr, "shipquiz: skipped (lockfiles or skip_paths only)")
		return 0
	}

	con := openConsole()
	defer con.close()

	key := quiz.Key(diff.Base, diff.Head, diff.Text, cfg.Model, cfg.QuestionStyle)
	entry, hit, err := quiz.LoadEntry(key)
	if err != nil {
		return fail(err)
	}
	if hit && entry.Passed && !again {
		fmt.Fprintln(os.Stderr, "shipquiz: this diff already passed")
		return 0
	}

	var generated quiz.Generated
	if hit && !again && len(entry.Generated.Questions) > 0 {
		generated = entry.Generated
		fmt.Fprintln(os.Stderr, "shipquiz: using cached questions")
	} else if cachedOnly {
		fmt.Fprintln(os.Stderr, "shipquiz: no cached quiz for this diff")
		return 1
	} else {
		fmt.Fprintln(os.Stderr, "shipquiz: generating questions…")
		apiKey, err := keychain.Get(cfg.Provider)
		if err != nil {
			fmt.Fprintln(os.Stderr, "shipquiz: no API key. Run shipquiz setup.")
			return 1
		}
		client, err := llm.New(cfg.Provider, cfg.BaseURL, apiKey)
		if err != nil {
			return fail(err)
		}
		genCtx, cancel := quiz.WithTimeout(context.Background(), cfg.Quiz.TimeoutSeconds)
		generated, err = quiz.Generate(genCtx, client, cfg.Model, diff, cfg.QuestionStyle, cfg.Quiz.MinQuestions, cfg.Quiz.MaxQuestions)
		cancel()
		if err != nil {
			if hook && timedOut(genCtx, err) {
				fmt.Fprintln(os.Stderr, "shipquiz: timed out generating questions; allowing this push")
				return 0
			}
			return fail(err)
		}
		entry = quiz.Entry{Generated: generated}
		if err := quiz.StoreEntry(key, entry); err != nil {
			fmt.Fprintf(os.Stderr, "shipquiz: cache: %v\n", err)
		}
	}

	ui.PrintIntro(con.out, generated)
	answers, err := ui.CollectAnswers(con.in, con.out, generated.Questions)
	if err != nil {
		if hook || errors.Is(err, io.EOF) {
			fmt.Fprintln(os.Stderr, "shipquiz: could not read answers. Push blocked. Use SHIPQUIZ_SKIP=1 or --no-verify.")
			return 1
		}
		return fail(err)
	}

	fmt.Fprintln(os.Stderr, "shipquiz: grading…")
	apiKey, err := keychain.Get(cfg.Provider)
	if err != nil {
		fmt.Fprintln(os.Stderr, "shipquiz: no API key. Run shipquiz setup.")
		return 1
	}
	client, err := llm.New(cfg.Provider, cfg.BaseURL, apiKey)
	if err != nil {
		return fail(err)
	}
	gradeCtx, cancel := quiz.WithTimeout(context.Background(), cfg.Quiz.TimeoutSeconds)
	ev, err := quiz.Grade(gradeCtx, client, cfg.Model, diff, generated.Questions, answers, cfg.Quiz.PassingAverage)
	cancel()
	if err != nil {
		if hook && timedOut(gradeCtx, err) {
			fmt.Fprintln(os.Stderr, "shipquiz: timed out grading; allowing this push")
			return 0
		}
		return fail(err)
	}
	entry.Generated = generated
	entry.Passed = ev.Passed()
	entry.Score = ev.Score
	if err := quiz.StoreEntry(key, entry); err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: cache: %v\n", err)
	}
	ui.PrintResult(con.out, ev)
	if !ev.Passed() {
		if hook {
			fmt.Fprintln(con.out, "push blocked. Explain the change and push again, or use --no-verify.")
		}
		return 1
	}
	return 0
}

func timedOut(ctx context.Context, err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded)
}

func fail(err error) int {
	fmt.Fprintf(os.Stderr, "shipquiz: %v\n", err)
	return 1
}

func readAll(r io.Reader) string {
	data, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	return string(data)
}
