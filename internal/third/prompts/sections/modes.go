package sections

import (
	"mind-weaver/internal/third/toolgroups"

	"encoding/json"
	"fmt"
	"os"
)

// ModeSlug represents the unique identifier for a mode.
type ModeSlug string

// DefaultModeSlug is the slug for the standard code mode.
const DefaultModeSlug ModeSlug = "code"

// FileRestriction defines constraints on file access for a tool group.
type FileRestriction struct {
	FileRegex   string `json:"fileRegex"` // Go regex syntax
	Description string `json:"description"`
}

// GroupEntry can be a simple group name or a tuple with restrictions.
// We use an interface or a struct with optional fields to represent this.
type GroupEntry struct {
	Name        toolgroups.ToolGroupName `json:"name"`                  // The core group name
	Restriction *FileRestriction         `json:"restriction,omitempty"` // Optional restriction
}

// ModeConfig defines the structure for a single mode.
type ModeConfig struct {
	Slug               ModeSlug     `json:"slug"`
	Name               string       `json:"name"`
	RoleDefinition     string       `json:"roleDefinition"`
	Groups             []GroupEntry `json:"groups"` // Array of allowed tool groups/restrictions
	CustomInstructions string       `json:"customInstructions,omitempty"`
}

// BuiltinModes defines the standard modes available.
// This should be populated with the Go equivalent of the TS `modes` array.
var BuiltinModes = []ModeConfig{
	{
		Slug:           "code",
		Name:           "Code",
		RoleDefinition: "You are mind-weaver, an AI programming assistant...", // Truncated for brevity
		Groups: []GroupEntry{
			{Name: toolgroups.GroupRead},
			{Name: toolgroups.GroupEdit},
			{Name: toolgroups.GroupCommand},
			// Add other default groups
		},
		CustomInstructions: "",
	},
	{
		Slug:           "architect",
		Name:           "Architect",
		RoleDefinition: "You are mind-weaver, an AI software architect...", // Truncated
		Groups: []GroupEntry{
			{Name: toolgroups.GroupRead},
			// Example restriction: only allow editing Markdown
			{Name: toolgroups.GroupEdit, Restriction: &FileRestriction{FileRegex: `\.md$`, Description: "Markdown files only"}},
			{Name: toolgroups.GroupCommand},
		},
		CustomInstructions: "Focus on high-level design, documentation...",
	},
	// Add other built-in modes (ask, test, review)...
}

// LoadCustomModes reads custom mode configurations from a JSON file.
func LoadCustomModes(filePath string) ([]ModeConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ModeConfig{}, nil // No custom modes file is fine
		}
		return nil, fmt.Errorf("reading custom modes file %s: %w", filePath, err)
	}

	var config struct {
		CustomModes []ModeConfig `json:"customModes"`
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling custom modes file %s: %w", filePath, err)
	}
	return config.CustomModes, nil
}

// GetAllModes combines built-in modes with custom modes, overriding by slug.
func GetAllModes(customModeFilePaths ...string) ([]ModeConfig, error) {
	allModesMap := make(map[ModeSlug]ModeConfig)

	// Load built-in modes first
	for _, mode := range BuiltinModes {
		allModesMap[mode.Slug] = mode
	}

	// Load custom modes, overriding built-in ones
	for _, filePath := range customModeFilePaths {
		customModes, err := LoadCustomModes(filePath)
		if err != nil {
			// Decide how to handle errors - log and continue, or return error?
			fmt.Printf("Warning: Failed to load custom modes from %s: %v\n", filePath, err)
			continue // Continue with other files or built-ins
		}
		for _, mode := range customModes {
			if mode.Slug == "" {
				fmt.Printf("Warning: Skipping custom mode with empty slug in %s\n", filePath)
				continue
			}
			allModesMap[mode.Slug] = mode // Override existing entry
		}
	}

	// Convert map back to slice
	finalModes := make([]ModeConfig, 0, len(allModesMap))
	for _, mode := range allModesMap {
		finalModes = append(finalModes, mode)
	}

	// Optional: Sort the final list by name or slug for consistency
	// sort.Slice(finalModes, func(i, j int) bool { return finalModes[i].Slug < finalModes[j].Slug })

	return finalModes, nil
}

// GetModeBySlug finds a specific mode configuration by its slug from a list.
func GetModeBySlug(slug ModeSlug, allModes []ModeConfig) *ModeConfig {
	for _, mode := range allModes {
		if mode.Slug == slug {
			return &mode // Return a pointer to the found mode
		}
	}
	// Fallback to built-in if not found in the provided list (optional)
	for _, mode := range BuiltinModes {
		if mode.Slug == slug {
			return &mode
		}
	}
	return nil // Not found
}

// GetDefaultMode returns the configuration for the default mode.
func GetDefaultMode() *ModeConfig {
	// Assuming BuiltinModes always contains the default
	for _, mode := range BuiltinModes {
		if mode.Slug == DefaultModeSlug {
			return &mode
		}
	}
	// Should not happen if BuiltinModes is defined correctly
	panic("Default mode configuration not found")
}

// GetGroupName extracts the group name from a GroupEntry.
func GetGroupName(entry GroupEntry) toolgroups.ToolGroupName {
	return entry.Name
}

// TODO: Implement IsToolAllowedForMode based on ModeConfig.Groups and restrictions
// func IsToolAllowedForMode(tool toolgroups.ToolName, modeSlug ModeSlug, allModes []ModeConfig, experiments map[string]bool) bool { ... }
// func IsFileAllowedForGroup(filePath string, groupEntry GroupEntry) bool { ... } // Needs regex matching
