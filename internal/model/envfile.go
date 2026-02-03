package model

import (
	"path/filepath"
	"strings"
)

func (ef *EnvFile) GetEntry(key string) *Entry {
	for _, entry := range ef.Entries {
		if entry.Type == KeyValueEntry && entry.Key == key {
			return entry
		}
	}
	return nil
}

func (ef *EnvFile) AddEntry(entry *Entry) {
	ef.Entries = append(ef.Entries, entry)
}

func (ef *EnvFile) UpdateEntry(key, value string) bool {
	for _, entry := range ef.Entries {
		if entry.Type == KeyValueEntry && entry.Key == key {
			entry.Value = value
			return true
		}
	}
	return false
}

func (ef *EnvFile) DeleteEntry(key string) bool {
	for i, entry := range ef.Entries {
		if entry.Type == KeyValueEntry && entry.Key == key {
			ef.Entries = append(ef.Entries[:i], ef.Entries[i+1:]...)
			return true
		}
	}
	return false
}

func (ef *EnvFile) FilterEntries(query string) []*Entry {
	var kvEntries []*Entry
	for _, entry := range ef.Entries {
		if entry.Type == KeyValueEntry {
			kvEntries = append(kvEntries, entry)
		}
	}

	if query == "" {
		return kvEntries
	}

	query = strings.ToLower(query)
	var filtered []*Entry

	for _, entry := range kvEntries {
		// Simple fuzzy matching: check if all characters in query appear in order
		key := strings.ToLower(entry.Key)
		value := strings.ToLower(entry.Value)

		if fuzzyMatch(key, query) || fuzzyMatch(value, query) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func fuzzyMatch(text, pattern string) bool {
	if pattern == "" {
		return true
	}

	pi := 0 // pattern index
	for ti := 0; ti < len(text) && pi < len(pattern); ti++ {
		if text[ti] == pattern[pi] {
			pi++
		}
	}

	return pi == len(pattern)
}

// FileDiff represents a comparison between two env files
type FileDiff struct {
	Key           string
	CurrentValue  string
	OtherValue    string
	OnlyInCurrent bool
	OnlyInOther   bool
	Different     bool
}

// CompareWith compares current env file with another and returns differences
type EnvFileCompare struct {
	OtherFile       string
	Differences     []FileDiff
	TotalKeys       int
	MatchingKeys    int
	DifferentValues int
	OnlyInCurrent   int
	OnlyInOther     int
}

// CompareWith compares this env file with another env file
func (ef *EnvFile) CompareWith(other *EnvFile) *EnvFileCompare {
	compare := &EnvFileCompare{
		OtherFile: filepath.Base(other.Path),
	}

	currentEntries := make(map[string]string)
	otherEntries := make(map[string]string)

	for _, entry := range ef.Entries {
		if entry.Type == KeyValueEntry {
			currentEntries[entry.Key] = entry.Value
			compare.TotalKeys++
		}
	}

	for _, entry := range other.Entries {
		if entry.Type == KeyValueEntry {
			otherEntries[entry.Key] = entry.Value
		}
	}

	// Check all keys from both files
	allKeys := make(map[string]bool)
	for key := range currentEntries {
		allKeys[key] = true
	}
	for key := range otherEntries {
		allKeys[key] = true
	}

	for key := range allKeys {
		diff := FileDiff{Key: key}

		currentVal, hasCurrent := currentEntries[key]
		otherVal, hasOther := otherEntries[key]

		if hasCurrent && hasOther {
			diff.CurrentValue = currentVal
			diff.OtherValue = otherVal
			if currentVal != otherVal {
				diff.Different = true
				compare.DifferentValues++
			} else {
				compare.MatchingKeys++
			}
		} else if hasCurrent {
			diff.CurrentValue = currentVal
			diff.OnlyInCurrent = true
			compare.OnlyInCurrent++
		} else {
			diff.OtherValue = otherVal
			diff.OnlyInOther = true
			compare.OnlyInOther++
		}

		compare.Differences = append(compare.Differences, diff)
	}

	return compare
}

// HasDifferences returns true if there are any differences with the other file
func (ec *EnvFileCompare) HasDifferences() bool {
	return ec.DifferentValues > 0 || ec.OnlyInCurrent > 0 || ec.OnlyInOther > 0
}
