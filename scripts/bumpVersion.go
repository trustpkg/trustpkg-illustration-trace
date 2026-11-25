package scripts

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type VersionBumpCallback func(step string, details string)

type invalidCommitMessageError struct {
	subject string
	reasons []error
}

func (err *invalidCommitMessageError) Error() string {
	return fmt.Sprintf("invalid commit message %q: %v", err.subject, errors.Join(err.reasons...))
}

type VersionParts struct {
	Major int
	Minor int
	Patch int
}

func readVersionFile(versionFilePath string) (VersionParts, error) {
	data, err := os.ReadFile(versionFilePath)
	if err != nil {
		return VersionParts{}, fmt.Errorf("read version file: %w", err)
	}

	return parseVersion(string(data))
}

func parseVersion(data string) (VersionParts, error) {
	version := strings.TrimSpace(data)
	if !regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(version) {
		return VersionParts{}, fmt.Errorf("invalid version format: %q", version)
	}
	parts := strings.Split(version, ".")

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return VersionParts{}, fmt.Errorf("parse major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return VersionParts{}, fmt.Errorf("parse minor version: %w", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return VersionParts{}, fmt.Errorf("parse patch version: %w", err)
	}

	return VersionParts{Major: major, Minor: minor, Patch: patch}, nil
}

func writeVersionFile(versionFilePath string, version VersionParts) error {
	content := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
	if err := os.WriteFile(versionFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write version file: %w", err)
	}

	return nil
}

func determineVersionBump(commitTypeStr commitType) string {
	prefix := string(commitTypeStr)

	if commitTypeStr.isBreakingChange() {
		return "major"
	}

	baseType, _, _ := strings.Cut(prefix, "(")

	if baseType == "feat" {
		return "minor"
	}

	if baseType == "fix" {
		return "patch"
	}

	return "none"
}

func bumpVersionParts(version VersionParts, bumpType string) VersionParts {
	switch bumpType {
	case "major":
		return VersionParts{Major: version.Major + 1, Minor: 0, Patch: 0}
	case "minor":
		return VersionParts{Major: version.Major, Minor: version.Minor + 1, Patch: 0}
	case "patch":
		return VersionParts{Major: version.Major, Minor: version.Minor, Patch: version.Patch + 1}
	default:
		return version
	}
}

func BumpVersion(message string, callback VersionBumpCallback) error {
	if callback == nil {
		callback = func(step, details string) {}
	}

	kind, err := readCommitType(message, callback)
	if err != nil {
		return err
	}

	callback("validation", "Reading current version")
	currentVersion, err := previousVersion()
	if err != nil {
		return err
	}
	callback("validation", fmt.Sprintf("Current version: %d.%d.%d", currentVersion.Major, currentVersion.Minor, currentVersion.Patch))

	bumpType := determineVersionBump(kind)
	if bumpType == "none" {
		_, hasPrevious, err := versionFromHead()
		if err != nil {
			return err
		}
		if !hasPrevious {
			_, staged, err := stagedVersionIfPresent()
			if err != nil {
				return err
			}
			if !staged {
				if err := gitAdd("VERSION"); err != nil {
					return err
				}
			}
		}
		callback("validation", "No version bump needed for this commit type")
		return nil
	}
	callback("validation", fmt.Sprintf("Bump type: %s", bumpType))

	newVersion := bumpVersionParts(currentVersion, bumpType)
	if newVersion.Major < 0 || newVersion.Minor < 0 || newVersion.Patch < 0 {
		return fmt.Errorf("version overflow")
	}
	if err := validateVersionFiles(currentVersion, newVersion); err != nil {
		return err
	}

	callback("bump", fmt.Sprintf("Bumping version to %d.%d.%d", newVersion.Major, newVersion.Minor, newVersion.Patch))
	if err := writeVersionFile("VERSION", newVersion); err != nil {
		return err
	}
	if err := gitAdd("VERSION"); err != nil {
		return err
	}
	callback("bump", "Version file updated and staged")
	return nil
}

func readCommitType(message string, callback VersionBumpCallback) (commitType, error) {
	subject, _, _ := strings.Cut(message, "\n")
	subject = strings.TrimSpace(subject)
	callback("validation", fmt.Sprintf("Commit message: %s", subject))

	commit := commitStr(subject)
	if valid, validationErrors := commit.validate(); !valid {
		return "", &invalidCommitMessageError{subject: subject, reasons: validationErrors}
	}
	kind, _, err := commit.extractChunks()
	if err != nil {
		return "", fmt.Errorf("parse commit message: %w", err)
	}
	callback("validation", fmt.Sprintf("Commit type: %s", kind))
	return kind, nil
}

