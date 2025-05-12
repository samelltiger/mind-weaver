package tools

import (
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ApplyDiffTool applies a diff patch to a file.
func ApplyDiffTool(input ExecutorInput) (*ExecutorResult, error) {
	if input.DiffStrategy == nil {
		return &ExecutorResult{Result: prompts.FormatToolError("apply_diff tool is not available/enabled."), IsError: true}, nil
	}

	relPath, ok := input.ToolUse.Params[string(assistantmessage.Path)]
	if !ok || strings.TrimSpace(relPath) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Path))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	diffContent, ok := input.ToolUse.Params[string(assistantmessage.Diff)]
	if !ok || strings.TrimSpace(diffContent) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Diff))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	startLineStr, _ := input.ToolUse.Params[string(assistantmessage.StartLine)]
	endLineStr, _ := input.ToolUse.Params[string(assistantmessage.EndLine)]

	startLine := 0 // Default to 0 if not provided or invalid
	if i, err := strconv.Atoi(startLineStr); err == nil && i > 0 {
		startLine = i
	}
	endLine := 0 // Default to 0 if not provided or invalid
	if i, err := strconv.Atoi(endLineStr); err == nil && i > 0 {
		endLine = i
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

	// Read original content
	originalContentBytes, err := os.ReadFile(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("File does not exist at path: %s", relPath)), IsError: true}, nil
		}
		return nil, fmt.Errorf("reading file for diff %s: %w", absolutePath, err) // Internal error
	}
	originalContent := string(originalContentBytes)

	// Apply diff using the strategy
	diffResult, err := input.DiffStrategy.ApplyDiff(originalContent, diffContent, startLine, endLine)
	if err != nil {
		// Error during the diff application process itself (internal)
		return nil, fmt.Errorf("applying diff strategy to %s: %w", relPath, err)
	}

	if !diffResult.Success {
		// The diff strategy determined the patch couldn't be applied cleanly
		errorMsg := fmt.Sprintf("Unable to apply diff to file: %s.", relPath)
		if diffResult.Error != "" {
			errorMsg += "\nError: " + diffResult.Error
		}
		// Include fail parts if available
		if len(diffResult.FailParts) > 0 {
			errorMsg += "\nFailed Parts Details:"
			for i, part := range diffResult.FailParts {
				if !part.Success {
					errorMsg += fmt.Sprintf("\n [%d] %s", i+1, part.Error)
					// Add part.Details if needed
				}
			}
		}
		return &ExecutorResult{Result: prompts.FormatToolError(errorMsg), IsError: true}, nil
	}

	// --- Write the patched content ---
	// Ensure directory exists
	dir := filepath.Dir(absolutePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating directory for patched file %s: %w", dir, err)
	}
	// Write the file
	err = os.WriteFile(absolutePath, []byte(diffResult.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("writing patched file %s: %w", absolutePath, err)
	}

	// Success message
	successMsg := fmt.Sprintf("Changes successfully applied to %s.", relPath)
	if len(diffResult.FailParts) > 0 {
		// Check if *any* part failed, even if overall Success was true (partial success)
		hasFailures := false
		failDetails := ""
		for _, part := range diffResult.FailParts {
			if !part.Success {
				hasFailures = true
				failDetails += fmt.Sprintf("\n - %s", part.Error) // Add more detail if available
			}
		}
		if hasFailures {
			successMsg += fmt.Sprintf("\nWarning: Some diff parts failed to apply:%s\nPlease review the file.", failDetails)
		}
	}

	return &ExecutorResult{Result: successMsg}, nil

}
