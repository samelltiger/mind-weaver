package tools

import (
	"bufio"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"

	// "mind-weaver/internal/treesitter"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype" // For basic binary detection
)

const defaultMaxReadFileLine = 500 // Default, should be configurable

// ReadFileTool reads the content of a file, potentially a specific line range.
func ReadFileTool(input ExecutorInput) (*ExecutorResult, error) {
	relPath, ok := input.ToolUse.Params[string(assistantmessage.Path)]
	if !ok || strings.TrimSpace(relPath) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Path))
		return &ExecutorResult{Result: fmt.Sprintf(`<file><path></path><error>%s</error></file>`, errText), IsError: true}, nil
	}

	absolutePath := filepath.Join(input.Cwd, relPath) // Use Join for safety
	if !filepath.IsAbs(relPath) {
		// Clean the path to prevent directory traversal issues if needed
		absolutePath = filepath.Clean(absolutePath)
		// Add further validation/sandboxing if necessary
	} else {
		absolutePath = filepath.Clean(relPath) // Clean absolute paths too
	}

	// Check rooignore
	if input.RooIgnoreController != nil && !input.RooIgnoreController.ValidateAccess(absolutePath) {
		errText := prompts.FormatRooIgnoreError(relPath) // Use relPath for message consistency
		return &ExecutorResult{Result: fmt.Sprintf(`<file><path>%s</path><error>%s</error></file>`, relPath, errText), IsError: true}, nil
	}

	startLineStr, _ := input.ToolUse.Params[string(assistantmessage.StartLine)]
	endLineStr, _ := input.ToolUse.Params[string(assistantmessage.EndLine)]
	isRangeRead := startLineStr != "" || endLineStr != ""
	var startLine, endLine int = -1, -1 // 0-based index, -1 means not set
	var parseError error

	if startLineStr != "" {
		startLine, parseError = strconv.Atoi(startLineStr)
		if parseError != nil || startLine < 1 {
			return &ExecutorResult{Result: fmt.Sprintf(`<file><path>%s</path><error>Invalid start_line value: %s</error></file>`, relPath, startLineStr), IsError: true}, nil
		}
		startLine-- // Convert to 0-based
	}

	if endLineStr != "" {
		endLine, parseError = strconv.Atoi(endLineStr)
		if parseError != nil || endLine < 1 {
			return &ExecutorResult{Result: fmt.Sprintf(`<file><path>%s</path><error>Invalid end_line value: %s</error></file>`, relPath, endLineStr), IsError: true}, nil
		}
		endLine-- // Convert to 0-based
	}

	// Validate line range logic
	if startLine != -1 && endLine != -1 && startLine > endLine {
		return &ExecutorResult{Result: fmt.Sprintf(`<file><path>%s</path><error>start_line (%d) cannot be greater than end_line (%d)</error></file>`, relPath, startLine+1, endLine+1), IsError: true}, nil
	}

	// --- File Reading Logic ---
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ExecutorResult{Result: fmt.Sprintf(`<file><path>%s</path><error>File does not exist</error></file>`, relPath), IsError: true}, nil
		}
		return nil, fmt.Errorf("stating file %s: %w", absolutePath, err) // Internal error
	}
	if fileInfo.IsDir() {
		return &ExecutorResult{Result: fmt.Sprintf(`<file><path>%s</path><error>Path is a directory, not a file</error></file>`, relPath), IsError: true}, nil
	}

	// Get configuration (e.g., max lines) - placeholder
	maxLinesConfig := defaultMaxReadFileLine // Should come from config or input

	totalLines, countErr := countFileLines(absolutePath)
	if countErr != nil {
		fmt.Printf("Warning: could not count lines in %s: %v\n", absolutePath, countErr)
		totalLines = -1 // Indicate error or unknown
	}

	var contentBuilder strings.Builder
	var xmlInfoBuilder strings.Builder // For notices, definitions etc.
	var contentTag string = ""         // The <content> tag itself
	isFileTruncated := false
	isBinary := false

	// Check if binary using mimetype library
	mtype, err := mimetype.DetectFile(absolutePath)
	if err == nil && !strings.HasPrefix(mtype.String(), "text/") && mtype.String() != "application/json" && mtype.String() != "application/xml" {
		// Add more text-based MIME types if needed
		isBinary = true
	} else if err != nil {
		fmt.Printf("Warning: could not detect mime type for %s: %v\n", absolutePath, err)
		// Proceed assuming text, but maybe add a notice?
	}

	if isBinary {
		xmlInfoBuilder.WriteString("<notice>File appears to be binary. Content display might be limited or nonsensical.</notice>\n")
		// Optionally read only first few KB for binary files or skip content entirely
		contentBuilder.WriteString("[Binary file content omitted or truncated]") // Placeholder
	} else if isRangeRead {
		// Read specific line range
		lines, readErr := readLines(absolutePath, startLine, endLine) // readLines needs to handle 0-based indices
		if readErr != nil {
			return nil, fmt.Errorf("reading lines %d-%d from %s: %w", startLine+1, endLine+1, absolutePath, readErr)
		}
		// Add line numbers relative to the start line requested
		lineNumStart := 1
		if startLine != -1 {
			lineNumStart = startLine + 1
		}
		contentBuilder.WriteString(addLineNumbers(strings.Join(lines, "\n"), lineNumStart))

		// Determine actual end line read for attribute
		actualEndLine := len(lines) - 1 // 0-based index of last line read
		if startLine != -1 {
			actualEndLine += startLine
		}
		if endLine != -1 && actualEndLine > endLine { // Clamp if readLines read fewer than requested up to endLine
			actualEndLine = endLine
		}
		if totalLines != -1 && actualEndLine >= totalLines { // Clamp if we read past the actual end of file
			actualEndLine = totalLines - 1
		}

		contentTag = fmt.Sprintf(`<content lines="%d-%d">`, lineNumStart, actualEndLine+1)

	} else {
		// Read full file or truncate
		shouldTruncate := maxLinesConfig >= 0 && totalLines > maxLinesConfig && totalLines != -1

		if shouldTruncate {
			isFileTruncated = true
			// Read only maxLinesConfig lines
			lines, readErr := readLines(absolutePath, 0, maxLinesConfig-1) // Read lines 0 to maxLinesConfig-1
			if readErr != nil {
				return nil, fmt.Errorf("reading first %d lines from %s: %w", maxLinesConfig, absolutePath, readErr)
			}
			if len(lines) > 0 {
				contentBuilder.WriteString(addLineNumbers(strings.Join(lines, "\n"), 1))
				contentTag = fmt.Sprintf(`<content lines="1-%d">`, len(lines)) // Use actual lines read
			} else {
				// File might be smaller than maxLinesConfig but still truncated conceptually if totalLines > maxLinesConfig
				// Or readLines failed partially? Handle empty case.
				contentTag = `<content/>` // Represent as empty content if no lines read
			}

			// Attempt to parse definitions for truncated files
			// defs, defErr := treesitter.ParseSourceCodeDefinitionsForFile(absolutePath, nil)
			// if defErr != nil {
			// 	fmt.Printf("Warning: failed to parse definitions for %s: %v\n", absolutePath, defErr)
			// } else if defs != "" {
			// xmlInfoBuilder.WriteString(fmt.Sprintf("<list_code_definition_names>%s</list_code_definition_names>\n", defs))
			// }

		} else if maxLinesConfig == 0 {
			// Definitions only mode
			isFileTruncated = totalLines > 0 // Considered truncated if file has content but we show none
			// defs, defErr := treesitter.ParseSourceCodeDefinitionsForFile(absolutePath, nil)
			// if defErr != nil {
			// 	fmt.Printf("Warning: failed to parse definitions for %s: %v\n", absolutePath, defErr)
			// 	xmlInfoBuilder.WriteString("<notice>Could not extract definitions.</notice>\n")
			// } else if defs != "" {
			// 	xmlInfoBuilder.WriteString(fmt.Sprintf("<list_code_definition_names>%s</list_code_definition_names>\n", defs))
			// } else {
			xmlInfoBuilder.WriteString("<notice>No definitions found.</notice>\n")
			// }
			contentTag = "" // No content tag in this mode

		} else {
			// Read the entire file
			fileBytes, readErr := os.ReadFile(absolutePath)
			if readErr != nil {
				return nil, fmt.Errorf("reading full file %s: %w", absolutePath, readErr)
			}
			// We assume text file here because binary check passed earlier
			contentStr := string(fileBytes)
			if contentStr != "" {
				contentBuilder.WriteString(addLineNumbers(contentStr, 1))
				lineCount := totalLines
				if lineCount <= 0 {
					lineCount = 1
				} // Min 1 line if content exists
				contentTag = fmt.Sprintf(`<content lines="1-%d">`, lineCount)
			} else {
				contentTag = `<content/>` // File exists but is empty
			}
		}
	}

	// --- Assemble Final XML ---
	finalContent := contentBuilder.String()
	finalXmlInfo := xmlInfoBuilder.String()

	if isFileTruncated && totalLines > 0 { // Add truncation notice only if we know total lines
		finalXmlInfo += fmt.Sprintf("<notice>Showing only %d of %d total lines. Use start_line and end_line if you need to read more</notice>\n", maxLinesConfig, totalLines)
	} else if finalContent == "" && totalLines == 0 && !isRangeRead && maxLinesConfig != 0 {
		finalXmlInfo += "<notice>File is empty</notice>\n"
	}

	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("<file><path>%s</path>\n", relPath))
	if contentTag != "" {
		resultBuilder.WriteString(contentTag) // Opening tag with attributes
		if finalContent != "" {
			resultBuilder.WriteString("\n") // Newline before content if not empty
			resultBuilder.WriteString(finalContent)
			resultBuilder.WriteString("\n") // Newline after content
		}
		resultBuilder.WriteString("</content>\n") // Closing tag
	}
	resultBuilder.WriteString(finalXmlInfo) // Add notices, definitions
	resultBuilder.WriteString("</file>")

	return &ExecutorResult{Result: resultBuilder.String()}, nil
}

