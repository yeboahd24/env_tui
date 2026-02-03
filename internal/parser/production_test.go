package parser

import (
	"testing"
)

func TestProductionParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantVals []string
	}{
		{
			name:     "simple key-value",
			input:    "KEY=value",
			wantKeys: []string{"KEY"},
			wantVals: []string{"value"},
		},
		{
			name:     "quoted value",
			input:    `KEY="value with spaces"`,
			wantKeys: []string{"KEY"},
			wantVals: []string{"value with spaces"},
		},
		{
			name:     "escape sequences",
			input:    `KEY="line1\nline2\ttab"`,
			wantKeys: []string{"KEY"},
			wantVals: []string{"line1\nline2\ttab"},
		},
		{
			name:     "export keyword",
			input:    "export NODE_ENV=production",
			wantKeys: []string{"NODE_ENV"},
			wantVals: []string{"production"},
		},
		{
			name:     "empty value",
			input:    "KEY=",
			wantKeys: []string{"KEY"},
			wantVals: []string{""},
		},
		{
			name:     "comments and blanks",
			input:    "# Comment\n\nKEY=value",
			wantKeys: []string{"KEY"},
			wantVals: []string{"value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			kvEntries := envFile.FilterEntries("")
			if len(kvEntries) != len(tt.wantKeys) {
				t.Fatalf("got %d entries, want %d", len(kvEntries), len(tt.wantKeys))
			}

			for i, entry := range kvEntries {
				if entry.Key != tt.wantKeys[i] {
					t.Errorf("entry[%d].Key = %q, want %q", i, entry.Key, tt.wantKeys[i])
				}
				if entry.Value != tt.wantVals[i] {
					t.Errorf("entry[%d].Value = %q, want %q", i, entry.Value, tt.wantVals[i])
				}
			}
		})
	}
}

func TestValidation(t *testing.T) {
	input := `KEY1=value1
KEY1=value2
SECRET_KEY=changeme`

	envFile, _ := Parse(input)
	issues := envFile.Validate()

	if len(issues) == 0 {
		t.Error("expected validation issues, got none")
	}

	// Should detect duplicate key
	foundDuplicate := false
	for _, issue := range issues {
		if issue.Key == "KEY1" {
			foundDuplicate = true
		}
	}

	if !foundDuplicate {
		t.Error("expected duplicate key validation issue")
	}
}