package treesitter

import (
	"regexp"

	sitter "github.com/smacker/go-tree-sitter"
)

// EditorContext holds information passed from the web frontend/API.
type EditorContext struct {
	DirPath string // The root directory path to scan.
	WasmDir string // The directory containing tree-sitter WASM files.
}

// ParserInfo holds the parser and query for a specific language extension.
type ParserInfo struct {
	Parser *sitter.Parser
	Query  *sitter.Query
}

// LanguageParsers maps file extensions (without dot) to their parser info.
type LanguageParsers map[string]ParserInfo

// CaptureInfo represents a processed definition capture.
type CaptureInfo struct {
	StartLine      uint32
	EndLine        uint32
	DefinitionLine string // The first line of the definition text.
	CaptureName    string // The name of the capture (e.g., "definition.class").
}

// --- Markdown Specific Types ---

// MarkdownNode mimics the structure needed for markdown captures.
type MarkdownNode struct {
	StartPosition sitter.Point
	EndPosition   sitter.Point
	Text          string
}

// MarkdownCapture mimics the tree-sitter capture structure for markdown.
type MarkdownCapture struct {
	Node MarkdownNode
	Name string // e.g., "name.definition.header.h1", "definition.header.h1"
}

// --- Regex for Markdown Parsing ---
var (
	atxHeaderRegex       = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	setextH1Regex        = regexp.MustCompile(`^={3,}\s*$`)
	setextH2Regex        = regexp.MustCompile(`^-{3,}\s*$`)
	validSetextTextRegex = regexp.MustCompile(`^\s*[^#<>!\[\]` + "`" + `\t]+[^\n]$`) // Escaped backtick
)

// --- Supported Extensions ---
// We define this centrally for reuse.
var supportedExtensions = map[string]struct{}{
	".js": {}, ".jsx": {}, ".ts": {}, ".tsx": {}, ".py": {}, ".rs": {}, ".go": {},
	".c": {}, ".h": {}, ".cpp": {}, ".hpp": {}, ".cs": {}, ".rb": {}, ".java": {},
	".php": {}, ".swift": {}, ".kt": {}, ".kts": {}, ".md": {}, ".markdown": {},
	".json": {},
}

// isExtensionSupported checks if a file extension (including dot) is supported.
func isExtensionSupported(ext string) bool {
	_, ok := supportedExtensions[ext]
	return ok
}
