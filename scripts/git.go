package scripts

import (
	"fmt"
	"os/exec"
)

func gitAdd(files ...string) error {
	if err := exec.Command("git", append([]string{"add"}, files...)...).Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	return nil
}

func gitCommit(message string) error {
	if err := exec.Command("git", "commit", "-m", message).Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

func gitCommitByPreviousOne() error {
	if err := exec.Command("git", "commit", "amend", "--no-edit").Run(); err != nil {
		return fmt.Errorf("git commit by previous one: %w", err)
	}

	return nil
}
