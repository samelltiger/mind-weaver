package tools

import (
	"encoding/json"
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchAndReplaceTool performs find and replace operations on a file.
func SearchAndReplaceTool(input ExecutorInput) (*ExecutorResult, error) {
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

	// Check if file exists
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("File does not exist at path: %s.", relPath)), IsError: true}, nil
		}
		return nil, fmt.Errorf("stating file %s: %w", absolutePath, err) // Internal error
	}
	if info.IsDir() {
		return &ExecutorResult{Result: prompts.FormatToolError(fmt.Sprintf("Path is a directory, not a file: %s", relPath)), IsError: true}, nil
	}

	// Parse operations JSON
	var parsedOperations []struct {
		Search     string  `json:"search"`
		Replace    string  `json:"replace"`
		StartLine  *int    `json:"start_line,omitempty"` // Use pointers for optional ints (1-based)
		EndLine    *int    `json:"end_line,omitempty"`   // Use pointers for optional ints (1-based)
		UseRegex   bool    `json:"use_regex,omitempty"`
		IgnoreCase bool    `json:"ignore_case,omitempty"`
		RegexFlags *string `json:"regex_flags,omitempty"` // Use pointer for optional string
	}
	err = json.Unmarshal([]byte(operationsJSON), &parsedOperations)
	if err != nil {
		errText := fmt.Sprintf("Invalid operations JSON format: %v", err)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	}

	if len(parsedOperations) == 0 {
		return &ExecutorResult{Result: "No search/replace operations provided."}, nil
	}

	// Read original content
	originalContentBytes, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("reading file for search/replace %s: %w", absolutePath, err)
	}
	currentContent := string(originalContentBytes)
	originalContentForDiff := currentContent // Keep original for final diff message if needed

	// Apply operations sequentially
	for i, op := range parsedOperations {
		var searchRe *regexp.Regexp
		var compileErr error

		if op.UseRegex {
			// Build regex flags string
			flags := ""
			if op.RegexFlags != nil {
				flags = *op.RegexFlags // Use provided flags
			} else if op.IgnoreCase {
				flags = "i" // Default to ignore case if specified
			}
			// Ensure multiline (?m) and global (implicit via ReplaceAllString)
			// Go regex doesn't have a direct 'g' flag, ReplaceAll handles it.
			// Add (?m) for ^$ matching lines, (?i) for case-insensitivity.
			pattern := op.Search
			finalFlags := ""
			if !strings.Contains(flags, "m") {
				finalFlags += "m"
			}
			if !strings.Contains(flags, "i") && op.IgnoreCase {
				finalFlags += "i"
			}
			// Add other flags if needed, filtering duplicates
			// ... logic to parse and add flags from op.RegexFlags or flags string ...

			if finalFlags != "" {
				pattern = fmt.Sprintf("(?%s)%s", finalFlags, pattern)
			}

			searchRe, compileErr = regexp.Compile(pattern)
		} else {
			// Literal search - escape regex metacharacters
			pattern := regexp.QuoteMeta(op.Search)
			if op.IgnoreCase {
				pattern = "(?i)" + pattern // Add case-insensitive flag
			}
			searchRe, compileErr = regexp.Compile(pattern)
		}

		if compileErr != nil {
			errText := fmt.Sprintf("Invalid search pattern (operation %d): %v", i+1, compileErr)
			return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
		}

		// Handle line ranges if specified
		if op.StartLine != nil || op.EndLine != nil {
			lines := strings.Split(currentContent, "\n")
			startIdx := 0 // 0-based
			if op.StartLine != nil && *op.StartLine > 0 {
				startIdx = *op.StartLine - 1
			}
			endIdx := len(lines) - 1 // 0-based inclusive
			if op.EndLine != nil && *op.EndLine > 0 && *op.EndLine-1 < len(lines) {
				endIdx = *op.EndLine - 1
			}

			if startIdx > endIdx || startIdx >= len(lines) {
				// Invalid range or starts after file ends, skip operation? Or error?
				fmt.Printf("Warning: Invalid or out-of-bounds line range (%d-%d) for operation %d on file %s. Skipping.\n", startIdx+1, endIdx+1, i+1, relPath)
				continue
			}

			targetLines := lines[startIdx : endIdx+1]
			targetContent := strings.Join(targetLines, "\n")
			modifiedSection := searchRe.ReplaceAllString(targetContent, op.Replace)

			// Reconstruct the content
			var builder strings.Builder
			if startIdx > 0 {
				builder.WriteString(strings.Join(lines[:startIdx], "\n"))
				builder.WriteString("\n") // Add newline removed by split if needed
			}
			builder.WriteString(modifiedSection)
			if endIdx+1 < len(lines) {
				builder.WriteString("\n") // Add newline removed by split if needed
				builder.WriteString(strings.Join(lines[endIdx+1:], "\n"))
			}
			currentContent = builder.String()

		} else {
			// Global replace on the whole content
			currentContent = searchRe.ReplaceAllString(currentContent, op.Replace)
		}
	} // End loop through operations

	// --- Write the modified content ---
	if currentContent == originalContentForDiff {
		return &ExecutorResult{Result: fmt.Sprintf("No changes needed for '%s' after search/replace.", relPath)}, nil
	}

	err = os.WriteFile(absolutePath, []byte(currentContent), info.Mode()) // Preserve original permissions
	if err != nil {
		return nil, fmt.Errorf("writing search/replaced content to file %s: %w", absolutePath, err)
	}

	// Success message (maybe include a diff summary?)
	// diffSummary := prompts.CreatePrettyPatch(relPath, originalContentForDiff, currentContent)
	resultText := fmt.Sprintf("Search and replace operations successfully applied to %s.", relPath)
	// if diffSummary != "" { resultText += "\nChanges Applied:\n" + diffSummary }

	return &ExecutorResult{Result: resultText}, nil
}
