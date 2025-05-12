package assistantmessage

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// ParseAssistantMessage parses an assistant message into content blocks
func ParseAssistantMessage(assistantMessage string) []AssistantMessageContent {
	var contentBlocks []AssistantMessageContent
	var currentTextContent *TextContent
	var currentToolUse *ToolUse
	var currentToolUseStartIndex int
	var currentParamName ToolParamName
	var currentParamValueStartIndex int
	var inThinkingBlock bool
	var thinkingStartIndex int
	accumulator := ""

	runes := []rune(assistantMessage)
	for _, char := range runes {
		accumulator += string(char)

		// 检查thinking标签
		if !inThinkingBlock && strings.HasSuffix(accumulator, "<thinking>") {
			// 如果在thinking标签前有文本内容，保存它
			if currentTextContent != nil {
				currentTextContent.Partial = false
				prefixLength := len(accumulator) - len("<thinking>")
				if prefixLength > 0 {
					currentTextContent.Content = strings.TrimSpace(accumulator[:prefixLength])
					contentBlocks = append(contentBlocks, *currentTextContent)
				}
				currentTextContent = nil
			}

			// 开始一个thinking块
			inThinkingBlock = true
			thinkingStartIndex = len(accumulator)
			continue
		}

		if inThinkingBlock && strings.HasSuffix(accumulator, "</thinking>") {
			// 结束thinking块，提取内容
			thinkingContent := accumulator[thinkingStartIndex : len(accumulator)-len("</thinking>")]
			thinkingBlock := TextContent{
				Type:    "thinking",
				Content: thinkingContent,
				Partial: false,
			}
			contentBlocks = append(contentBlocks, thinkingBlock)
			inThinkingBlock = false

			// 重置累加器，准备处理下一个内容块
			accumulator = ""
			continue
		}

		// 如果在thinking块内，继续累积内容
		if inThinkingBlock {
			continue
		}

		// 以下是原有的工具使用解析逻辑
		// There should not be a param without a tool use
		if currentToolUse != nil && currentParamName != "" {
			currentParamValue := accumulator[currentParamValueStartIndex:]
			paramClosingTag := "</" + string(currentParamName) + ">"
			if strings.HasSuffix(currentParamValue, paramClosingTag) {
				// End of param value
				if currentToolUse.Params == nil {
					currentToolUse.Params = make(map[string]string)
				}
				valueLength := len(currentParamValue) - len(paramClosingTag)
				currentToolUse.Params[string(currentParamName)] = strings.TrimSpace(currentParamValue[:valueLength])
				currentParamName = ""
				continue
			} else {
				// Partial param value is accumulating
				continue
			}
		}

		// No currentParamName

		if currentToolUse != nil {
			currentToolValue := accumulator[currentToolUseStartIndex:]
			toolUseClosingTag := "</" + string(currentToolUse.Name) + ">"
			if strings.HasSuffix(currentToolValue, toolUseClosingTag) {
				// End of a tool use
				currentToolUse.Partial = false
				contentBlocks = append(contentBlocks, currentToolUse)
				currentToolUse = nil

				// Reset accumulator and text content tracking after complete tool use
				accumulator = ""
				currentTextContent = nil
				continue
			} else {
				// Check for parameter opening tags
				foundParam := false
				for _, paramName := range AllToolParamNames() {
					paramOpeningTag := "<" + string(paramName) + ">"
					if strings.HasSuffix(accumulator, paramOpeningTag) {
						// Start of a new parameter
						currentParamName = paramName
						currentParamValueStartIndex = len(accumulator)
						foundParam = true
						break
					}
				}

				if foundParam {
					continue
				}

				// Special case for write_to_file where file contents could contain the closing tag
				contentParamName := Content
				if currentToolUse.Name == WriteToFile && strings.HasSuffix(accumulator, "</"+string(contentParamName)+">") {
					toolContent := accumulator[currentToolUseStartIndex:]
					contentStartTag := "<" + string(contentParamName) + ">"
					contentEndTag := "</" + string(contentParamName) + ">"
					contentStartIndex := strings.Index(toolContent, contentStartTag) + len(contentStartTag)
					contentEndIndex := strings.LastIndex(toolContent, contentEndTag)
					if contentStartIndex != -1 && contentEndIndex != -1 && contentEndIndex > contentStartIndex {
						if currentToolUse.Params == nil {
							currentToolUse.Params = make(map[string]string)
						}
						currentToolUse.Params[string(contentParamName)] = strings.TrimSpace(toolContent[contentStartIndex:contentEndIndex])
					}
				}

				// Partial tool value is accumulating
				continue
			}
		}

		// No currentToolUse

		didStartToolUse := false
		for _, toolUseName := range AllToolUseNames() {
			toolUseOpeningTag := "<" + string(toolUseName) + ">"
			if strings.HasSuffix(accumulator, toolUseOpeningTag) {
				// If we have text before the tool use, save it
				if currentTextContent != nil {
					currentTextContent.Partial = false
					// Remove the opening tag from the content
					prefixLength := len(accumulator) - len(toolUseOpeningTag)
					if prefixLength > 0 {
						currentTextContent.Content = strings.TrimSpace(accumulator[:prefixLength])
						contentBlocks = append(contentBlocks, *currentTextContent)
					}
					currentTextContent = nil
				}

				// Start of a new tool use
				currentToolUse = &ToolUse{
					Type:    "tool_use",
					Name:    toolUseName,
					Params:  make(map[string]string),
					Partial: true,
				}
				currentToolUseStartIndex = len(accumulator)
				didStartToolUse = true
				break
			}
		}

		if !didStartToolUse {
			// No tool use, so it must be text either at the beginning or between tools
			if currentTextContent == nil {
				currentTextContent = &TextContent{
					Type:    "text",
					Content: string(char), // Start with just the current character
					Partial: true,
				}
			} else {
				// Only append the current character, not the whole accumulator
				currentTextContent.Content += string(char)
			}
		}
	}

	// 处理末尾的thinking块
	if inThinkingBlock {
		thinkingContent := accumulator[thinkingStartIndex:]
		thinkingBlock := TextContent{
			Type:    "thinking",
			Content: thinkingContent,
			Partial: true, // 标记为部分，因为没有找到结束标签
		}
		contentBlocks = append(contentBlocks, thinkingBlock)
	} else if currentToolUse != nil {
		// Stream did not complete tool call, add it as partial
		if currentParamName != "" {
			// Tool call has a parameter that was not completed
			if currentToolUse.Params == nil {
				currentToolUse.Params = make(map[string]string)
			}
			currentToolUse.Params[string(currentParamName)] = strings.TrimSpace(accumulator[currentParamValueStartIndex:])
		}
		contentBlocks = append(contentBlocks, currentToolUse)
	} else if currentTextContent != nil {
		// Stream did not complete text content, add it as partial
		contentBlocks = append(contentBlocks, *currentTextContent)
	}

	return contentBlocks
}

