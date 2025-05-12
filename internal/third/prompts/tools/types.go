package tools

import (
	"mind-weaver/internal/third/diff" // Adjust
)

// ToolDescriptionGenArgs holds arguments for generating tool descriptions.
type ToolDescriptionGenArgs struct {
	Cwd                 string
	SupportsComputerUse bool
	DiffStrategy        diff.DiffStrategy // Can be nil
	BrowserViewportSize string
	// McpHub // Add back if needed
	// ToolOptions map[string]string // Keep if specific options are needed per tool description
}