func validateCommitVersion(messageFile string, callback VersionBumpCallback) error {
	callback("validation", "Reading commit message")
	data, err := os.ReadFile(messageFile)
	if err != nil {
		return fmt.Errorf("read commit message: %w", err)
	}
	kind, err := readCommitType(string(data), callback)
	if err != nil {
		return err
	}

	callback("validation", "Reading current version")
	currentVersion, err := previousVersion()
	if err != nil {
		return err
	}
	callback("validation", fmt.Sprintf("Current version: %d.%d.%d", currentVersion.Major, currentVersion.Minor, currentVersion.Patch))

	bumpType := determineVersionBump(kind)
	expected := bumpVersionParts(currentVersion, bumpType)
	if expected.Major < 0 || expected.Minor < 0 || expected.Patch < 0 {
		return fmt.Errorf("version overflow")
	}
	if err := validateVersionFiles(currentVersion, currentVersion); err != nil {
		return err
	}
	if bumpType == "none" {
		callback("validation", "No version bump needed for this commit type")
	} else {
		callback("validation", fmt.Sprintf("Version after commit: %d.%d.%d", expected.Major, expected.Minor, expected.Patch))
	}
	return nil
}

func validateVersionFiles(currentVersion, newVersion VersionParts) error {
	worktreeVersion, err := readVersionFile("VERSION")
	if err != nil {
		return err
	}

	stagedVersion, staged, err := stagedVersionIfPresent()
	if err != nil {
		return err
	}
	if !staged {
		_, hasPrevious, err := versionFromHead()
		if err != nil {
			return err
		}
		if hasPrevious {
			return fmt.Errorf("VERSION is missing from the index; resolve the staged deletion before committing")
		}
		stagedVersion = worktreeVersion
	}

	if (worktreeVersion != currentVersion && worktreeVersion != newVersion) ||
		(stagedVersion != currentVersion && stagedVersion != newVersion) {
		return fmt.Errorf("VERSION has unrelated changes: HEAD %d.%d.%d, staged %d.%d.%d, working tree %d.%d.%d; resolve them before committing",
			currentVersion.Major, currentVersion.Minor, currentVersion.Patch,
			stagedVersion.Major, stagedVersion.Minor, stagedVersion.Patch,
			worktreeVersion.Major, worktreeVersion.Minor, worktreeVersion.Patch)
	}
	return nil
}

func readStagedVersion() (VersionParts, error) {
	staged, err := exec.Command("git", "show", ":VERSION").Output()
	if err != nil {
		return VersionParts{}, fmt.Errorf("read staged VERSION: %w", err)
	}
	stagedVersion, err := parseVersion(string(staged))
	if err != nil {
		return VersionParts{}, fmt.Errorf("invalid staged VERSION: %w", err)
	}
	return stagedVersion, nil
}

func stagedVersionIfPresent() (VersionParts, bool, error) {
	output, err := exec.Command("git", "ls-files", "--stage", "-z", "--", "VERSION").Output()
	if err != nil {
		return VersionParts{}, false, fmt.Errorf("check staged VERSION: %w", err)
	}
	if len(output) == 0 {
		return VersionParts{}, false, nil
	}
	version, err := readStagedVersion()
	if err != nil {
		return VersionParts{}, false, err
	}
	return version, true, nil
}

func readInitialVersion() (VersionParts, error) {
	version, staged, err := stagedVersionIfPresent()
	if err != nil {
		return VersionParts{}, err
	}
	if staged {
		return version, nil
	}
	return readVersionFile("VERSION")
}

func previousVersion() (VersionParts, error) {
	version, found, err := versionFromHead()
	if err != nil {
		return VersionParts{}, err
	}
	if found {
		return version, nil
	}
	return readInitialVersion()
}

func versionFromHead() (VersionParts, bool, error) {
	hasHead, err := hasHeadCommit()
	if err != nil {
		return VersionParts{}, false, err
	}
	if !hasHead {
		return VersionParts{}, false, nil
	}
	paths, err := exec.Command("git", "ls-tree", "-z", "--name-only", "HEAD", "--", "VERSION").Output()
	if err != nil {
		return VersionParts{}, false, fmt.Errorf("check VERSION in HEAD: %w", err)
	}
	if len(paths) == 0 {
		return VersionParts{}, false, nil
	}
	data, err := exec.Command("git", "show", "HEAD:VERSION").Output()
	if err != nil {
		return VersionParts{}, false, fmt.Errorf("read VERSION from HEAD: %w", err)
	}
	version, err := parseVersion(string(data))
	if err != nil {
		return VersionParts{}, false, fmt.Errorf("invalid VERSION in HEAD: %w", err)
	}
	return version, true, nil
}

func hasHeadCommit() (bool, error) {
	command := exec.Command("git", "rev-parse", "--verify", "--quiet", "HEAD")
	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return false, fmt.Errorf("check HEAD: %w", err)
		}
		return false, nil
	}
	return true, nil
}