// GenerateMarkdown 从解析后的助手消息内容生成 markdown 文本
func GenerateMarkdown(contents []AssistantMessageContent) string {
	var result strings.Builder

	for _, content := range contents {
		switch c := content.(type) {
		case TextContent:
			if c.Type == "thinking" {
				// 为 thinking 内容添加特殊样式
				result.WriteString("<div class=\"thinking-block\">\n")
				result.WriteString("<div class=\"thinking-header\">🧠 思考过程</div>\n")
				result.WriteString("<div class=\"thinking-content\">\n\n")
				// 保留原始格式但转义 HTML
				// formattedContent := strings.ReplaceAll(c.Content, "\n", "\n\n")
				result.WriteString(c.Content)
				if !c.Partial {
					result.WriteString("\n\n</div>\n</div>\n\n")
				}
			} else {
				// 普通文本直接输出
				result.WriteString(c.Content)
				result.WriteString("\n\n")
			}
		case *ToolUse:
			// 处理不同的工具使用，生成适当的 markdown
			switch c.Name {
			case WriteToFile:
				// 对于 write_to_file，将内容显示为代码块并进行语言检测
				if content, ok := c.Params["content"]; ok {
					path := c.Params["path"]
					lang := detectLanguage(path)
					result.WriteString(fmt.Sprintf("📄 **写入文件**: `%s`\n\n", path))
					result.WriteString("```")
					result.WriteString(lang)
					result.WriteString("\n")
					result.WriteString(content)
					if !c.Partial {
						result.WriteString("\n")
						result.WriteString("```\n\n")
					}
				}
			case ExecuteCommand:
				// 对于 execute_command，显示为 shell 代码块
				if command, ok := c.Params["command"]; ok {
					result.WriteString("🖥️ **执行命令**:\n\n")
					result.WriteString("```shell\n")
					result.WriteString(command)
					if !c.Partial {
						result.WriteString("\n```\n\n")
					}
				}
			case ReadFile:
				// 对于 read_file，提及正在读取的文件
				if path, ok := c.Params["path"]; ok {
					result.WriteString(fmt.Sprintf("📖 **读取文件**: `%s`\n\n", path))
				}
			case SearchFiles:
				// 对于 search_files，显示搜索模式
				if pattern, ok := c.Params["file_pattern"]; ok {
					result.WriteString(fmt.Sprintf("🔍 **搜索文件**: `%s`", pattern))
					if regex, ok := c.Params["regex"]; ok && regex != "" {
						result.WriteString(fmt.Sprintf(" (正则表达式: `%s`)", regex))
					}
					result.WriteString("\n\n")
				}
			case ListFiles:
				// 对于 list_files，显示列表路径
				if path, ok := c.Params["path"]; ok {
					recursive := "否"
					if rec, ok := c.Params["recursive"]; ok && rec == "true" {
						recursive = "是"
					}
					result.WriteString(fmt.Sprintf("📁 **列出文件**: 路径 `%s` (递归: %s)\n\n", path, recursive))
				}
			case ApplyDiff:
				// 对于 apply_diff，显示差异内容
				if path, ok := c.Params["path"]; ok {
					result.WriteString(fmt.Sprintf("✏️ **应用差异**: 文件 `%s`\n\n", path))
					if diff, ok := c.Params["diff"]; ok {
						result.WriteString("```diff\n")
						result.WriteString(diff)
						if !c.Partial {
							result.WriteString("\n```\n\n")
						}
					}
				}
			case InsertContent:
				// 对于 insert_content，显示插入内容
				if path, ok := c.Params["path"]; ok {
					result.WriteString(fmt.Sprintf("➕ **插入内容**: 文件 `%s`\n\n", path))
					if content, ok := c.Params["content"]; ok {
						lang := detectLanguage(path)
						result.WriteString("```")
						result.WriteString(lang)
						result.WriteString("\n")
						result.WriteString(content)
						if !c.Partial {
							result.WriteString("\n```\n\n")
						}
					}
				}
			case SearchAndReplace:
				// 对于 search_and_replace，显示搜索和替换内容
				if path, ok := c.Params["path"]; ok {
					result.WriteString(fmt.Sprintf("🔄 **搜索替换**: 文件 `%s`\n\n", path))
				}
			case AttemptCompletion:
				// 对于 attempt_completion，显示结果
				if res, ok := c.Params["result"]; ok {
					result.WriteString(fmt.Sprintf("✅ **完成操作**: %s\n\n", res))
					if command, ok := c.Params["command"]; ok {
						result.WriteString(fmt.Sprintf("建议执行: `%s`\n\n", command))
					}
				}
			default:
				// 对于其他工具，显示通用表示
				result.WriteString(fmt.Sprintf("🔧 **工具**: %s\n\n", c.Name))
				for param, value := range c.Params {
					result.WriteString(fmt.Sprintf("- **%s**: %s\n", param, value))
				}
				result.WriteString("\n")
			}
		}
	}

	return strings.TrimSpace(result.String())
}

