package treesitter

import (
	"fmt"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ParseMarkdown parses markdown content and returns captures mimicking tree-sitter.
func ParseMarkdown(content string) []MarkdownCapture {
	if content == "" || strings.TrimSpace(content) == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	captures := []MarkdownCapture{}

	for i, line := range lines {
		// Check for ATX headers (# Header)
		atxMatch := atxHeaderRegex.FindStringSubmatch(line)
		if len(atxMatch) == 3 {
			level := len(atxMatch[1])
			text := strings.TrimSpace(atxMatch[2])
			node := MarkdownNode{
				StartPosition: sitter.Point{Row: uint32(i), Column: 0},
				EndPosition:   sitter.Point{Row: uint32(i), Column: uint32(len(line))},
				Text:          text,
			}
			name := fmt.Sprintf("name.definition.header.h%d", level)
			defName := fmt.Sprintf("definition.header.h%d", level)
			captures = append(captures, MarkdownCapture{Node: node, Name: name})
			captures = append(captures, MarkdownCapture{Node: node, Name: defName})
			continue
		}

		// Check for setext headers (underlined headers)
		if i > 0 {
			prevLine := lines[i-1]
			isH1 := setextH1Regex.MatchString(line) && validSetextTextRegex.MatchString(prevLine)
			isH2 := setextH2Regex.MatchString(line) && validSetextTextRegex.MatchString(prevLine)

			if isH1 || isH2 {
				text := strings.TrimSpace(prevLine)
				level := 1
				if isH2 {
					level = 2
				}

				node := MarkdownNode{
					StartPosition: sitter.Point{Row: uint32(i - 1), Column: 0},
					// Initial end position is the underline
					EndPosition: sitter.Point{Row: uint32(i), Column: uint32(len(line))},
					Text:        text,
				}
				name := fmt.Sprintf("name.definition.header.h%d", level)
				defName := fmt.Sprintf("definition.header.h%d", level)
				captures = append(captures, MarkdownCapture{Node: node, Name: name})
				captures = append(captures, MarkdownCapture{Node: node, Name: defName})
				continue
			}
		}
	}

	// Calculate section ranges
	// Sort captures by their start position
	sort.SliceStable(captures, func(i, j int) bool {
		return captures[i].Node.StartPosition.Row < captures[j].Node.StartPosition.Row
	})

	// Group captures by header (name and definition pairs)
	headerCaptures := [][]MarkdownCapture{}
	for i := 0; i < len(captures); i += 2 {
		if i+1 < len(captures) {
			// Ensure they are a pair for the same header before grouping
			if captures[i].Node.StartPosition.Row == captures[i+1].Node.StartPosition.Row &&
				strings.Replace(captures[i].Name, "name.", "", 1) == captures[i+1].Name {
				headerCaptures = append(headerCaptures, []MarkdownCapture{captures[i], captures[i+1]})
			} else {
				// Handle potential single capture if pairing failed
				headerCaptures = append(headerCaptures, []MarkdownCapture{captures[i]})
				i-- // Adjust index as we only consumed one
			}
		} else {
			headerCaptures = append(headerCaptures, []MarkdownCapture{captures[i]})
		}
	}

	// Update end positions for section ranges
	numHeaders := len(headerCaptures)
	for i := 0; i < numHeaders; i++ {
		headerPair := headerCaptures[i]
		var endRow uint32

		if i < numHeaders-1 {
			// End position is the start of the next header minus 1
			nextHeaderStartRow := headerCaptures[i+1][0].Node.StartPosition.Row
			if nextHeaderStartRow > 0 {
				endRow = nextHeaderStartRow - 1
			} else {
				endRow = 0 // Handle case where next header is on the first line
			}
		} else {
			// Last header extends to the end of the file
			endRow = uint32(len(lines) - 1)
		}

		// Ensure endRow is not less than startRow
		if endRow < headerPair[0].Node.StartPosition.Row {
			endRow = headerPair[0].Node.StartPosition.Row
		}

		// Update end position for both captures in the pair
		for j := range headerPair {
			// Keep the original EndPosition column if needed, or set to 0/end-of-line
			headerPair[j].Node.EndPosition.Row = endRow
			// Optionally adjust column:
			// if i < numHeaders-1 {
			//     headerPair[j].Node.EndPosition.Column = 0 // Or length of the endRow line
			// } else {
			//     headerPair[j].Node.EndPosition.Column = uint32(len(lines[endRow]))
			// }
		}
	}

	// Flatten the grouped captures back to a single array
	finalCaptures := []MarkdownCapture{}
	for _, pair := range headerCaptures {
		finalCaptures = append(finalCaptures, pair...)
	}

	return finalCaptures
}
