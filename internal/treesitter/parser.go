package treesitter

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

const maxFilesToList = 200
const maxFilesToParse = 50

// fileExists checks if a file or directory exists and is accessible.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// listFiles lists files in a directory, respecting the ignore controller.
// Returns up to 'limit' files.
func listFiles(dirPath string, ignoreCtrl IgnoreController, limit int) ([]string, error) {
	var files []string
	var count int = 0
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log or handle permission errors etc. if needed
			fmt.Printf("Warning: Error accessing path %q: %v\n", path, err)
			// Decide if you want to skip the error or stop walking
			if d != nil && d.IsDir() {
				return fs.SkipDir // Skip directory if error occurs
			}
			return nil // Skip file if error occurs
		}

		// Skip the root directory itself
		if path == dirPath {
			return nil
		}

		// Check if ignored *before* checking if it's a directory
		// This allows ignoring entire directories
		if ignoreCtrl != nil && ignoreCtrl.Match(path) {
			fmt.Printf("Ignoring path: %s\n", path) // Debugging
			if d.IsDir() {
				return fs.SkipDir // Skip ignored directories efficiently
			}
			return nil // Skip ignored files
		}

		if !d.IsDir() {
			files = append(files, path)
			count++
			if count >= limit {
				return fs.SkipAll // Stop walking once limit is reached
			}
		}
		return nil
	})

	if err != nil && err != fs.SkipAll { // fs.SkipAll is not a real error here
		return nil, fmt.Errorf("error walking directory %s: %w", dirPath, err)
	}
	return files, nil

}

// parseFileInternal parses a single file using the appropriate tree-sitter parser.
// Assumes the parser for the file's extension is already loaded in 'parsers'.
func parseFileInternal(ctx context.Context, filePath string, content []byte, parsers LanguageParsers, ignoreCtrl IgnoreController) (string, error) {
	// Double check ignore status (might be redundant if called from ParseDirectory which already filters)
	if ignoreCtrl != nil && ignoreCtrl.Match(filePath) {
		fmt.Printf("Skipping ignored file in parseFileInternal: %s\n", filePath) // Debugging
		return "", nil                                                           // File ignored
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if len(ext) < 1 {
		fmt.Printf("Skipping file with no extension: %s\n", filePath)
		return "", nil // No extension
	}
	ext = ext[1:] // Remove dot

	parserInfo, ok := parsers[ext]
	if !ok || parserInfo.Parser == nil || parserInfo.Query == nil {
		fmt.Printf("No parser/query found for extension '%s' in file: %s\n", ext, filePath)
		return "", nil // Silently skip unsupported types or if parser wasn't loaded
	}

	// Parse the file content
	tree, err := parserInfo.Parser.ParseCtx(ctx, nil, content)
	if err != nil {
		// Log parsing errors but don't fail the whole process
		fmt.Printf("Error parsing file %s: %v\n", filePath, err)
		return "", nil // Return empty string on error, indicating no definitions found/error
	}
	if tree == nil {
		fmt.Printf("Parsing returned nil tree for file %s\n", filePath)
		return "", nil
	}
	defer tree.Close() // Ensure tree resources are released

	rootNode := tree.RootNode()
	if rootNode == nil {
		fmt.Printf("Parsed tree has nil root node for file %s\n", filePath)
		return "", nil
	}

	// Execute the query
	qc := sitter.NewQueryCursor()
	defer qc.Close() // Ensure cursor resources are released

	qc.Exec(parserInfo.Query, rootNode)

	// Collect captures
	var captures []*sitter.QueryCapture
	for {
		match, ok := qc.NextMatch()
		if !ok || match == nil {
			break // No more matches
		}
		// Filter captures within the match if needed, though usually not necessary
		// for simple definition queries.
		captures = append(captures, match.Captures...)
	}

	// Process captures
	// Use the default min lines unless specified otherwise
	definitions := ProcessCaptures(captures, content, minComponentLinesDefault)

	return definitions, nil
}

// ParseFile parses a single file for definitions.
// It handles reading the file, checking ignores, loading the necessary parser,
// and formatting the output.
func ParseFile(ctx context.Context, filePath string, ignoreCtrl IgnoreController) (string, error) {
	if ignoreCtrl == nil {
		ignoreCtrl = NewNoopIgnoreController()
	}

	// 1. Check existence and ignore status
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}
	if !fileExists(absPath) {
		return fmt.Sprintf("File does not exist or permission denied: %s", filePath), nil // Return message as string, not error
	}
	if ignoreCtrl.Match(absPath) {
		fmt.Printf("Ignoring file: %s\n", filePath) // Debugging
		return "", nil                              // File is ignored
	}

	// 2. Check extension support
	ext := strings.ToLower(filepath.Ext(absPath))
	if !isExtensionSupported(ext) {
		fmt.Printf("Unsupported extension '%s' for file: %s\n", ext, filePath)
		return "", nil // Unsupported extension
	}

	// 3. Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// 4. Handle Markdown special case
	if ext == ".md" || ext == ".markdown" {
		mdCaptures := ParseMarkdown(string(content))
		definitions := ProcessMarkdownCaptures(mdCaptures, string(content), minComponentLinesDefault)
		if definitions != "" {
			// Use base name for single file parsing
			return fmt.Sprintf("# %s\n%s", filepath.Base(filePath), definitions), nil
		}
		return "", nil // No definitions found in markdown
	}

	// 5. Handle other supported files (Tree-sitter)
	// Load parser for this specific file type
	parsers, err := LoadRequiredLanguageParsers(ctx, []string{absPath})
	if err != nil {
		return "", fmt.Errorf("failed to load parser for %s: %w", filePath, err)
	}

	// Parse using tree-sitter
	definitions, err := parseFileInternal(ctx, absPath, content, parsers, ignoreCtrl)
	if err != nil {
		// parseFileInternal logs errors, return empty string here
		return "", nil
	}

	if definitions != "" {
		// Use base name for single file parsing
		return fmt.Sprintf("# %s\n%s", filepath.Base(filePath), definitions), nil
	}

	return "", nil // No definitions found
}

