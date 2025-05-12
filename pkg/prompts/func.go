package prompts

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	contentDir = "./pkg/prompts/content"
	CodeAnalys = ``
)

var (
	//go:embed content/*
	contentFS     embed.FS
	contentCache  = make(map[string]string)
	cacheMutex    sync.RWMutex
	useEmbeddedFS = false // 切换嵌入式文件系统标志
)

// Init 初始化服务（决定是否使用嵌入式文件系统）
func Init(useEmbedded bool) {
	useEmbeddedFS = useEmbedded
}

// GetPrompt 获取指定名称的prompt内容
func GetPrompt(name string) (string, error) {
	// 检查缓存
	if cached, ok := getFromCache(name); ok {
		return cached, nil
	}

	// 读取文件
	content, err := readPromptFile(name)
	if err != nil {
		return "", err
	}

	// 存入缓存
	setToCache(name, content)
	return content, nil
}

// GetPromptWithParams 获取并替换参数的prompt
func GetPromptWithParams(name string, params map[string]string) (string, error) {
	content, err := GetPrompt(name)
	if err != nil {
		return "", err
	}

	for k, v := range params {
		content = strings.ReplaceAll(content, "{{"+k+"}}", v)
	}

	return content, nil
}

// ListPrompts 列出所有可用的prompt名称
func ListPrompts() ([]string, error) {
	if useEmbeddedFS {
		return listEmbeddedPrompts()
	}
	return listLocalPrompts()
}

// 私有辅助函数
func readPromptFile(name string) (string, error) {
	filename := filepath.Join(contentDir, name+".txt")

	if useEmbeddedFS {
		return readEmbeddedFile(filename)
	}
	return readLocalFile(filename)
}

func readLocalFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", wrapError("failed to read local file", err)
	}
	return string(content), nil
}

func readEmbeddedFile(path string) (string, error) {
	content, err := contentFS.ReadFile(strings.TrimPrefix(path, "./"))
	if err != nil {
		return "", wrapError("failed to read embedded file", err)
	}
	return string(content), nil
}

func listLocalPrompts() ([]string, error) {
	files, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, wrapError("failed to list local prompts", err)
	}

	var prompts []string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}
		prompts = append(prompts, strings.TrimSuffix(file.Name(), ".txt"))
	}

	return prompts, nil
}

func listEmbeddedPrompts() ([]string, error) {
	entries, err := fs.ReadDir(contentFS, "content")
	if err != nil {
		return nil, wrapError("failed to list embedded prompts", err)
	}

	var prompts []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		prompts = append(prompts, strings.TrimSuffix(entry.Name(), ".txt"))
	}

	return prompts, nil
}

func getFromCache(name string) (string, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	content, ok := contentCache[name]
	return content, ok
}

func setToCache(name, content string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	contentCache[name] = content
}

func wrapError(context string, err error) error {
	return errors.New(context + ": " + err.Error())
}
