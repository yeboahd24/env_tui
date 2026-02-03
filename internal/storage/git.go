package storage

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitStatus represents the git status of a file
type GitStatus int

const (
	GitStatusNone GitStatus = iota
	GitStatusUntracked
	GitStatusModified
	GitStatusStaged
	GitStatusClean
)

func (gs GitStatus) String() string {
	switch gs {
	case GitStatusUntracked:
		return "untracked"
	case GitStatusModified:
		return "modified"
	case GitStatusStaged:
		return "staged"
	case GitStatusClean:
		return "clean"
	default:
		return ""
	}
}

// IsGitRepository checks if the given path is in a git repository
func IsGitRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = filepath.Dir(path)
	err := cmd.Run()
	return err == nil
}

// GetGitStatus returns the git status of a file
func GetGitStatus(path string) GitStatus {
	if !IsGitRepository(path) {
		return GitStatusNone
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	// Check if file is tracked
	cmd := exec.Command("git", "ls-files", "--error-unmatch", base)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// File is not tracked (untracked)
		return GitStatusUntracked
	}

	// Check status
	cmd = exec.Command("git", "status", "--porcelain", base)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return GitStatusNone
	}

	status := strings.TrimSpace(string(output))
	if status == "" {
		return GitStatusClean
	}

	// Parse porcelain output
	// Format: XY filename where X is index status and Y is working tree status
	if len(status) >= 2 {
		indexStatus := status[0]
		workTreeStatus := status[1]

		if indexStatus != ' ' && indexStatus != '?' {
			return GitStatusStaged
		}
		if workTreeStatus == 'M' || workTreeStatus == 'D' {
			return GitStatusModified
		}
	}

	return GitStatusClean
}

// GetGitStatusIcon returns an icon representing the git status
func GetGitStatusIcon(status GitStatus) string {
	switch status {
	case GitStatusUntracked:
		return "?"
	case GitStatusModified:
		return "M"
	case GitStatusStaged:
		return "S"
	case GitStatusClean:
		return "✓"
	default:
		return ""
	}
}

// GetGitBranch returns the current git branch
func GetGitBranch(path string) string {
	if !IsGitRepository(path) {
		return ""
	}

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = filepath.Dir(path)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// FileGitInfo holds git information for a file
type FileGitInfo struct {
	Status GitStatus
	Branch string
	Icon   string
}

// GetFileGitInfo returns complete git information for a file
func GetFileGitInfo(path string) FileGitInfo {
	status := GetGitStatus(path)
	return FileGitInfo{
		Status: status,
		Branch: GetGitBranch(path),
		Icon:   GetGitStatusIcon(status),
	}
}

// FormatGitStatusForTab returns a formatted string for file tabs
func FormatGitStatusForTab(status GitStatus) string {
	icon := GetGitStatusIcon(status)
	if icon == "" {
		return ""
	}
	return fmt.Sprintf(" [%s]", icon)
}