// ParseDirectory parses all supported files in a directory for definitions.
// It respects ignore rules and limits the number of files processed.
func ParseDirectory(ctx context.Context, editorCtx EditorContext, ignoreCtrl IgnoreController) (string, error) {
	if ignoreCtrl == nil {
		ignoreCtrl = NewNoopIgnoreController()
	}

	// 1. Check directory existence
	absDirPath, err := filepath.Abs(editorCtx.DirPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", editorCtx.DirPath, err)
	}
	if !fileExists(absDirPath) {
		return "Directory does not exist or permission denied.", nil // Return message, not error
	}

	// 2. List files (limited)
	allFiles, err := listFiles(absDirPath, ignoreCtrl, maxFilesToList)
	if err != nil {
		return "", fmt.Errorf("failed to list files in %s: %w", absDirPath, err)
	}

	// 3. Filter and Separate Files
	var filesToParse []string
	markdownFiles := []string{}
	otherFiles := []string{}

	// Filter by supported extensions first
	for _, file := range allFiles {
		ext := strings.ToLower(filepath.Ext(file))
		if isExtensionSupported(ext) {
			filesToParse = append(filesToParse, file)
		}
	}

	// Apply ignore controller filtering *after* listing (listFiles already does some filtering)
	// This ensures the final list respects the controller precisely.
	allowedFiles := ignoreCtrl.FilterPaths(filesToParse)

	// Apply the parsing limit and separate markdown/other
	count := 0
	for _, file := range allowedFiles {
		if count >= maxFilesToParse {
			break
		}
		ext := strings.ToLower(filepath.Ext(file))
		if ext == ".md" || ext == ".markdown" {
			markdownFiles = append(markdownFiles, file)
		} else {
			otherFiles = append(otherFiles, file)
		}
		count++
	}

	if len(markdownFiles) == 0 && len(otherFiles) == 0 {
		return "No source code definitions found (no supported files or all ignored/limited).", nil
	}

	// 4. Load Parsers for non-markdown files
	languageParsers, err := LoadRequiredLanguageParsers(ctx, otherFiles)
	if err != nil {
		return "", fmt.Errorf("failed to load language parsers: %w", err)
	}

	// 5. Process Files and Aggregate Results
	var finalResult strings.Builder
	var wg sync.WaitGroup
	resultChan := make(chan string, len(markdownFiles)+len(otherFiles))
	errorChan := make(chan error, len(markdownFiles)+len(otherFiles)) // To collect errors non-blockingly

	// Process Markdown Files Concurrently
	for _, file := range markdownFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return // Respect context cancellation
			default:
				content, err := os.ReadFile(f)
				if err != nil {
					fmt.Printf("Error reading markdown file %s: %v\n", f, err)
					errorChan <- fmt.Errorf("read %s: %w", f, err) // Report error
					return
				}
				mdCaptures := ParseMarkdown(string(content))
				definitions := ProcessMarkdownCaptures(mdCaptures, string(content), minComponentLinesDefault)
				if definitions != "" {
					relPath, _ := filepath.Rel(absDirPath, f)
					resultChan <- fmt.Sprintf("# %s\n%s\n", filepath.ToSlash(relPath), definitions)
				}
			}
		}(file)
	}

	// Process Other Files Concurrently
	for _, file := range otherFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return // Respect context cancellation
			default:
				content, err := os.ReadFile(f)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", f, err)
					errorChan <- fmt.Errorf("read %s: %w", f, err) // Report error
					return
				}
				// Pass ignoreCtrl=nil here because filtering already happened
				definitions, err := parseFileInternal(ctx, f, content, languageParsers, nil)
				if err != nil {
					// parseFileInternal already logs, report error for tracking
					errorChan <- fmt.Errorf("parse %s: %w", f, err)
					return
				}
				if definitions != "" {
					relPath, _ := filepath.Rel(absDirPath, f)
					resultChan <- fmt.Sprintf("# %s\n%s\n", filepath.ToSlash(relPath), definitions)
				}
			}
		}(file)
	}

	// Wait for all processing goroutines to finish
	wg.Wait()
	close(resultChan)
	close(errorChan) // Close error channel after wait group finishes

	// Collect results
	for res := range resultChan {
		finalResult.WriteString(res)
	}

	// Collect and potentially report errors (optional)
	var processingErrors []error
	for err := range errorChan {
		processingErrors = append(processingErrors, err)
	}
	if len(processingErrors) > 0 {
		// Log these errors or handle them as needed
		fmt.Printf("Encountered %d errors during file processing:\n", len(processingErrors))
		for _, e := range processingErrors {
			fmt.Printf("- %v\n", e)
		}
		// Decide if errors should halt the process or just be logged
	}

	// 6. Return final result
	output := finalResult.String()
	if output == "" {
		return "No source code definitions found.", nil
	}

	return output, nil
}

