package tools

import (
	"fmt"
	"sort"
	"strings"

	"mind-weaver/internal/third/prompts/sections"
	"mind-weaver/internal/third/toolgroups"
)

func GetExecuteCommandDescription(args ToolDescriptionGenArgs) string {
	return fmt.Sprintf(`## execute_command
Description: Request to execute a CLI command on the system. Use this when you need to perform system operations or run specific commands to accomplish any step in the user's task. You must tailor your command to the user's system and provide a clear explanation of what the command does. For command chaining, use the appropriate chaining syntax for the user's shell. Prefer to execute complex CLI commands over creating executable scripts, as they are more flexible and easier to run. Prefer relative commands and paths that avoid location sensitivity for terminal consistency, e.g: `+"`touch ./testdata/example.file`"+`, `+"`dir ./examples/model1/data/yaml`"+`, or ·`+"`go test ./cmd/front --config ./cmd/front/config.yml`"+`. If directed by the user, you may open a terminal in a different directory by using the `+"`cwd`"+` parameter.
Parameters:
- command: (required) The CLI command to execute. This should be valid for the current operating system. Ensure the command is properly formatted and does not contain any harmful instructions.
- cwd: (optional) The working directory to execute the command in (default: %s)
Usage:
<execute_command>
<command>Your command here</command>
<cwd>Working directory path (optional)</cwd>
</execute_command>

Example: Requesting to execute npm run dev
<execute_command>
<command>npm run dev</command>
</execute_command>

Example: Requesting to execute ls in a specific directory if directed
<execute_command>
<command>ls -la</command>
<cwd>/home/user/projects</cwd>
</execute_command>`, args.Cwd)
}

func GetReadFileDescription(args ToolDescriptionGenArgs) string {
	return fmt.Sprintf(`## read_file
Description: Request to read the contents of a file at the specified path. Use this when you need to examine the contents of an existing file you do not know the contents of, for example to analyze code, review text files, or extract information from configuration files. The output includes line numbers prefixed to each line (e.g. "1 | const x = 1"), making it easier to reference specific lines when creating diffs or discussing code. By specifying start_line and end_line parameters, you can efficiently read specific portions of large files without loading the entire file into memory. Automatically extracts raw text from PDF and DOCX files. May not be suitable for other types of binary files, as it returns the raw content as a string.
Parameters:
- path: (required) The path of the file to read (relative to the current workspace directory %s)
- start_line: (optional) The starting line number to read from (1-based). If not provided, it starts from the beginning of the file.
- end_line: (optional) The ending line number to read to (1-based, inclusive). If not provided, it reads to the end of the file.
Usage:
<read_file>
<path>File path here</path>
<start_line>Starting line number (optional)</start_line>
<end_line>Ending line number (optional)</end_line>
</read_file>

Examples:

1. Reading an entire file:
<read_file>
<path>frontend-config.json</path>
</read_file>

2. Reading the first 1000 lines of a large log file:
<read_file>
<path>logs/application.log</path>
<end_line>1000</end_line>
</read_file>

3. Reading lines 500-1000 of a CSV file:
<read_file>
<path>data/large-dataset.csv</path>
<start_line>500</start_line>
<end_line>1000</end_line>
</read_file>

4. Reading a specific function in a source file:
<read_file>
<path>src/app.ts</path>
<start_line>46</start_line>
<end_line>68</end_line>
</read_file>

Note: When both start_line and end_line are provided, this tool efficiently streams only the requested lines, making it suitable for processing large files like logs, CSV files, and other large datasets without memory issues.`, args.Cwd)
}

