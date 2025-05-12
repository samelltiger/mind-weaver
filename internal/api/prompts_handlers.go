// internal/command_handlers.go
package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
	"mind-weaver/internal/third/prompts"
	"mind-weaver/internal/third/prompts/sections"
)

// TestPromptRequest represents the request for testing prompts
type TestPromptRequest struct {
	Prompt             string                     `json:"prompt"`
	Mode               string                     `json:"mode"`
	GlobalInstructions string                     `json:"globalInstructions"`
	EnvContext         prompts.EnvironmentContext `json:"envContext"`
	CustomModeConfigs  []sections.ModeConfig      `json:"customModeConfigs"`
}

// TestPromptResponse represents the response for generating a system prompt
type TestPromptResponse struct {
	Result string `json:"result"`
}

// TestPrompt godoc
// @Summary Generate a system prompt using the prompt engine
// @Description Builds a system prompt using the given parameters
// @Tags Prompts
// @Accept json
// @Produce json
// @Param data body TestPromptRequest true "Request data"
// @Success 200 {object} base.Response{data=TestPromptResponse}
// @Router /prompts/test [post]
func (h *Handler) TestPrompt(c *gin.Context) {
	var req TestPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}

	// Build args for BuildSystemPrompt
	args := prompts.BuildSystemPromptArgs{
		EnvCtx:             req.EnvContext,
		Mode:               sections.ModeSlug(req.Mode),
		CustomModeConfigs:  req.CustomModeConfigs,
		GlobalInstructions: req.GlobalInstructions,
		// Note: DiffStrategy and RooIgnoreController are nil here
	}

	// Generate the system prompt
	result, err := prompts.BuildSystemPrompt(args)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to add user message: %v", err))
		return
	}

	// Return the generated prompt
	base.SuccessResponse(c, TestPromptResponse{
		Result: result,
	})
}
