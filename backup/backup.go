package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Result holds the outcome of a git backup operation.
type Result struct {
	Message string
	IsError bool
}

// Perform executes the git backup workflow:
// 1. Initialize repo if needed
// 2. Configure remote
// 3. Stage all files
// 4. Commit with timestamp
// 5. Push if remote configured
func Perform(dir, remote string) Result {
	isRepo := isGitRepo(dir)

	if !isRepo {
		if err := runGitSilent(dir, "init"); err != nil {
			return Result{Message: fmt.Sprintf("git init failed: %v", err), IsError: true}
		}
		_ = runGitSilent(dir, "checkout", "-b", "main")
	}

	if remote != "" {
		if remoteExists(dir) {
			if err := runGitSilent(dir, "remote", "set-url", "origin", remote); err != nil {
				return Result{Message: fmt.Sprintf("set-url failed: %v", err), IsError: true}
			}
		} else {
			if err := runGitSilent(dir, "remote", "add", "origin", remote); err != nil {
				return Result{Message: fmt.Sprintf("remote add failed: %v", err), IsError: true}
			}
		}
	}

	if err := runGitSilent(dir, "add", "."); err != nil {
		return Result{Message: fmt.Sprintf("git add failed: %v", err), IsError: true}
	}

	now := time.Now()
	commitMsg := fmt.Sprintf("backup: %s-%03d", now.Format("2006-01-02-15-04-05"), now.Nanosecond()/1e6)
	if err := runGitSilent(dir, "commit", "-m", commitMsg); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "nothing to commit") || strings.Contains(errStr, "nothing added") {
			return Result{Message: "Nothing to commit — backup up to date", IsError: false}
		}
		return Result{Message: fmt.Sprintf("git commit failed: %v", err), IsError: true}
	}

	if remote != "" || remoteExists(dir) {
		if err := runGitSilent(dir, "push", "origin", "main"); err != nil {
			return Result{Message: fmt.Sprintf("Committed (push failed: %v)", err), IsError: false}
		}
		return Result{Message: "Backup committed and pushed", IsError: false}
	}

	return Result{Message: "Backup committed (no remote configured)", IsError: false}
}

// StatusInfo holds a summary of the git state for a data directory.
type StatusInfo struct {
	IsRepo     bool
	HasRemote  bool
	Modified   int
	Untracked  int
	LastCommit string
	Clean      bool
}

// Status checks the git state of a directory and returns a summary.
func Status(dir string) StatusInfo {
	info := StatusInfo{}

	if !isGitRepo(dir) {
		return info
	}
	info.IsRepo = true
	info.HasRemote = remoteExists(dir)

	out, err := runGitOutput(dir, "status", "--porcelain")
	if err != nil {
		return info
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		if line[:2] == "??" {
			info.Untracked++
		} else {
			info.Modified++
		}
	}

	info.Clean = info.Modified == 0 && info.Untracked == 0

	commitOut, err := runGitOutput(dir, "log", "-1", "--format=%s")
	if err == nil {
		info.LastCommit = strings.TrimSpace(commitOut)
	}

	return info
}

func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

func remoteExists(dir string) bool {
	out, err := runGitOutput(dir, "remote")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "origin" {
			return true
		}
	}
	return false
}

func runGitSilent(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
