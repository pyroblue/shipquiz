package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"

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
	if hit && entry.Passed && quiz.MultipleChoiceReady(entry.Generated) && !again {
		fmt.Fprintln(os.Stderr, "shipquiz: this diff already passed")
		return 0
	}

	var generated quiz.Generated
	if hit && !again && quiz.MultipleChoiceReady(entry.Generated) {
		generated = entry.Generated
		fmt.Fprintln(os.Stderr, "shipquiz: using cached questions")
	} else if cachedOnly {
		fmt.Fprintln(os.Stderr, "shipquiz: no cached quiz for this diff")
		return 1
	} else {
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
		stop := ui.Spin(os.Stderr, "generating questions")
		generated, err = quiz.Generate(genCtx, client, cfg.Model, diff, cfg.QuestionStyle, cfg.Quiz.MinQuestions, cfg.Quiz.MaxQuestions)
		stop()
		cancel()
		if err != nil {
			if timedOut(genCtx, err) {
				if hook {
					fmt.Fprintln(os.Stderr, "shipquiz: timed out generating questions; allowing this push")
					return 0
				}
				return fail(fmt.Errorf("timed out generating questions after %ds. Raise quiz.timeout_seconds and try again", timeoutSeconds(cfg)))
			}
			if out := modelOutput(err); out != nil && !hook {
				fmt.Fprintf(os.Stderr, "shipquiz: %s\n", out.Error())
				offerCopy(con, out.Error())
				return 1
			}
			return fail(err)
		}
		entry = quiz.Entry{Generated: generated}
		if err := quiz.StoreEntry(key, entry); err != nil {
			fmt.Fprintf(os.Stderr, "shipquiz: cache: %v\n", err)
		}
	}

	ui.PrintIntro(con.out, generated)
	answers, err := ui.CollectChoices(con.in, con.file, con.out, generated.Questions)
	if err != nil {
		if hook || errors.Is(err, io.EOF) {
			fmt.Fprintln(os.Stderr, "shipquiz: could not read answers. Push blocked. Use SHIPQUIZ_SKIP=1 or --no-verify.")
			return 1
		}
		return fail(err)
	}

	for {
		ev := quiz.GradeChoices(generated.Questions, answers)
		entry.Generated = generated
		entry.Passed = ev.Passed()
		entry.Score = ev.Score
		if err := quiz.StoreEntry(key, entry); err != nil {
			fmt.Fprintf(os.Stderr, "shipquiz: cache: %v\n", err)
		}
		ui.PrintResult(con.out, ev)
		if ev.Passed() {
			return 0
		}

		missed := missedQuestions(generated.Questions, ev)
		if len(missed) == 0 {
			break
		}
		retry, promptErr := con.confirm("Retake missed questions?", false)
		if promptErr != nil && !errors.Is(promptErr, io.EOF) {
			return fail(promptErr)
		}
		if !retry {
			break
		}
		retried, answerErr := ui.CollectChoices(con.in, con.file, con.out, missed)
		if answerErr != nil {
			if hook || errors.Is(answerErr, io.EOF) {
				fmt.Fprintln(os.Stderr, "shipquiz: could not read answers. Push blocked. Use SHIPQUIZ_SKIP=1 or --no-verify.")
				return 1
			}
			return fail(answerErr)
		}
		answers = mergeAnswers(answers, retried)
	}
	if hook {
		fmt.Fprintln(con.out, ui.Fail(con.out, "Push blocked.")+" Explain the change and push again, or use --no-verify.")
	}
	return 1
}

func missedQuestions(questions []quiz.Question, ev quiz.Evaluation) []quiz.Question {
	scores := make(map[string]int, len(ev.PerQuestion))
	for _, item := range ev.PerQuestion {
		scores[item.ID] = item.Score
	}
	missed := make([]quiz.Question, 0, len(questions))
	for _, question := range questions {
		if scores[question.ID] < 2 {
			missed = append(missed, question)
		}
	}
	return missed
}

func mergeAnswers(current, retried []quiz.Answer) []quiz.Answer {
	replacements := make(map[string]quiz.Answer, len(retried))
	for _, answer := range retried {
		replacements[answer.ID] = answer
	}
	merged := append([]quiz.Answer(nil), current...)
	for i := range merged {
		if answer, ok := replacements[merged[i].ID]; ok {
			merged[i] = answer
			delete(replacements, merged[i].ID)
		}
	}
	for _, answer := range retried {
		if _, ok := replacements[answer.ID]; ok {
			merged = append(merged, answer)
			delete(replacements, answer.ID)
		}
	}
	return merged
}

func timedOut(ctx context.Context, err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded)
}

func timeoutSeconds(cfg config.Config) int {
	if cfg.Quiz.TimeoutSeconds > 0 {
		return cfg.Quiz.TimeoutSeconds
	}
	return 60
}

func modelOutput(err error) *quiz.ModelOutput {
	var out *quiz.ModelOutput
	if errors.As(err, &out) {
		return out
	}
	return nil
}

// offerCopy waits for c, then copies the model response. Other keys leave it on screen.
func offerCopy(con console, text string) {
	if con.file == nil || !term.IsTerminal(int(con.file.Fd())) {
		return
	}
	fmt.Fprintln(con.out, ui.Muted(con.out, "press c to copy this response"))
	state, err := term.MakeRaw(int(con.file.Fd()))
	if err != nil {
		return
	}
	defer term.Restore(int(con.file.Fd()), state)
	b, err := con.in.ReadByte()
	fmt.Fprint(con.out, "\r\n")
	if err != nil || (b != 'c' && b != 'C') {
		return
	}
	if err := copyText(text); err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: could not copy: %v\n", err)
		return
	}
	fmt.Fprintln(os.Stderr, "shipquiz: copied")
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
