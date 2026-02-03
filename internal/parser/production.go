package parser

import (
	"strings"
	"unicode"

	"github.com/envtui/envtui/internal/model"
)

func Parse(input string) (*model.EnvFile, error) {
	envFile := &model.EnvFile{Entries: make([]*model.Entry, 0)}
	lines := strings.Split(input, "\n")
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		
		// Blank line
		if trimmed == "" {
			envFile.Entries = append(envFile.Entries, &model.Entry{
				Type: model.BlankEntry,
				Line: i + 1,
			})
			continue
		}
		
		// Comment
		if strings.HasPrefix(trimmed, "#") {
			envFile.Entries = append(envFile.Entries, &model.Entry{
				Type:    model.CommentEntry,
				Comment: line,
				Line:    i + 1,
			})
			continue
		}
		
		// Handle export
		exported := false
		if strings.HasPrefix(trimmed, "export ") {
			exported = true
			trimmed = strings.TrimPrefix(trimmed, "export ")
			trimmed = strings.TrimSpace(trimmed)
		}
		
		// Key=Value
		eqIdx := strings.Index(trimmed, "=")
		if eqIdx == -1 {
			continue // Skip invalid lines
		}
		
		key := strings.TrimSpace(trimmed[:eqIdx])
		if key == "" || !isValidKey(key) {
			continue // Skip invalid keys
		}
		
		valueStr := trimmed[eqIdx+1:]
		value, consumed := parseValue(valueStr, lines, i)
		i += consumed // Skip consumed lines for multiline values
		
		envFile.Entries = append(envFile.Entries, &model.Entry{
			Type:     model.KeyValueEntry,
			Key:      key,
			Value:    value,
			Line:     i + 1,
			Exported: exported,
			IsSecret: isSecretKey(key),
		})
	}
	
	return envFile, nil
}

func parseValue(valueStr string, lines []string, currentLine int) (string, int) {
	valueStr = strings.TrimSpace(valueStr)
	
	// Empty value
	if valueStr == "" {
		return "", 0
	}
	
	// Quoted value (single or double)
	if len(valueStr) > 0 && (valueStr[0] == '"' || valueStr[0] == '\'') {
		quote := valueStr[0]
		return parseQuotedValue(valueStr, quote, lines, currentLine)
	}
	
	// Unquoted value - read until comment or end
	if idx := strings.Index(valueStr, "#"); idx != -1 {
		valueStr = strings.TrimSpace(valueStr[:idx])
	}
	
	return valueStr, 0
}

func parseQuotedValue(valueStr string, quote byte, lines []string, currentLine int) (string, int) {
	var result strings.Builder
	i := 1 // Skip opening quote
	linesConsumed := 0
	currentLineStr := valueStr
	
	for {
		for i < len(currentLineStr) {
			ch := currentLineStr[i]
			
			if ch == '\\' && i+1 < len(currentLineStr) {
				// Handle escape sequences
				next := currentLineStr[i+1]
				switch next {
				case 'n':
					result.WriteByte('\n')
				case 't':
					result.WriteByte('\t')
				case 'r':
					result.WriteByte('\r')
				case '\\':
					result.WriteByte('\\')
				case '"', '\'':
					result.WriteByte(next)
				default:
					result.WriteByte(next)
				}
				i += 2
				continue
			}
			
			if ch == quote {
				return result.String(), linesConsumed
			}
			
			result.WriteByte(ch)
			i++
		}
		
		// Multiline value - continue to next line
		if currentLine+linesConsumed+1 < len(lines) {
			linesConsumed++
			currentLineStr = lines[currentLine+linesConsumed]
			result.WriteByte('\n')
			i = 0
		} else {
			break
		}
	}
	
	return result.String(), linesConsumed
}

func isValidKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	
	for i, ch := range key {
		if i == 0 && !unicode.IsLetter(ch) && ch != '_' {
			return false
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}
	
	return true
}