package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupInfo holds information about a backup file
type BackupInfo struct {
	Path      string
	Timestamp time.Time
	Size      int64
}

// ListBackups returns a list of backup files for the given env file
func ListBackups(path string) ([]BackupInfo, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	pattern := base + ".backup.*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}

	var backups []BackupInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		// Parse timestamp from filename
		timestamp, err := parseBackupTimestamp(match)
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      match,
			Timestamp: timestamp,
			Size:      info.Size(),
		})
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// parseBackupTimestamp extracts the timestamp from a backup filename
func parseBackupTimestamp(path string) (time.Time, error) {
	base := filepath.Base(path)
	parts := strings.Split(base, ".backup.")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid backup filename format")
	}

	timestamp := parts[1]
	return time.Parse("20060102-150405", timestamp)
}

// RestoreBackup restores a backup file to the original env file
func RestoreBackup(backupPath, originalPath string) error {
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer src.Close()

	// Create a backup of the current file first (just in case)
	if _, err := os.Stat(originalPath); err == nil {
		timestamp := time.Now().Format("20060102-150405")
		safetyBackupPath := fmt.Sprintf("%s.backup.pre-restore.%s", originalPath, timestamp)
		if err := copyFile(originalPath, safetyBackupPath); err != nil {
			return fmt.Errorf("failed to create safety backup: %w", err)
		}
	}

	dst, err := os.Create(originalPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy backup: %w", err)
	}

	return nil
}

// DeleteBackup removes a backup file
func DeleteBackup(backupPath string) error {
	return os.Remove(backupPath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// CreateBackup creates a backup of the given file
func CreateBackup(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // no file to backup
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", path, timestamp)

	return copyFile(path, backupPath)
}
