package tools

import (
	"context"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/diff"
	"mind-weaver/internal/third/ignore"
)

// ExecutorInput holds all necessary context for executing a tool.
type ExecutorInput struct {
	Ctx                 context.Context // For cancellation
	ToolUse             assistantmessage.ToolUse
	Cwd                 string
	RooIgnoreController *ignore.RooIgnoreController // Can be nil
	DiffStrategy        diff.DiffStrategy           // Can be nil
	Confirmed           bool
	// Add any other required context (e.g., UserID, SessionID)
}

// ExecutorResult represents the outcome of a tool execution.
// This is the data that will be formatted (using prompts/responses) and sent back to the LLM.
type ExecutorResult struct {
	Result string // The primary text result for the LLM
	Error  error  // Any error that occurred during execution
	// Add fields for specific tool outputs if needed (e.g., file list, search results)
	IsError bool // Indicates if 'Result' is an error message for the LLM
}

// ToolExecutor defines the interface for a tool execution function.
type ToolExecutor func(input ExecutorInput) (*ExecutorResult, error)
