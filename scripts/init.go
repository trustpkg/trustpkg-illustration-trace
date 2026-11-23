package scripts

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/trustpkg/trustpkg-cli-forge/pkg/logger"
)

func Init(logHandler *logger.Logger) {
	if err := initScripts(logHandler); err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

func initScripts(logHandler *logger.Logger) (result error) {
	if err := logHandler.StartFooter("Initializing git-hooks scripts"); err != nil {
		if !errors.Is(err, logger.ErrFooterUnsupported) {
			return err
		}
		fmt.Fprintln(os.Stderr, "Warning:", err)
	} else {
		defer func() { result = errors.Join(result, logHandler.StopFooter()) }()
	}

	finish, err := logHandler.StartSection("Initialize git-hooks scripts", logger.SeverityInfo)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, finish()) }()

	preCommitFile, err := os.Open("hooks/pre-commit.sh")
	if err != nil {
		logHandler.Print("Failed to Open hooks/pre-commit.sh", logger.SeverityError)

		return err
	}
	defer preCommitFile.Close()

	destinationFile, err := os.Create(".git/hooks/pre-commit")
	if err != nil {
		logHandler.Print("Failed to create pre-commit file", logger.SeverityError)

		return err
	}
	defer destinationFile.Close()

	err = os.Chmod(destinationFile.Name(), 0755)
	if err != nil {
		logHandler.Print("Failed to set executable permission for pre-commit file", logger.SeverityError)

		return err
	}

	if _, err := io.Copy(destinationFile, preCommitFile); err != nil {
		logHandler.Print("Failed to copy to pre-commit file", logger.SeverityError)

		return err
	}

	logHandler.Print("Successfully initialized hooks/pre-commit script", logger.SeveritySuccess)

	return nil
}
