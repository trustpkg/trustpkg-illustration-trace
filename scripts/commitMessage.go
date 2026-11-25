package scripts

import (
	"errors"
	"fmt"

	"github.com/trustpkg/trustpkg-cli-forge/pkg/logger"
)

func CommitMessage(logHandler *logger.Logger, messageFile string) error {
	return runPrecommit(logHandler, messageFile)
}

func printCommitError(logHandler *logger.Logger, err error) error {
	var invalid *invalidCommitMessageError
	if !errors.As(err, &invalid) {
		return logHandler.Print(err.Error(), logger.SeverityError, logger.WithTimestamp)
	}

	logErr := logHandler.Print(fmt.Sprintf("Invalid commit message: %q", invalid.subject), logger.SeverityError, logger.WithTimestamp)
	for _, reason := range invalid.reasons {
		logErr = errors.Join(logErr, logHandler.Print(reason.Error(), logger.SeverityError, logger.WithTimestamp))
	}
	logErr = errors.Join(logErr, logHandler.Print(""))
	logErr = errors.Join(logErr, logHandler.Print("Commit message guidelines:"))
	for _, help := range commitHelp {
		logErr = errors.Join(logErr, logHandler.Print(help))
	}
	return errors.Join(logErr, logHandler.Print(""))
}
