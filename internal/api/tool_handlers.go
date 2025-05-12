// internal/command_handlers.go
package api

import (
	"mind-weaver/internal/api/base"
	"mind-weaver/pkg/logger"
	"mind-weaver/pkg/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

// JsInspector 解析html文件并判断是否出错
// @Summary      解析html文件并判断是否出错
// @Description  解析html文件并判断是否出错
// @Tags         tools
// @Accept       json
// @Produce      json
// @Param        path  query  string  true  "html文件的绝对路径"
// @Success      200  {object}  base.Response{data=services.ContextInfo}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /tools/jsinspector [get]
func (h *Handler) JsInspector(c *gin.Context) {
	path := c.Query("path")

	result, outputs, err := h.commandService.JsInspector(path)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, err.Error())
		return
	}

	// Return results to client
	response := CommandResponse{
		Output:   outputs,
		ExitCode: result.ExitCode,
		Success:  result.Success,
		ErrorMsg: result.ErrorMessage,
	}

	base.SuccessResponse(c, response)
}

// JsInspector 解析html文件并判断是否出错
// @Summary      解析html文件并判断是否出错
// @Description  解析html文件并判断是否出错
// @Tags         tools
// @Accept       json
// @Produce      json
// @Success      200  {object}  base.Response{data=services.ContextInfo}
// @Failure      500  {object}  base.Response
// @Router       /tools/handle-llm-response [get]
func (h *Handler) HandleLlmResponseError(c *gin.Context) {
	aiResList := []string{
		util.ReadFileToString("./test_data/aiNewAddRes-4.txt"),
		util.ReadFileToString("./test_data/aiNewAddRes-last.txt"),
	}

	err := h.handleLlmResponseError(aiResList, "/mnt/h/code/test_project/tetris2-test.html")
	if err != nil {
		logger.Error("HandleLlmResponseError error: ", err)
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, err.Error())
		return
	}

	base.SuccessResponse(c, "success")
}
