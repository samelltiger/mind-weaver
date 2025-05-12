package tools

import (
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
)

// ToolExecutorMap maps tool names to their execution functions.
var ToolExecutorMap = map[assistantmessage.ToolUseName]ToolExecutor{
	assistantmessage.ExecuteCommand:      ExecuteCommandTool,
	assistantmessage.ReadFile:            ReadFileTool,
	assistantmessage.WriteToFile:         WriteToFileTool,
	assistantmessage.ApplyDiff:           ApplyDiffTool,
	assistantmessage.ListFiles:           ListFilesTool,
	assistantmessage.SearchFiles:         SearchFilesTool,
	assistantmessage.AskFollowupQuestion: AskFollowupQuestionTool,
	assistantmessage.AttemptCompletion:   AttemptCompletionTool,
	// assistantmessage.ListCodeDefinitionNames: ListCodeDefinitionNamesTool,
	assistantmessage.InsertContent: InsertContentTool, // Added
	// assistantmessage.ToolSearchAndReplace:        SearchAndReplaceTool, // Added
	// Add BrowserActionTool if implemented
}

// ExecuteTool selects and runs the appropriate tool executor.
func ExecuteTool(input ExecutorInput) (*ExecutorResult, error) {
	executor, ok := ToolExecutorMap[input.ToolUse.Name]
	if !ok {
		errText := fmt.Sprintf("Unknown tool requested: %s", input.ToolUse.Name)
		return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil // Format as tool error for LLM
	}

	// Check if the tool is enabled via experiments if that logic is needed here
	// if !isToolEnabled(input.ToolUse.Name, input.Experiments) {
	// 	errText := fmt.Sprintf("Tool '%s' is experimental and not enabled.", input.ToolUse.Name)
	// 	return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
	// }

	return executor(input)
}

// isToolEnabled (Example placeholder - implement based on how experiments are managed)
func isToolEnabled(toolName assistantmessage.ToolUseName, experiments map[string]bool) bool {
	switch toolName {
	case assistantmessage.InsertContent:
		return experiments["insert_content"] // Check experiment flag
	case assistantmessage.SearchAndReplace:
		return experiments["search_and_replace"] // Check experiment flag
	default:
		return true // Assume other tools are always enabled
	}
}
