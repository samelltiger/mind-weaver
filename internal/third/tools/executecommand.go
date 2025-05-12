package tools

import (
	"context"
	"fmt"
	"html" // For unescaping
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
	"os/exec"
	"strings"
	"syscall"
)

// ExecuteCommandTool executes a shell command.
func ExecuteCommandTool(input ExecutorInput) (*ExecutorResult, error) {
	commandStr, ok := input.ToolUse.Params[string(assistantmessage.Command)]
	if !ok || strings.TrimSpace(commandStr) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Command))
		return &ExecutorResult{Result: errText, IsError: true}, nil // Return error message for LLM
	}
	customCwd, _ := input.ToolUse.Params[string(assistantmessage.Cwd)]

	// Unescape HTML entities
	commandStr = html.UnescapeString(commandStr)

	// Validate against rooignore (simplified)
	if input.RooIgnoreController != nil {
		ignoredPath := input.RooIgnoreController.ValidateCommand(commandStr)
		if ignoredPath != "" {
			errText := prompts.FormatRooIgnoreError(ignoredPath)
			// Wrap in a tool error format for the LLM
			return &ExecutorResult{Result: prompts.FormatToolError(errText), IsError: true}, nil
		}
	}

	// --- Actual Command Execution ---
	// Warning: Executing arbitrary commands is dangerous!
	// Ensure proper sandboxing, validation, and security measures are in place.
	// This is a simplified example.
	ctx := input.Ctx
	if ctx == nil {
		ctx = context.Background() // Default context
	}

	// Consider using a shell explicitly for complex commands/chaining
	// e.g., exec.CommandContext(ctx, "sh", "-c", commandStr)
	// or exec.CommandContext(ctx, "bash", "-c", commandStr)
	cmd := exec.CommandContext(ctx, "bash", "-c", commandStr) // Using sh -c for basic shell features

	if customCwd != "" {
		// TODO: Validate customCwd path before using
		cmd.Dir = customCwd
	} else {
		cmd.Dir = input.Cwd
	}

	outputBytes, err := cmd.CombinedOutput() // Get stdout and stderr
	output := string(outputBytes)

	result := fmt.Sprintf("Command executed: %s\nWorking Directory: %s\nOutput:\n%s", commandStr, cmd.Dir, output)

	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		result += fmt.Sprintf("\nError: %v (Exit Code: %d)", err, exitCode)
		// Decide if the execution error itself should be returned as err,
		// or just included in the result string for the LLM.
		// Returning it in Result for now.
		return &ExecutorResult{Result: result}, nil // Don't set IsError=true, let LLM see the command failed
	}

	return &ExecutorResult{Result: result}, nil
}