func GetWriteToFileDescription(args ToolDescriptionGenArgs) string {
	return fmt.Sprintf(`## write_to_file
Description: Request to write full content to a file at the specified path. If the file exists, it will be overwritten with the provided content. If the file doesn't exist, it will be created. This tool will automatically create any directories needed to write the file.
Parameters:
- path: (required) The path of the file to write to (relative to the current workspace directory %s)
- content: (required) The content to write to the file. ALWAYS provide the COMPLETE intended content of the file, without any truncation or omissions. You MUST include ALL parts of the file, even if they haven't been modified. Do NOT include the line numbers in the content though, just the actual content of the file.
- line_count: (required) The number of lines in the file. Make sure to compute this based on the actual content of the file, not the number of lines in the content you're providing.
Usage:
<write_to_file>
<path>File path here</path>
<content>
Your file content here
</content>
<line_count>total number of lines in the file, including empty lines</line_count>
</write_to_file>

Example: Requesting to write to frontend-config.json
<write_to_file>
<path>frontend-config.json</path>
<content>
{
  "apiEndpoint": "https://api.example.com",
  "theme": {
    "primaryColor": "#007bff",
    "secondaryColor": "#6c757d",
    "fontFamily": "Arial, sans-serif"
  },
  "features": {
    "darkMode": true,
    "notifications": true,
    "analytics": false
  },
  "version": "1.0.0"
}
</content>
<line_count>14</line_count>
</write_to_file>`, args.Cwd)
}

func GetAskFollowupQuestionDescription(args ToolDescriptionGenArgs) string {
	return `## ask_followup_question
Description: Ask the user a question to gather additional information needed to complete the task. This tool should be used when you encounter ambiguities, need clarification, or require more details to proceed effectively. It allows for interactive problem-solving by enabling direct communication with the user. Use this tool judiciously to maintain a balance between gathering necessary information and avoiding excessive back-and-forth.
Parameters:
- question: (required) The question to ask the user. This should be a clear, specific question that addresses the information you need.
- follow_up: (required) A list of 2-4 suggested answers that logically follow from the question, ordered by priority or logical sequence. Each suggestion must:
  1. Be provided in its own <suggest> tag
  2. Be specific, actionable, and directly related to the completed task
  3. Be a complete answer to the question - the user should not need to provide additional information or fill in any missing details. DO NOT include placeholders with brackets or parentheses.
Usage:
<ask_followup_question>
<question>Your question here</question>
<follow_up>
<suggest>
Your suggested answer here
</suggest>
</follow_up>
</ask_followup_question>

Example: Requesting to ask the user for the path to the frontend-config.json file
<ask_followup_question>
<question>What is the path to the frontend-config.json file?</question>
<follow_up>
<suggest>./src/frontend-config.json</suggest>
<suggest>./config/frontend-config.json</suggest>
<suggest>./frontend-config.json</suggest>
</follow_up>
</ask_followup_question>`
}

func GetAttemptCompletionDescription(args ToolDescriptionGenArgs) string {
	return `## attempt_completion
Description: After each tool use, the user will respond with the result of that tool use, i.e. if it succeeded or failed, along with any reasons for failure. Once you've received the results of tool uses and can confirm that the task is complete, use this tool to present the result of your work to the user. Optionally you may provide a CLI command to showcase the result of your work. The user may respond with feedback if they are not satisfied with the result, which you can use to make improvements and try again.
IMPORTANT NOTE: This tool CANNOT be used until you've confirmed from the user that any previous tool uses were successful. Failure to do so will result in code corruption and system failure. Before using this tool, you must ask yourself in <thinking></thinking> tags if you've confirmed from the user that any previous tool uses were successful. If not, then DO NOT use this tool.
Parameters:
- result: (required) The result of the task. Formulate this result in a way that is final and does not require further input from the user. Don't end your result with questions or offers for further assistance.
- command: (optional) A CLI command to execute to show a live demo of the result to the user. For example, use ` + "open index.html" +
		` to display a created html website, or ` + "open localhost:3000" + ` to display a locally running development server. But DO NOT use commands like ` + "echo" +
		` or ` + "cat" + ` that merely print text. This command should be valid for the current operating system. Ensure the command is properly formatted and does not contain any harmful instructions.
Usage:
<attempt_completion>
<result>
Your final result description here
</result>
<command>Command to demonstrate result (optional)</command>
</attempt_completion>

Example: Requesting to attempt completion with a result and command
<attempt_completion>
<result>
I've updated the CSS
</result>
<command>open index.html</command>
</attempt_completion>`
}

