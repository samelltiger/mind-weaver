package prompts

import (
	"fmt"
	"mind-weaver/internal/third/ignore"           // Adjust
	"mind-weaver/internal/third/prompts/sections" // Adjust pathutils
	"path/filepath"

	// Go's path package
	"sort"
	"strings"
	// "github.com/sergi/go-diff/diffmatchpatch" // For diff generation if needed
)

const toolUseInstructionsReminder = `# Reminder: Instructions for Tool Use

Tool uses are formatted using XML-style tags. The tool name is enclosed in opening and closing tags, and each parameter is similarly enclosed within its own set of tags. Here's the structure:

<tool_name>
<parameter1_name>value1</parameter1_name>
<parameter2_name>value2</parameter2_name>
...
</tool_name>

For example:

<attempt_completion>
<result>
I have completed the task...
</result>
</attempt_completion>

Always adhere to this format for all tool uses to ensure proper parsing and execution.`

// FormatToolResult formats a successful tool execution result for the LLM.
func FormatToolResult(resultText string) string {
	// Simple wrapping for now, can add more structure if needed
	return fmt.Sprintf("<tool_result>\n%s\n</tool_result>", resultText)
}

// FormatToolError formats an error message for the LLM.
func FormatToolError(errorMessage string) string {
	return fmt.Sprintf("<tool_error>\n%s\n</tool_error>", errorMessage)
}

// FormatRooIgnoreError formats the specific error for ignored files.
func FormatRooIgnoreError(filePath string) string {
	// Using Posix path for consistency in the message
	posixPath := sections.ToPosix(filePath)
	return fmt.Sprintf("Access to %s is blocked by the .rooignore file settings. You must try to continue in the task without using this file, or ask the user to update the .rooignore file.", posixPath)
}

// FormatFilesList formats a list of files, optionally marking ignored ones.
func FormatFilesList(basePath string, files []string, didHitLimit bool, rooIgnore *ignore.RooIgnoreController, showIgnored bool) string {
	if len(files) == 0 && !didHitLimit {
		return "No files found."
	}

	relativeFiles := make([]string, 0, len(files))
	for _, file := range files {
		relPath, err := filepath.Rel(basePath, file)
		if err != nil {
			relPath = file // Use absolute if Rel fails
		}
		relPath = sections.ToPosix(relPath) // Convert to Posix for display
		if strings.HasSuffix(file, string(filepath.Separator)) || strings.HasSuffix(file, "/") {
			// Ensure trailing slash for directories if original had it
			if !strings.HasSuffix(relPath, "/") {
				relPath += "/"
			}
		}
		relativeFiles = append(relativeFiles, relPath)
	}

	// Sort like the TS version
	sort.SliceStable(relativeFiles, func(i, j int) bool {
		aParts := strings.Split(relativeFiles[i], "/")
		bParts := strings.Split(relativeFiles[j], "/")
		minLen := len(aParts)
		if len(bParts) < minLen {
			minLen = len(bParts)
		}
		for k := 0; k < minLen; k++ {
			if aParts[k] != bParts[k] {
				// Directory vs file check
				isADir := k+1 == len(aParts) && k+1 < len(bParts) && strings.HasSuffix(relativeFiles[i], "/")
				isBDir := k+1 == len(bParts) && k+1 < len(aParts) && strings.HasSuffix(relativeFiles[j], "/")
				if isADir && !isBDir {
					return true
				} // a is dir, b is file
				if !isADir && isBDir {
					return false
				} // a is file, b is dir
				// Alphabetical sort
				return strings.Compare(aParts[k], bParts[k]) < 0
			}
		}
		// Shorter path comes first
		return len(aParts) < len(bParts)
	})

	var finalLines []string
	if rooIgnore != nil {
		for _, relPath := range relativeFiles {
			// ValidateAccess needs absolute or relative-to-CWD path
			absPath := filepath.Join(basePath, relPath) // Reconstruct absolute path for check
			isIgnored := !rooIgnore.ValidateAccess(absPath)

			if isIgnored {
				if !showIgnored {
					continue
				}
				finalLines = append(finalLines, ignore.LOCK_TEXT_SYMBOL+" "+relPath)
			} else {
				finalLines = append(finalLines, relPath)
			}
		}
	} else {
		finalLines = relativeFiles // No ignore controller, show all
	}

	if len(finalLines) == 0 && !didHitLimit {
		return "No files found." // Possible if all files were ignored and showIgnored=false
	}

	result := strings.Join(finalLines, "\n")
	if didHitLimit {
		result += "\n\n(File list truncated. Use list_files on specific subdirectories if you need to explore further.)"
	}
	return result

}

