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

	terminal.Print("Unknown repo script: "+os.Args[1], logger.SeverityError)

	os.Exit(1)
}
