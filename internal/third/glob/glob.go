package glob

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mind-weaver/internal/third/ignore" // Assuming this provides necessary interfaces/constructors
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	// Default timeout for recursive listing to prevent excessive runtime.
	defaultListTimeout = 10 * time.Second
)

// --- Default Ignore Patterns (similar to TS example, using gitignore syntax) ---
// These are applied *in addition* to .gitignore rules when recursive is true.
var defaultRecursiveIgnorePatterns = []string{
	// Common directories
	"node_modules/",
	"__pycache__/",
	"env/",
	"venv/",
	"dist/",
	"out/",
	"bundle/",
	"vendor/",
	"tmp/",
	"temp/",
	"deps/",
	"pkg/",
	"Pods/",

	// Specific nested build/dependency directories
	"target/dependency/",  // e.g., Maven/Gradle
	"build/dependencies/", // e.g., Custom build systems

	// Hidden files and directories (match anywhere)
	// Using '.*' is common but broad. '/.*' anchors to the root.
	// '**/.*/' is more explicit for hidden dirs anywhere.
	// Let's use '.*' for simplicity, matching common .gitignore usage.
	".*",
}

// IgnoreController defines the interface needed for checking paths.
// We define it here so glob doesn't strictly depend on the concrete type
// from the ignore package, allowing for wrappers like combinedIgnorer.
type IgnoreController interface {
	ValidateAccess(path string) bool
}

// combinedIgnorer wraps an optional original ignorer and a default ignorer.
// It ensures a path is allowed by *both* (if the original exists).
type combinedIgnorer struct {
	original IgnoreController // The ignorer passed by the user (e.g., from .gitignore)
	defaults IgnoreController // The ignorer with our hardcoded default patterns
}

// ValidateAccess checks against defaults first, then the original ignorer.
// Access is only granted if *neither* ignorer rejects the path.
func (c *combinedIgnorer) ValidateAccess(path string) bool {
	// If default rules ignore the path, access is denied.
	if !c.defaults.ValidateAccess(path) {
		return false
	}
	// If default rules allow it, check the original ignorer (if it exists).
	if c.original != nil && !c.original.ValidateAccess(path) {
		return false
	}
	// If both allow access (or original doesn't exist), access is granted.
	return true
}

// arePathsEqual checks if two paths are equivalent.
// (Implementation remains the same as before)
func arePathsEqual(path1, path2 string) (bool, error) {
	absPath1, err := filepath.Abs(path1)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for %s: %w", path1, err)
	}
	absPath2, err := filepath.Abs(path2)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for %s: %w", path2, err)
	}

	cleanPath1 := filepath.Clean(absPath1)
	cleanPath2 := filepath.Clean(absPath2)

	if runtime.GOOS == "windows" {
		return strings.EqualFold(cleanPath1, cleanPath2), nil
	}
	return cleanPath1 == cleanPath2, nil
}

// ListFiles lists files and directories, potentially recursively, respecting ignore rules.
// Includes default ignores for recursive scans.
func ListFiles(dirPath string, recursive bool, limit int, ignorer *ignore.RooIgnoreController) ([]string, bool, error) {
	absolutePath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get absolute path for %s: %w", dirPath, err)
	}

	// --- Safety Checks (remain the same) ---
	var root string
	if runtime.GOOS == "windows" {
		vol := filepath.VolumeName(absolutePath)
		if vol != "" {
			root = vol + string(filepath.Separator)
		} else {
			root = string(filepath.Separator)
		}
	} else {
		root = "/"
	}

	isRoot, err := arePathsEqual(absolutePath, root)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check if path is root: %w", err)
	}
	if isRoot {
		return []string{absolutePath}, false, nil
	}

	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		isHome, err := arePathsEqual(absolutePath, homeDir)
		if err != nil {
			fmt.Printf("Warning: failed to check if path is home directory: %v\n", err)
		} else if isHome {
			return []string{absolutePath}, false, nil
		}
	} else if err != nil {
		fmt.Printf("Warning: failed to get user home directory: %v\n", err)
	}
	// --- End Safety Checks ---

	// --- Prepare the Effective Ignorer ---
	var effectiveIgnorer IgnoreController
	if ignorer != nil {
		effectiveIgnorer = ignorer // Start with the passed ignorer
	}

	if recursive {
		// Create an ignorer for the default patterns
		defaultIgnorer := ignore.NewRooIgnoreController(dirPath)
		// Add default patterns
		err := defaultIgnorer.AddPatterns(".", defaultRecursiveIgnorePatterns)
		if err != nil {
			fmt.Printf("Warning: failed to add default ignore patterns: %v\n", err)
		}

		// Combine the original ignorer with the default ignorer
		effectiveIgnorer = &combinedIgnorer{
			original: ignorer,
			defaults: defaultIgnorer,
		}
	}
	// --- End Prepare Ignorer ---

	// --- Call the appropriate listing function with the effective ignorer ---
	if !recursive {
		// Non-recursive uses the original ignorer (or nil if none passed)
		return listFilesNonRecursive(absolutePath, limit, ignorer) // Pass original ignorer
	} else {
		// Recursive uses the combined ignorer
		return listFilesRecursiveBFS(absolutePath, limit, effectiveIgnorer) // Pass combined ignorer
	}
}

