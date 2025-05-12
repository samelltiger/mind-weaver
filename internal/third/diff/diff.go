package diff

import "mind-weaver/internal/third/diff/multireplace"

type DiffStrategy interface {
	GetName() string
	GetToolDescription(args multireplace.ToolDescriptionArgs) string
	ApplyDiff(originalContent, diffContent string, startLine, endLine int) (*multireplace.DiffResult, error)
	// GetProgressStatus can be added if needed, returning a simple struct/map
}
