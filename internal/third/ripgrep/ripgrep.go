package ripgrep

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RooIgnoreController is an interface for path access validation
type RooIgnoreController interface {
	ValidateAccess(path string) bool
}

// RegexSearchFiles executes ripgrep to search for a regex pattern in files
func RegexSearchFiles(cwd string, searchPath string, regexPattern string, filePattern string, rooIgnore RooIgnoreController) (string, error) {
	// Prepare ripgrep command
	args := []string{
		"--line-number",    // Show line numbers
		"--no-heading",     // Don't group matches by file
		"--color", "never", // No color codes in output
		"--with-filename", // Show filenames
	}

	// Add file pattern if provided
	if filePattern != "" {
		args = append(args, "--glob", filePattern)
	}

	// Add the regex pattern and search path
	args = append(args, regexPattern, searchPath)

	// Determine ripgrep executable name based on OS
	rgCmd := "rg"
	if runtime.GOOS == "windows" {
		rgCmd = "rg.exe"
	}

	// Create command
	cmd := exec.Command(rgCmd, args...)
	cmd.Dir = cwd

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Check for ripgrep-specific errors
	if err != nil {
		// Exit code 1 means no matches found (not an error)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil
		}

		// Check stderr for regex parsing errors
		stderrOutput := stderr.String()
		if strings.Contains(stderrOutput, "regex parse error") {
			return "", fmt.Errorf("regex parse error: %s", stderrOutput)
		}

		return "", fmt.Errorf("ripgrep execution error: %v - %s", err, stderrOutput)
	}

	// Process results to ensure they're relative to cwd
	results := stdout.String()
	if results != "" {
		processedLines := []string{}
		for _, line := range strings.Split(results, "\n") {
			if line == "" {
				continue
			}

			// The typical ripgrep output format is: filename:line:content
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				filePath := parts[0]

				// Convert absolute paths to relative paths
				if filepath.IsAbs(filePath) {
					rel, err := filepath.Rel(cwd, filePath)
					if err == nil {
						parts[0] = rel
						line = strings.Join(parts, ":")
					}
				}
			}
			processedLines = append(processedLines, line)
		}
		results = strings.Join(processedLines, "\n")
	}

	return results, nil
}
