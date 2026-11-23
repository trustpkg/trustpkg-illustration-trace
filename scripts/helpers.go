package scripts

import (
	"fmt"
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
