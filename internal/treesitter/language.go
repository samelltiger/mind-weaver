package treesitter

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	js "github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/swift"
	tsx "github.com/smacker/go-tree-sitter/typescript/tsx"
	ts "github.com/smacker/go-tree-sitter/typescript/typescript"
	// "github.com/smacker/go-tree-sitter/typescript/tsx"
	// "github.com/smacker/go-tree-sitter/typescript/typescript"
	// Note: WASM loading is different in go-tree-sitter. It typically uses
	// pre-compiled bindings or CGO. If WASM is strictly required,
	// a different approach using a WASM runtime like wasmer-go would be needed.
	// For simplicity, this translation uses the standard go-tree-sitter bindings.
)

var (
	languageMap = map[string]*sitter.Language{
		"js":  js.GetLanguage(),
		"jsx": tsx.GetLanguage(), // Use JS parser for JSX
		// "json":  javascript.GetLanguage(), // Use JS parser for JSON
		"ts":    ts.GetLanguage(),
		"tsx":   tsx.GetLanguage(),
		"py":    python.GetLanguage(),
		"rs":    rust.GetLanguage(),
		"go":    golang.GetLanguage(),
		"cpp":   cpp.GetLanguage(),
		"hpp":   cpp.GetLanguage(),
		"c":     c.GetLanguage(),
		"h":     c.GetLanguage(),
		"cs":    csharp.GetLanguage(),
		"rb":    ruby.GetLanguage(),
		"java":  java.GetLanguage(),
		"php":   php.GetLanguage(),
		"swift": swift.GetLanguage(),
		"kt":    kotlin.GetLanguage(),
		"kts":   kotlin.GetLanguage(),
	}

	queryMap = map[string]string{
		"js":   javascriptQuery,
		"jsx":  javascriptQuery, // Reusing JS query
		"json": javascriptQuery, // Reusing JS query
		"ts":   typescriptQuery,
		"tsx":  tsxQuery,
		"py":   pythonQuery,
		// "rs":    rustQuery,
		"go":    goQuery,
		"cpp":   cppQuery,
		"hpp":   cppQuery,
		"c":     cQuery,
		"h":     cQuery,
		"cs":    csharpQuery,
		"rb":    rubyQuery,
		"java":  javaQuery,
		"php":   phpQuery,
		"swift": swiftQuery,
		"kt":    kotlinQuery,
		"kts":   kotlinQuery,
	}

	loadedParsers = make(LanguageParsers)
	parserMutex   sync.RWMutex
)

// LoadRequiredLanguageParsers loads and caches parsers for the required file extensions.
// It uses pre-compiled go-tree-sitter bindings instead of dynamically loading WASM.
func LoadRequiredLanguageParsers(ctx context.Context, filesToParse []string) (LanguageParsers, error) {
	parserMutex.Lock()
	defer parserMutex.Unlock()

	requiredExts := make(map[string]struct{})
	for _, file := range filesToParse {
		ext := strings.ToLower(filepath.Ext(file))
		if len(ext) > 1 { // Ensure there's an extension
			requiredExts[ext[1:]] = struct{}{} // Store extension without dot
		}
	}

	parsersToReturn := make(LanguageParsers)
	for ext := range requiredExts {
		// Check cache first
		if info, exists := loadedParsers[ext]; exists {
			parsersToReturn[ext] = info
			continue
		}

		// Load language and query based on extension
		lang, langOk := languageMap[ext]
		queryStr, queryOk := queryMap[ext]

		if !langOk || !queryOk {
			// Skip unsupported extensions silently, or return error if preferred
			// return nil, fmt.Errorf("unsupported language extension: %s", ext)
			continue
		}

		// Compile the query
		query, err := sitter.NewQuery([]byte(queryStr), lang)
		if err != nil {
			return nil, fmt.Errorf("failed to compile query for %s: %w", ext, err)
		}

		// Create a new parser instance
		parser := sitter.NewParser()
		parser.SetLanguage(lang)

		// Store in cache and return map
		info := ParserInfo{Parser: parser, Query: query}
		loadedParsers[ext] = info
		parsersToReturn[ext] = info
	}

	return parsersToReturn, nil
}