// ParseSourceCodeForDefinitionsTopLevel parses only the top-level files in a directory.
// It returns formatted code definitions and respects the provided ignore controller.
func ParseSourceCodeForDefinitionsTopLevel(dirPath string, ignoreCtrl IgnoreController) (string, error) {
	if ignoreCtrl == nil {
		ignoreCtrl = NewNoopIgnoreController()
	}

	// Check directory existence
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", dirPath, err)
	}
	if !fileExists(absDirPath) {
		return "Directory does not exist or permission denied.", nil
	}

	// List only top-level files (not recursive)
	entries, err := os.ReadDir(absDirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", absDirPath, err)
	}

	var filesToParse []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		filePath := filepath.Join(absDirPath, entry.Name())
		// Apply ignore filter
		if ignoreCtrl.Match(filePath) {
			continue
		}

		// Check if file extension is supported
		ext := strings.ToLower(filepath.Ext(filePath))
		if isExtensionSupported(ext) {
			filesToParse = append(filesToParse, filePath)
		}
	}

	if len(filesToParse) == 0 {
		return "No source code files found at the top level.", nil
	}

	// Limit the number of files to parse if necessary
	if len(filesToParse) > maxFilesToParse {
		filesToParse = filesToParse[:maxFilesToParse]
	}

	// Group by file type
	markdownFiles := []string{}
	otherFiles := []string{}
	for _, file := range filesToParse {
		ext := strings.ToLower(filepath.Ext(file))
		if ext == ".md" || ext == ".markdown" {
			markdownFiles = append(markdownFiles, file)
		} else {
			otherFiles = append(otherFiles, file)
		}
	}

	// Load parsers for non-markdown files
	languageParsers, err := LoadRequiredLanguageParsers(context.Background(), otherFiles)
	if err != nil {
		return "", fmt.Errorf("failed to load language parsers: %w", err)
	}

	// Process files and aggregate results
	var finalResult strings.Builder
	var wg sync.WaitGroup
	resultChan := make(chan string, len(markdownFiles)+len(otherFiles))
	errorChan := make(chan error, len(markdownFiles)+len(otherFiles))

	// Process Markdown Files Concurrently
	for _, file := range markdownFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			content, err := os.ReadFile(f)
			if err != nil {
				errorChan <- fmt.Errorf("read %s: %w", f, err)
				return
			}
			mdCaptures := ParseMarkdown(string(content))
			definitions := ProcessMarkdownCaptures(mdCaptures, string(content), minComponentLinesDefault)
			if definitions != "" {
				relPath, _ := filepath.Rel(absDirPath, f)
				resultChan <- fmt.Sprintf("# %s\n%s\n", filepath.ToSlash(relPath), definitions)
			}
		}(file)
	}

	// Process Other Files Concurrently
	for _, file := range otherFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			content, err := os.ReadFile(f)
			if err != nil {
				errorChan <- fmt.Errorf("read %s: %w", f, err)
				return
			}
			// Parse file and process captures
			definitions, err := parseFileInternal(context.Background(), f, content, languageParsers, nil)
			if err != nil {
				errorChan <- fmt.Errorf("parse %s: %w", f, err)
				return
			}
			if definitions != "" {
				relPath, _ := filepath.Rel(absDirPath, f)
				resultChan <- fmt.Sprintf("# %s\n%s\n", filepath.ToSlash(relPath), definitions)
			}
		}(file)
	}

	// Wait for all processing goroutines to finish
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	for res := range resultChan {
		finalResult.WriteString(res)
	}

	// Log errors but don't fail the entire process
	var processingErrors []error
	for err := range errorChan {
		processingErrors = append(processingErrors, err)
	}
	if len(processingErrors) > 0 {
		fmt.Printf("Encountered %d errors during file processing\n", len(processingErrors))
		for _, e := range processingErrors {
			fmt.Printf("- %v\n", e)
		}
	}

	output := finalResult.String()
	if output == "" {
		return "No source code definitions found in top-level files.", nil
	}

	return output, nil
}

// ParseSourceCodeDefinitionsForFile is an alias for ParseFile for API consistency.
func ParseSourceCodeDefinitionsForFile(filePath string, ignoreCtrl IgnoreController) (string, error) {
	return ParseFile(context.Background(), filePath, ignoreCtrl)
}
