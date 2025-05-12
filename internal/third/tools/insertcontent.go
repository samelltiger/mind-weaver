package tools

import (
	"encoding/json"
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/diff"
	"mind-weaver/internal/third/prompts"
	"os"
	"path/filepath"
	"strings"
)

// InsertContentTool inserts content at specified lines in a file.
func InsertContentTool(input ExecutorInput) (*ExecutorResult, error) {
	relPath, ok := input.ToolUse.Params[string(assistantmessage.Path)]
	if !ok || strings.TrimSpace(relPath) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Path))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	operationsJSON, ok := input.ToolUse.Params[string(assistantmessage.Operations)]
	if !ok || strings.TrimSpace(operationsJSON) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Operations))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}

	absolutePath := filepath.Join(input.Cwd, relPath)
	if !filepath.IsAbs(relPath) {
		absolutePath = filepath.Clean(absolutePath)
	} else {
		absolutePath = filepath.Clean(relPath)
	}

	// Check rooignore
	if input.RooIgnoreController != nil && !input.RooIgnoreController.ValidateAccess(absolutePath) {
		errText := prompts.FormatRooIgnoreError(relPath)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	}

	// Check if file exists (insert only works on existing files)
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("File does not exist at path: %s. insert_content only works on existing files.", relPath)), IsError: true}, nil
		}
		return nil, fmt.Errorf("stating file %s: %w", absolutePath, err) // Internal error
	}
	if info.IsDir() {
		return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("Path is a directory, not a file: %s", relPath)), IsError: true}, nil
	}

	// Parse operations JSON
	var parsedOperations []struct {
		StartLine int    `json:"start_line"` // 1-based line number
		Content   string `json:"content"`
	}
	err = json.Unmarshal([]byte(operationsJSON), &parsedOperations)
	if err != nil {
		errText := fmt.Sprintf("Invalid operations JSON format: %v", err)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	}

	if len(parsedOperations) == 0 {
		return &ExecutorResult{Result: "No insertion operations provided."}, nil // Not an error, just nothing to do
	}

	// Read original content
	originalContentBytes, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("reading file for insert %s: %w", absolutePath, err)
	}
	originalContent := string(originalContentBytes)
	originalLines := strings.Split(originalContent, "\n")

	// Prepare insert groups for the diff helper
	insertGroups := make([]diff.InsertGroup, 0, len(parsedOperations))
	for _, op := range parsedOperations {
		if op.StartLine < 1 {
			errText := fmt.Sprintf("Invalid start_line: %d. Line numbers must be 1 or greater.", op.StartLine)
			return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
		}
		insertGroups = append(insertGroups, diff.InsertGroup{
			Index:    op.StartLine - 1, // Convert to 0-based index
			Elements: strings.Split(op.Content, "\n"),
		})
	}

	// Apply insertions
	newLines := diff.InsertGroups(originalLines, insertGroups)
	newContent := strings.Join(newLines, "\n")

	// --- Write the modified content ---
	err = os.WriteFile(absolutePath, []byte(newContent), info.Mode()) // Preserve original permissions
	if err != nil {
		return nil, fmt.Errorf("writing inserted content to file %s: %w", absolutePath, err)
	}

	// Success message
	resultText := fmt.Sprintf("The content was successfully inserted in %s.", relPath)
	return &ExecutorResult{Result: resultText}, nil
}
