package diff

import "sort"

type InsertGroup struct {
	Index    int      // 0-based index where insertion should happen
	Elements []string // Lines to insert
}

// InsertGroups inserts multiple groups of elements at specified indices.
// Assumes indices are 0-based.
func InsertGroups(original []string, insertGroups []InsertGroup) []string {
	if len(insertGroups) == 0 {
		return original
	}

	// Sort groups by index to process them in order
	sort.Slice(insertGroups, func(i, j int) bool {
		return insertGroups[i].Index < insertGroups[j].Index
	})

	// Estimate capacity to reduce reallocations
	newSize := len(original)
	for _, group := range insertGroups {
		newSize += len(group.Elements)
	}
	result := make([]string, 0, newSize)
	lastIndex := 0

	for _, group := range insertGroups {
		// Ensure index is within bounds
		insertIndex := group.Index
		if insertIndex < 0 {
			insertIndex = 0
		}
		// Clamp insertIndex if it's beyond the *current* end of the original slice part being processed
		if insertIndex > len(original) {
			insertIndex = len(original)
		}

		// Add elements from original array up to the insertion point
		if insertIndex > lastIndex {
			result = append(result, original[lastIndex:insertIndex]...)
		}

		// Add the group of elements to insert
		result = append(result, group.Elements...)

		// Update lastIndex to the current insertion point in the *original* array
		lastIndex = insertIndex
	}

	// Add remaining elements from original array
	if lastIndex < len(original) {
		result = append(result, original[lastIndex:]...)
	}

	return result
}
