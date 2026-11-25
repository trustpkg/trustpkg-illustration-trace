package scripts

import (
	"fmt"
	"slices"
	"strings"
)

var (
	commitHelp = []string{
		"Expected: type(scope)!: message (scope and ! are optional)",
		"Types: feat, fix, refactor, build, ci, test, chore, docs",
		"Breaking changes (!): feat, fix, build",
		"Examples: fix: correct a bug | feat(api): add endpoint | feat!: change API",
	}

	allowedTypes = []string{
		"feat",
		"fix",
		"refactor",
		"build",
		"ci",
		"test",
		"chore",
		"docs",
	}

	typesWithAllowedBreakingChange = []string{
		"feat",
		"fix",
		"build",
	}
)

type commitStr string
type commitType string

func (commit *commitStr) extractChunks() (commitType, string, error) {
	if commit == nil || *commit == "" {
		return "", "", fmt.Errorf("Commit is nil or empty")
	}

	prefix, message, isFound := strings.Cut(string(*commit), ":")

	if !isFound {
		return "", "", fmt.Errorf("Commit does not contain a \":\" or has an invalid format")
	}

	return commitType(prefix), message, nil
}

func (commit *commitStr) validate() (bool, []error) {
	prefix, message, err := commit.extractChunks()
	if err != nil {
		return false, []error{err}
	}

	var allErrors []error

	baseType := strings.TrimSuffix(string(prefix), "!")
	if before, scope, found := strings.Cut(baseType, "("); found {
		baseType = before
		if !strings.HasSuffix(scope, ")") || len(scope) < 2 || strings.ContainsAny(scope[:len(scope)-1], "()") {
			allErrors = append(allErrors, fmt.Errorf("Commit scope is invalid"))
		}
	}

	isAllowed := false
	for _, currentType := range allowedTypes {
		if baseType == currentType {
			isAllowed = true

			break
		}
	}

	if !isAllowed {
		allErrors = append(allErrors, fmt.Errorf("Commit type is not allowed"))
	}

	if prefix.isBreakingChange() {
		isAllowed := slices.Contains(typesWithAllowedBreakingChange, baseType)
		if !isAllowed {
			allErrors = append(allErrors, fmt.Errorf("Breaking change is not allowed for this commit type"))
		}
	}

	if strings.TrimSpace(message) == "" {
		allErrors = append(allErrors, fmt.Errorf("Commit message is empty"))
	}

	return len(allErrors) == 0, allErrors
}

func (prefix commitType) isBreakingChange() bool {
	return strings.HasSuffix(string(prefix), "!")
}
