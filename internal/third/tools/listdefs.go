package tools

// import (
// 	"mind-weaver/internal/third/assistantmessage"
// 	"mind-weaver/internal/third/prompts"
// 	"mind-weaver/internal/treesitter"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// // ListCodeDefinitionNamesTool lists definitions from source files.
// func ListCodeDefinitionNamesTool(input ExecutorInput) (*ExecutorResult, error) {
// 	relPath, ok := input.ToolUse.Params[string(assistantmessage.Path)]
// 	if !ok || strings.TrimSpace(relPath) == "" {
// 		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Path))
// 		return &ExecutorResult{Result: errText, IsError: true}, nil
// 	}

// 	absolutePath := filepath.Join(input.Cwd, relPath)
// 	if !filepath.IsAbs(relPath) {
// 		absolutePath = filepath.Clean(absolutePath)
// 	} else {
// 		absolutePath = filepath.Clean(relPath)
// 	}

// 	// Check rooignore (Tree-sitter service should ideally handle this)
// 	if input.RooIgnoreController != nil && !input.RooIgnoreController.ValidateAccess(absolutePath) {
// 		errText := prompts.FormatRooIgnoreError(relPath)
// 		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
// 	}

// 	// Call the hypothetical tree-sitter service
// 	var results string
// 	var err error

// 	info, statErr := os.Stat(absolutePath)
// 	if statErr != nil {
// 		if os.IsNotExist(statErr) {
// 			results = fmt.Sprintf("Path does not exist or cannot be accessed: %s", relPath)
// 		} else {
// 			return nil, fmt.Errorf("stating path %s: %w", absolutePath, statErr) // Internal error
// 		}
// 	} else {
// 		if info.IsDir() {
// 			// Parse all top-level files in the directory
// 			results, err = treesitter.ParseSourceCodeForDefinitionsTopLevel(absolutePath, nil)
// 		} else {
// 			// Parse a single file
// 			results, err = treesitter.ParseSourceCodeDefinitionsForFile(absolutePath, nil)
// 		}

// 		if err != nil {
// 			// Handle specific errors from tree-sitter if possible
// 			return nil, fmt.Errorf("parsing definitions for %s: %w", absolutePath, err) // Internal error
// 		}
// 	}

// 	if results == "" {
// 		results = "No source code definitions found."
// 	}

// 	return &ExecutorResult{Result: results}, nil

// }
