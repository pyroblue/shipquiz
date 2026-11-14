# shipquiz

You're the captain. Ship only what you can explain.

shipquiz asks a few questions about your change and blocks `git push` until the answers show you understand it. You use your own API key and model. The diff never goes to a shipquiz server.

Use a pre-push hook, a GitHub Action, or an agent skill. All three run the same command-line tool, which calls your model provider. shipquiz does not store API keys or diffs.

macOS and Linux. The site is [shipquiz.com](https://shipquiz.com). Source is [github.com/pyroblue/shipquiz](https://github.com/pyroblue/shipquiz).

## Install

Paste the install command on a Mac or Linux machine. It downloads shipquiz into `~/.shipquiz/bin` and adds that directory to `.zshrc` or `.bashrc` if it is not already there. Re-running it is safe. Setup then asks for your API key, a model, and whether to quiz this repo before a push.

```bash
curl -fsSL https://shipquiz.com/install.sh | sh
export PATH="$HOME/.shipquiz/bin:$PATH"
```

Run the path line if you want to stay in the same terminal. A new one already knows where shipquiz is.

Releases are not published yet, so the installer cannot download a binary. Have the repo checked out? Build it from there instead. Go is required.

```bash
mkdir -p ~/.shipquiz/bin
go build -o ~/.shipquiz/bin/shipquiz ./cmd/shipquiz
shipquiz setup
shipquiz install
```

`make install` builds the same binary into `~/.shipquiz/bin`. `make build` writes `dist/shipquiz`.

Setup does this:

1. If a config already exists, setup asks before replacing it.
2. Pick anthropic, openai, or gemini. Gemini uses Google's version of the OpenAI API.
3. The API key is read from the terminal and stored in the macOS Keychain, or in a file only you can read on Linux.
4. shipquiz lists that provider's models and lets you pick one. The default is a fast model for that provider.
5. You can send a one-token test to check that the key and model work.
6. Inside a git repo, setup offers to install the pre-push hook.

Set `base_url` when you use a host that speaks the OpenAI API. For Anthropic, `base_url` is the site origin, such as `https://api.anthropic.com`. For OpenAI and Gemini it is the API root, and it already includes the version, such as `https://api.openai.com/v1`.

`SHIPQUIZ_API_KEY` overrides the saved key. That is how CI supplies the secret.

## Commands

| Command | What it does |
| --- | --- |
| `shipquiz setup` | Choose a provider, store the API key, and pick a model. |
| `shipquiz install` | Install the pre-push hook in the current repo. |
| `shipquiz uninstall` | Remove the hook and restore the previous one. |
| `shipquiz quiz` | Ask about the current change. Saved questions for this diff are reused. |
| `shipquiz quiz --cached` | Use a saved quiz only. Exits with an error if there isn't one. |
| `shipquiz quiz --again` | Ask again even if this diff already passed. |
| `shipquiz status` | Show the config, the hook, and whether the API key works. |
| `shipquiz config` | Print the config that will be used, including repo overrides. |
| `shipquiz config set KEY VALUE` | Update one global setting, such as `quiz.timeout_seconds`. |
| `shipquiz github-action` | Print the GitHub Actions workflow. |
| `shipquiz github-action --write` | Write `.github/workflows/shipquiz.yml`. |
| `shipquiz update` | Replace this binary with the latest GitHub release. |

The hook calls `shipquiz hook-run`. GitHub Actions calls `shipquiz action-run`. Neither is a command you type.

## Config

Global config is `~/.config/shipquiz/config.toml`. A repo may add `.shipquiz.toml`. Repo values override the global file for the keys they set.

```toml
provider = "anthropic"
model = "claude-sonnet-4-5"
# base_url = "" # optional OpenAI-compatible host
# The API key lives in the macOS Keychain, or a 0600 file on Linux.

[quiz]
min_questions = 1
max_questions = 7
passing_average = 1.5
timeout_seconds = 20

[hooks]
pre_push = true
```

A repo file can override the model, skip paths, and question style:

```toml
model = "claude-haiku-4-5"
skip_paths = ["**/*_test.go", "docs/**", "*.md"]
question_style = "default" # default | failure-modes | data-flow
```

`question_style` is `default`, `failure-modes`, or `data-flow`. Default asks about the change as a whole. Failure-modes asks only what can break. Data-flow asks you to follow one value from one side of the diff to the other.

`shipquiz config set quiz.timeout_seconds 15` writes the global file. Keys you can set: `provider`, `model`, `base_url`, `question_style`, `quiz.min_questions`, `quiz.max_questions`, `quiz.passing_average`, `quiz.timeout_seconds`, `hooks.pre_push`.

## Hook

`shipquiz install` adds a pre-push hook. Git sends it the commits you are about to push. If you pass, the push goes through. If you don't, it stops.

The hook is a short shell script. It runs the CLI with the commits git was about to push. If you already had a pre-push hook, shipquiz moves it to `pre-push.previous` and runs it after you pass. `shipquiz uninstall` puts the old hook back.

- Empty diffs, lockfiles, and paths in `skip_paths` skip the quiz and let the push continue.
- Questions are saved by the two commits, the diff, the model, and the question style. A diff you already passed is not asked again.
- A failed quiz stops the push.
- If the model times out, the hook prints a warning and lets the push through.
- `SHIPQUIZ_SKIP=1` prints a skip line and lets the push through. `git push --no-verify` skips the hook entirely.
- Set `hooks.pre_push` to false to leave the hook installed but skip the quiz.

You answer in the terminal, one question at a time. A blank line submits the answer. If you fail and try the same diff again, shipquiz reuses the questions and only calls the model to grade.

## GitHub Action

On a pull request, the Action posts the questions as a comment and opens a check named `shipquiz` with the result `action_required`. Require that check in branch protection if you want it to block merging. The diff is sent to your model provider, not to shipquiz.

Reply in the thread. Start each answer with its id:

```text
q1: the token is checked in the handler, so a new route that skips it is public
q2: a cache miss stamps the origin because the lock is process-local
```

When you reply, the workflow grades the answers, posts what was missing, and marks the check passed or failed. A failed quiz also fails the job. Add a repository secret named `SHIPQUIZ_API_KEY`.

`shipquiz github-action` prints the workflow. `shipquiz github-action --write` writes `.github/workflows/shipquiz.yml`, and refuses to overwrite an existing one. This repo already has that file. The Action installs with `install.sh --no-setup` and runs `shipquiz action-run` with `SHIPQUIZ_API_KEY`, `SHIPQUIZ_PROVIDER`, and `SHIPQUIZ_MODEL`.

## Skill

The skill tells your agent to run `shipquiz quiz` when a change is about to commit, push, or open a pull request. You can also invoke it by name. It is not a gate. The hook is what blocks the push. The skill never sees your API key, and it does not grade the diff itself. If the CLI is missing, the skill stops and shows the install commands.

Claude Code and Cursor both read `.claude/skills`. Cursor also reads `.agents/skills`, which covers the other agents that look there.

```bash
mkdir -p .claude/skills/shipquiz .agents/skills
curl -fsSL https://shipquiz.com/skill/SKILL.md -o .claude/skills/shipquiz/SKILL.md
cp -R .claude/skills/shipquiz .agents/skills/
```

The same file is in this repo at `skills/shipquiz/SKILL.md` and on the site at [shipquiz.com/skill/SKILL.md](https://shipquiz.com/skill/SKILL.md). Commit both copies if the whole team should get it.

## Scoring

Each answer is 0 (wrong or missing), 1 (partial), or 2 (clear). The quiz passes only when the average is at least `passing_average` (default 1.5) and no question scored 0. shipquiz checks this itself after the model replies, so one wrong answer still fails the quiz.

Questions have to be about this diff. They are not questions about syntax or style. Question count follows the size of the change: 1 trivial, 2 tiny, 3 small, 4–5 medium, 6–7 large. The hard cap is 7. Even a trivial change gets one question.

The quiz is skipped only when the diff is empty, when it only touches lockfiles, or when every file matches `skip_paths`.

## Website

The marketing site and docs live in `website/`. Point Vercel at that directory. The docs page is the long form of this file.

```bash
cd website
npm install
npm run dev
```

## Tests

`go test ./...` covers config merging, skipped paths, grading, the hook installer, and comment parsing.
