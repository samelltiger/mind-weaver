package tools

import (
	"fmt"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/prompts"
	"mind-weaver/internal/utils"
	"strings"
)

// AskFollowupQuestionTool signals that the AI needs to ask a question.
// The actual question asking and response handling occurs on the frontend/API layer.
// This tool's result tells the *calling system* (not the LLM directly) that a question was intended.
func AskFollowupQuestionTool(input ExecutorInput) (*ExecutorResult, error) {
	question, ok := input.ToolUse.Params[string(assistantmessage.Question)]
	if !ok || strings.TrimSpace(question) == "" {
		errText := prompts.FormatMissingParamError(string(input.ToolUse.Name), string(assistantmessage.Question))
		// This error is tricky - the LLM *tried* to ask but failed format.
		// We might return the error for the LLM to retry, or signal failure differently.
		// Returning error message for LLM for now.
		return &ExecutorResult{Result: errText, IsError: true}, nil
	}

	followUpXml, _ := input.ToolUse.Params[string(assistantmessage.FollowUp)]
	suggestions := []string{}
	if followUpXml != "" {
		// Basic XML parsing to extract suggestions
		// Using a simple regex approach here, a proper XML parser is better.
		parsedSuggestions, err := utils.ParseSimpleXmlSuggest(followUpXml)
		if err != nil {
			fmt.Printf("Warning: Failed to parse suggestions XML: %v\nXML: %s\n", err, followUpXml)
			// Proceed without suggestions if parsing fails
		} else {
			suggestions = parsedSuggestions
		}
	}

	// The backend doesn't *ask* the question. It processes the LLM's *request* to ask.
	// The result should inform the system controlling the LLM interaction.
	// Option 1: Return structured data (requires calling system to handle)
	// type FollowUpData struct { Question string; Suggestions []string }
	// return &ExecutorResult{Result: "", Data: FollowUpData{Question: question, Suggestions: suggestions}}, nil

	// Option 2: Return a formatted string indicating the intent (simpler for LLM loop)
	// This tells the system "The AI used ask_followup_question". The system then needs
	// to present the question/suggestions to the user and feed the answer back.
	// The string itself isn't directly shown back to the LLM in the next turn usually.
	resultText := fmt.Sprintf("SYSTEM_SIGNAL: ASK_FOLLOWUP\nQuestion: %s", question)
	if len(suggestions) > 0 {
		resultText += "\nSuggestions:\n- " + strings.Join(suggestions, "\n- ")
	}

	// We don't return an error *to the LLM*. This tool use was validly formatted.
	return &ExecutorResult{Result: resultText}, nil
}