func GetListFilesDescription(args ToolDescriptionGenArgs) string {
	return fmt.Sprintf(`## list_files
Description: Request to list files and directories within the specified directory. If recursive is true, it will list all files and directories recursively. If recursive is false or not provided, it will only list the top-level contents. Do not use this tool to confirm the existence of files you may have created, as the user will let you know if the files were created successfully or not.
Parameters:
- path: (required) The path of the directory to list contents for (relative to the current workspace directory  %s)
- recursive: (optional) Whether to list files recursively. Use true for recursive listing, false or omit for top-level only.
Usage:
<list_files>
<path>Directory path here</path>
<recursive>true or false (optional)</recursive>
</list_files>

Example: Requesting to list all files in the current directory
<list_files>
<path>.</path>
<recursive>false</recursive>
</list_files>`, args.Cwd)
}

func GetSearchFilesDescription(args ToolDescriptionGenArgs) string {
	return fmt.Sprintf(`## search_files
Description: Request to perform a regex search across files in a specified directory, providing context-rich results. This tool searches for patterns or specific content across multiple files, displaying each match with encapsulating context.
Parameters:
- path: (required) The path of the directory to search in (relative to the current workspace directory %s). This directory will be recursively searched.
- regex: (required) The regular expression pattern to search for. Uses Rust regex syntax.
- file_pattern: (optional) Glob pattern to filter files (e.g., '*.ts' for TypeScript files). If not provided, it will search all files (*).
Usage:
<search_files>
<path>Directory path here</path>
<regex>Your regex pattern here</regex>
<file_pattern>file pattern here (optional)</file_pattern>
</search_files>

Example: Requesting to search for all .ts files in the current directory
<search_files>
<path>.</path>
<regex>.*</regex>
<file_pattern>*.ts</file_pattern>
</search_files>`, args.Cwd)
}

func GetListCodeDefinitionNamesDescription(args ToolDescriptionGenArgs) string {
	return fmt.Sprintf(`## list_code_definition_names
Description: Request to list definition names (classes, functions, methods, etc.) from source code. This tool can analyze either a single file or all files at the top level of a specified directory. It provides insights into the codebase structure and important constructs, encapsulating high-level concepts and relationships that are crucial for understanding the overall architecture.
Parameters:
- path: (required) The path of the file or directory (relative to the current working directory %s) to analyze. When given a directory, it lists definitions from all top-level source files.
Usage:
<list_code_definition_names>
<path>Directory path here</path>
</list_code_definition_names>

Examples:

1. List definitions from a specific file:
<list_code_definition_names>
<path>src/main.ts</path>
</list_code_definition_names>

2. List definitions from all files in a directory:
<list_code_definition_names>
<path>src/</path>
</list_code_definition_names>`, args.Cwd)
}

func GetInsertContentDescription(args ToolDescriptionGenArgs) string {
	return fmt.Sprintf(`## insert_content
Description: Inserts content at specific line positions in a file. This is the primary tool for adding new content and code (functions/methods/classes, imports, attributes etc.) as it allows for precise insertions without overwriting existing content. The tool uses an efficient line-based insertion system that maintains file integrity and proper ordering of multiple insertions. Beware to use the proper indentation. This tool is the preferred way to add new content and code to files.
Parameters:
- path: (required) The path of the file to insert content into (relative to the current workspace directory %s)
- operations: (required) A JSON array of insertion operations. Each operation is an object with:
    * start_line: (required) The line number where the content should be inserted.  The content currently at that line will end up below the inserted content.
    * content: (required) The content to insert at the specified position. IMPORTANT NOTE: If the content is a single line, it can be a string. If it's a multi-line content, it should be a string with newline characters (\n) for line breaks. Make sure to include the correct indentation for the content.
Usage:
<insert_content>
<path>File path here</path>
<operations>[
  {
    "start_line": 10,
    "content": "Your content here"
  }
]</operations>
</insert_content>
Example: Insert a new function and its import statement
<insert_content>
<path>File path here</path>
<operations>[
  {
    "start_line": 1,
    "content": "import { sum } from './utils';"
  },
  {
    "start_line": 10,
    "content": "function calculateTotal(items: number[]): number {\n    return items.reduce((sum, item) => sum + item, 0);\n}"
  }
]</operations>
</insert_content>`, args.Cwd)
}

