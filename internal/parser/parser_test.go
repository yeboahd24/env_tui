package parser

import (
	"testing"
	"github.com/envtui/envtui/internal/model"
)

func TestLexer(t *testing.T) {
	input := `KEY=value`

	lexer := NewLexer(input)
	
	// Test basic key-value parsing
	token1 := lexer.NextToken()
	if token1.Type != KEY || token1.Value != "KEY" {
		t.Errorf("expected KEY token with value 'KEY', got %v with value '%s'", token1.Type, token1.Value)
	}

	// The value should be read separately by the parser
	value := lexer.ReadValue()
	if value != "value" {
		t.Errorf("expected value 'value', got '%s'", value)
	}
}

func TestParser(t *testing.T) {
	input := `# Database config
DB_HOST=localhost
DB_PASSWORD=secret123
export NODE_ENV=development`

	parser := NewParser(input)
	envFile, err := parser.Parse()
	
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	// Find key-value entries
	var kvEntries []*model.Entry
	for _, entry := range envFile.Entries {
		if entry.Type == model.KeyValueEntry {
			kvEntries = append(kvEntries, entry)
		}
	}

	if len(kvEntries) != 3 {
		t.Fatalf("expected 3 key-value entries, got %d", len(kvEntries))
	}

	// Check entries
	dbHost := kvEntries[0]
	if dbHost.Key != "DB_HOST" || dbHost.Value != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %s=%s", dbHost.Key, dbHost.Value)
	}

	// Check secret detection
	dbPassword := kvEntries[1]
	if !dbPassword.IsSecret {
		t.Errorf("expected DB_PASSWORD to be detected as secret")
	}

	// Check export
	nodeEnv := kvEntries[2]
	if !nodeEnv.Exported {
		t.Errorf("expected NODE_ENV to be exported")
	}
}