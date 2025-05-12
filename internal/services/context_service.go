package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ContextService struct {
	fileService *FileService
}

type FileContext struct {
	FilePath     string   `json:"file_path"`
	Content      string   `json:"content"`
	Language     string   `json:"language"`
	LineCount    int      `json:"line_count"`
	ImportedDeps []string `json:"imported_deps,omitempty"`
}

func NewContextService(fileService *FileService) *ContextService {
	return &ContextService{
		fileService: fileService,
	}
}

func (cs *ContextService) GetFileContext(filePath string) (*FileContext, error) {
	// Read the file content
	content, err := cs.fileService.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Determine language based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := getLanguageFromExt(ext)

	// Count lines
	lineCount := strings.Count(content, "\n") + 1

	// Create the context
	fileContext := &FileContext{
		FilePath:  filePath,
		Content:   content,
		Language:  language,
		LineCount: lineCount,
	}

	// Extract imports (basic implementation - would be more sophisticated in a real app)
	fileContext.ImportedDeps = cs.extractImports(content, language)

	return fileContext, nil
}

func (cs *ContextService) GetCurrentFileContext(projectPath, relativePath string) (*FileContext, error) {
	fullPath := filepath.Join(projectPath, relativePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("file does not exist")
	}

	return cs.GetFileContext(fullPath)
}

func (cs *ContextService) GetRelatedFiles(projectPath, filePath string, maxFiles int) ([]*FileContext, error) {
	// Get the current file context
	currentContext, err := cs.GetFileContext(filePath)
	if err != nil {
		return nil, err
	}
	fmt.Println(currentContext)

	// Get directory of the current file
	dir := filepath.Dir(filePath)

	// List files in the same directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	relatedContexts := []*FileContext{}
	count := 0

	// Add files from the same directory (excluding the current file)
	for _, file := range files {
		if count >= maxFiles {
			break
		}

		if file.IsDir() {
			continue
		}

		fullPath := filepath.Join(dir, file.Name())
		if fullPath == filePath {
			continue // Skip the current file
		}

		// Only include files with code extensions
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !isCodeFile(ext) {
			continue
		}

		fileContext, err := cs.GetFileContext(fullPath)
		if err == nil {
			relatedContexts = append(relatedContexts, fileContext)
			count++
		}
	}

	// In a real implementation, we would also follow imports and include related files
	// based on the dependency graph, but that's beyond the scope of this example

	return relatedContexts, nil
}

func (cs *ContextService) extractImports(content, language string) []string {
	// This is a simplified implementation
	// In a real app, we would use language-specific parsers
	imports := []string{}

	switch language {
	case "go":
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import") {
				// Extract import paths
				if strings.Contains(line, "(") {
					continue // Multi-line import, would need more complex parsing
				}
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					importPath := strings.Trim(parts[1], "\"")
					imports = append(imports, importPath)
				}
			}
		}
	case "javascript", "typescript":
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "require(") {
				// Very basic extraction - would need a proper parser in a real app
				imports = append(imports, line)
			}
		}
	case "python":
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
				imports = append(imports, line)
			}
		}
	}

	return imports
}

func getLanguageFromExt(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".jsx":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cpp", ".cc":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".rb":
		return "ruby"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".sh":
		return "shell"
	case ".sql":
		return "sql"
	default:
		return "text"
	}
}

func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		".go":   true,
		".js":   true,
		".jsx":  true,
		".ts":   true,
		".tsx":  true,
		".py":   true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
		".cs":   true,
		".php":  true,
		".rb":   true,
	}
	return codeExts[ext]
}