// detectLanguage attempts to detect the programming language based on file extension
func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".html", ".htm":
		return "html"
	case ".js":
		return "javascript"
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".java":
		return "java"
	case ".c", ".cpp", ".cc":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".rb":
		return "ruby"
	case ".rs":
		return "rust"
	case ".ts":
		return "typescript"
	case ".sh":
		return "shell"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".sql":
		return "sql"
	case ".xml":
		return "xml"
	case ".yaml", ".yml":
		return "yaml"
	case ".css":
		return "css"
	case ".txt":
		return ""
	default:
		return ""
	}
}

// 去掉文本节点
func RemoveTextNode(contents []AssistantMessageContent) []AssistantMessageContent {
	var result []AssistantMessageContent

	for _, content := range contents {
		switch c := content.(type) {
		case TextContent:
			continue
		case *ToolUse:
			result = append(result, c)
		}
	}

	return result
}

// StartsWithCodeBlock checks if the input string starts with a code block marker
// after trimming whitespace, case insensitive.
func StartsWithCodeBlock(input string) bool {
	trimmed := strings.TrimLeftFunc(input, unicode.IsSpace)
	trimmed = strings.ToLower(trimmed)

	codeBlocks := []string{
		"```css",
		"```xml",
		"```html",
		"```svg",
		"```javascript",
	}

	for _, cb := range codeBlocks {
		if strings.HasPrefix(trimmed, strings.ToLower(cb)) {
			return true
		}
	}

	return false
}
