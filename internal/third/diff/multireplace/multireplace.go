package multireplace

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/adrg/strutil/metrics"
)

// --- multireplace/multireplace.go (ApplyDiff continued) ---

var diffBlockRegex = regexp.MustCompile(
	`(?sm)(?:^|\n)<<<<<<< SEARCH\s*\n` + // Start marker
		`(:start_line:\s*(\d+)\s*\n)?` + // Optional start_line (group 2 = number)
		`(:end_line:\s*(\d+)\s*\n)?` + // Optional end_line (group 4 = number)
		`(-------\s*\n)?` + // Optional metadata separator (group 5)
		`([\s\S]*?)(?:\n)?` + // Search content (group 6)
		`(?:.*?\n=======\s*\n)` + // Separator
		`([\s\S]*?)(?:\n)?` + // Replace content (group 7)
		`(?:.*?\n>>>>>>> REPLACE)(?:\n|$)`, // End marker
)

// var diffBlockRegex = regexp.MustCompile(
// 	`(?s)(?:^|\n)<<<<<<< SEARCH\s*\n` + // Start marker
// 		`(:start_line:\s*(\d+)\s*\n)?` + // Optional start_line (group 2 = number)
// 		`(:end_line:\s*(\d+)\s*\n)?` + // Optional end_line (group 4 = number)
// 		`(-------\s*\n)?` + // Optional metadata separator (group 5)
// 		`(.*?)(?:\n)?` + // Search content (group 6)
// 		`(?:.*?\n=======\s*\n)` + // Separator
// 		`(.*?)(?:\n)?` + // Replace content (group 7)
// 		`(?:.*?\n>>>>>>> REPLACE)(?:\n|$)`, // End marker
// )

const defaultBufferLines = 40

type MultiSearchReplaceDiffStrategy struct {
	fuzzyThreshold float64
	bufferLines    int
	levenshtein    *metrics.Levenshtein // Pre-configure metric
}

func NewMultiSearchReplaceDiffStrategy(fuzzyThreshold *float64, bufferLines *int) *MultiSearchReplaceDiffStrategy {
	threshold := 1.0
	if fuzzyThreshold != nil {
		threshold = *fuzzyThreshold
	}
	bufLines := defaultBufferLines
	if bufferLines != nil {
		bufLines = *bufferLines
	}
	return &MultiSearchReplaceDiffStrategy{
		fuzzyThreshold: threshold,
		bufferLines:    bufLines,
		levenshtein:    metrics.NewLevenshtein(), // Can customize options if needed
	}
}

func (s *MultiSearchReplaceDiffStrategy) GetName() string {
	return "MultiSearchReplace"
}

func (s *MultiSearchReplaceDiffStrategy) GetToolDescription(args ToolDescriptionArgs) string {
	return fmt.Sprintf(`## apply_diff
Description: Request to replace existing code using a search and replace block.
This tool allows for precise, surgical replaces to files by specifying exactly what content to search for and what to replace it with.
The tool will maintain proper indentation and formatting while making changes.
Only a single operation is allowed per tool use.
The SEARCH section must exactly match existing content including whitespace and indentation.
If you're not confident in the exact content to search for, use the read_file tool first to get the exact content.
When applying the diffs, be extra careful to remember to change any closing brackets or other syntax that may be affected by the diff farther down in the file.
ALWAYS make as many changes in a single 'apply_diff' request as possible using multiple SEARCH/REPLACE blocks

Parameters:
- path: (required) The path of the file to modify (relative to the current workspace directory %s)
- diff: (required) The search/replace block defining the changes.

Diff format:
`+"```"+`
<<<<<<< SEARCH
:start_line: (required) The line number of original content where the search block starts.
:end_line: (required) The line number of original content  where the search block ends.
-------
[exact content to find including whitespace]
=======
[new content to replace with]
>>>>>>> REPLACE

`+"```"+`


Example:

Original file:
`+"```"+`
1 | def calculate_total(items):
2 |     total = 0
3 |     for item in items:
4 |         total += item
5 |     return total
`+"```"+`

Search/Replace content:
`+"```"+`
<<<<<<< SEARCH
:start_line:1
:end_line:5
-------
def calculate_total(items):
    total = 0
    for item in items:
        total += item
    return total
=======
def calculate_total(items):
    """Calculate total with 10%% markup"""
    return sum(item * 1.1 for item in items)
>>>>>>> REPLACE

`+"```"+`

Search/Replace content with multi edits:
`+"```"+`
<<<<<<< SEARCH
:start_line:1
:end_line:2
-------
def calculate_total(items):
    sum = 0
=======
def calculate_sum(items):
    sum = 0
>>>>>>> REPLACE

<<<<<<< SEARCH
:start_line:4
:end_line:5
-------
        total += item
    return total
=======
        sum += item
    return sum 
>>>>>>> REPLACE
`+"```"+`

Usage:
<apply_diff>
<path>File path here</path>
<diff>
Your search/replace content here
You can use multi search/replace block in one diff block, but make sure to include the line numbers for each block.
Only use a single line of '=======' between search and replacement content, because multiple '=======' will corrupt the file.
</diff>
</apply_diff>`, args.Cwd)
}

