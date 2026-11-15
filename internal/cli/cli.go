// Package cli is the shipquiz command line.
package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/shipquiz/shipquiz/internal/action"
	"github.com/shipquiz/shipquiz/internal/version"
)

// Run executes one shipquiz command and returns the process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		usage(os.Stdout)
		return 0
	}
	if interactiveCommand(args[0]) {
		offerUpdate(args[0])
	}
	switch args[0] {
	case "help", "-h", "--help":
		usage(os.Stdout)
		return 0
	case "version", "-v", "--version":
		fmt.Println(version.Current)
		return 0
	case "setup":
		return cmdSetup()
	case "install":
		return cmdInstall()
	case "uninstall":
		return cmdUninstall()
	case "quiz":
		return cmdQuiz(args[1:])
	case "status":
		return cmdStatus()
	case "config":
		return cmdConfig(args[1:])
	case "github-action":
		return cmdGitHubAction(args[1:])
	case "update":
		return cmdUpdate()
	case "hook-run":
		return cmdHookRun()
	case "action-run":
		return action.Run(context.Background())
	default:
		fmt.Fprintf(os.Stderr, "shipquiz: unknown command %q\n\n", args[0])
		usage(os.Stderr)
		return 1
	}
}

func usage(out *os.File) {
	fmt.Fprintf(out, `shipquiz %s

Quiz yourself on the architecture of a change before it ships.

  shipquiz setup                 Choose a provider, key, and model
  shipquiz install               Install the pre-push hook in this repo
  shipquiz uninstall             Remove the pre-push hook
  shipquiz quiz [--cached] [--again]
                                 Quiz the current change
  shipquiz status                Show config, hook, and key validity
  shipquiz config                Print the resolved config
  shipquiz config set KEY VALUE  Update the global config
  shipquiz github-action         Ask for the model and key, write the workflow,
                                 and print the SHIPQUIZ_LLM_API_KEY secret
  shipquiz github-action --print Print the workflow without asking
  shipquiz update                Replace this binary with the latest release

The hook calls shipquiz hook-run. GitHub Actions calls shipquiz action-run.
Keys stay in the macOS Keychain or a mode-0600 file. Diffs are not uploaded
to shipquiz.

Skip one push:  SHIPQUIZ_SKIP=1 git push
                git push --no-verify

A newer release asks before it replaces this binary.
SHIPQUIZ_NO_UPDATE=1 skips that check.
`, version.Current)
}
