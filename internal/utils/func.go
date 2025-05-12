package utils

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// getLanguageFromExtension 根据文件扩展名返回编程语言
func GetLanguageFromExtension(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".rb":
		return "ruby"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".rs":
		return "rust"
	case ".sh":
		return "shell"
	case ".sql":
		return "sql"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	default:
		return "text"
	}
}

// FileFilterFunc 文件过滤函数类型
type FileFilterFunc func(path string, info os.FileInfo) bool

// GetFilesInDirectory 递归获取目录下所有文件
// dirPath: 目录路径
// filter: 文件过滤函数，可为nil
// maxDepth: 最大递归深度，0表示不限制
func GetFilesInDirectory(dirPath string, filter FileFilterFunc, maxDepth int) ([]string, error) {
	var files []string
	depth := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算当前深度
		// relPath, _ := filepath.Rel(dirPath, path)
		// depth = strings.Count(relPath, string(filepath.Separator)) + 1
		depth += 1

		// 检查深度限制
		if maxDepth > 0 && depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理文件
		if !info.IsDir() {
			// 应用过滤函数
			if filter == nil || filter(path, info) {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

func GetAllowedExtensions() map[string]bool {
	return map[string]bool{
		".go":    true,
		".js":    true,
		".jsx":   true,
		".ts":    true,
		".tsx":   true,
		".py":    true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".cs":    true,
		".hpp":   true,
		".php":   true,
		".rb":    true,
		".html":  true,
		".css":   true,
		".json":  true,
		".xml":   true,
		".yaml":  true,
		".yml":   true,
		".md":    true,
		".txt":   true,
		".sql":   true,
		".sh":    true,
		".bash":  true,
		".swift": true,
		".kt":    true,
		".rs":    true,
	}
}

// DefaultCodeFilter 默认代码文件过滤器
func DefaultCodeFilter() FileFilterFunc {
	codeExts := GetAllowedExtensions()

	return func(path string, info os.FileInfo) bool {
		ext := strings.ToLower(filepath.Ext(path))
		return codeExts[ext]
	}
}

// 将传入的结构体序列化为json字符串，并返回json字符串，失败错误时返回空字符串
func StructToJSON(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

// countFileLines 统计文件行数
func CountFileLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		// 跳过空行
		if strings.TrimSpace(scanner.Text()) != "" {
			lineCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lineCount, nil
}

// Helper to check if a path is a subpath of another
func IsSubPath(parent, child string) bool {
	// Normalize paths
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	// Check if child starts with parent
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", it's not a subpath
	return !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..")
}

// CreateDirectoryIfNotExist creates a directory if it doesn't exist
func CreateDirectoryIfNotExist(path string) error {
	if path == "" {
		return nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}

	return nil
}

// WriteJSONToFile writes JSON data to a file
func WriteJSONToFile(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, jsonData, 0644)
}

// InterfacesToJSON converts interfaces to JSON
func InterfacesToJSON(interfaces []ApiInterface, base64Encode bool) (string, error) {
	jsonData, err := json.MarshalIndent(interfaces, "", "  ")
	if err != nil {
		return "", err
	}

	if base64Encode {
		return base64.StdEncoding.EncodeToString(jsonData), nil
	}

	return string(jsonData), nil
}

// 辅助函数：从Markdown中提取代码块
func ExtractCodeFromMarkdown(markdown string) string {
	re := regexp.MustCompile("```(?:javascript|js|html|css|svg)?\n((?s).*?)\n```")
	matches := re.FindAllStringSubmatch(markdown, -1)

	if len(matches) > 0 && len(matches[0]) > 1 {
		return matches[0][1]
	}

	// 如果没有指定语言的代码块，尝试查找没有语言标记的代码块
	re = regexp.MustCompile("```\n((?s).*?)\n```")
	matches = re.FindAllStringSubmatch(markdown, -1)

	if len(matches) > 0 && len(matches[0]) > 1 {
		return matches[0][1]
	}

	return ""
}
