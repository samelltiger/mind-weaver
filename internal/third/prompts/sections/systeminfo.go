package sections

import (
	"os"
	"runtime"
	"strings"
)

// Note: You'll need to implement or import these helper functions
// toPosix - converts path to POSIX style (forward slashes)
// getShell - gets the default shell
// osName - gets the OS name

func GetSystemInfoSection(cwd string, currentMode ModeSlug, customModes []ModeConfig) string {
	findModeBySlug := func(slug ModeSlug, modes []ModeConfig) string {
		for _, m := range modes {
			if m.Slug == slug {
				return m.Name
			}
		}
		return ""
	}

	currentModeName := findModeBySlug(currentMode, customModes)
	if currentModeName == "" {
		currentModeName = string(currentMode)
	}

	codeModeName := findModeBySlug(defaultModeSlug, customModes)
	if codeModeName == "" {
		codeModeName = "Code"
	}

	homeDir, _ := os.UserHomeDir()
	details := `====

SYSTEM INFORMATION

Operating System: ` + osName() + `
Default Shell: ` + getShell() + `
Home Directory: ` + ToPosix(homeDir) + `
Current Workspace Directory: ` + ToPosix(cwd) + `

The Current Workspace Directory is the active VS Code project directory, and is therefore the default directory for all tool operations. New terminals will be created in the current workspace directory, however if you change directories in a terminal it will then have a different working directory; changing directories in a terminal does not modify the workspace directory, because you do not have access to change the workspace directory. When the user initially gives you a task, a recursive list of all filepaths in the current workspace directory ('/test/path') will be included in environment_details. This provides an overview of the project's file structure, offering key insights into the project from directory/file names (how developers conceptualize and organize their code) and file extensions (the language used). This can also guide decision-making on which files to explore further. If you need to further explore directories such as outside the current workspace directory, you can use the list_files tool. If you pass 'true' for the recursive parameter, it will list files recursively. Otherwise, it will list files at the top level, which is better suited for generic directories where you don't necessarily need the nested structure, like the Desktop.`

	return details
}

// Helper function to convert path to POSIX style
func ToPosix(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// You'll need to implement these functions or import them from other packages
func osName() string {
	// Implement OS name detection or use a third-party package
	return runtime.GOOS // This just returns "linux", "windows", etc.
}

func getShell() string {
	// Implement shell detection logic
	return os.Getenv("SHELL") // Unix systems
}

const defaultModeSlug = "code" // or whatever your default is
