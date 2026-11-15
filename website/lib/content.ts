/** One-line installer. Serves from this site at /install.sh. */
export const installCommand = "curl -fsSL https://shipquiz.com/install.sh | sh";

/** Path export for the shell that just ran the installer. */
export const pathExport = 'export PATH="$HOME/.shipquiz/bin:$PATH"';

/** Build the CLI from a checkout of this repo. */
export const sourceBuild = `mkdir -p ~/.shipquiz/bin
go build -o ~/.shipquiz/bin/shipquiz ./cmd/shipquiz
shipquiz setup
shipquiz install`;

/** Install the agent skill where Claude Code and Cursor both look. */
export const skillInstall = `mkdir -p .claude/skills/shipquiz .agents/skills
curl -fsSL https://shipquiz.com/skill/SKILL.md -o .claude/skills/shipquiz/SKILL.md
cp -R .claude/skills/shipquiz .agents/skills/`;

/** Global config written by shipquiz setup. */
export const globalConfig = `provider = "anthropic"
model = "claude-sonnet-4-5"
# base_url = "" # optional OpenAI-compatible host
# The API key lives in the macOS Keychain, or a 0600 file on Linux.

[quiz]
min_questions = 1
max_questions = 5
passing_average = 1.5
timeout_seconds = 60

[hooks]
pre_push = true
`;

/** Optional per-repo override committed next to the code. */
export const repoConfig = `model = "claude-haiku-4-5"
skip_paths = ["**/*_test.go", "docs/**", "*.md"]
question_style = "default" # default | failure-modes | data-flow
`;

/** Workflow written by `shipquiz github-action`. */
export const workflowYaml = `name: shipquiz

on:
  pull_request:
    types: [opened, synchronize, reopened]
  issue_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write
  checks: write

jobs:
  quiz:
    if: github.event_name == 'pull_request' || (github.event_name == 'issue_comment' && github.event.issue.pull_request && !contains(github.event.comment.body, '<!-- shipquiz'))
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install shipquiz
        run: |
          curl -fsSL https://shipquiz.com/install.sh | sh -s -- --no-setup
          echo "$HOME/.shipquiz/bin" >> "$GITHUB_PATH"

      - name: Run the quiz
        env:
          SHIPQUIZ_LLM_API_KEY: \${{ secrets.SHIPQUIZ_LLM_API_KEY }}
          SHIPQUIZ_PROVIDER: \${{ secrets.SHIPQUIZ_PROVIDER || 'anthropic' }}
          SHIPQUIZ_MODEL: \${{ secrets.SHIPQUIZ_MODEL || 'claude-sonnet-4-5' }}
          SHIPQUIZ_BASE_URL: \${{ secrets.SHIPQUIZ_BASE_URL }}
          GITHUB_TOKEN: \${{ secrets.GITHUB_TOKEN }}
        run: |
          if [ -z "$SHIPQUIZ_LLM_API_KEY" ]; then
            echo "::warning::shipquiz: SHIPQUIZ_LLM_API_KEY is not set, so this quiz was skipped"
            exit 0
          fi
          shipquiz action-run
`;

export const commands: { command: string; detail: string }[] = [
  { command: "shipquiz setup", detail: "Choose a provider, store the API key, and pick a model." },
  { command: "shipquiz install", detail: "Install the pre-push hook in the current repo." },
  { command: "shipquiz uninstall", detail: "Remove the hook and restore the previous one." },
  { command: "shipquiz quiz", detail: "Ask about the current change. Saved questions for this diff are reused." },
  { command: "shipquiz quiz --cached", detail: "Use a saved quiz only. Exits with an error if there isn't one." },
  { command: "shipquiz quiz --again", detail: "Ask again even if this diff already passed." },
  { command: "shipquiz status", detail: "Show the config, the hook, and whether the API key works." },
  { command: "shipquiz config", detail: "Print the config that will be used, including repo overrides." },
  { command: "shipquiz config set KEY VALUE", detail: "Update one global setting, such as quiz.timeout_seconds." },
  { command: "shipquiz github-action", detail: "Ask for the provider, model, and LLM API key, write the workflow, and show the secret to set." },
  { command: "shipquiz github-action --print", detail: "Print the workflow without asking." },
  { command: "shipquiz update", detail: "Replace this binary with the latest GitHub release." },
];
