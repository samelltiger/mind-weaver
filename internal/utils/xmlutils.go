package utils

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Basic structure to capture suggestions from the simple XML format.
type suggestionsWrapper struct {
	XMLName xml.Name `xml:"follow_up"` // Assuming suggestions are wrapped in <follow_up>
	Suggest []string `xml:"suggest"`
}

// ParseSimpleXmlSuggest extracts text content from <suggest> tags within a parent tag.
// This is a basic parser; robust XML handling might need encoding/xml with proper structs.
func ParseSimpleXmlSuggest(xmlString string) ([]string, error) {
	// Wrap the input string to ensure it has a root element if it's just fragments.
	// The expected input format is `<suggest>...</suggest><suggest>...</suggest>`
	// We need a dummy root for the decoder.
	fullXml := "<follow_up>" + xmlString + "</follow_up>"

	decoder := xml.NewDecoder(strings.NewReader(fullXml))
	var wrapper suggestionsWrapper

	err := decoder.Decode(&wrapper)
	if err != nil {
		// Attempt fallback if Decode fails (e.g., malformed XML) - might indicate LLM error
		// A simple regex might be a fallback, but brittle.
		return nil, fmt.Errorf("failed to decode suggestions XML: %w", err)
	}

	// Trim whitespace from each suggestion
	cleanedSuggestions := make([]string, len(wrapper.Suggest))
	for i, s := range wrapper.Suggest {
		cleanedSuggestions[i] = strings.TrimSpace(s)
	}

	return cleanedSuggestions, nil
}
