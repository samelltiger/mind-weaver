package utils

// import (
// 	"path/filepath"
// 	"regexp"
// 	"strings"
// )

// // CodeParser 提供代码解析和分析功能
// type CodeParser struct {
// 	// 语言特定的解析规则
// 	importPatterns   map[string]*regexp.Regexp
// 	functionPatterns map[string]*regexp.Regexp
// 	classPatterns    map[string]*regexp.Regexp
// }

// // CodeElement 表示代码中的一个元素
// type CodeElement struct {
// 	Type      string `json:"type"`      // "import", "function", "class", "variable", etc.
// 	Name      string `json:"name"`      // 元素名称
// 	Signature string `json:"signature"` // 函数签名或类定义等
// 	Content   string `json:"content"`   // 完整内容
// 	StartLine int    `json:"start_line"`
// 	EndLine   int    `json:"end_line"`
// }

// // NewCodeParser 创建一个新的代码解析器
// func NewCodeParser() *CodeParser {
// 	parser := &CodeParser{
// 		importPatterns:   make(map[string]*regexp.Regexp),
// 		functionPatterns: make(map[string]*regexp.Regexp),
// 		classPatterns:    make(map[string]*regexp.Regexp),
// 	}

// 	// 初始化语言特定的正则表达式
// 	// Go
// 	parser.importPatterns["go"] = regexp.MustCompile(`import\s+(?:\(\s*((?:.|\n)*?)\s*\)|([^\s(]+))`)
// 	parser.functionPatterns["go"] = regexp.MustCompile(`(?m)^func\s+(\w+)\s*\((.*?)\)(?:\s*\(?(.*?)\)?)?\s*\{`)
// 	parser.classPatterns["go"] = regexp.MustCompile(`(?m)^type\s+(\w+)\s+struct\s*\{`)

// 	// JavaScript/TypeScript
// 	parser.importPatterns["javascript"] = regexp.MustCompile(`(?:import|require)\s*\(?.*?['"]([^'"]+)['"]`)
// 	parser.importPatterns["typescript"] = parser.importPatterns["javascript"]
// 	parser.functionPatterns["javascript"] = regexp.MustCompile(`(?m)(?:function\s+(\w+)|const\s+(\w+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>|(\w+)\s*:\s*(?:async\s*)?\([^)]*\)\s*=>|(?:async\s*)?function\s*\*?\s*(\w+))`)
// 	parser.functionPatterns["typescript"] = parser.functionPatterns["javascript"]
// 	parser.classPatterns["javascript"] = regexp.MustCompile(`(?m)class\s+(\w+)(?:\s+extends\s+(\w+))?`)
// 	parser.classPatterns["typescript"] = parser.classPatterns["javascript"]

// 	// Python
// 	parser.importPatterns["python"] = regexp.MustCompile(`(?:from\s+(\S+)\s+import|import\s+(\S+))`)
// 	parser.functionPatterns["python"] = regexp.MustCompile(`(?m)^def\s+(\w+)\s*\(`)
// 	parser.classPatterns["python"] = regexp.MustCompile(`(?m)^class\s+(\w+)(?:\((.*?)\))?:`)

// 	return parser
// }

// // ParseCode 解析代码内容，提取导入、函数和类
// func (p *CodeParser) ParseCode(content string, language string) []CodeElement {
// 	elements := []CodeElement{}

// 	// 如果没有语言特定的解析器，使用通用解析
// 	if language == "" {
// 		language = "text"
// 	}

// 	// 提取导入
// 	if pattern, ok := p.importPatterns[language]; ok {
// 		matches := pattern.FindAllStringSubmatch(content, -1)
// 		for _, match := range matches {
// 			var importStr string
// 			if len(match) > 1 && match[1] != "" {
// 				importStr = match[1]
// 			} else if len(match) > 2 {
// 				importStr = match[2]
// 			}

// 			if importStr != "" {
// 				element := CodeElement{
// 					Type:    "import",
// 					Name:    importStr,
// 					Content: match[0],
// 				}
// 				elements = append(elements, element)
// 			}
// 		}
// 	}

// 	// 提取函数
// 	if pattern, ok := p.functionPatterns[language]; ok {
// 		matches := pattern.FindAllStringSubmatchIndex(content, -1)
// 		for _, match := range matches {
// 			if len(match) >= 2 {
// 				startIdx := match[0]
// 				// 找到函数体的结束位置（简化版，实际需要考虑嵌套括号）
// 				endIdx := findClosingBrace(content, match[1])
// 				if endIdx > startIdx {
// 					functionContent := content[startIdx : endIdx+1]
// 					functionName := extractName(pattern, content, match)

// 					element := CodeElement{
// 						Type:      "function",
// 						Name:      functionName,
// 						Content:   functionContent,
// 						Signature: extractSignature(functionContent, language),
// 						StartLine: countLines(content[:startIdx]),
// 						EndLine:   countLines(content[:endIdx]),
// 					}
// 					elements = append(elements, element)
// 				}
// 			}
// 		}
// 	}

// 	// 提取类
// 	if pattern, ok := p.classPatterns[language]; ok {
// 		matches := pattern.FindAllStringSubmatchIndex(content, -1)
// 		for _, match := range matches {
// 			if len(match) >= 2 {
// 				startIdx := match[0]
// 				// 找到类体的结束位置
// 				endIdx := findClosingBrace(content, match[1])
// 				if endIdx > startIdx {
// 					classContent := content[startIdx : endIdx+1]
// 					className := extractName(pattern, content, match)