// ... Add functions for apply_diff (using args.DiffStrategy.GetToolDescription),
// ... insert_content, search_and_replace, browser_action (checking args.SupportsComputerUse)

// Map tool names to their description functions
var toolDescriptionMap = map[toolgroups.ToolName]func(ToolDescriptionGenArgs) string{

	toolgroups.ToolExecuteCommand:          GetExecuteCommandDescription,
	toolgroups.ToolReadFile:                GetReadFileDescription,
	toolgroups.ToolWriteToFile:             GetWriteToFileDescription,
	toolgroups.ToolListFiles:               GetListFilesDescription,
	toolgroups.ToolSearchFiles:             GetSearchFilesDescription,
	toolgroups.ToolListCodeDefinitionNames: GetListCodeDefinitionNamesDescription,
	toolgroups.ToolAskFollowupQuestion:     GetAskFollowupQuestionDescription,
	toolgroups.ToolAttemptCompletion:       GetAttemptCompletionDescription,
	toolgroups.ToolInsertContent:           GetInsertContentDescription,
	// Add other tools here...
	// toolgroups.ToolApplyDiff: func(args ToolDescriptionGenArgs) string {
	//     if args.DiffStrategy != nil {
	//         // Assuming DiffStrategy.GetToolDescription takes diff.ToolDescriptionArgs
	//         diffArgs := diff.ToolDescriptionArgs{Cwd: args.Cwd}
	//         return args.DiffStrategy.GetToolDescription(diffArgs)
	//     }
	//     return ""
	// },
	// ... browser_action, insert_content, search_and_replace ...

}

// GetToolDescriptionsForMode generates the combined tool description string for a given mode.
func GetToolDescriptionsForMode(mode sections.ModeSlug, args ToolDescriptionGenArgs, customModes []sections.ModeConfig) string {

	config := sections.GetModeBySlug(mode, customModes)
	if config == nil {
		// Handle error or fallback
		config = sections.GetDefaultMode() // Example fallback
	}

	allowedTools := make(map[toolgroups.ToolName]bool)

	// Add tools from mode's groups
	for _, groupEntry := range config.Groups {
		groupName := sections.GetGroupName(groupEntry) // Needs implementation
		toolGroup, ok := toolgroups.TOOL_GROUPS[groupName]
		if ok {
			for _, tool := range toolGroup.Tools {
				// Add experiment check here if needed: isToolAllowedForMode(tool, mode, customModes, experiments)
				allowedTools[tool] = true
			}
		}
	}

	// Add always available tools
	for _, tool := range toolgroups.ALWAYS_AVAILABLE_TOOLS {
		allowedTools[tool] = true
	}

	var descriptions []string
	toolNames := make([]toolgroups.ToolName, 0, len(allowedTools))
	for toolName := range allowedTools {
		toolNames = append(toolNames, toolName)
	}
	// Sort tool names for consistent output
	sort.Slice(toolNames, func(i, j int) bool {
		return toolNames[i] < toolNames[j]
	})

	for _, toolName := range toolNames {
		if descFunc, ok := toolDescriptionMap[toolName]; ok {
			desc := descFunc(args)
			if desc != "" {
				descriptions = append(descriptions, desc)
			}
		} else {
			fmt.Printf("Warning: No description function found for tool: %s\n", toolName)
		}
	}

	return fmt.Sprintf("# Tools\n\n%s", strings.Join(descriptions, "\n\n"))
}