// Helper to safely parse int, defaulting to 0 on error
func safeAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func (s *MultiSearchReplaceDiffStrategy) ApplyDiff(originalContent, diffContent string, paramStartLine, paramEndLine int) (*DiffResult, error) {
	// 1. Validate Marker Sequencing (translate validateMarkerSequencing)
	validSeq, errStr := s.validateMarkerSequencing(diffContent) // Assuming validateMarkerSequencing exists
	if !validSeq {
		return &DiffResult{Success: false, Error: errStr}, nil
	}

	// 2. Parse diff blocks using regex
	matches := diffBlockRegex.FindAllStringSubmatch(diffContent, -1)
	if len(matches) == 0 {
		return &DiffResult{Success: false, Error: "Invalid diff format - no valid SEARCH/REPLACE blocks found"}, nil
	}

	lineEnding := "\n"
	if strings.Contains(originalContent, "\r\n") {
		lineEnding = "\r\n"
	}
	resultLines := strings.Split(originalContent, lineEnding) // Use detected line ending
	delta := 0                                                // Line number offset due to previous replacements
	appliedCount := 0
	var failedParts []*DiffResult

	// Prepare replacements (parse start/end lines, sort)
	type replacement struct {
		OriginalIndex  int // Keep track of original match order if needed
		StartLine      int
		EndLine        int
		SearchContent  string
		ReplaceContent string
	}
	replacements := make([]replacement, 0, len(matches))

	for i, match := range matches {
		startLine := safeAtoi(match[2])
		endLine := safeAtoi(match[4])
		if endLine == 0 { // If end_line wasn't present or valid, default to length (adjust later)
			endLine = math.MaxInt32 // Use a large number initially
		}
		replacements = append(replacements, replacement{
			OriginalIndex:  i,
			StartLine:      startLine,
			EndLine:        endLine,
			SearchContent:  match[6],
			ReplaceContent: match[7],
		})
	}

	// Sort by start line to apply changes sequentially
	sort.Slice(replacements, func(i, j int) bool {
		if replacements[i].StartLine != replacements[j].StartLine {
			return replacements[i].StartLine < replacements[j].StartLine
		}
		// Maintain original order for blocks starting on the same line
		return replacements[i].OriginalIndex < replacements[j].OriginalIndex
	})

	// 3. Iterate and Apply/Fail each block
	for _, rep := range replacements {
		currentStartLine := rep.StartLine
		currentEndLine := rep.EndLine

		// Adjust line numbers based on previous modifications
		if currentStartLine != 0 { // Don't adjust if start line wasn't specified
			currentStartLine += delta
		}
		if currentEndLine != math.MaxInt32 { // Don't adjust the 'max' placeholder
			currentEndLine += delta
		} else {
			currentEndLine = len(resultLines) // Now set to actual end if not specified
		}

		searchContent := unescapeMarkers(rep.SearchContent)
		replaceContent := unescapeMarkers(rep.ReplaceContent)

		// Strip line numbers
		if (everyLineHasLineNumbers(searchContent) && everyLineHasLineNumbers(replaceContent)) ||
			(everyLineHasLineNumbers(searchContent) && strings.TrimSpace(replaceContent) == "") {
			searchContent = stripLineNumbers(searchContent)
			replaceContent = stripLineNumbers(replaceContent)
		}

		// Validate: Empty search requires start line
		if searchContent == "" && currentStartLine == 0 {
			failedParts = append(failedParts, &DiffResult{
				Success: false,
				Error:   "Empty search content requires start_line to be specified",
			})
			continue
		}
		// Validate: Empty search requires start == end
		if searchContent == "" && currentStartLine != 0 && currentEndLine != 0 && currentStartLine != currentEndLine {
			failedParts = append(failedParts, &DiffResult{
				Success: false,
				Error:   fmt.Sprintf("Empty search content requires start_line and end_line to be the same (got %d-%d)", currentStartLine, currentEndLine),
			})
			continue
		}

		searchLines := strings.Split(searchContent, "\n")
		if searchContent == "" {
			searchLines = []string{}
		}
		replaceLines := strings.Split(replaceContent, "\n")
		if replaceContent == "" {
			replaceLines = []string{}
		}

		matchIndex := -1
		bestMatchScore := 0.0
		var bestMatchContent string
		searchChunk := strings.Join(searchLines, "\n")

		// Determine search bounds based on provided lines and buffer
		searchStartIndex := 0
		searchEndIndex := len(resultLines)

		hasLineHint := currentStartLine > 0 // Use provided start line as a hint

		if hasLineHint {
			// Convert 1-based hint to 0-based index
			hintStartIndex := currentStartLine - 1
			hintEndIndex := currentEndLine - 1 // End line is inclusive in spec, so index is also inclusive

			// Clamp hint indices to valid range
			if hintStartIndex < 0 {
				hintStartIndex = 0
			}
			if hintEndIndex >= len(resultLines) {
				hintEndIndex = len(resultLines) - 1
			}
			if hintStartIndex > hintEndIndex && len(searchLines) > 0 { // Invalid range if start > end, unless it's an insertion
				failedParts = append(failedParts, &DiffResult{
					Success: false,
					Error:   fmt.Sprintf("Invalid line range %d-%d (file has %d lines)", currentStartLine, currentEndLine, len(resultLines)),
				})
				continue
			}

			// Try exact range first if search content is not empty
			if len(searchLines) > 0 && hintStartIndex <= hintEndIndex {
				if hintEndIndex+1-hintStartIndex == len(searchLines) { // Check if range length matches search length
					originalChunk := strings.Join(resultLines[hintStartIndex:hintEndIndex+1], "\n")
					similarity := getSimilarity(originalChunk, searchChunk, s.levenshtein)
					if similarity >= s.fuzzyThreshold {
						matchIndex = hintStartIndex
						bestMatchScore = similarity
						bestMatchContent = originalChunk
					}
				}
			} else if len(searchLines) == 0 { // Handle insertion case
				if hintStartIndex >= 0 && hintStartIndex <= len(resultLines) { // Valid insertion point
					matchIndex = hintStartIndex // Mark the insertion point
					bestMatchScore = 1.0        // Insertion always "matches" the location
				} else {
					failedParts = append(failedParts, &DiffResult{
						Success: false,
						Error:   fmt.Sprintf("Invalid insertion line %d (file has %d lines)", currentStartLine, len(resultLines)),
					})
					continue
				}
			}

			// If exact range didn't match (or wasn't applicable), define buffered search area
			if matchIndex == -1 {
				searchStartIndex = int(math.Max(0, float64(hintStartIndex-s.bufferLines)))
				searchEndIndex = int(math.Min(float64(len(resultLines)), float64(hintEndIndex+1+s.bufferLines))) // +1 because slice end is exclusive
			}
		}

		// Perform search within bounds if no match yet or if it's an insertion
		if matchIndex == -1 && len(searchLines) > 0 {
			// Simple linear scan within the buffer for now (can optimize with middle-out later if needed)
			for i := searchStartIndex; i <= searchEndIndex-len(searchLines); i++ {
				originalChunk := strings.Join(resultLines[i:i+len(searchLines)], "\n")
				similarity := getSimilarity(originalChunk, searchChunk, s.levenshtein)

				if similarity > bestMatchScore {
					bestMatchScore = similarity
					matchIndex = i
					bestMatchContent = originalChunk
				}
				// Optimization: if perfect match found, stop searching
				if bestMatchScore == 1.0 {
					break
				}
			}
		}

		// Check if match meets threshold (or if it's a valid insertion)
		if matchIndex == -1 || (len(searchLines) > 0 && bestMatchScore < s.fuzzyThreshold) {
			// Format error message (similar to TS version)
			lineRangeStr := ""
			if hasLineHint {
				lineRangeStr = fmt.Sprintf(" near lines %d-%d", currentStartLine, currentEndLine)
			}
			errMsg := fmt.Sprintf("No sufficiently similar match found%s (%.0f%% similar, needs %.0f%%)",
				lineRangeStr, bestMatchScore*100, s.fuzzyThreshold*100)
			failedParts = append(failedParts, &DiffResult{
				Success: false,
				Error:   errMsg,
				Details: &Details{
					Similarity:    bestMatchScore,
					Threshold:     s.fuzzyThreshold,
					SearchContent: searchChunk,
					BestMatch:     bestMatchContent,
				},
			})
			continue // Skip applying this block
		}

		// --- Apply the replacement ---
		var matchedLines []string
		if len(searchLines) > 0 {
			matchedLines = resultLines[matchIndex : matchIndex+len(searchLines)]
		} else {
			matchedLines = []string{} // For insertions
		}

		// Preserve indentation logic (simplified example)
		var baseIndent string
		if len(matchedLines) > 0 {
			// Use indent of the first matched line as the base for the replacement block
			baseIndent = GetIndent(matchedLines[0])
		} else if matchIndex > 0 && matchIndex <= len(resultLines) {
			// For insertion, try to use indent of the line *before* the insertion point
			baseIndent = GetIndent(resultLines[matchIndex-1])
		} else {
			// Default to no indent (start of file or empty file)
			baseIndent = ""
		}

		indentedReplaceLines := make([]string, len(replaceLines))
		for i, line := range replaceLines {
			// Basic indentation: apply the base indent to all replacement lines
			// More complex logic needed to handle relative indents within the replace block itself
			indentedReplaceLines[i] = baseIndent + strings.TrimLeft(line, " \t") // Apply base, remove original leading whitespace
		}

		// Reconstruct resultLines
		var newResultLines []string
		newResultLines = append(newResultLines, resultLines[:matchIndex]...)
		newResultLines = append(newResultLines, indentedReplaceLines...)
		if matchIndex+len(matchedLines) <= len(resultLines) {
			newResultLines = append(newResultLines, resultLines[matchIndex+len(matchedLines):]...)
		}

		resultLines = newResultLines
		delta += len(indentedReplaceLines) - len(matchedLines) // Update line delta
		appliedCount++
	} // End of loop through replacements

	// 4. Final Result
	if appliedCount == 0 && len(failedParts) > 0 {
		// If no blocks were applied successfully, but there were failures
		return &DiffResult{Success: false, FailParts: failedParts, Error: "No diff blocks could be applied."}, nil
	}

	finalContent := strings.Join(resultLines, lineEnding)
	return &DiffResult{Success: true, Content: finalContent, FailParts: failedParts}, nil
}

