package tools

import (
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/glob"
	"mind-weaver/internal/third/prompts"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ListFilesTool lists files and directories.
func ListFilesTool(input ExecutorInput) (*ExecutorResult, error) {
	relPath, ok := input.ToolUse.Params[string(assistantmessage.Path)]
	if !ok || strings.TrimSpace(relPath) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Path))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	recursiveStr, _ := input.ToolUse.Params[string(assistantmessage.Recursive)]
	recursive, _ := strconv.ParseBool(recursiveStr) // Defaults to false on error

	absolutePath := filepath.Join(input.Cwd, relPath)
	if !filepath.IsAbs(relPath) {
		absolutePath = filepath.Clean(absolutePath)
	} else {
		absolutePath = filepath.Clean(relPath)
	}

	// Check rooignore (important for listing)
	// Note: The glob service itself should ideally handle ignores for efficiency
	if input.RooIgnoreController != nil && !input.RooIgnoreController.ValidateAccess(absolutePath) {
		// Technically, listing *within* an ignored dir might be blocked,
		// but accessing the dir *itself* is the primary check here.
		// If the glob service doesn't handle ignores, we might need to filter results later.
		errText := prompts.FormatRooIgnoreError(relPath)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	}

	// Check if path exists and is a directory
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("Directory does not exist: %s", relPath)), IsError: true}, nil
		}
		return nil, fmt.Errorf("stating directory %s: %w", absolutePath, err) // Internal error
	}
	if !info.IsDir() {
		return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("Path is not a directory: %s", relPath)), IsError: true}, nil
	}

	// Call the hypothetical glob service
	// This service should handle recursion and limits.
	// It should ideally also integrate RooIgnoreController.
	files, didHitLimit, err := glob.ListFiles(absolutePath, recursive, 200, input.RooIgnoreController) // Example limit
	if err != nil {
		// Handle specific errors from ListFiles if necessary
		return nil, fmt.Errorf("listing files in %s (recursive=%t): %w", absolutePath, recursive, err) // Internal error
	}

	// Format the results using the prompt formatter
	// Assuming showIgnored = true for now, make configurable if needed
	resultText := prompts.FormatFilesList(absolutePath, files, didHitLimit, input.RooIgnoreController, true)

	return &ExecutorResult{Result: resultText}, nil

}
