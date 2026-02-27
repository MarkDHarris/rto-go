package backup

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

func setGitIdentity(t *testing.T, dir string) {
	t.Helper()
	_ = runGitSilent(dir, "config", "user.email", "test@test.com")
	_ = runGitSilent(dir, "config", "user.name", "Test")
}

func TestPerformNewRepo(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Init so we can set local config before Perform does it
	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)
	// Remove .git so Perform re-initializes
	os.RemoveAll(filepath.Join(dir, ".git"))

	result := Perform(dir, "")
	if result.IsError {
		t.Errorf("expected success, got error: %s", result.Message)
	}
	if !isGitRepo(dir) {
		t.Error("expected git repo to be initialized")
	}
}

func TestPerformNothingToCommit(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)

	Perform(dir, "")

	result := Perform(dir, "")
	if result.IsError {
		t.Errorf("expected no error for nothing-to-commit, got: %s", result.Message)
	}
	if !strings.Contains(result.Message, "up to date") {
		t.Errorf("expected 'up to date' message, got: %s", result.Message)
	}
}

func TestPerformCommitMessageFormat(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)

	result := Perform(dir, "")
	if result.IsError {
		t.Fatalf("backup failed: %s", result.Message)
	}

	out, err := runGitOutput(dir, "log", "-1", "--format=%s")
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	msg := strings.TrimSpace(out)
	if !strings.HasPrefix(msg, "backup: ") {
		t.Errorf("expected commit message starting with 'backup: ', got: %s", msg)
	}
	// Verify YYYY-MM-DD-HH-MM-SS-MMM format (e.g. backup: 2026-02-27-14-30-05-123)
	ts := strings.TrimPrefix(msg, "backup: ")
	parts := strings.Split(ts, "-")
	if len(parts) != 7 {
		t.Errorf("expected 7 dash-separated parts in timestamp %q, got %d", ts, len(parts))
	}
}

func TestIsGitRepo(t *testing.T) {
	dir := t.TempDir()
	if isGitRepo(dir) {
		t.Error("empty temp dir should not be a git repo")
	}

	if !hasGit(t) {
		return
	}
	_ = runGitSilent(dir, "init")
	if !isGitRepo(dir) {
		t.Error("should be a git repo after init")
	}
}

func TestRemoteExists(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()
	_ = runGitSilent(dir, "init")

	if remoteExists(dir) {
		t.Error("should not have remote before adding one")
	}

	remoteDir := t.TempDir()
	_ = runGitSilent(remoteDir, "init", "--bare")
	_ = runGitSilent(dir, "remote", "add", "origin", remoteDir)

	if !remoteExists(dir) {
		t.Error("should have remote after adding one")
	}
}

func TestPerformWithRemote(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()
	remoteDir := t.TempDir()

	if err := runGitSilent(remoteDir, "init", "--bare"); err != nil {
		t.Fatalf("init bare: %v", err)
	}

	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("backup"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := Perform(dir, remoteDir)
	if result.IsError {
		t.Errorf("expected success, got error: %s", result.Message)
	}
	if !strings.Contains(result.Message, "pushed") {
		t.Errorf("expected 'pushed' in message, got: %s", result.Message)
	}
}

func TestPerformWithExistingRemote(t *testing.T) {
	if !hasGit(t) {
		return
	}
	dir := t.TempDir()
	remoteDir := t.TempDir()

	_ = runGitSilent(remoteDir, "init", "--bare")
	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)
	_ = runGitSilent(dir, "remote", "add", "origin", remoteDir)

	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("backup"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Should detect existing remote and push
	result := Perform(dir, "")
	if result.IsError {
		t.Errorf("expected success, got error: %s", result.Message)
	}
	if !strings.Contains(result.Message, "pushed") {
		t.Errorf("expected push with existing remote, got: %s", result.Message)
	}
}

func TestStatus_NotARepo(t *testing.T) {
	dir := t.TempDir()
	info := Status(dir)
	if info.IsRepo {
		t.Error("expected IsRepo=false for non-git dir")
	}
}

func TestStatus_CleanRepo(t *testing.T) {
	dir := t.TempDir()
	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("init"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = runGitSilent(dir, "add", ".")
	_ = runGitSilent(dir, "commit", "-m", "initial")

	info := Status(dir)
	if !info.IsRepo {
		t.Error("expected IsRepo=true")
	}
	if !info.Clean {
		t.Errorf("expected Clean=true, got modified=%d untracked=%d", info.Modified, info.Untracked)
	}
	if info.LastCommit != "initial" {
		t.Errorf("expected LastCommit='initial', got %q", info.LastCommit)
	}
}

func TestStatus_ModifiedAndUntracked(t *testing.T) {
	dir := t.TempDir()
	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("init"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = runGitSilent(dir, "add", ".")
	_ = runGitSilent(dir, "commit", "-m", "initial")

	_ = os.WriteFile(filepath.Join(dir, "data.txt"), []byte("changed"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644)

	info := Status(dir)
	if info.Clean {
		t.Error("expected Clean=false")
	}
	if info.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", info.Modified)
	}
	if info.Untracked != 1 {
		t.Errorf("expected 1 untracked, got %d", info.Untracked)
	}
}

func TestStatus_WithRemote(t *testing.T) {
	dir := t.TempDir()
	remoteDir := t.TempDir()
	_ = runGitSilent(remoteDir, "init", "--bare")
	_ = runGitSilent(dir, "init")
	setGitIdentity(t, dir)
	_ = runGitSilent(dir, "remote", "add", "origin", remoteDir)

	info := Status(dir)
	if !info.HasRemote {
		t.Error("expected HasRemote=true")
	}
}
