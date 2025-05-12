package sections

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// safeReadFile reads a file, returns empty string on ENOENT/EISDIR, propagates other errors.
func safeReadFile(filePath string) (string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) || err.Error() == "read "+filePath+": is a directory" { // Crude EISDIR check
			return "", nil // Not found or is directory is okay, return empty
		}
		return "", err // Propagate other errors
	}
	return strings.TrimSpace(string(contentBytes)), nil
}

// directoryExists checks if a path exists and is a directory.
func directoryExists(dirPath string) (bool, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// readTextFilesFromDirectory reads all files recursively (basic implementation).
// Returns map[filename]content. Needs error handling refinement.
func readTextFilesFromDirectory(dirPath string) (map[string]string, error) {
	files := make(map[string]string)
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Stop walking on error
		}
		if !d.IsDir() {
			// Check if it's a symlink and resolve if necessary (more complex)
			// For simplicity, just read if it's reported as a file for now
			content, readErr := safeReadFile(path)
			if readErr != nil {
				fmt.Printf("Warning: could not read rule file %s: %v\n", path, readErr) // Log error but continue
				return nil
			}
			if content != "" {
				// Use relative path from the base rules dir for the key if possible
				relPath, relErr := filepath.Rel(dirPath, path)
				if relErr == nil {
					files[relPath] = content
				} else {
					files[path] = content // Fallback to absolute path
				}
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) { // Don't fail if the dir itself doesn't exist
		return nil, err
	}
	return files, nil
}

// formatDirectoryContent formats content from multiple files.
func formatDirectoryContent(dirPath string, files map[string]string) string {
	if len(files) == 0 {
		return ""
	}
	var builder strings.Builder
	// Sort keys for consistent output
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		// Use the original key which might be relative or absolute
		fullPath := filepath.Join(dirPath, k) // Reconstruct full path for display if key is relative
		builder.WriteString(fmt.Sprintf("\n\n# Rules from %s:\n%s", fullPath, files[k]))
	}
	return builder.String()
}

// loadRuleFiles loads rules from .roo/rules/ or fallback files.
func loadRuleFiles(cwd string) (string, string, error) { // Returns content, source, error
	rooRulesDir := filepath.Join(cwd, ".roo", "rules")
	exists, err := directoryExists(rooRulesDir)
	if err != nil {
		return "", "", fmt.Errorf("checking rules dir %s: %w", rooRulesDir, err)
	}

	if exists {
		files, err := readTextFilesFromDirectory(rooRulesDir)
		if err != nil {
			return "", "", fmt.Errorf("reading rules dir %s: %w", rooRulesDir, err)
		}
		if len(files) > 0 {
			return formatDirectoryContent(rooRulesDir, files), rooRulesDir, nil
		}
	}

	// Fallback
	ruleFiles := []string{".roorules", ".clinerules"}
	for _, file := range ruleFiles {
		filePath := filepath.Join(cwd, file)
		content, err := safeReadFile(filePath)
		if err != nil {
			return "", "", fmt.Errorf("reading rule file %s: %w", filePath, err) // Propagate unexpected errors
		}
		if content != "" {
			return fmt.Sprintf("\n# Rules from %s:\n%s\n", file, content), file, nil
		}
	}
	return "", "", nil // No rules found
}

// AddCustomInstructions builds the custom instructions section.
func AddCustomInstructions(modeInstructions, globalInstructions, cwd, mode, rooIgnoreInstructions string) (string, error) {
	var sections []string

	// Mode-specific rules
	modeRuleContent := ""
	usedRuleFile := ""
	// var loadErr error

	if mode != "" {
		modeRulesDir := filepath.Join(cwd, ".roo", fmt.Sprintf("rules-%s", mode))
		exists, err := directoryExists(modeRulesDir)
		if err != nil {
			return "", fmt.Errorf("checking mode rules dir %s: %w", modeRulesDir, err)
		}
		if exists {
			files, err := readTextFilesFromDirectory(modeRulesDir)
			if err != nil {
				return "", fmt.Errorf("reading mode rules dir %s: %w", modeRulesDir, err)
			}
			if len(files) > 0 {
				modeRuleContent = formatDirectoryContent(modeRulesDir, files)
				usedRuleFile = modeRulesDir // Source is the directory itself
			}
		}

		if modeRuleContent == "" { // Fallback if directory empty or doesn't exist
			modeRuleFiles := []string{fmt.Sprintf(".roorules-%s", mode), fmt.Sprintf(".clinerules-%s", mode)}
			for _, file := range modeRuleFiles {
				filePath := filepath.Join(cwd, file)
				content, err := safeReadFile(filePath)
				if err != nil {
					return "", fmt.Errorf("reading mode rule file %s: %w", filePath, err)
				}
				if content != "" {
					modeRuleContent = content
					usedRuleFile = file
					break
				}
			}
		}
	}

	// Language Preference
	langName := "中文" // Assuming FormatLanguage exists
	sections = append(sections, fmt.Sprintf(`Language Preference:

You should always speak and think in the "%s" (%s) language unless the user gives you instructions below to do otherwise.`, langName, "zh-cn"))

	// Global Instructions
	trimmedGlobal := strings.TrimSpace(globalInstructions)
	if trimmedGlobal != "" {
		sections = append(sections, fmt.Sprintf("Global Instructions:\n%s", trimmedGlobal))
	}

	// Mode-specific Instructions
	trimmedMode := strings.TrimSpace(modeInstructions)
	if trimmedMode != "" {
		sections = append(sections, fmt.Sprintf("Mode-specific Instructions:\n%s", trimmedMode))
	}

	// Rules
	var rules []string
	if modeRuleContent != "" {
		if usedRuleFile == filepath.Join(cwd, ".roo", fmt.Sprintf("rules-%s", mode)) || strings.HasPrefix(usedRuleFile, filepath.Join(cwd, ".roo")) {
			// Content already formatted with headers if from directory
			rules = append(rules, strings.TrimSpace(modeRuleContent))
		} else {
			rules = append(rules, fmt.Sprintf("# Rules from %s:\n%s", usedRuleFile, modeRuleContent))
		}
	}

	if rooIgnoreInstructions != "" {
		rules = append(rules, strings.TrimSpace(rooIgnoreInstructions))
	}

	genericRuleContent, genericRuleSource, err := loadRuleFiles(cwd)
	if err != nil {
		return "", err // Propagate error from loading generic rules
	}
	if genericRuleContent != "" && genericRuleSource != usedRuleFile { // Avoid duplicating if mode rule *was* the generic rule
		rules = append(rules, strings.TrimSpace(genericRuleContent))
	}

	if len(rules) > 0 {
		sections = append(sections, fmt.Sprintf("Rules:\n\n%s", strings.Join(rules, "\n\n")))
	}

	joinedSections := strings.Join(sections, "\n\n")
	if joinedSections == "" {
		return "", nil
	}

	return fmt.Sprintf(`

====

USER'S CUSTOM INSTRUCTIONS

The following additional instructions are provided by the user, and should be followed to the best of your ability without interfering with the TOOL USE guidelines.

%s`, joinedSections), nil
}