// --- Helper Function Stubs/Implementations ---

// addLineNumbers adds 1-based line numbers to each line of a string.
func addLineNumbers(content string, startNum int) string {
	lines := strings.Split(content, "\n")
	var builder strings.Builder
	for i, line := range lines {
		// Handle potential empty last line from Split
		if i == len(lines)-1 && line == "" {
			break
		}
		builder.WriteString(fmt.Sprintf("%d | %s\n", startNum+i, line))
	}
	// Remove trailing newline if present
	return strings.TrimSuffix(builder.String(), "\n")
}

// countFileLines counts the number of lines in a file.
func countFileLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	count := 0
	for {
		_, err := reader.ReadString('\n')
		count++
		if err == io.EOF {
			// Check if the file ends without a newline
			// Need to read the last part to see if it's non-empty
			file.Seek(-1, io.SeekEnd) // Go back one byte
			lastByte := make([]byte, 1)
			_, readErr := file.Read(lastByte)
			if readErr == nil && lastByte[0] != '\n' {
				// Last line didn't end with newline, but wasn't empty, already counted.
			} else if count == 1 {
				// If only one line read and it ended in EOF, check if file was actually empty
				info, statErr := file.Stat()
				if statErr == nil && info.Size() == 0 {
					count = 0 // Empty file
				}
			}
			break
		}
		if err != nil {
			return 0, err // Return other errors
		}
	}
	return count, nil
}

