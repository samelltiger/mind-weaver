package assistantmessage

// AssistantMessageContent represents either text content or tool use
type AssistantMessageContent interface {
	IsPartial() bool
	GetType() string
}

// TextContent represents simple text from the assistant
type TextContent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Partial bool   `json:"partial"`
}

func (t TextContent) IsPartial() bool {
	return t.Partial
}

func (t TextContent) GetType() string {
	return t.Type
}

// ToolUse represents a tool call from the assistant
type ToolUse struct {
	Type    string            `json:"type"`
	Name    ToolUseName       `json:"name"`
	Params  map[string]string `json:"params"`
	Partial bool              `json:"partial"`
}

func (t ToolUse) IsPartial() bool {
	return t.Partial
}

func (t ToolUse) GetType() string {
	return t.Type
}

type ToolUseReq struct {
	ToolUse      ToolUse `json:"tool_use"`
	ProjectPath  string  `json:"project_path"`
	ContextFiles []any   `json:"context_files"`
	Model        string  `json:"model"`
	Confirmed    bool    `json:"confirmed"`
}

// ToolUseName is the name of a tool
type ToolUseName string

// ToolParamName is the name of a parameter for a tool
type ToolParamName string

// Tool use name constants
const (
	ExecuteCommand          ToolUseName = "execute_command"
	ReadFile                ToolUseName = "read_file"
	WriteToFile             ToolUseName = "write_to_file"
	ApplyDiff               ToolUseName = "apply_diff"
	InsertContent           ToolUseName = "insert_content"
	SearchAndReplace        ToolUseName = "search_and_replace"
	SearchFiles             ToolUseName = "search_files"
	ListFiles               ToolUseName = "list_files"
	ListCodeDefinitionNames ToolUseName = "list_code_definition_names"
	BrowserAction           ToolUseName = "browser_action"
	UseMcpTool              ToolUseName = "use_mcp_tool"
	AccessMcpResource       ToolUseName = "access_mcp_resource"
	AskFollowupQuestion     ToolUseName = "ask_followup_question"
	AttemptCompletion       ToolUseName = "attempt_completion"
	SwitchMode              ToolUseName = "switch_mode"
	NewTask                 ToolUseName = "new_task"
	FetchInstructions       ToolUseName = "fetch_instructions"
)

// Tool parameter name constants
const (
	Command     ToolParamName = "command"
	Path        ToolParamName = "path"
	Content     ToolParamName = "content"
	LineCount   ToolParamName = "line_count"
	Regex       ToolParamName = "regex"
	FilePattern ToolParamName = "file_pattern"
	Recursive   ToolParamName = "recursive"
	Action      ToolParamName = "action"
	URL         ToolParamName = "url"
	Coordinate  ToolParamName = "coordinate"
	Text        ToolParamName = "text"
	ServerName  ToolParamName = "server_name"
	ToolName    ToolParamName = "tool_name"
	Arguments   ToolParamName = "arguments"
	URI         ToolParamName = "uri"
	Question    ToolParamName = "question"
	Result      ToolParamName = "result"
	Diff        ToolParamName = "diff"
	StartLine   ToolParamName = "start_line"
	EndLine     ToolParamName = "end_line"
	ModeSlug    ToolParamName = "mode_slug"
	Reason      ToolParamName = "reason"
	Operations  ToolParamName = "operations"
	Mode        ToolParamName = "mode"
	Message     ToolParamName = "message"
	Cwd         ToolParamName = "cwd"
	FollowUp    ToolParamName = "follow_up"
	Task        ToolParamName = "task"
	Size        ToolParamName = "size"
)

// AllToolUseNames returns all tool use names as a slice
func AllToolUseNames() []ToolUseName {
	return []ToolUseName{
		ExecuteCommand,
		ReadFile,
		WriteToFile,
		ApplyDiff,
		InsertContent,
		SearchAndReplace,
		SearchFiles,
		ListFiles,
		ListCodeDefinitionNames,
		BrowserAction,
		UseMcpTool,
		AccessMcpResource,
		AskFollowupQuestion,
		AttemptCompletion,
		SwitchMode,
		NewTask,
		FetchInstructions,
	}
}

// AllToolParamNames returns all tool parameter names as a slice
func AllToolParamNames() []ToolParamName {
	return []ToolParamName{
		Command,
		Path,
		Content,
		LineCount,
		Regex,
		FilePattern,
		Recursive,
		Action,
		URL,
		Coordinate,
		Text,
		ServerName,
		ToolName,
		Arguments,
		URI,
		Question,
		Result,
		Diff,
		StartLine,
		EndLine,
		ModeSlug,
		Reason,
		Operations,
		Mode,
		Message,
		Cwd,
		FollowUp,
		Task,
		Size,
	}
}
