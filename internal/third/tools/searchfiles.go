package tools

import (
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
	"mind-weaver/internal/third/ripgrep"
	"os"
	"path/filepath"
	"strings"
)

// SearchFilesTool performs regex search in files.
func SearchFilesTool(input ExecutorInput) (*ExecutorResult, error) {
	relPath, ok := input.ToolUse.Params[string(assistantmessage.Path)]
	if !ok || strings.TrimSpace(relPath) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Path))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	regexStr, ok := input.ToolUse.Params[string(assistantmessage.Regex)]
	if !ok || strings.TrimSpace(regexStr) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Regex))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	filePattern, _ := input.ToolUse.Params[string(assistantmessage.FilePattern)] // Optional

	absolutePath := filepath.Join(input.Cwd, relPath)
	if !filepath.IsAbs(relPath) {
		absolutePath = filepath.Clean(absolutePath)
	} else {
		absolutePath = filepath.Clean(relPath)
	}

	// Check rooignore (ripgrep service should ideally handle this)
	if input.RooIgnoreController != nil && !input.RooIgnoreController.ValidateAccess(absolutePath) {
		errText := prompts.FormatRooIgnoreError(relPath)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	}

	// Check if path exists and is a directory (ripgrep usually handles this, but good practice)
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("Search path does not exist: %s", relPath)), IsError: true}, nil
		}
		return nil, fmt.Errorf("stating search path %s: %w", absolutePath, err) // Internal error
	}
	// Allow searching single files too? If so, remove this check. Ripgrep handles both.
	if !info.IsDir() {
		return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("Search path is not a directory: %s", relPath)), IsError: true}, nil
	}

	// Call the hypothetical ripgrep service
	// This service should handle invoking ripgrep with appropriate flags
	// and integrating RooIgnoreController.
	results, err := ripgrep.RegexSearchFiles(input.Cwd, absolutePath, regexStr, filePattern, input.RooIgnoreController)
	if err != nil {
		// Handle specific errors from ripgrep if needed
		// e.g., invalid regex might be an LLM error vs internal error
		if strings.Contains(err.Error(), "regex parse error") {
			return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("Invalid regex pattern provided: %v", err)), IsError: true}, nil
		}
		return nil, fmt.Errorf("searching files in %s with regex '%s': %w", absolutePath, regexStr, err) // Internal error
	}

	if results == "" {
		results = "No matches found."
	}

	return &ExecutorResult{Result: results}, nil
}
