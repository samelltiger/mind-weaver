package tools

import (
	"fmt"
	"html" // For unescaping potential LLM artifacts
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
	"mind-weaver/pkg/logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// WriteToFileTool writes content to a file, overwriting or creating it.
func WriteToFileTool(input ExecutorInput) (*ExecutorResult, error) {
	relPath, ok := input.ToolUse.Params[string(assistantmessage.Path)]
	if !ok || strings.TrimSpace(relPath) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Path))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}

	content, ok := input.ToolUse.Params[string(assistantmessage.Content)]
	if !ok { // Content can be empty, but the parameter must exist
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Content))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}

	lineCountStr, ok := input.ToolUse.Params[string(assistantmessage.LineCount)]
	if !ok || strings.TrimSpace(lineCountStr) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(string(assistantmessage.LineCount)))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	predictedLineCount, err := strconv.Atoi(lineCountStr)
	if err != nil || predictedLineCount < 0 {
		errText := fmt.Sprintf("Invalid line_count value: %s", lineCountStr)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	}

	absolutePath := filepath.Join(input.Cwd, relPath) // Use Join for safety
	if !filepath.IsAbs(relPath) {
		absolutePath = filepath.Clean(absolutePath)
	} else {
		absolutePath = filepath.Clean(relPath)
	}
	logger.Infof("Writing to file: %s", absolutePath)

	// Check rooignore
	if input.RooIgnoreController != nil && !input.RooIgnoreController.ValidateAccess(absolutePath) {
		errText := prompts.FormatRooIgnoreError(relPath)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	}

	// --- Pre-processing Content ---
	content = preprocessLLMContent(content)
	actualLineCount := strings.Count(content, "\n")
	// Add 1 if the content is not empty and doesn't end with a newline
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		actualLineCount++
	}
	if len(content) == 0 { // Empty content is 0 lines
		actualLineCount = 0
	}

	// --- Omission Detection (Simplified) ---
	// More robust detection needed based on TS logic (checking comments, comparing predicted vs actual lines)
	omissionDetected := false
	if predictedLineCount > 0 && actualLineCount > 0 && predictedLineCount > actualLineCount+5 { // Heuristic: Significant mismatch
		omissionDetected = true
		fmt.Printf("Warning: Potential content omission detected for %s. Predicted: %d, Actual: %d\n", relPath, predictedLineCount, actualLineCount)
	}
	// Add comment checking logic here if desired
	// omissionDetected = omissionDetected || detectCodeOmissionComments(content)

	if omissionDetected {
		// In the Go backend, we likely just perform the write but could log a warning
		// or potentially return an error/notice to the LLM if configured to do so.
		// For now, just proceed but log.
		fmt.Printf("Proceeding with write for %s despite potential omission.\n", relPath)
		// If apply_diff is available and preferred on omission:
		// if input.DiffStrategy != nil {
		//     errText := fmt.Sprintf("Content appears truncated or omits code for %s (predicted %d lines, got %d). Use 'apply_diff' instead.", relPath, predictedLineCount, actualLineCount)
		//     return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
		// }
	}

	logger.Infof("WriteToFileTool file content length: %v", len(content))
	// --- Perform Write ---
	// Ensure directory exists
	dir := filepath.Dir(absolutePath)
	if err := os.MkdirAll(dir, 0755); err != nil { // Use appropriate permissions
		return nil, fmt.Errorf("creating directory %s: %w", dir, err) // Internal error
	}

	// Write the file (overwrite)
	err = os.WriteFile(absolutePath, []byte(content), 0644) // Use appropriate permissions
	if err != nil {
		return nil, fmt.Errorf("writing file %s: %w", absolutePath, err) // Internal error
	}

	// Success message for LLM
	resultText := fmt.Sprintf("The content was successfully saved to %s.", relPath) // Use relPath for message
	return &ExecutorResult{Result: resultText}, nil
}

// preprocessLLMContent cleans up common artifacts from LLM output.
func preprocessLLMContent(content string) string {
	// Remove markdown code blocks
	content = strings.TrimPrefix(content, "```")
	// Handle potential language specifier after ```
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) > 1 && strings.TrimSpace(lines[0]) != "" && !strings.Contains(lines[0], " ") { // Heuristic: language specifier likely has no spaces
		content = lines[1]
	}
	content = strings.TrimSuffix(content, "\n```")
	content = strings.TrimSpace(content) // Trim leading/trailing whitespace

	// Unescape common HTML entities that might appear erroneously
	content = html.UnescapeString(content) // Handles <, >, &, ", etc.

	// Strip line numbers if present on every line (assuming helper exists)
	// if everyLineHasLineNumbers(content) {
	//    content = stripLineNumbers(content)
	// }

	return content
}
