package cmd

import (
	"os"

	"rto/backup"
)

// BackupResult is an alias for backward compatibility.
type BackupResult = backup.Result

// RunBackup performs a git backup of the data directory.
func RunBackup(remote, targetDir string) BackupResult {
	if targetDir == "" {
		if dir, err := os.Getwd(); err == nil {
			targetDir = dir
		}
	}
	return PerformGitBackup(targetDir, remote)
}

// PerformGitBackup delegates to the backup package.
func PerformGitBackup(dir, remote string) BackupResult {
	return backup.Perform(dir, remote)
}
