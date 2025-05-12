package utils

import (
	"path/filepath"
	"strings"
)

// ToPosix converts a path to use forward slashes, typical for display.
func ToPosix(path string) string {
	return strings.ReplaceAll(path, string(filepath.Separator), "/")
}

// GetIndent returns the leading whitespace (spaces/tabs) of a string.
func GetIndent(line string) string {
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return line[:i]
		}
	}
	// If the line is all whitespace or empty, return the whole line
	return line
}

// GetReadablePath returns a path relative to the CWD if possible, otherwise absolute.
// Ensures Posix-style separators for display.
func GetReadablePath(cwd, targetPath string) string {
	if targetPath == "" {
		return "." // Or handle empty path as needed
	}
	// Ensure CWD is absolute for reliable Rel calculation
	absCwd := cwd
	if !filepath.IsAbs(cwd) {
		var err error
		absCwd, err = filepath.Abs(cwd)
		if err != nil {
			absCwd = cwd // Fallback if Abs fails
		}
	}

	absPath := targetPath
	if !filepath.IsAbs(targetPath) {
		absPath = filepath.Join(absCwd, targetPath)
	}
	absPath = filepath.Clean(absPath)

	relPath, err := filepath.Rel(absCwd, absPath)
	// Prefer relative if it's within cwd subtree and doesn't start with '..'
	// Also handle the case where relPath is "."
	if err == nil && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) && relPath != ".." {
		return ToPosix(relPath)
	}

	// Otherwise, return the cleaned absolute path
	return ToPosix(absPath)
}

// IsPathOutsideWorkspace checks if a target path is outside the CWD.
// It compares the cleaned absolute paths.
func IsPathOutsideWorkspace(cwd, targetPath string) bool {
	if targetPath == "" {
		return false // Empty path is considered within
	}

	// Ensure CWD is absolute and clean
	absCwd := cwd
	if !filepath.IsAbs(cwd) {
		var err error
		absCwd, err = filepath.Abs(cwd)
		if err != nil {
			// If we can't get absolute CWD, assume target might be outside for safety?
			// Or handle error appropriately. Let's assume outside if CWD is problematic.
			return true
		}
	}
	absCwd = filepath.Clean(absCwd)

	absTarget := targetPath
	if !filepath.IsAbs(targetPath) {
		absTarget = filepath.Join(absCwd, targetPath)
	}
	absTarget = filepath.Clean(absTarget)

	// Check if the absolute target path starts with the absolute CWD path.
	// Add separator to avoid matching /path/to/workspac vs /path/to/workspace-extra
	prefix := absCwd + string(filepath.Separator)
	return !strings.HasPrefix(absTarget, prefix) && absTarget != absCwd
}
