package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"rto/backup"
)

func hasGit(t *testing.T) bool {
	t.Helper()
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available, skipping test")
		return false
	}
	return true
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	cmd.Run()
}

func TestPerformGitBackupWrapper(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()
	gitInit(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result := PerformGitBackup(dir, "")
	if result.IsError {
		t.Errorf("expected success, got error: %s", result.Message)
	}
}

func TestRunBackupDefaultDir(t *testing.T) {
	// RunBackup with empty dir uses cwd — just verify it doesn't panic
	result := RunBackup("", "/nonexistent-path-that-wont-exist")
	_ = result
}

func TestBackupResultAlias(t *testing.T) {
	var r BackupResult = backup.Result{Message: "test", IsError: false}
	if r.Message != "test" {
		t.Error("alias should work")
	}
	if r.IsError {
		t.Error("expected no error")
	}
}

func TestPerformGitBackupNothingToCommit(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()
	gitInit(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	PerformGitBackup(dir, "")

	result := PerformGitBackup(dir, "")
	if result.IsError {
		t.Errorf("expected no error for nothing-to-commit, got: %s", result.Message)
	}
	if !strings.Contains(result.Message, "up to date") {
		t.Errorf("expected 'up to date' message, got: %s", result.Message)
	}
}
