package treesitter

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

const minComponentLinesDefault = 4

// htmlElementRegex is a simplified check for common HTML elements at the start of a line.
// Adjust as needed for more complex cases.
var htmlElementRegex = regexp.MustCompile(`^\s*<\/?(div|span|button|input|h[1-6]|p|a|img|ul|li|form)\b`)

// isNotHtmlElement checks if a line likely starts with a common HTML element.
func isNotHtmlElement(line string) bool {
	return !htmlElementRegex.MatchString(line)
}

// ProcessCaptures processes tree-sitter captures and formats them into a string.
func ProcessCaptures(captures []*sitter.QueryCapture, content []byte, minComponentLines int) string {
	if len(captures) == 0 {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	var processedInfos []CaptureInfo
	processedLineKeys := make(map[string]struct{}) // Track processed line ranges (e.g., "5-10")

	// Sort captures by start position
	sort.SliceStable(captures, func(i, j int) bool {
		return captures[i].Node.StartPoint().Row < captures[j].Node.StartPoint().Row
	})

	for _, capture := range captures {
		node := capture.Node
		captureName := capture.Settings.Query.CaptureNameForId(capture.Index) // Get capture name

		// Skip captures that don't represent definitions or names
		if !strings.Contains(captureName, "definition") && !strings.Contains(captureName, "name") {
			continue
		}

		// Get the parent node that contains the full definition
		// For 'name' captures, the parent usually holds the full block.
		// For 'definition' captures, the node itself might be the block.
		definitionNode := node
		if strings.Contains(captureName, "name") && node.Parent() != nil {
			// Heuristic: Use parent for name captures unless it's a tiny node type
			// You might need more sophisticated logic based on specific grammars
			if node.Parent().EndPoint().Row > node.EndPoint().Row || node.Parent().StartPoint().Row < node.StartPoint().Row {
				definitionNode = node.Parent()
			}
		}

		if definitionNode == nil {
			continue
		}

		startLine := definitionNode.StartPoint().Row
		endLine := definitionNode.EndPoint().Row
		lineCount := endLine - startLine + 1

		// Skip components that don't span enough lines
		if lineCount < uint32(minComponentLines) {
			continue
		}

		// Create unique key for this definition based on line range
		lineKey := fmt.Sprintf("%d-%d", startLine, endLine)

		// Skip already processed lines/ranges
		if _, exists := processedLineKeys[lineKey]; exists {
			continue
		}

		// Get the first line of the definition node's content
		firstLineIndex := int(startLine)
		var definitionLine string
		if firstLineIndex >= 0 && firstLineIndex < len(lines) {
			definitionLine = strings.TrimSpace(lines[firstLineIndex])
		} else {
			definitionLine = strings.TrimSpace(node.Content(content)) // Fallback
		}

		// Filter out common HTML elements if desired (less critical in backend)
		// if !isNotHtmlElement(definitionLine) {
		//     continue
		// }

		// Add to processed list
		processedInfos = append(processedInfos, CaptureInfo{
			StartLine:      startLine + 1, // 1-based indexing for output
			EndLine:        endLine + 1,   // 1-based indexing for output
			DefinitionLine: definitionLine,
			CaptureName:    captureName,
		})
		processedLineKeys[lineKey] = struct{}{} // Mark range as processed

	}

	// Format the output string
	var builder strings.Builder
	for _, info := range processedInfos {
		// Use the first line of the actual definition content for display
		builder.WriteString(fmt.Sprintf("%d--%d | %s\n", info.StartLine, info.EndLine, info.DefinitionLine))
	}

	return builder.String()
}

// ProcessMarkdownCaptures processes markdown captures and formats them.
func ProcessMarkdownCaptures(captures []MarkdownCapture, content string, minComponentLines int) string {
	if len(captures) == 0 {
		return ""
	}

	lines := strings.Split(content, "\n")
	var builder strings.Builder
	processedLineKeys := make(map[string]struct{}) // Track processed line ranges

	// Process only the definition captures (every other capture)
	for i := 1; i < len(captures); i += 2 {
		capture := captures[i]
		startLine := capture.Node.StartPosition.Row
		endLine := capture.Node.EndPosition.Row

		// Defensive check for valid range
		if endLine < startLine {
			endLine = startLine
		}

		sectionLength := endLine - startLine + 1

		// Only include sections that span at least minComponentLines lines
		if sectionLength >= uint32(minComponentLines) {
			lineKey := fmt.Sprintf("%d-%d", startLine, endLine)
			if _, exists := processedLineKeys[lineKey]; exists {
				continue
			}

			// Extract header level from the name
			headerLevel := 1
			headerMatch := regexp.MustCompile(`\.h(\d)$`).FindStringSubmatch(capture.Name)
			if len(headerMatch) == 2 {
				fmt.Sscan(headerMatch[1], &headerLevel)
			}
			if headerLevel < 1 {
				headerLevel = 1
			}
			if headerLevel > 6 {
				headerLevel = 6
			}

			headerPrefix := strings.Repeat("#", headerLevel)

			// Format: startLine--endLine | # Header Text
			// Output lines are 1-based
			builder.WriteString(fmt.Sprintf("%d--%d | %s %s\n", startLine+1, endLine+1, headerPrefix, capture.Node.Text))
			processedLineKeys[lineKey] = struct{}{}
		}
	}

	return builder.String()
}
