package tools

import (
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
	"strings"
)

// AttemptCompletionTool signals the AI believes the task is complete.
// The backend verifies parameters and passes the signal + content up.
func AttemptCompletionTool(input ExecutorInput) (*ExecutorResult, error) {
	resultText, ok := input.ToolUse.Params[string(assistantmessage.Result)]
	if !ok || strings.TrimSpace(resultText) == "" { // Result text is mandatory
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Result))
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}
	commandStr, _ := input.ToolUse.Params[string(assistantmessage.Command)] // Optional

	// Similar to AskFollowup, the backend signals completion attempt.
	// The controlling system handles presenting result/command to user.
	signal := fmt.Sprintf("SYSTEM_SIGNAL: ATTEMPT_COMPLETION\nResult: %s", resultText)
	if commandStr != "" {
		signal += fmt.Sprintf("\nCommand: %s", commandStr)
		// TODO: Optionally execute the commandStr here if the backend is responsible for demos.
		// Be very careful with security if executing commands based on LLM output.
		// _, cmdErr := ExecuteCommandTool(ExecutorInput{..., ToolUse: ... with command ...}) // Example
		// if cmdErr != nil { signal += fmt.Sprintf("\nCommand Execution Error: %v", cmdErr)}
	}

	return &ExecutorResult{Result: signal}, nil
}
