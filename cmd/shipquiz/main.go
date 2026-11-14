// Command shipquiz quizzes a developer on a diff before git push.
package main

import (
	"os"

	"github.com/shipquiz/shipquiz/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