// 					element := CodeElement{
// 						Type:      "class",
// 						Name:      className,
// 						Content:   classContent,
// 						Signature: extractSignature(classContent, language),
// 						StartLine: countLines(content[:startIdx]),
// 						EndLine:   countLines(content[:endIdx]),
// 					}
// 					elements = append(elements, element)
// 				}
// 			}
// 		}
// 	}

// 	return elements
// }

// // ExtractImports 提取代码中的导入语句
// func (p *CodeParser) ExtractImports(content string, language string) []string {
// 	imports := []string{}

// 	if pattern, ok := p.importPatterns[language]; ok {
// 		matches := pattern.FindAllStringSubmatch(content, -1)
// 		for _, match := range matches {
// 			var importStr string
// 			if len(match) > 1 && match[1] != "" {
// 				// 多行导入 (Go)
// 				importBlock := match[1]
// 				for _, line := range strings.Split(importBlock, "\n") {
// 					line = strings.TrimSpace(line)
// 					if line != "" && !strings.HasPrefix(line, "//") {
// 						// 提取引号中的内容
// 						importMatch := regexp.MustCompile(`"([^"]+)"`).FindStringSubmatch(line)
// 						if len(importMatch) > 1 {
// 							imports = append(imports, importMatch[1])
// 						}
// 					}
// 				}
// 			} else if len(match) > 2 && match[2] != "" {
// 				// 单行导入
// 				importStr = strings.Trim(match[2], `"'`)
// 				imports = append(imports, importStr)
// 			} else if len(match) > 0 {
// 				// JavaScript/TypeScript 或 Python 导入
// 				importMatch := regexp.MustCompile(`['"]([^'"]+)['"]`).FindStringSubmatch(match[0])
// 				if len(importMatch) > 1 {
// 					imports = append(imports, importMatch[1])
// 				}
// 			}
// 		}
// 	}

// 	return imports
// }

// // GetCodeContext 获取光标位置的代码上下文
// func (p *CodeParser) GetCodeContext(content string, language string, cursorPosition int) *CodeElement {
// 	// 找到光标所在的行
// 	cursorLine := countLines(content[:cursorPosition])

// 	// 解析代码
// 	elements := p.ParseCode(content, language)

// 	// 找到包含光标位置的代码元素
// 	for _, element := range elements {
// 		if cursorLine >= element.StartLine && cursorLine <= element.EndLine {
// 			return &element
// 		}
// 	}

// 	return nil
// }

// // DetectLanguage 根据文件扩展名检测语言
// func (p *CodeParser) DetectLanguage(filename string) string {
// 	ext := strings.ToLower(filepath.Ext(filename))

// 	switch ext {
// 	case ".go":
// 		return "go"
// 	case ".js":
// 		return "javascript"
// 	case ".jsx":
// 		return "javascript"
// 	case ".ts":
// 		return "typescript"
// 	case ".tsx":
// 		return "typescript"
// 	case ".py":
// 		return "python"
// 	case ".java":
// 		return "java"
// 	case ".c":
// 		return "c"
// 	case ".cpp", ".cc", ".cxx":
// 		return "cpp"
// 	case ".cs":
// 		return "csharp"
// 	case ".php":
// 		return "php"
// 	case ".rb":
// 		return "ruby"
// 	case ".swift":
// 		return "swift"
// 	case ".kt":
// 		return "kotlin"
// 	case ".rs":
// 		return "rust"
// 	default:
// 		return "text"
// 	}
// }

// // 辅助函数

// // findClosingBrace 查找匹配的闭合括号
// func findClosingBrace(content string, startPos int) int {
// 	if startPos >= len(content) {
// 		return -1
// 	}

// 	// 找到开始的 '{'
// 	bracePos := strings.Index(content[startPos:], "{")
// 	if bracePos == -1 {
// 		return -1
// 	}

// 	bracePos += startPos
// 	braceCount := 1

// 	// 寻找匹配的闭合括号
// 	for i := bracePos + 1; i < len(content); i++ {
// 		if content[i] == '{' {
// 			braceCount++
// 		} else if content[i] == '}' {
// 			braceCount--
// 			if braceCount == 0 {
// 				return i
// 			}
// 		}
// 	}

// 	return -1
// }

// // countLines 计算文本中的行数
// func countLines(text string) int {
// 	return strings.Count(text, "\n") + 1
// }

// // extractName 从正则表达式匹配中提取名称
// func extractName(pattern *regexp.Regexp, content string, match []int) string {
// 	// 获取完整匹配的子字符串
// 	fullMatch := content[match[0]:match[1]]

// 	// 根据正则表达式重新匹配以获取捕获组
// 	subMatches := pattern.FindStringSubmatch(fullMatch)

// 	// 查找第一个非空的捕获组作为名称
// 	for i := 1; i < len(subMatches); i++ {
// 		if subMatches[i] != "" {
// 			return subMatches[i]
// 		}
// 	}

// 	return ""
// }

// // extractSignature 提取函数或类的签名
// func extractSignature(content string, language string) string {
// 	// 简化版，仅提取第一行作为签名
// 	lines := strings.Split(content, "\n")
// 	if len(lines) > 0 {
// 		firstLine := strings.TrimSpace(lines[0])

// 		// 移除尾部的 { 或 :
// 		firstLine = strings.TrimRight(firstLine, " {:")

// 		return firstLine
// 	}

// 	return ""
// }
