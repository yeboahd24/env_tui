package storage

import (
	"fmt"
	"os"

	"github.com/envtui/envtui/internal/model"
	"github.com/envtui/envtui/internal/parser"
)

func ReadFile(path string) (*model.EnvFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	envFile, err := parser.Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	envFile.Path = path
	return envFile, nil
}

func WriteFile(envFile *model.EnvFile) error {
	// Create backup first
	if err := createBackup(envFile.Path); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write to temporary file
	tempPath := envFile.Path + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Write content
	for _, entry := range envFile.Entries {
		if _, err := tempFile.WriteString(entry.String() + "\n"); err != nil {
			return fmt.Errorf("failed to write entry: %w", err)
		}
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, envFile.Path); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func createBackup(path string) error {
	return CreateBackup(path)
}
