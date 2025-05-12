package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptBuilder 用于构建发送给AI的提示词
type PromptBuilder struct {
	systemPrompt   string
	contextFiles   []ContextFile
	codeSelection  string
	cursorPosition string
	language       string
	maxTokens      int
}

// ContextFile 表示上下文中的文件
type ContextFile struct {
	Path     string
	Content  string
	Language string
	IsMain   bool
}

// NewPromptBuilder 创建一个新的提示词构建器
func NewPromptBuilder(maxTokens int) *PromptBuilder {
	return &PromptBuilder{
		systemPrompt: "You are MindWeaver AI, an intelligent coding assistant. Help the user write clean, efficient, and correct code.",
		contextFiles: []ContextFile{},
		maxTokens:    maxTokens,
	}
}

// SetSystemPrompt 设置系统提示词
func (b *PromptBuilder) SetSystemPrompt(prompt string) *PromptBuilder {
	b.systemPrompt = prompt
	return b
}

// AddContextFile 添加上下文文件
func (b *PromptBuilder) AddContextFile(path, content, language string, isMain bool) *PromptBuilder {
	b.contextFiles = append(b.contextFiles, ContextFile{
		Path:     path,
		Content:  content,
		Language: language,
		IsMain:   isMain,
	})
	return b
}

// AddCodeFile 添加上下文文件
func (b *PromptBuilder) AddCodeFile(path string) (*PromptBuilder, error) {
	// 检查文件路径是否为空
	if path == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// 获取文件扩展名以确定语言
	ext := strings.ToLower(filepath.Ext(path))
	language := GetLanguageFromExtension(ext)

	b.contextFiles = append(b.contextFiles, ContextFile{
		Path:     path,
		Content:  string(content),
		Language: language,
		IsMain:   false, // 默认false，可通过其他方法设置
	})

	return b, nil
}

// SetCodeSelection 设置当前选中的代码
func (b *PromptBuilder) SetCodeSelection(selection string) *PromptBuilder {
	b.codeSelection = selection
	return b
}

// SetCursorPosition 设置光标位置描述
func (b *PromptBuilder) SetCursorPosition(position string) *PromptBuilder {
	b.cursorPosition = position
	return b
}

// SetLanguage 设置目标语言
func (b *PromptBuilder) SetLanguage(language string) *PromptBuilder {
	b.language = language
	return b
}

// BuildSystemPrompt 构建系统提示词
func (b *PromptBuilder) BuildSystemPrompt(addSysPrompt bool) string {
	var sb strings.Builder

	// 基础系统提示词
	sb.WriteString(b.systemPrompt)
	sb.WriteString("\n\n")

	if addSysPrompt {
		// 添加语言指示
		if b.language != "" {
			sb.WriteString(fmt.Sprintf("You are helping with %s code. ", b.language))
		}

		// 添加编码指南
		sb.WriteString("Follow these guidelines:\n")
		sb.WriteString("1. Write clean, efficient, and well-documented code\n")
		sb.WriteString("2. Follow best practices and design patterns\n")
		sb.WriteString("3. Include error handling where appropriate\n")
		sb.WriteString("4. Be concise but thorough in your explanations\n\n")
		sections := fmt.Sprintf(`Language Preference:
You should always speak and think in the "%s" (%s) language unless the user gives you instructions below to do otherwise.`, "中文", "zh-cn")
		sb.WriteString(sections + "\n\n")

	}

	// 添加上下文文件
	if len(b.contextFiles) > 0 {
		sb.WriteString("CONTEXT FILES:\n")

		// 先添加主要文件
		for _, file := range b.contextFiles {
			if file.IsMain {
				sb.WriteString(fmt.Sprintf("FILE (MAIN): %s\n", file.Path))
				sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", file.Language, file.Content))
			}
		}

		// 再添加其他上下文文件
		for _, file := range b.contextFiles {
			if !file.IsMain {
				sb.WriteString(fmt.Sprintf("FILE: %s\n", file.Path))
				sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", file.Language, file.Content))
			}
		}
	}

	// 添加当前选中的代码（如果有）
	if b.codeSelection != "" {
		sb.WriteString("SELECTED CODE:\n")
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", b.codeSelection))
	}

	// 添加光标位置（如果有）
	if b.cursorPosition != "" {
		sb.WriteString(fmt.Sprintf("CURSOR POSITION: %s\n\n", b.cursorPosition))
	}

	return sb.String()
}

// BuildUserPrompt 优化用户提示词
func (b *PromptBuilder) BuildUserPrompt(userPrompt string) string {
	// 如果用户提示已经很详细，就直接返回
	if len(userPrompt) > 100 {
		return userPrompt
	}

	// 否则，增强用户提示
	var sb strings.Builder

	// 如果提示不是以动词开始，添加一个动作词
	lowerPrompt := strings.ToLower(userPrompt)
	if !strings.HasPrefix(lowerPrompt, "write") &&
		!strings.HasPrefix(lowerPrompt, "create") &&
		!strings.HasPrefix(lowerPrompt, "implement") &&
		!strings.HasPrefix(lowerPrompt, "generate") &&
		!strings.HasPrefix(lowerPrompt, "add") {
		sb.WriteString("Implement ")
	}

	sb.WriteString(userPrompt)

	// 如果没有指定语言，添加语言信息
	if b.language != "" && !strings.Contains(lowerPrompt, b.language) {
		sb.WriteString(fmt.Sprintf(" in %s", b.language))
	}

	// 添加质量期望
	if !strings.Contains(lowerPrompt, "clean") && !strings.Contains(lowerPrompt, "efficient") {
		sb.WriteString(". The code should be clean, efficient, and follow best practices")
	}

	return sb.String()
}