// FormatMissingParamError

// FormatMissingParamError formats the error for a missing required tool parameter.
func FormatMissingParamError(toolName, paramName string) string {
	return fmt.Sprintf("Missing value for required parameter '%s' in tool '%s'. Please retry with complete response.\n\n%s", paramName, toolName, toolUseInstructionsReminder)
}

// FormatNoToolsUsedError formats the error when the AI response lacks a tool use.
func FormatNoToolsUsedError() string {
	// Note: The "Next Steps" part might be less relevant if the backend doesn't drive the completion/follow-up logic directly.
	// Adjust based on how your application handles task state.
	return fmt.Sprintf("[ERROR] You did not use a tool in your previous response! Please retry with a tool use.\n\n%s\n\n# Next Steps\nIf you have completed the user's task, use the attempt_completion tool. If you require additional information, use the ask_followup_question tool. Otherwise, proceed with the next step.", toolUseInstructionsReminder)
}

// FormatTooManyMistakes formats a message when the AI makes repeated errors.
func FormatTooManyMistakes(feedback string) string {
	fbSection := ""
	if strings.TrimSpace(feedback) != "" {
		fbSection = fmt.Sprintf("\nThe user has provided the following feedback to help guide you:\n<feedback>\n%s\n</feedback>", feedback)
	}
	return fmt.Sprintf("You seem to be having trouble proceeding.%s", fbSection)
}

// CreatePrettyPatch generates a human-readable diff patch.
// Requires a diff library. Returns empty string if no changes or on error.
func CreatePrettyPatch(filename, oldStr, newStr string) string {
	// --- Option 1: Using github.com/sergi/go-diff/diffmatchpatch ---
	// dmp := diffmatchpatch.New()
	// diffs := dmp.DiffMain(oldStr, newStr, true) // Check lines = true
	// patches := dmp.PatchMake(oldStr, diffs)
	// patchText := dmp.PatchToText(patches)
	// // Basic cleanup to resemble the TS output (remove header)
	// lines := strings.Split(patchText, "\n")
	// if len(lines) > 4 {
	//     // Heuristic: Remove the standard patch header lines
	//     // Adjust indices if the library's output format differs
	//     return strings.Join(lines[4:], "\n")
	// }
	// return patchText // Return raw patch if cleanup fails

	// --- Option 2: Placeholder if no diff library is used server-side ---
	fmt.Println("Warning: CreatePrettyPatch requires a diff library implementation.")
	if oldStr == newStr {
		return ""
	}
	// Simple indication of change if no library
	return fmt.Sprintf("--- a/%s\n+++ b/%s\n@@ ... @@\n- [Content differs, diff generation not implemented server-side]\n+ [Content differs, diff generation not implemented server-side]\n", filename, filename)
}

// FormatToolDenied formats the message when the user denies an operation.
func FormatToolDenied() string {
	return "The user denied this operation."
}

// FormatToolDeniedWithFeedback formats the denial message with user feedback.
func FormatToolDeniedWithFeedback(feedback string) string {
	fb := strings.TrimSpace(feedback)
	if fb == "" {
		return FormatToolDenied()
	}
	return fmt.Sprintf("The user denied this operation and provided the following feedback:\n<feedback>\n%s\n</feedback>", fb)
}

// FormatToolApprovedWithFeedback formats the approval message with user context.
func FormatToolApprovedWithFeedback(feedback string) string {
	fb := strings.TrimSpace(feedback)
	if fb == "" {
		// This case might not be used often, but handle it.
		return "The user approved this operation."
	}
	return fmt.Sprintf("The user approved this operation and provided the following context:\n<feedback>\n%s\n</feedback>", fb)
}
