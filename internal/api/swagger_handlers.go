// internal/command_handlers.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
	"mind-weaver/internal/utils"
)

// ListInterfacesRequest represents the request for listing interfaces
type ListInterfacesRequest struct {
	SwaggerSource string `json:"swagger_file"`
}

// GenerateDocRequest represents the request for generating documentation
type GenerateDocRequest struct {
	SwaggerSource string               `json:"swagger_file"`
	ApiList       []utils.ApiInterface `json:"api_list"`
}

// GenerateDocResponse represents the response for generating documentation
type GenerateDocResponse struct {
	Result string `json:"result"`
}

// ListInterfaces godoc
// @Summary List all interfaces in a Swagger document
// @Description Returns a list of all interfaces in the Swagger document
// @Tags Swagger
// @Accept json
// @Produce json
// @Param data body ListInterfacesRequest true "Request data"
// @Success 200 {object} base.Response{data=[]utils.ApiInterface}
// @Router /swaggers/list [post]
func (h *Handler) ListInterfaces(c *gin.Context) {
	var req ListInterfacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid request")
		return
	}

	result, err := h.swaggerService.ListInterfaces(
		req.SwaggerSource,
		false,
		"",
		false,
	)

	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, err.Error())
		return
	}

	base.SuccessResponse(c, result)
}

// GenerateDoc godoc
// @Summary Generate API documentation from a Swagger document
// @Description Generates Markdown documentation for specified interfaces
// @Tags Swagger
// @Accept json
// @Produce json
// @Param data body GenerateDocRequest true "Request data"
// @Success 200 {object} base.Response{data=GenerateDocResponse}
// @Router /swaggers/doc [post]
func (h *Handler) GenerateDoc(c *gin.Context) {
	var req GenerateDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid request")
		return
	}

	result, err := h.swaggerService.GenerateDoc(
		req.SwaggerSource,
		false,
		req.ApiList,
		false,
		"",
		false,
	)

	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, err.Error())
		return
	}

	base.SuccessResponse(c, GenerateDocResponse{
		Result: result,
	})
}