// BuildCompletionPrompt 构建完整的提示词
func (b *PromptBuilder) BuildCompletionPrompt(userPrompt string, addSysPrompt bool) (string, string) {
	systemPrompt := b.BuildSystemPrompt(addSysPrompt)
	enhancedUserPrompt := b.BuildUserPrompt(userPrompt)

	return systemPrompt, enhancedUserPrompt
}

// EstimateTokenCount 估计提示词的token数量
// 这是一个简化的估算，实际token数量需要用tokenizer计算
func (b *PromptBuilder) EstimateTokenCount(text string) int {
	// 粗略估计：英文平均每4个字符约为1个token
	return len(text) / 4
}

// TrimContextToFit 裁剪上下文以适应token限制
func (b *PromptBuilder) TrimContextToFit() {
	// 如果没有上下文文件，直接返回
	if len(b.contextFiles) == 0 {
		return
	}

	// 计算系统提示的基本token数（不包括文件内容）
	basePrompt := b.systemPrompt + "\n\n"
	if b.language != "" {
		basePrompt += fmt.Sprintf("You are helping with %s code. ", b.language)
	}
	basePrompt += "Follow these guidelines:\n1. Write clean, efficient, and well-documented code\n2. Follow best practices and design patterns\n3. Include error handling where appropriate\n4. Be concise but thorough in your explanations\n\nCONTEXT FILES:\n"

	baseTokens := b.EstimateTokenCount(basePrompt)

	// 为用户提示和回复预留tokens
	reservedTokens := 1000
	availableTokens := b.maxTokens - baseTokens - reservedTokens

	// 如果可用token不足，需要裁剪上下文
	if availableTokens < 0 {
		// 极端情况，无法容纳任何上下文
		b.contextFiles = []ContextFile{}
		return
	}

	// 计算所有文件内容的token
	totalContentTokens := 0
	fileTokens := make([]int, len(b.contextFiles))

	for i, file := range b.contextFiles {
		fileHeader := fmt.Sprintf("FILE: %s\n```%s\n", file.Path, file.Language)
		fileFooter := "```\n\n"
		content := file.Content

		tokens := b.EstimateTokenCount(fileHeader + content + fileFooter)
		fileTokens[i] = tokens
		totalContentTokens += tokens
	}

	// 如果总token数超过可用token，需要裁剪
	if totalContentTokens > availableTokens {
		// 优先保留主文件
		var mainFiles []ContextFile
		var otherFiles []ContextFile
		var mainTokens int

		for i, file := range b.contextFiles {
			if file.IsMain {
				mainFiles = append(mainFiles, file)
				mainTokens += fileTokens[i]
			} else {
				otherFiles = append(otherFiles, file)
			}
		}

		// 如果主文件已经超过可用token，需要裁剪主文件内容
		if mainTokens > availableTokens {
			// 简单策略：只保留第一个主文件，并裁剪其内容
			if len(mainFiles) > 0 {
				file := mainFiles[0]
				fileHeader := fmt.Sprintf("FILE (MAIN): %s\n```%s\n", file.Path, file.Language)
				fileFooter := "```\n\n"
				headerFooterTokens := b.EstimateTokenCount(fileHeader + fileFooter)

				// 计算可用于内容的token
				contentTokens := availableTokens - headerFooterTokens
				if contentTokens > 0 {
					// 按字符数粗略裁剪（实际应该按token裁剪）
					maxChars := contentTokens * 4
					if len(file.Content) > maxChars {
						file.Content = file.Content[:maxChars] + "\n// ... content truncated to fit token limit"
					}

					b.contextFiles = []ContextFile{file}
				} else {
					// 极端情况，无法容纳任何文件内容
					b.contextFiles = []ContextFile{}
				}
			}
		} else {
			// 主文件可以完全保留，裁剪其他文件
			remainingTokens := availableTokens - mainTokens

			// 按重要性排序其他文件（简化版：保持原顺序）
			b.contextFiles = mainFiles

			// 尽可能添加其他文件
			for _, file := range otherFiles {
				fileHeader := fmt.Sprintf("FILE: %s\n```%s\n", file.Path, file.Language)
				fileFooter := "```\n\n"
				content := file.Content

				tokens := b.EstimateTokenCount(fileHeader + content + fileFooter)

				if tokens <= remainingTokens {
					b.contextFiles = append(b.contextFiles, file)
					remainingTokens -= tokens
				}
			}
		}
	}
}
