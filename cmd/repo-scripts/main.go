package main

import (
	"os"
	"trace/scripts"

	"github.com/trustpkg/trustpkg-cli-forge/pkg/logger"
)

func main() {
	terminal := scripts.RunTerminal()

	if len(os.Args) == 1 || os.Args[1] == "init" {
		scripts.Init(terminal)

		return
	}

	if len(os.Args) > 1 && os.Args[1] == "pre-commit" {
		scripts.Precommit(terminal)

		return
	}

	if len(os.Args) > 1 && os.Args[1] == "commit-msg" {
		if len(os.Args) != 3 {
			terminal.Print("Expected commit message file", logger.SeverityError)
			os.Exit(1)
		}
		if err := scripts.CommitMessage(terminal, os.Args[2]); err != nil {
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "post-commit" {
		if err := scripts.PostCommit(terminal); err != nil {
			os.Exit(1)
		}
		return
	}

	terminal.Print("Unknown repo script: "+os.Args[1], logger.SeverityError)

	os.Exit(1)
}
