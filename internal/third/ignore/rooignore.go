package ignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/denormal/go-gitignore" // Using a library
)

const LOCK_TEXT_SYMBOL = "🔒" // Or use a different indicator

type RooIgnoreController struct {
	cwd     string
	parser  gitignore.GitIgnore // Use a gitignore parsing library
	rules   []string            // Store raw rules for instructions
	enabled bool
}

func NewRooIgnoreController(cwd string) *RooIgnoreController {
	return &RooIgnoreController{
		cwd:     cwd,
		enabled: false, // Disabled until initialized successfully
	}
}

// Initialize reads and parses the .rooignore file.
func (c *RooIgnoreController) Initialize() error {
	ignorePath := filepath.Join(c.cwd, ".rooignore")
	parser, err := gitignore.NewFromFile(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.enabled = false // File doesn't exist, controller is disabled
			c.rules = nil
			c.parser = nil
			return nil // Not an error if the file doesn't exist
		}
		return fmt.Errorf("failed to read .rooignore at %s: %w", ignorePath, err)
	}

	// Read raw rules for instructions
	contentBytes, err := os.ReadFile(ignorePath)
	if err != nil {
		// Should not happen if NewFromFile succeeded, but check anyway
		return fmt.Errorf("failed to re-read .rooignore content: %w", err)
	}
	c.rules = strings.Split(string(contentBytes), "\n")

	c.parser = parser
	c.enabled = true
	return nil
}

// ValidateAccess checks if a given path (relative to CWD or absolute) is allowed.
// Returns true if allowed, false if ignored.
func (c *RooIgnoreController) ValidateAccess(filePath string) bool {
	if !c.enabled || c.parser == nil {
		return true // Allowed if ignore is disabled or failed to init
	}

	// Library expects path relative to the ignore file's location (c.cwd)
	relPath, err := filepath.Rel(c.cwd, filePath)
	if err != nil {
		// If it's not relative (e.g., different drive on Windows),
		// treat it as allowed unless an absolute pattern matches.
		// The library might handle absolute paths, check its docs.
		relPath = filePath // Use the original path for matching absolute patterns
	}

	// Check if the path matches any ignore pattern
	match := c.parser.Match(relPath)
	return match == nil // No match means it's allowed
}

// ValidateCommand checks if a command attempts to access ignored files.
// Returns the first ignored path found, or empty string if none.
// This is a simplified check. A robust solution needs proper shell parsing.
func (c *RooIgnoreController) ValidateCommand(command string) string {
	if !c.enabled {
		return ""
	}
	// Basic check: Split command by spaces and check potential file paths.
	// This is very naive and won't handle complex commands, quotes, etc.
	parts := strings.Fields(command)
	for _, part := range parts {
		// Crude check if it looks like a path - might need refinement
		if strings.Contains(part, "/") || strings.Contains(part, "\\") || strings.HasPrefix(part, ".") {
			absPath := part
			if !filepath.IsAbs(part) {
				absPath = filepath.Join(c.cwd, part) // Assume relative to cwd
			}
			if !c.ValidateAccess(absPath) {
				return part // Return the part identified as an ignored path
			}
		}
	}
	return ""
}

// GetInstructions returns the formatted .rooignore rules for the system prompt.
func (c *RooIgnoreController) GetInstructions() string {
	if !c.enabled || len(c.rules) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("# .rooignore Rules\n")
	builder.WriteString(fmt.Sprintf("# Files matching these patterns are blocked (indicated by %s):\n", LOCK_TEXT_SYMBOL))
	for _, rule := range c.rules {
		trimmedRule := strings.TrimSpace(rule)
		if trimmedRule != "" && !strings.HasPrefix(trimmedRule, "#") { // Ignore empty lines/comments
			builder.WriteString(fmt.Sprintf("- %s\n", trimmedRule))
		}
	}
	return builder.String()
}

// AddPatterns adds ignore patterns programmatically
func (c *RooIgnoreController) AddPatterns(contextDir string, patterns []string) error {
	// Initialize a memory-based gitignore parser
	content := strings.Join(patterns, "\n")
	reader := strings.NewReader(content)
	c.parser = gitignore.New(reader, c.cwd, nil)
	// Store the patterns
	c.rules = append(c.rules, patterns...)
	c.enabled = true
	return nil
}
