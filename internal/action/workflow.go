// Package action runs the pull-request quiz and renders the GitHub Actions workflow.
package action

import (
	"strconv"
	"strings"
)

// Workflow is the GitHub Action for one provider and model.
// Repository secrets override the baked-in defaults when set.
// The model provider key belongs in the SHIPQUIZ_LLM_API_KEY repository secret.
func Workflow(provider, model, baseURL string) string {
	if strings.TrimSpace(provider) == "" {
		provider = "anthropic"
	}
	if strings.TrimSpace(model) == "" {
		model = "claude-sonnet-4-5"
	}
	baseLine := ""
	if url := strings.TrimSpace(baseURL); url != "" {
		baseLine = "\n          SHIPQUIZ_BASE_URL: ${{ secrets.SHIPQUIZ_BASE_URL || " + strconv.Quote(url) + " }}"
	} else {
		baseLine = "\n          SHIPQUIZ_BASE_URL: ${{ secrets.SHIPQUIZ_BASE_URL }}"
	}
	return strings.NewReplacer(
		"__PROVIDER__", provider,
		"__MODEL__", model,
		"__BASE_URL_LINE__", baseLine,
	).Replace(workflowTemplate)
}

const workflowTemplate = `name: shipquiz

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
      - name: Resolve checkout ref
        id: target
        run: |
          if [ "${{ github.event_name }}" = "pull_request" ]; then
            echo "ref=${{ github.event.pull_request.head.sha }}" >> "$GITHUB_OUTPUT"
          else
            echo "ref=refs/pull/${{ github.event.issue.number }}/head" >> "$GITHUB_OUTPUT"
          fi

      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          ref: ${{ steps.target.outputs.ref }}

      - name: Install shipquiz
        run: |
          curl -fsSL https://shipquiz.com/install.sh | sh -s -- --no-setup
          echo "$HOME/.shipquiz/bin" >> "$GITHUB_PATH"

      - name: Run the quiz
        env:
          SHIPQUIZ_LLM_API_KEY: ${{ secrets.SHIPQUIZ_LLM_API_KEY }}
          SHIPQUIZ_PROVIDER: ${{ secrets.SHIPQUIZ_PROVIDER || '__PROVIDER__' }}
          SHIPQUIZ_MODEL: ${{ secrets.SHIPQUIZ_MODEL || '__MODEL__' }}__BASE_URL_LINE__
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          if [ -z "$SHIPQUIZ_LLM_API_KEY" ]; then
            echo "::warning::shipquiz: SHIPQUIZ_LLM_API_KEY is not set, so this quiz was skipped"
            exit 0
          fi
          shipquiz action-run
`