// listFilesNonRecursive lists files and directories in a single directory level.
// Uses the passed IgnoreController.
func listFilesNonRecursive(dirPath string, limit int, ignorer IgnoreController) ([]string, bool, error) {
	// (Implementation remains largely the same, but uses the IgnoreController interface type)
	var files []string
	hitLimit := false
	count := 0

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Printf("Warning: permission denied accessing path %s: %v\n", dirPath, err)
			// Check if the dir itself is ignored before returning it
			if ignorer == nil || ignorer.ValidateAccess(dirPath) {
				return []string{dirPath}, false, nil
			}
			return []string{}, false, nil // Ignored directory
		}
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, fmt.Errorf("directory not found: %s", dirPath)
		}
		_, errNotDir := err.(*fs.PathError)
		if errNotDir && strings.Contains(err.Error(), "not a directory") {
			if info, statErr := os.Stat(dirPath); statErr == nil && !info.IsDir() {
				if ignorer == nil || ignorer.ValidateAccess(dirPath) {
					return []string{dirPath}, false, nil
				} else {
					return []string{}, false, nil // Ignored file
				}
			}
			return nil, false, fmt.Errorf("path is not a directory: %s: %w", dirPath, err)
		}
		return nil, false, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if limit > 0 && count >= limit {
			hitLimit = true
			break
		}

		entryPath := filepath.Join(dirPath, entry.Name())

		// Use the effective ignorer
		if ignorer != nil && !ignorer.ValidateAccess(entryPath) {
			continue
		}

		files = append(files, entryPath)
		count++
	}

	return files, hitLimit, nil
}

// listFilesRecursiveBFS lists files and directories recursively using Breadth-First Search.
// Uses the passed IgnoreController.
func listFilesRecursiveBFS(startPath string, limit int, ignorer IgnoreController) ([]string, bool, error) {
	// (Implementation remains largely the same, but uses the IgnoreController interface type)
	var results []string
	hitLimit := false
	count := 0

	queue := make([]string, 0)
	visited := make(map[string]struct{})

	// Add start path *only* if not ignored by the effective controller
	if ignorer == nil || ignorer.ValidateAccess(startPath) {
		if limit > 0 && count < limit {
			results = append(results, startPath)
			count++
			queue = append(queue, startPath) // Only queue if not ignored and within limit
			visited[startPath] = struct{}{}
		} else if limit > 0 {
			hitLimit = true
		} else { // limit <= 0 means no limit
			results = append(results, startPath)
			count++
			queue = append(queue, startPath)
			visited[startPath] = struct{}{}
		}
	} else {
		// Start directory is ignored, return empty
		return []string{}, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultListTimeout)
	defer cancel()

	processedDirs := 0

	for len(queue) > 0 && !hitLimit {
		levelSize := len(queue)
		processedDirs = 0

		for i := 0; i < levelSize && !hitLimit; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("Warning: Globbing timed out, returning partial results.")
				sort.Strings(results)
				// Use context.Cause(ctx) for potentially more specific error in Go 1.20+
				cause := context.Cause(ctx)
				if cause == nil {
					cause = ctx.Err()
				} // Fallback for older Go versions
				return results, true, fmt.Errorf("glob operation timed out or was canceled: %w", cause)
			default:
			}

			currentPath := queue[i]

			entries, err := os.ReadDir(currentPath)
			if err != nil {
				fmt.Printf("Warning: could not read directory %s: %v\n", currentPath, err)
				continue
			}
			processedDirs++

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Name() < entries[j].Name()
			})

			for _, entry := range entries {
				if limit > 0 && count >= limit {
					hitLimit = true
					break
				}

				entryPath := filepath.Join(currentPath, entry.Name())

				// Use the effective ignorer
				if ignorer != nil && !ignorer.ValidateAccess(entryPath) {
					continue
				}

				results = append(results, entryPath)
				count++

				if entry.IsDir() {
					if _, found := visited[entryPath]; !found {
						queue = append(queue, entryPath)
						visited[entryPath] = struct{}{}
					}
				}
			}
		}

		if levelSize > 0 {
			queue = queue[levelSize:]
		}
		if hitLimit {
			break
		}
	}

	sort.Strings(results)
	finalHitLimit := hitLimit || (limit > 0 && count >= limit)
	finalErr := context.Cause(ctx)
	if finalErr != nil && finalErr != context.Canceled && finalErr != context.DeadlineExceeded {
		// Keep underlying error if it's not just timeout/cancel
	} else if finalErr != nil {
		finalHitLimit = true // Ensure limit is true on timeout
		finalErr = fmt.Errorf("glob operation timed out or was canceled: %w", finalErr)
	} else {
		finalErr = nil // No error if context was ok
	}

	return results, finalHitLimit, finalErr
}
