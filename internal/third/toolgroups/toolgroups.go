package toolgroups

// ToolName represents the name of a tool.
type ToolName string

// Constants for all known tool names
const (
	ToolExecuteCommand          ToolName = "execute_command"
	ToolReadFile                ToolName = "read_file"
	ToolWriteToFile             ToolName = "write_to_file"
	ToolApplyDiff               ToolName = "apply_diff"
	ToolInsertContent           ToolName = "insert_content"
	ToolSearchAndReplace        ToolName = "search_and_replace"
	ToolSearchFiles             ToolName = "search_files"
	ToolListFiles               ToolName = "list_files"
	ToolListCodeDefinitionNames ToolName = "list_code_definition_names"
	ToolBrowserAction           ToolName = "browser_action"
	ToolAskFollowupQuestion     ToolName = "ask_followup_question"
	ToolAttemptCompletion       ToolName = "attempt_completion"
	// Add any other tools like use_mcp_tool if re-added
)

// ToolGroupName identifies a category of tools.
type ToolGroupName string

// Constants for tool group names
const (
	GroupRead    ToolGroupName = "read"
	GroupEdit    ToolGroupName = "edit"
	GroupCommand ToolGroupName = "command"
	GroupBrowser ToolGroupName = "browser"
	GroupMcp     ToolGroupName = "mcp" // Keep if MCP tools might be re-added
)

// ToolGroup defines the tools belonging to a specific group.
type ToolGroup struct {
	Name  ToolGroupName
	Tools []ToolName
}

// TOOL_GROUPS maps group names to their definitions.
var TOOL_GROUPS = map[ToolGroupName]ToolGroup{
	GroupRead: {
		Name: GroupRead,
		Tools: []ToolName{
			ToolReadFile,
			ToolSearchFiles,
			ToolListFiles,
			ToolListCodeDefinitionNames,
			// Add fetch_instructions if re-added
		},
	},
	GroupEdit: {
		Name: GroupEdit,
		Tools: []ToolName{
			ToolWriteToFile,
			ToolApplyDiff,        // Diff strategy availability checked elsewhere
			ToolInsertContent,    // Experiment availability checked elsewhere
			ToolSearchAndReplace, // Experiment availability checked elsewhere
		},
	},
	GroupCommand: {
		Name:  GroupCommand,
		Tools: []ToolName{ToolExecuteCommand},
	},
	GroupBrowser: {
		Name:  GroupBrowser,
		Tools: []ToolName{ToolBrowserAction}, // Browser support checked elsewhere
	},
	// Add GroupMcp if needed
}

// ALWAYS_AVAILABLE_TOOLS lists tools accessible regardless of mode group configuration.
var ALWAYS_AVAILABLE_TOOLS = []ToolName{
	ToolAskFollowupQuestion,
	ToolAttemptCompletion,
	// Add switch_mode, new_task if re-added
}
