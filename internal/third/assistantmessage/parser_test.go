package assistantmessage

import (
	"reflect"
	"testing"
)

func TestParseAssistantMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []AssistantMessageContent
	}{
		{
			name:  "Simple text message",
			input: "Hello, how can I help you?",
			expected: []AssistantMessageContent{
				TextContent{
					Type:    "text",
					Content: "Hello, how can I help you?",
					Partial: true,
				},
			},
		},
		{
			name:  "Simple thinking message",
			input: "好的，HTML 文件已创建成功。\n\n<thinking>\n1. 分析用户需求：用户需要一个简单的HTML \"Hello World\"演示\n2. 检查环境信息：当前工作目录是/mnt/h/code/codepilot\n3. 确定解决方案：创建一个包含基本HTML结构的index.html文件\n4. 选择工具：使用write_to_file工具创建新文件\n5. 参数确认：\n   - path: ./index.html (相对路径)\n   - content: 完整的HTML代码\n   - line_count: 计算内容行数\n6. 无需额外信息，可以直接执行\n</thinking>\n\n将其设置成TextContent，type为thinking，content为thinking标签中的文本内容",
			expected: []AssistantMessageContent{
				TextContent{
					Type:    "text",
					Content: "好的，HTML 文件已创建成功。",
					Partial: false,
				},
				TextContent{
					Type:    "thinking",
					Content: "\n1. 分析用户需求：用户需要一个简单的HTML \"Hello World\"演示\n2. 检查环境信息：当前工作目录是/mnt/h/code/codepilot\n3. 确定解决方案：创建一个包含基本HTML结构的index.html文件\n4. 选择工具：使用write_to_file工具创建新文件\n5. 参数确认：\n   - path: ./index.html (相对路径)\n   - content: 完整的HTML代码\n   - line_count: 计算内容行数\n6. 无需额外信息，可以直接执行\n",
					Partial: false,
				},
				TextContent{
					Type:    "text",
					Content: "\n\n将其设置成TextContent，type为thinking，content为thinking标签中的文本内容",
					Partial: true,
				},
			},
		},
		{
			name:  "Simple thinking message2",
			input: "<thinking>\n1. 分析用户需求：用户需要一个简单的HTML \"Hello World\"演示\n2. 检查环境信息：当前工作目录是/mnt/h/code/codepilot\n3. 确定解决方案：创建一个包含基本HTML结构的index.html文件\n4. 选择工具：使用write_to_file工具创建新文件\n5. 参数确认：\n   - path: ./index.html (相对路径)\n   - content: 完整的HTML代码\n   - line_count: 计算内容行数\n6. 无需额外信息，可以直接执行",
			expected: []AssistantMessageContent{
				TextContent{
					Type:    "thinking",
					Content: "\n1. 分析用户需求：用户需要一个简单的HTML \"Hello World\"演示\n2. 检查环境信息：当前工作目录是/mnt/h/code/codepilot\n3. 确定解决方案：创建一个包含基本HTML结构的index.html文件\n4. 选择工具：使用write_to_file工具创建新文件\n5. 参数确认：\n   - path: ./index.html (相对路径)\n   - content: 完整的HTML代码\n   - line_count: 计算内容行数\n6. 无需额外信息，可以直接执行\n",
					Partial: true,
				},
			},
		},
		{
			name:  "Simple tool use",
			input: "<execute_command><command>ls -la</command></execute_command>",
			expected: []AssistantMessageContent{
				&ToolUse{
					Type: "tool_use",
					Name: "execute_command",
					Params: map[string]string{
						"command": "ls -la",
					},
					Partial: false,
				},
			},
		},
		{
			name:  "Text followed by tool use",
			input: "Let me check the files. <execute_command><command>ls -la</command></execute_command>",
			expected: []AssistantMessageContent{
				TextContent{
					Type:    "text",
					Content: "Let me check the files.",
					Partial: false,
				},
				&ToolUse{
					Type: "tool_use",
					Name: "execute_command",
					Params: map[string]string{
						"command": "ls -la",
					},
					Partial: false,
				},
			},
		},
		{
			name:  "Tool use followed by text",
			input: "<execute_command><command>ls -la</command></execute_command> Here are your files.",
			expected: []AssistantMessageContent{
				&ToolUse{
					Type: "tool_use",
					Name: "execute_command",
					Params: map[string]string{
						"command": "ls -la",
					},
					Partial: false,
				},
				TextContent{
					Type:    "text",
					Content: " Here are your files.",
					Partial: true,
				},
			},
		},
		{
			name:  "Write to file special case",
			input: "<write_to_file><path>test.txt</path><content>This is a test\nwith multiple lines\n</content></write_to_file>",
			expected: []AssistantMessageContent{
				&ToolUse{
					Type: "tool_use",
					Name: "write_to_file",
					Params: map[string]string{
						"path":    "test.txt",
						"content": "This is a test\nwith multiple lines",
					},
					Partial: false,
				},
			},
		},
		{
			name:  "Partial tool use",
			input: "<execute_command><command>ls -la",
			expected: []AssistantMessageContent{
				&ToolUse{
					Type: "tool_use",
					Name: "execute_command",
					Params: map[string]string{
						"command": "ls -la",
					},
					Partial: true,
				},
			},
		},
		{
			name:  "Multiple tool uses",
			input: "<read_file><path>file.txt</path></read_file><write_to_file><path>output.txt</path><content>New content</content></write_to_file>",
			expected: []AssistantMessageContent{
				&ToolUse{
					Type: "tool_use",
					Name: "read_file",
					Params: map[string]string{
						"path": "file.txt",
					},
					Partial: false,
				},
				&ToolUse{
					Type: "tool_use",
					Name: "write_to_file",
					Params: map[string]string{
						"path":    "output.txt",
						"content": "New content",
					},
					Partial: false,
				},
			},
		},
		{
			name:  "Broken tool tag in text",
			input: "<execute_command",
			expected: []AssistantMessageContent{
				TextContent{
					Type:    "text",
					Content: "<execute_command",
					Partial: true,
				},
			},
		},
		{
			name:  "attempt completion tag in text",
			input: "好的，HTML 文件已创建成功。\n\n<attempt_completion>\n<result>我创建了一个名为 `index.html` 的文件，其中包含绘制正弦波所需的 HTML、CSS 和 JavaScript。</result><command>open index.html</command></attempt_completion>",
			expected: []AssistantMessageContent{
				TextContent{
					Type:    "text",
					Content: "好的，HTML 文件已创建成功。",
					Partial: false,
				},
				&ToolUse{
					Type: "tool_use",
					Name: "attempt_completion",
					Params: map[string]string{
						"command": "open index.html",
						"result":  "我创建了一个名为 `index.html` 的文件，其中包含绘制正弦波所需的 HTML、CSS 和 JavaScript。",
					},
					Partial: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseAssistantMessage(tc.input)

			if len(result) != len(tc.expected) {
				t.Fatalf("Expected %d content blocks, got %d", len(tc.expected), len(result))
			}

			for i, expected := range tc.expected {
				actual := result[i]

				if expected.GetType() != actual.GetType() {
					t.Errorf("Content block %d: expected type %s, got %s", i, expected.GetType(), actual.GetType())
				}

				if expected.IsPartial() != actual.IsPartial() {
					t.Errorf("Content block %d: expected partial %v, got %v", i, expected.IsPartial(), actual.IsPartial())
				}

				if expected.GetType() == "text" {
					expectedText, ok := expected.(TextContent)
					if !ok {
						t.Errorf("Content block %d: expected TextContent type assertion failed", i)
						continue
					}

					actualText, ok := actual.(TextContent)
					if !ok {
						t.Errorf("Content block %d: expected TextContent, got %T", i, actual)
						continue
					}

					if expectedText.Content != actualText.Content {
						t.Errorf("TextContent %d: expected content %q, got %q", i, expectedText.Content, actualText.Content)
					}
				} else if expected.GetType() == "tool_use" {
					expectedTool, ok := expected.(*ToolUse)
					if !ok {
						t.Errorf("Content block %d: expected *ToolUse type assertion failed", i)
						continue
					}

					actualTool, ok := actual.(*ToolUse)
					if !ok {
						t.Errorf("Content block %d: expected ToolUse, got %T", i, actual)
						continue
					}

					if expectedTool.Name != actualTool.Name {
						t.Errorf("ToolUse %d: expected name %q, got %q", i, expectedTool.Name, actualTool.Name)
					}

					if !reflect.DeepEqual(expectedTool.Params, actualTool.Params) {
						t.Errorf("ToolUse %d: params don't match\nExpected: %v\nActual: %v", i, expectedTool.Params, actualTool.Params)
					}
				}
			}
		})
	}
}
