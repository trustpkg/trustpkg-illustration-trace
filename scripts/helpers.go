package scripts

import (
	"fmt"
	"io"
	"os"

	"github.com/trustpkg/trustpkg-cli-forge/pkg/facts"
	"github.com/trustpkg/trustpkg-cli-forge/pkg/logger"
	"github.com/trustpkg/trustpkg-cli-forge/pkg/theme"
)

func RunTerminal() *logger.Logger {
	terminalFacts, err := facts.GatherTerminalFacts(nil)
	if err != nil {
		fmt.Print("Failed to init terminal")

		os.Exit(1)
	}

	handler, err := theme.InitTheme(theme.ThemeInitInput{TerminalFacts: terminalFacts})
	if err != nil {
		fmt.Print("Failed to init terminal")

		os.Exit(1)
	}

	maxWidth := maxTerminalWidth
	logHandler := logger.NewLogger(logger.NewLoggerInput{Facts: terminalFacts, Theme: handler.Theme(), MaxWidth: &maxWidth})

	return logHandler
}

func copyToGitHooks(source string, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(fmt.Sprintf("%s/%s", destinationDirectory, destination))
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	err = os.Chmod(destinationFile.Name(), 0755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return err
	}

	return nil
}
