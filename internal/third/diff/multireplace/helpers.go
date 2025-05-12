package multireplace

import (
	"regexp"
	"strings"

	"github.com/adrg/strutil" // Using a library for fuzzy matching
	"github.com/adrg/strutil/metrics"
)

// Using a library for fuzzy matching
type DiffResult struct {
	Success   bool          `json:"success"`
	Content   string        `json:"content,omitempty"`
	Error     string        `json:"error,omitempty"`
	FailParts []*DiffResult `json:"failParts,omitempty"` // For multi-block diffs
	// Simplified details compared to TS
	Details *Details `json:"details,omitempty"`
}

type Details struct {
	Similarity    float64 `json:"similarity,omitempty"`
	Threshold     float64 `json:"threshold,omitempty"`
	SearchContent string  `json:"searchContent,omitempty"`
	BestMatch     string  `json:"bestMatch,omitempty"`
}

// ToolDescriptionArgs holds arguments needed for GetToolDescription
type ToolDescriptionArgs struct {
	Cwd string
	// ToolOptions map[string]string // Keep if needed
}

func normalizeStr(str string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(str, " "))
}

func getSimilarity(original, search string, metric *metrics.Levenshtein) float64 {
	if search == "" {
		return 1.0
	}
	normOriginal := normalizeStr(original)
	normSearch := normalizeStr(search)

	if normOriginal == normSearch {
		return 1.0
	}

	// Use strutil for similarity calculation
	// Note: strutil's Similarity uses Levenshtein distance internally
	// Ensure the metric options match the desired normalization (case sensitivity etc.)
	// metric.CaseSensitive = false // Example if case-insensitivity is needed
	similarity := strutil.Similarity(normOriginal, normSearch, metric)
	return similarity
}

func unescapeMarkers(content string) string {
	content = regexp.MustCompile(`(?m)^\\<<<<<<<`).ReplaceAllString(content, "<<<<<<<")
	content = regexp.MustCompile(`(?m)^\\=======`).ReplaceAllString(content, "=======")
	content = regexp.MustCompile(`(?m)^\\>>>>>>>`).ReplaceAllString(content, ">>>>>>>")
	content = regexp.MustCompile(`(?m)^\\-------`).ReplaceAllString(content, "-------")
	content = regexp.MustCompile(`(?m)^\\:start_line:`).ReplaceAllString(content, ":start_line:")
	content = regexp.MustCompile(`(?m)^\\:end_line:`).ReplaceAllString(content, ":end_line:")
	return content
}

// GetIndent returns the leading whitespace characters (indentation) of a string.
func GetIndent(line string) string {
	// Iterate through each character in the line until we find a non-whitespace character
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	// The entire line consists of whitespace
	return line
}
