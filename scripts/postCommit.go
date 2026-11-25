package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/trustpkg/trustpkg-cli-forge/pkg/logger"
)

func PostCommit(logHandler *logger.Logger) error {
	if err := logHandler.Print("Updating VERSION in the new commit", logger.SeverityStart, logger.WithTimestamp); err != nil {
		return err
	}
	if err := bumpLastCommit(versionBumpCallback(logHandler)); err != nil {
		logErr := logHandler.Print("Commit was created, but VERSION was not updated: "+err.Error(), logger.SeverityError, logger.WithTimestamp)
		if logErr != nil {
			fmt.Fprintln(os.Stderr, "Failed to write post-commit error:", logErr)
		}
		return err
	}
	return logHandler.Print("Post-commit version update completed", logger.SeveritySuccess, logger.WithTimestamp)
}

func bumpLastCommit(callback VersionBumpCallback) error {
	subject, err := exec.Command("git", "log", "-1", "--format=%s").Output()
	if err != nil {
		return fmt.Errorf("read new commit message: %w", err)
	}
	kind, err := readCommitType(strings.TrimSpace(string(subject)), callback)
	if err != nil {
		return err
	}
	if determineVersionBump(kind) == "none" {
		callback("validation", "No version bump needed for this commit type")
		return nil
	}

	head, err := exec.Command("git", "rev-parse", "--verify", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("read new commit hash: %w", err)
	}
	commit, err := exec.Command("git", "cat-file", "-p", strings.TrimSpace(string(head))).Output()
	if err != nil {
		return fmt.Errorf("read new commit: %w", err)
	}
	headers, _, _ := strings.Cut(string(commit), "\n\n")
	for _, line := range strings.Split(headers, "\n") {
		if strings.HasPrefix(line, "gpgsig ") || strings.HasPrefix(line, "gpgsig-sha256 ") {
			return fmt.Errorf("signed commit cannot be amended automatically; update VERSION and re-sign it manually")
		}
	}
	parents, err := exec.Command("git", "rev-list", "--parents", "-n", "1", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("read new commit parents: %w", err)
	}
	if len(strings.Fields(string(parents))) > 2 {
		return fmt.Errorf("merge commit cannot be amended automatically; update VERSION manually")
	}

	if err := BumpVersion(strings.TrimSpace(string(subject)), callback); err != nil {
		return err
	}
	currentHead, err := exec.Command("git", "rev-parse", "--verify", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("check new commit hash: %w", err)
	}
	if string(currentHead) != string(head) {
		return fmt.Errorf("HEAD changed while updating VERSION; inspect staged changes before retrying")
	}

	command := exec.Command("git", "commit", "--amend", "--no-edit", "--no-verify", "--only", "--", "VERSION")
	command.Env = append(os.Environ(), "TRUSTPKG_VERSION_AMEND=1")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("amend commit (VERSION remains staged): %w: %s", err, strings.TrimSpace(string(output)))
	}
	callback("bump", "VERSION added to amended commit")
	return nil
}