// readLines reads a specific range of lines (0-based indices) from a file.
// Reads from startLine up to and including endLine.
// If endLine is -1, reads until EOF.
func readLines(filePath string, startLine, endLine int) ([]string, error) {
	if startLine < 0 {
		startLine = 0
	} // Default start

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	currentLine := 0

	for scanner.Scan() {
		if currentLine >= startLine {
			lines = append(lines, scanner.Text())
		}
		// Stop if we've reached the endLine (or read enough lines if endLine is -1)
		if endLine != -1 && currentLine >= endLine {
			break
		}
		// Stop if we have read the requested number of lines when endLine is specified
		if endLine != -1 && len(lines) > (endLine-startLine) {
			break
		}

		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Ensure we don't return more lines than requested if endLine was set
	if endLine != -1 {
		maxLines := endLine - startLine + 1
		if len(lines) > maxLines {
			lines = lines[:maxLines]
		}
	}

	return lines, nil
}

// extractTextFromFile - Placeholder for potential future complex extraction (PDF, DOCX)
// For now, just reads text files.
func extractTextFromFile(filePath string) (string, error) {
	// Basic implementation: just read the file as text
	// Real implementation would check extension and use appropriate libraries (like for PDF, DOCX)
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Basic check for UTF-8 validity might be good here
	return string(contentBytes), nil
}
