package model

import (
	"fmt"
	"strings"
)

type ValidationLevel int

const (
	ValidationError ValidationLevel = iota
	ValidationWarning
	ValidationInfo
)

type ValidationIssue struct {
	Level   ValidationLevel
	Message string
	Line    int
	Key     string
}

func (e *Entry) Validate() []ValidationIssue {
	var issues []ValidationIssue
	
	if e.Type != KeyValueEntry {
		return issues
	}
	
	// Check for empty key
	if e.Key == "" {
		issues = append(issues, ValidationIssue{
			Level:   ValidationError,
			Message: "Key cannot be empty",
			Line:    e.Line,
		})
	}
	
	// Check for spaces in unquoted values
	if strings.Contains(e.Value, " ") && !e.Exported {
		issues = append(issues, ValidationIssue{
			Level:   ValidationWarning,
			Message: fmt.Sprintf("Value contains spaces, consider quoting: %s", e.Key),
			Line:    e.Line,
			Key:     e.Key,
		})
	}
	
	// Check for suspicious patterns
	if e.IsSecret && (e.Value == "" || e.Value == "changeme" || e.Value == "password") {
		issues = append(issues, ValidationIssue{
			Level:   ValidationWarning,
			Message: fmt.Sprintf("Suspicious secret value: %s", e.Key),
			Line:    e.Line,
			Key:     e.Key,
		})
	}
	
	// Check for duplicate keys (requires context from EnvFile)
	
	return issues
}

func (ef *EnvFile) Validate() []ValidationIssue {
	var issues []ValidationIssue
	keysSeen := make(map[string]int)
	
	for _, entry := range ef.Entries {
		// Validate individual entry
		issues = append(issues, entry.Validate()...)
		
		// Check for duplicates
		if entry.Type == KeyValueEntry {
			if prevLine, exists := keysSeen[entry.Key]; exists {
				issues = append(issues, ValidationIssue{
					Level:   ValidationError,
					Message: fmt.Sprintf("Duplicate key '%s' (first seen at line %d)", entry.Key, prevLine),
					Line:    entry.Line,
					Key:     entry.Key,
				})
			}
			keysSeen[entry.Key] = entry.Line
		}
	}
	
	return issues
}