package app

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"testing"
)

func TestAddEntryWithTyping(t *testing.T) {
	// Create test file
	testFile := "/tmp/test_add_type.env"
	os.WriteFile(testFile, []byte("EXISTING=value\n"), 0644)
	defer os.Remove(testFile)

	// Create app and press 'a' to enter add mode
	m := New(testFile)
	mUpdate, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = mUpdate.(Model)

	fmt.Printf("After 'a' - viewMode: %d, editMode: %d\n", m.viewMode, m.editView.GetMode())
	fmt.Printf("Initial key='%s', value='%s'\n", m.editView.GetKey(), m.editView.GetValue())

	// Type "TESTKEY" - send each character
	testKey := "TESTKEY"
	for _, r := range testKey {
		mUpdate, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		_ = cmd
		m = mUpdate.(Model)
	}

	fmt.Printf("After typing key='%s' - key='%s', value='%s'\n", testKey, m.editView.GetKey(), m.editView.GetValue())

	// Press Tab to switch to value field
	mUpdate, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mUpdate.(Model)

	// Type "testvalue"
	testValue := "testvalue"
	for _, r := range testValue {
		mUpdate, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mUpdate.(Model)
	}

	fmt.Printf("After typing value='%s' - key='%s', value='%s'\n", testValue, m.editView.GetKey(), m.editView.GetValue())

	// Press Enter to save
	mUpdate, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mUpdate.(Model)

	envFile := m.GetCurrentEnvFile()
	fmt.Printf("After Enter - viewMode: %d, entries: %d\n", m.viewMode, len(envFile.Entries))

	// Check file content
	content, _ := os.ReadFile(testFile)
	fmt.Printf("File content:\n%s\n", string(content))

	// Verify the new entry was added
	found := false
	for _, e := range envFile.Entries {
		if e.Key == "TESTKEY" && e.Value == "testvalue" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("New entry TESTKEY=testvalue was not found")
	}
}

func TestTemplateKeyInEditMode(t *testing.T) {
	// Create test file
	testFile := "/tmp/test_template.env"
	os.WriteFile(testFile, []byte("EXISTING=value\n"), 0644)
	defer os.Remove(testFile)

	// Create app and press 'a' to enter add mode
	m := New(testFile)
	mUpdate, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = mUpdate.(Model)

	fmt.Printf("After 'a' - viewMode: %d\n", m.viewMode)

	// Press 't' to open template picker
	mUpdate, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = mUpdate.(Model)

	// Check if template mode is active
	view := m.View()
	fmt.Printf("After 't' - View output:\n%s\n", view)

	if !contains(view, "TEMPLATE") {
		t.Errorf("Template picker should be visible after pressing 't', got:\n%s", view)
	}
}

func TestCopyKeyWithMultipleFiles(t *testing.T) {
	// Create two test files
	testFile1 := "/tmp/test_copy1.env"
	testFile2 := "/tmp/test_copy2.env"
	os.WriteFile(testFile1, []byte("KEY1=value1\n"), 0644)
	os.WriteFile(testFile2, []byte("KEY2=value2\n"), 0644)
	defer os.Remove(testFile1)
	defer os.Remove(testFile2)

	// Create app with multiple files
	m := NewMultiFile([]string{testFile1, testFile2})

	// Set window size so view renders properly
	mUpdate, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdate.(Model)

	fmt.Printf("Initial state - viewMode: %d, files: %d\n", m.viewMode, len(m.envFiles))

	// Press 'y' to enter copy mode
	mUpdate, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = mUpdate.(Model)

	// Check if copy mode is active
	view := m.View()
	fmt.Printf("After 'y' - View output:\n%s\n", view)

	if !contains(view, "COPY") {
		t.Errorf("Copy mode banner should be visible after pressing 'y', got:\n%s", view)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
