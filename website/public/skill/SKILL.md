---
name: shipquiz
description: >
  Quizzes the developer on the architecture of a diff before a push or pull
  request. Use when the user is about to commit, push, open a PR, or asks
  for a ship quiz, pre-push hook, architecture quiz, or shipquiz.
---

# shipquiz

shipquiz asks architecture questions about the current diff and blocks `git push` until the answers pass. The CLI calls the user's own model provider. Do not send the diff or the API key anywhere else. Do not invent a quiz yourself when the CLI is available.

## Before a change ships

1. Check whether `shipquiz` is on `PATH` (`shipquiz version`).
2. If it is missing, stop and give the user the install commands. Do not skip ahead.
3. If it is installed, run `shipquiz quiz` in the repository and let the user answer in the terminal.
4. A passing quiz exits 0. A failing quiz exits 1. Do not treat a failure as done.
5. Do not set `SHIPQUIZ_SKIP` or pass `--no-verify` unless the user explicitly asks to skip.

## Install

Hook (local gate):

```bash
curl -fsSL https://shipquiz.com/install.sh | sh
export PATH="$HOME/.shipquiz/bin:$PATH"
shipquiz install
```

GitHub Action (pull request gate):

```bash
shipquiz github-action
```

The command asks for the provider, model, and LLM API key, writes `.github/workflows/shipquiz.yml`, and tells you to set the repository secret `SHIPQUIZ_LLM_API_KEY`. That secret is the model provider key. The key is not written into the workflow. Require the `shipquiz` check if a failed quiz should block merging. Reviewers reply to the quiz comment in their own words.

Agent skill (this file), from the repository root. Claude Code reads `.claude/skills`. Cursor reads that folder and `.agents/skills`.

```bash
mkdir -p .claude/skills/shipquiz .agents/skills
curl -fsSL https://shipquiz.com/skill/SKILL.md -o .claude/skills/shipquiz/SKILL.md
cp -R .claude/skills/shipquiz .agents/skills/
```

## Rules the quiz enforces

- Questions stay at the system and product level: purpose, ownership, boundaries, guarantees, failure behavior, and major trade-offs.
- In the terminal, Ctrl-G shows a reference answer. A blank line submits.
- Ask one question per meaningful architectural decision. Most quizzes have 1–3 questions. Use 4–5 only when the diff has that many distinct architectural decisions. Hard cap is 5.
- Pass only when the average score is at least 1.5 and no question scored 0. Scores are 0, 1, or 2.
- Lockfiles and `skip_paths` in `.shipquiz.toml` skip the quiz. A trivial code change still gets one question.
- The same diff is cached. A passed diff is not asked again unless the user runs `shipquiz quiz --again`.
- On a hook timeout, the push is allowed and a warning is printed. A failed quiz still blocks the push.

## Config

Global config is `~/.config/shipquiz/config.toml`. The API key is in the macOS Keychain, or a mode-0600 file on Linux. Per-repo overrides live in `.shipquiz.toml` (`model`, `skip_paths`, `question_style`).

`question_style` is `default`, `failure-modes`, or `data-flow`.