func (s *MultiSearchReplaceDiffStrategy) validateMarkerSequencing(diffContent string) (bool, string) {
	type State int
	const (
		START State = iota
		AFTER_SEARCH
		AFTER_SEPARATOR
	)

	state := START
	line := 0

	const SEARCH = "<<<<<<< SEARCH"
	const SEP = "======="
	const REPLACE = ">>>>>>> REPLACE"
	const SEARCH_PREFIX = "<<<<<<"
	const REPLACE_PREFIX = ">>>>>>>"

	reportMergeConflictError := func(found, expected string) string {
		return fmt.Sprintf(`ERROR: Special marker '%s' found in your diff content at line %d:

When removing merge conflict markers like '%s' from files, you MUST escape them
in your SEARCH section by prepending a backslash (\) at the beginning of the line:

CORRECT FORMAT:

<<<<<<< SEARCH
content before
\%s    <-- Note the backslash here in this example
content after
=======
replacement content
>>>>>>> REPLACE

Without escaping, the system confuses your content with diff syntax markers.
You may use multiple diff blocks in a single diff request, but ANY of ONLY the following separators that occur within SEARCH or REPLACE content must be escaped, as follows:
\%s
\%s
\%s`, found, line, found, found, SEARCH, SEP, REPLACE)
	}

	reportInvalidDiffError := func(found, expected string) string {
		return fmt.Sprintf(`ERROR: Diff block is malformed: marker '%s' found in your diff content at line %d. Expected: %s

CORRECT FORMAT:

<<<<<<< SEARCH
:start_line: (required) The line number of original content where the search block starts.
:end_line: (required) The line number of original content  where the search block ends.
-------
[exact content to find including whitespace]
=======
[new content to replace with]
>>>>>>> REPLACE`, found, line, expected)
	}

	lines := strings.Split(diffContent, "\n")
	searchCount := 0
	sepCount := 0
	replaceCount := 0

	for _, l := range lines {
		if strings.TrimSpace(l) == SEARCH {
			searchCount++
		} else if strings.TrimSpace(l) == SEP {
			sepCount++
		} else if strings.TrimSpace(l) == REPLACE {
			replaceCount++
		}
	}

	likelyBadStructure := searchCount != replaceCount || sepCount < searchCount

	for _, lineContent := range lines {
		line++
		marker := strings.TrimSpace(lineContent)

		switch state {
		case START:
			if marker == SEP {
				if likelyBadStructure {
					return false, reportInvalidDiffError(SEP, SEARCH)
				}
				return false, reportMergeConflictError(SEP, SEARCH)
			}
			if marker == REPLACE {
				return false, reportInvalidDiffError(REPLACE, SEARCH)
			}
			if strings.HasPrefix(marker, REPLACE_PREFIX) {
				return false, reportMergeConflictError(marker, SEARCH)
			}
			if marker == SEARCH {
				state = AFTER_SEARCH
			} else if strings.HasPrefix(marker, SEARCH_PREFIX) {
				return false, reportMergeConflictError(marker, SEARCH)
			}

		case AFTER_SEARCH:
			if marker == SEARCH {
				return false, reportInvalidDiffError(SEARCH, SEP)
			}
			if strings.HasPrefix(marker, SEARCH_PREFIX) {
				return false, reportMergeConflictError(marker, SEARCH)
			}
			if marker == REPLACE {
				return false, reportInvalidDiffError(REPLACE, SEP)
			}
			if strings.HasPrefix(marker, REPLACE_PREFIX) {
				return false, reportMergeConflictError(marker, SEARCH)
			}
			if marker == SEP {
				state = AFTER_SEPARATOR
			}

		case AFTER_SEPARATOR:
			if marker == SEARCH {
				return false, reportInvalidDiffError(SEARCH, REPLACE)
			}
			if strings.HasPrefix(marker, SEARCH_PREFIX) {
				return false, reportMergeConflictError(marker, REPLACE)
			}
			if marker == SEP {
				if likelyBadStructure {
					return false, reportInvalidDiffError(SEP, REPLACE)
				}
				return false, reportMergeConflictError(SEP, REPLACE)
			}
			if marker == REPLACE {
				state = START
			} else if strings.HasPrefix(marker, REPLACE_PREFIX) {
				return false, reportMergeConflictError(marker, REPLACE)
			}
		}
	}

	if state == START {
		return true, ""
	}

	expected := SEP
	if state == AFTER_SEPARATOR {
		expected = REPLACE
	}

	return false, fmt.Sprintf(`ERROR: Unexpected end of sequence: Expected '%s' was not found.`, expected)
}

// everyLineHasLineNumbers checks if every non-empty line in the content starts with a line number pattern.
func everyLineHasLineNumbers(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}

	lines := strings.Split(content, "\n")
	lineNumberRegex := regexp.MustCompile(`^\s*\d+\s*\|\s`)

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // Skip empty lines
		}

		if !lineNumberRegex.MatchString(line) {
			return false
		}
	}

	return true
}

// stripLineNumbers removes line numbers from content where each line starts with a pattern like "123 | ".
func stripLineNumbers(content string) string {
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	lineNumberRegex := regexp.MustCompile(`^\s*\d+\s*\|\s`)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, line) // Keep empty lines as-is
			continue
		}

		// Remove line number pattern
		strippedLine := lineNumberRegex.ReplaceAllString(line, "")
		result = append(result, strippedLine)
	}

	return strings.Join(result, "\n")
}
