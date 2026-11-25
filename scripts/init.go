package scripts

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/trustpkg/trustpkg-cli-forge/pkg/logger"
)

type hook struct {
	source      string
	destination string
}

var hooks = []hook{
	{source: "git-hooks/pre-commit.sh", destination: "pre-commit"},
	{source: "git-hooks/commit-msg.sh", destination: "commit-msg"},
	{source: "git-hooks/post-commit.sh", destination: "post-commit"},
}

const (
	destinationDirectory = ".git/hooks"
)

func Init(logHandler *logger.Logger) {
	if err := initScripts(logHandler); err != nil {
		logHandler.Print(fmt.Sprintf("Failed to initialize git-hooks scripts: %v", err), logger.SeverityError)

		os.Exit(1)
	}
}

func initScripts(logHandler *logger.Logger) (result error) {
	footerActive := false
	if err := logHandler.StartFooter("Initializing git-hooks scripts"); err != nil {
		if !errors.Is(err, logger.ErrFooterUnsupported) {
			return err
		}

		fmt.Fprintln(os.Stderr, "Warning:", err)
	} else {
		footerActive = true
		defer func() { result = errors.Join(result, logHandler.StopFooter()) }()
	}

	finish, err := logHandler.StartSection("Initialize git-hooks scripts", logger.SeverityInfo)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, finish()) }()

	for _, hook := range hooks {
		if footerActive {
			if err := logHandler.UpdateFooter("Installing " + hook.source); err != nil {
				return err
			}
		}

		if err := copyToGitHooks(hook.source, hook.destination); err != nil {
			logHandler.Print(fmt.Sprintf("Failed to install %s", hook.source), logger.SeverityError)

			return err
		}

		time.Sleep(time.Second / 2)
	}

	logHandler.Print("Successfully initialized git-hooks scripts", logger.SeveritySuccess)

	return nil
}
