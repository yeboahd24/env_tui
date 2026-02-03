package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/envtui/envtui/internal/model"
)

// ExportToShell exports env file entries as shell commands
func ExportToShell(envFile *model.EnvFile, exportFormat string) string {
	var sb strings.Builder

	for _, entry := range envFile.Entries {
		if entry.Type != model.KeyValueEntry {
			continue
		}

		// Escape special characters in value
		value := escapeShellValue(entry.Value)

		if entry.Exported || exportFormat == "export" {
			// export KEY=value format
			sb.WriteString(fmt.Sprintf("export %s=%s\n", entry.Key, value))
		} else {
			// KEY=value format (for sourcing)
			sb.WriteString(fmt.Sprintf("%s=%s\n", entry.Key, value))
		}
	}

	return sb.String()
}

// escapeShellValue escapes a value for safe shell usage
func escapeShellValue(value string) string {
	// If value contains spaces or special characters, wrap in quotes
	if strings.ContainsAny(value, " \\t\"'$()<>|&;`") {
		// Escape double quotes
		value = strings.ReplaceAll(value, "\"", "\\\"")
		// Wrap in double quotes
		return fmt.Sprintf("\"%s\"", value)
	}
	return value
}

// GenerateShellAlias generates shell alias/function for easy envtui usage
func GenerateShellAlias() string {
	return `# EnvTUI Shell Integration
# Add this to your .bashrc, .zshrc, or .bash_profile

# Function to load env file and launch envtui
envtui() {
    local files=".env"
    
    # Check if files argument provided
    if [ -n "$1" ]; then
        files="$1"
    fi
    
    # Launch envtui with files
    /path/to/envtui --files "$files"
}

# Export environment variables from env file
envtui-export() {
    local file=".env"
    
    if [ -n "$1" ]; then
        file="$1"
    fi
    
    eval $(/path/to/envtui --files "$file" --export - --format shell)
}

# Quick alias for common operations
alias et='envtui'
alias et-export='envtui-export'
`
}

// PrintShellCompletion prints shell completion scripts
func PrintShellCompletion(shell string) string {
	switch shell {
	case "bash":
		return generateBashCompletion()
	case "zsh":
		return generateZshCompletion()
	case "fish":
		return generateFishCompletion()
	default:
		return "# Shell completion not available for: " + shell + "\n# Supported shells: bash, zsh, fish\n"
	}
}

func generateBashCompletion() string {
	return `_envtui_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="--files --export --format --import --merge --overwrite --help"
    
    case "${prev}" in
        --files)
            COMPREPLY=( $(compgen -f -X '!*.env' -- "${cur}") )
            return 0
            ;;
        --format)
            COMPREPLY=( $(compgen -W "json yaml shell" -- "${cur}") )
            return 0
            ;;
        *)
            ;;
    esac
    
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
    return 0
}

complete -F _envtui_completions envtui
`
}

func generateZshCompletion() string {
	return `#compdef envtui

_arguments \
    '--files[Comma-separated env files]:files:_files -g "*.env"' \
    '--export[Export to file]:output file:_files' \
    '--format[Export format]:format:(json yaml shell)' \
    '--import[Import from file]:input file:_files -g "*.{json,yaml,yml}"' \
    '--merge[Merge imported entries]' \
    '--overwrite[Overwrite existing entries when importing]' \
    '--help[Show help]'
`
}

func generateFishCompletion() string {
	return `complete -c envtui -l files -d "Comma-separated env files" -r -F
complete -c envtui -l export -d "Export to file" -r -F
complete -c envtui -l format -d "Export format" -x -a "json yaml shell"
complete -c envtui -l import -d "Import from file" -r -F
complete -c envtui -l merge -d "Merge imported entries"
complete -c envtui -l overwrite -d "Overwrite existing entries"
complete -c envtui -l help -d "Show help"
`
}

// SaveShellIntegration saves shell integration script to a file
func SaveShellIntegration(outputPath string) error {
	content := GenerateShellAlias()
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// GetShellConfigPath returns the appropriate shell config file path
func GetShellConfigPath() string {
	shell := os.Getenv("SHELL")
	home := os.Getenv("HOME")

	if strings.Contains(shell, "zsh") {
		return filepath.Join(home, ".zshrc")
	} else if strings.Contains(shell, "bash") {
		// Check for .bash_profile first (macOS), then .bashrc
		bashProfile := filepath.Join(home, ".bash_profile")
		if _, err := os.Stat(bashProfile); err == nil {
			return bashProfile
		}
		return filepath.Join(home, ".bashrc")
	} else if strings.Contains(shell, "fish") {
		return filepath.Join(home, ".config", "fish", "config.fish")
	}

	return ""
}
