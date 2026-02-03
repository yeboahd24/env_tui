package parser

import (
	"fmt"
	"strings"

	"github.com/envtui/envtui/internal/model"
)

type Parser struct {
	lexer        *Lexer
	currentToken Token
	peekToken    Token
}

func NewParser(input string) *Parser {
	p := &Parser{lexer: NewLexer(input)}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) Parse() (*model.EnvFile, error) {
	envFile := &model.EnvFile{
		Entries: make([]*model.Entry, 0),
	}

	for p.currentToken.Type != EOF {
		switch p.currentToken.Type {
		case EXPORT:
			p.nextToken()
			if p.currentToken.Type == KEY {
				entry, err := p.parseEntry(true)
				if err != nil {
					return nil, err
				}
				envFile.Entries = append(envFile.Entries, entry)
			} else {
				p.nextToken()
			}
		case KEY:
			entry, err := p.parseEntry(false)
			if err != nil {
				return nil, err
			}
			envFile.Entries = append(envFile.Entries, entry)
		case COMMENT:
			entry := &model.Entry{
				Type:    model.CommentEntry,
				Comment: p.currentToken.Value,
				Line:    p.currentToken.Line,
			}
			envFile.Entries = append(envFile.Entries, entry)
			p.nextToken()
		case NEWLINE:
			entry := &model.Entry{
				Type: model.BlankEntry,
				Line: p.currentToken.Line,
			}
			envFile.Entries = append(envFile.Entries, entry)
			p.nextToken()
		default:
			p.nextToken()
		}
	}

	return envFile, nil
}

func (p *Parser) parseEntry(exported bool) (*model.Entry, error) {
	if p.currentToken.Type != KEY {
		return nil, fmt.Errorf("expected key, got %v", p.currentToken.Type)
	}

	key := p.currentToken.Value
	line := p.currentToken.Line
	
	// Read the value directly from lexer
	value := p.lexer.ReadValue()
	
	// Resync parser tokens with lexer
	p.currentToken = p.lexer.NextToken()
	p.peekToken = p.lexer.NextToken()

	var inlineComment string
	// Check for inline comment
	if p.currentToken.Type == COMMENT {
		inlineComment = p.currentToken.Value
		p.nextToken()
	}

	entry := &model.Entry{
		Type:          model.KeyValueEntry,
		Key:           key,
		Value:         value,
		Comment:       inlineComment,
		Line:          line,
		Exported:      exported,
		IsSecret:      isSecretKey(key),
	}

	return entry, nil
}

func isSecretKey(key string) bool {
	secretKeywords := []string{
		"PASSWORD", "SECRET", "TOKEN", "KEY", "PRIVATE",
		"API_KEY", "AUTH", "CREDENTIAL", "CERT",
	}
	
	upperKey := strings.ToUpper(key)
	for _, keyword := range secretKeywords {
		if strings.Contains(upperKey, keyword) {
			return true
		}
	}
	return false
}

func IsSecretKey(key string) bool {
	return isSecretKey(key)
}