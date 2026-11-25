package scripts

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/trustpkg/trustpkg-cli-forge/pkg/logger"
)

func versionBumpCallback(logHandler *logger.Logger) VersionBumpCallback {
	return func(step string, details string) {
		switch step {
		case "validation":
			if err := logHandler.Print(details, logger.SeverityInfo, logger.WithTimestamp); err != nil {
				fmt.Fprintf(os.Stderr, "Error logging: %v\n", err)
			}
		case "bump":
			if err := logHandler.Print(details, logger.SeveritySuccess, logger.WithTimestamp); err != nil {
				fmt.Fprintf(os.Stderr, "Error logging: %v\n", err)
			}
		}
	}
}

func runPrecommit(logHandler *logger.Logger, messageFile string) (result error) {
	footerActive := false
	if err := logHandler.StartFooter("Running pre-commit hook"); err != nil {
		if !errors.Is(err, logger.ErrFooterUnsupported) {
			return err
		}

		fmt.Fprintln(os.Stderr, "Warning:", err)
	} else {
		footerActive = true
		defer func() { result = errors.Join(result, logHandler.StopFooter()) }()
	}

	finish, err := logHandler.StartSection("Running pre-commit hook", logger.SeverityInfo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer func() {
		message, severity := "Pre-commit hook completed", logger.SeveritySuccess
		if result != nil {
			message, severity = "Pre-commit hook failed", logger.SeverityError
			result = errors.Join(result, printCommitError(logHandler, result))
		}
		result = errors.Join(result, logHandler.Print(message, severity, logger.WithTimestamp), finish())
	}()

	if err := logHandler.Print(""); err != nil {
		return err
	}

	commands := [][]string{
		{"go", "vet", "./..."},
		{"go", "build", "./..."},
		{"govulncheck", "./..."},
		{"gitleaks", "git", "--staged", "--verbose", "--redact", "--no-color", "--exit-code", "1"},
	}

	for index, command := range commands {
		if index > 0 {
			if err := logHandler.Print(""); err != nil {
				return err
			}
		}
		if footerActive {
			if err := logHandler.UpdateFooter("Running " + strings.Join(command, " ")); err != nil {
				return err
			}
		}
		if err := runCommand(logHandler, command[0], command[1:]...); err != nil {
			return err
		}
	}

	if messageFile != "" {
		if footerActive {
			if err := logHandler.UpdateFooter("Validating commit and bumping version"); err != nil {
				return err
			}
		}
		if err := logHandler.Print(""); err != nil {
			return err
		}
		if err := logHandler.Print("Validating commit and version", logger.SeverityStart, logger.WithTimestamp); err != nil {
			return err
		}
		if err := validateCommitVersion(messageFile, versionBumpCallback(logHandler)); err != nil {
			return err
		}
		if err := logHandler.Print("Commit and version validation completed", logger.SeveritySuccess, logger.WithTimestamp); err != nil {
			return err
		}
	}

	return nil
}

func runCommand(logHandler *logger.Logger, name string, args ...string) (result error) {
	command := strings.Join(append([]string{name}, args...), " ")
	defer func() {
		message, severity := command, logger.SeveritySuccess
		if result != nil {
			severity = logger.SeverityError
		}
		result = errors.Join(result, logHandler.Print(message, severity, logger.WithTimestamp))
	}()

	if err := logHandler.Print(command, logger.SeverityStart, logger.WithTimestamp); err != nil {
		return err
	}

	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		commandErr := fmt.Errorf("%s: %w", command, err)
		logErr := logHandler.Print(fmt.Sprintf("%s\n%s", commandErr, output), logger.SeverityError, logger.WithTimestamp)

		return errors.Join(commandErr, logErr)
	}
	if len(output) > 0 {
		return logHandler.Print(string(output), logger.SeverityInfo, logger.WithTimestamp)
	}

	return nil
}

func Precommit(logHandler *logger.Logger) {
	if err := runPrecommit(logHandler, ""); err != nil {
		os.Exit(1)
	}
}
