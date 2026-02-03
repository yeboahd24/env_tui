package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/envtui/envtui/internal/model"
)

// ExportFormat represents the format for export/import
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatYAML ExportFormat = "yaml"
)

// ExportEntry represents a single entry for export
type ExportEntry struct {
	Key      string `json:"key" yaml:"key"`
	Value    string `json:"value" yaml:"value"`
	Exported bool   `json:"exported,omitempty" yaml:"exported,omitempty"`
	IsSecret bool   `json:"is_secret,omitempty" yaml:"is_secret,omitempty"`
}

// ExportData represents the full export structure
type ExportData struct {
	File    string        `json:"file" yaml:"file"`
	Entries []ExportEntry `json:"entries" yaml:"entries"`
	Count   int           `json:"count" yaml:"count"`
}

// ExportToFile exports an EnvFile to JSON or YAML format
func ExportToFile(envFile *model.EnvFile, format ExportFormat, outputPath string) error {
	data := ExportData{
		File:  envFile.Path,
		Count: 0,
	}

	for _, entry := range envFile.Entries {
		if entry.Type == model.KeyValueEntry {
			data.Entries = append(data.Entries, ExportEntry{
				Key:      entry.Key,
				Value:    entry.Value,
				Exported: entry.Exported,
				IsSecret: entry.IsSecret,
			})
			data.Count++
		}
	}

	var content []byte
	var err error

	switch format {
	case FormatJSON:
		content, err = json.MarshalIndent(data, "", "  ")
	case FormatYAML:
		content = []byte(exportToYAML(data))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(outputPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// exportToYAML converts ExportData to YAML format manually
func exportToYAML(data ExportData) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("file: %s\n", data.File))
	sb.WriteString(fmt.Sprintf("count: %d\n", data.Count))
	sb.WriteString("entries:\n")

	for _, entry := range data.Entries {
		sb.WriteString("  - key: " + entry.Key + "\n")
		sb.WriteString("    value: " + entry.Value + "\n")
		if entry.Exported {
			sb.WriteString("    exported: true\n")
		}
		if entry.IsSecret {
			sb.WriteString("    is_secret: true\n")
		}
	}

	return sb.String()
}

// ImportFromFile imports entries from a JSON file
func ImportFromFile(inputPath string) (*model.EnvFile, error) {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data ExportData
	ext := strings.ToLower(filepath.Ext(inputPath))

	switch ext {
	case ".json":
		err = json.Unmarshal(content, &data)
	case ".yaml", ".yml":
		return nil, fmt.Errorf("YAML import not yet implemented - please use JSON format")
	default:
		// Try JSON format
		err = json.Unmarshal(content, &data)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Create EnvFile from imported data
	envFile := &model.EnvFile{
		Path:    data.File,
		Entries: make([]*model.Entry, 0, len(data.Entries)),
	}

	for _, expEntry := range data.Entries {
		entry := &model.Entry{
			Type:     model.KeyValueEntry,
			Key:      expEntry.Key,
			Value:    expEntry.Value,
			Exported: expEntry.Exported,
			IsSecret: expEntry.IsSecret,
		}
		envFile.Entries = append(envFile.Entries, entry)
	}

	return envFile, nil
}

// MergeImport merges imported entries with existing env file
func MergeImport(envFile *model.EnvFile, imported *model.EnvFile, overwrite bool) error {
	for _, importedEntry := range imported.Entries {
		if importedEntry.Type != model.KeyValueEntry {
			continue
		}

		existing := envFile.GetEntry(importedEntry.Key)
		if existing == nil {
			// Entry doesn't exist, add it
			envFile.AddEntry(&model.Entry{
				Type:     model.KeyValueEntry,
				Key:      importedEntry.Key,
				Value:    importedEntry.Value,
				Exported: importedEntry.Exported,
				IsSecret: importedEntry.IsSecret,
			})
		} else if overwrite {
			// Entry exists, update if overwrite is true
			existing.Value = importedEntry.Value
			existing.Exported = importedEntry.Exported
			existing.IsSecret = importedEntry.IsSecret
		}
	}

	return nil
}
