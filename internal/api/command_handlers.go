// internal/command_handlers.go
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
	"mind-weaver/pkg/logger"
	"mind-weaver/pkg/util"
)

// ExecuteCommandRequest represents a request to execute a shell command
type ExecuteCommandRequest struct {
	Command string `json:"command" binding:"required"`
}

// ExecuteCodeRequest represents a request to execute code
type ExecuteCodeRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
}

// CommandResponse represents the response from command execution
type CommandResponse struct {
	Output   []util.CommandOutput `json:"output"`
	ExitCode int                  `json:"exitCode"`
	Success  bool                 `json:"success"`
	ErrorMsg string               `json:"errorMessage,omitempty"`
}

// ExecuteCommand 处理 shell 命令的执行
// @Summary 执行 shell 命令
// @Description 执行 shell 命令并返回结果
// @Tags 命令
// @Accept json
// @Produce json
// @Param request body ExecuteCommandRequest true "要执行的命令"
// @Success 200 {object} base.Response{data=CommandResponse}
// @Failure 400 {object} base.Response
// @Failure 500 {object} base.Response
// @Router /commands/execute [post]
func (h *Handler) ExecuteCommand(c *gin.Context) {
	var req ExecuteCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}
	commandService := util.NewCommandService()

	// Validate command
	safe, reason := commandService.ValidateCommand(req.Command)
	if !safe {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, reason)
		return
	}

	// Collect command output
	var outputs []util.CommandOutput
	outputChan := make(chan util.CommandOutput, 100)

	done := make(chan struct{})
	go func() {
		for output := range outputChan {
			outputs = append(outputs, output)
		}
		close(done)
	}()

	// Execute command
	result, err := commandService.ExecuteCommand(c.Request.Context(), req.Command, outputChan)
	if err != nil {
		logger.Errorf("ExecuteCommand error: %v", err)

		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, reason)
		return
	}

	// Wait for output collection to complete
	<-done

	// Return results to client
	response := CommandResponse{
		Output:   outputs,
		ExitCode: result.ExitCode,
		Success:  result.Success,
		ErrorMsg: result.ErrorMessage,
	}

	base.SuccessResponse(c, response)
}

// ExecuteCode 处理代码片段的执行
// @Summary 执行指定语言的代码
// @Description 执行指定编程语言的代码并返回结果
// @Tags 命令
// @Accept json
// @Produce json
// @Param request body ExecuteCodeRequest true "要执行的代码"
// @Success 200 {object} base.Response{data=CommandResponse}
// @Failure 400 {object} base.Response
// @Failure 500 {object} base.Response
// @Router /commands/execute-code [post]
func (h *Handler) ExecuteCode(c *gin.Context) {
	var req ExecuteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	commandService := util.NewCommandService()

	// Collect command output
	var outputs []util.CommandOutput
	outputChan := make(chan util.CommandOutput, 100)

	// Start a goroutine to collect outputs
	done := make(chan struct{})
	go func() {
		for output := range outputChan {
			outputs = append(outputs, output)
		}
		close(done)
	}()

	// Execute code
	result, err := commandService.ExecuteCode(ctx, req.Code, req.Language, outputChan)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, err.Error())
		return
	}

	// Wait for output collection to complete
	<-done

	// Return results to client
	response := CommandResponse{
		Output:   outputs,
		ExitCode: result.ExitCode,
		Success:  result.Success,
		ErrorMsg: result.ErrorMessage,
	}

	base.SuccessResponse(c, response)
}
