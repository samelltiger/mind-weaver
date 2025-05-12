package prompts

import (
	"fmt"
	"mind-weaver/internal/third/diff"             // Adjust
	"mind-weaver/internal/third/ignore"           // Adjust
	"mind-weaver/internal/third/prompts/sections" // Adjust
	"mind-weaver/internal/third/prompts/tools"    // Adjust
	"strings"
)

// EnvironmentContext holds info passed from the web frontend/API
type EnvironmentContext struct {
	Cwd                 string
	SupportsComputerUse bool   // For browser actions etc.
	BrowserViewportSize string // e.g., "1280x800"
	Language            string // e.g., "en", "fr"
	// Potentially add OS, Shell info if needed by prompts
	// Could also include Experiments map[string]bool
}

// BuildSystemPromptArgs holds all arguments for building the prompt
type BuildSystemPromptArgs struct {
	EnvCtx              EnvironmentContext
	Mode                sections.ModeSlug     // Current mode slug
	CustomModeConfigs   []sections.ModeConfig // User-defined modes
	GlobalInstructions  string
	DiffStrategy        diff.DiffStrategy           // Can be nil if diff not enabled/available
	RooIgnoreController *ignore.RooIgnoreController // Can be nil
	// Add Experiments map[string]bool if needed
}

// BuildSystemPrompt constructs the main system prompt for the LLM.
func BuildSystemPrompt(args BuildSystemPromptArgs) (string, error) {
	var builder strings.Builder

	// 1. Get current mode configuration
	modeConfig := sections.GetModeBySlug(args.Mode, args.CustomModeConfigs) // Needs implementation in modes package
	if modeConfig == nil {
		// Fallback to default mode or return error
		return "", fmt.Errorf("mode '%s' not found", args.Mode)
		// modeConfig = modes.GetDefaultMode() // Assuming a default mode exists
	}

	// --- File-based Custom Prompt (Optional) ---
	// customPromptPath := filepath.Join(args.EnvCtx.Cwd, ".roo", "system-prompt-"+string(args.Mode))
	// customPromptContent, err := safeReadFile(customPromptPath) // Implement safeReadFile
	// if err == nil && customPromptContent != "" {
	//    builder.WriteString(modeConfig.RoleDefinition)
	//    builder.WriteString("\n\n")
	//    builder.WriteString(customPromptContent)
	//    // Add custom instructions section only
	//    customInstructionsSection, err := sections.AddCustomInstructions(...)
	//    if err != nil { return "", err }
	//    builder.WriteString(customInstructionsSection)
	//    return builder.String(), nil
	// }
	// --- End Optional File-based ---

	// 2. Role Definition
	builder.WriteString(modeConfig.RoleDefinition)
	builder.WriteString("\n\n")

	// 3. Shared Tool Use Intro
	builder.WriteString(sections.GetSharedToolUseSection())
	builder.WriteString("\n\n")

	// 4. Tool Descriptions for Current Mode
	toolDescArgs := tools.ToolDescriptionGenArgs{
		Cwd:                 args.EnvCtx.Cwd,
		SupportsComputerUse: args.EnvCtx.SupportsComputerUse,
		DiffStrategy:        args.DiffStrategy,
		BrowserViewportSize: args.EnvCtx.BrowserViewportSize,
		// Pass experiments if needed
	}
	builder.WriteString(tools.GetToolDescriptionsForMode(args.Mode, toolDescArgs, args.CustomModeConfigs)) // Needs implementation
	builder.WriteString("\n\n")

	// 5. Tool Use Guidelines
	builder.WriteString(sections.GetToolUseGuidelinesSection())
	builder.WriteString("\n\n")

	// 6. Capabilities Section
	builder.WriteString(sections.GetCapabilitiesSection(args.EnvCtx.Cwd, args.EnvCtx.SupportsComputerUse, args.DiffStrategy)) // Add MCP hub arg if re-added
	builder.WriteString("\n\n")

	// 7. Modes Section (if needed, less relevant if mode switching is removed)
	// builder.WriteString(sections.GetModesSection(args.CustomModeConfigs)) // Needs implementation
	// builder.WriteString("\n\n")

	// 8. Rules Section
	builder.WriteString(sections.GetRulesSection(args.EnvCtx.Cwd, args.EnvCtx.SupportsComputerUse, args.DiffStrategy)) // Add experiments arg if needed
	builder.WriteString("\n\n")

	// 9. System Info Section
	builder.WriteString(sections.GetSystemInfoSection(args.EnvCtx.Cwd, args.Mode, args.CustomModeConfigs)) // Needs implementation
	builder.WriteString("\n\n")

	// 10. Objective Section
	builder.WriteString(sections.GetObjectiveSection())
	builder.WriteString("\n\n")

	// 11. Custom Instructions Section
	rooIgnoreInstructions := ""
	if args.RooIgnoreController != nil {
		rooIgnoreInstructions = args.RooIgnoreController.GetInstructions()
	}
	customInstructionsSection, err := sections.AddCustomInstructions(
		modeConfig.CustomInstructions, // Mode specific
		args.GlobalInstructions,       // Global
		args.EnvCtx.Cwd,
		string(args.Mode),
		rooIgnoreInstructions,
	)
	if err != nil {
		return "", fmt.Errorf("failed to build custom instructions: %w", err)
	}
	builder.WriteString(customInstructionsSection)

	return builder.String(), nil

}
