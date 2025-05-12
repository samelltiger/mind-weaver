package sections

import (
	"fmt"
	"mind-weaver/internal/third/diff" // Adjust
	"strings"
)

func GetCapabilitiesSection(cwd string, supportsComputerUse bool, diffStrategy diff.DiffStrategy) string {
	var b strings.Builder
	b.WriteString("====\n\nCAPABILITIES\n\n")
	b.WriteString(fmt.Sprintf("- You have access to tools that let you execute CLI commands on the user's computer, list files, view source code definitions, regex search%s, read and write files, and ask follow-up questions...\n",
		map[bool]string{true: ", use the browser", false: ""}[supportsComputerUse])) // Go doesn't have ternary
	b.WriteString(fmt.Sprintf("- When the user initially gives you a task, a recursive list of all filepaths in the current workspace directory ('%s') will be included in environment_details. This provides an overview of the project's file structure, offering key insights into the project from directory/file names (how developers conceptualize and organize their code) and file extensions (the language used). This can also guide decision-making on which files to explore further. If you need to further explore directories such as outside the current workspace directory, you can use the list_files tool. If you pass 'true' for the recursive parameter, it will list files recursively. Otherwise, it will list files at the top level, which is better suited for generic directories where you don't necessarily need the nested structure, like the Desktop.\n", cwd)) // Use actual CWD
	b.WriteString("- You can use search_files to perform regex searches across files in a specified directory, outputting context-rich results that include surrounding lines. This is particularly useful for understanding code patterns, finding specific implementations, or identifying areas that need refactoring.\n")
	b.WriteString("- You can use the list_code_definition_names tool to get an overview of source code definitions for all files at the top level of a specified directory. This can be particularly useful when you need to understand the broader context and relationships between certain parts of the code. You may need to call this tool multiple times to understand various parts of the codebase related to the task.\n")

	editTools := "the write_to_file"
	if diffStrategy != nil {
		editTools = "the apply_diff or write_to_file"
	}
	b.WriteString(fmt.Sprintf("    - For example, when asked to make edits or improvements you might analyze the file structure in the initial environment_details to get an overview of the project, then use list_code_definition_names to get further insight using source code definitions for files located in relevant directories, then read_file to examine the contents of relevant files, analyze the code and suggest improvements or make necessary edits, then use  %s tool to apply the changes. If you refactored code that could affect other parts of the codebase, you could use search_files to ensure you update other files as needed.\n", editTools))

	b.WriteString("- You can use the execute_command tool to run commands on the user's computer whenever you feel it can help accomplish the user's task. When you need to execute a CLI command, you must provide a clear explanation of what the command does. Prefer to execute complex CLI commands over creating executable scripts, since they are more flexible and easier to run. Interactive and long-running commands are allowed, since the commands are run in the user's VSCode terminal. The user may keep commands running in the background and you will be kept updated on their status along the way. Each command you execute is run in a new terminal instance.\n")

	if supportsComputerUse {
		b.WriteString("\n- You can use the browser_action tool to interact with websites (including html files and locally running development servers) through a Puppeteer-controlled browser when you feel it is necessary in accomplishing the user's task. This tool is particularly useful for web development tasks as it allows you to launch a browser, navigate to pages, interact with elements through clicks and keyboard input, and capture the results through screenshots and console logs. This tool may be useful at key stages of web development tasks-such as after implementing new features, making substantial changes, when troubleshooting issues, or to verify the result of your work. You can analyze the provided screenshots to ensure correct rendering or identify errors, and review console logs for runtime issues.\n  - For example, if asked to add a component to a react website, you might create the necessary files, use execute_command to run the site locally, then use browser_action to launch the browser, navigate to the local server, and verify the component renders & functions correctly before closing the browser.\n")
	}

	// Removed MCP section

	return b.String()
}
