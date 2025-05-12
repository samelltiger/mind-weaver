package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
	"mind-weaver/internal/utils"
)

type ReadFileReq struct {
	ProjectID int64  `json:"project_id" binding:"required"`
	FilePath  string `json:"file_path" binding:"required"`
}

type ReadFileResp struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// ReadFile 读取文件内容
// @Summary      读取文件内容
// @Description  读取指定项目中的文件内容
// @Tags         file
// @Accept       json
// @Produce      json
// @Param        request  body      ReadFileReq  true  "文件请求参数"
// @Success      200      {object}  base.Response{data=ReadFileResp}
// @Failure      400      {object}  base.Response
// @Failure      403      {object}  base.Response
// @Failure      404      {object}  base.Response
// @Failure      500      {object}  base.Response
// @Router       /files/read [post]
func (h *Handler) ReadFile(c *gin.Context) {
	var req ReadFileReq

	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// Get the project
	project, err := h.database.GetProject(req.ProjectID)
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, base.ErrCodeNotFound, "Project not found")
		return
	}

	// Determine the full file path
	var fullPath string
	if filepath.IsAbs(req.FilePath) {
		// Ensure the file is within the project directory
		if !utils.IsSubPath(project.Path, req.FilePath) {
			base.ErrorResponse(c, http.StatusForbidden, base.ErrCodeInvalidParams, "File is outside of project directory")
			return
		}
		fullPath = req.FilePath
	} else {
		// Relative path - join with project path
		fullPath = filepath.Join(project.Path, req.FilePath)
	}

	// Read the file
	content, err := h.fileService.ReadFile(fullPath)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to read file: %v", err))
		return
	}

	base.SuccessResponse(c, ReadFileResp{
		FilePath: req.FilePath,
		Content:  content,
	})
}

// ReadFile 读取文件内容
// @Summary      读取文件内容
// @Description  读取指定项目中的文件内容
// @Tags         file
// @Accept       text/html
// @Produce      json
// @Param        path  query  string  true  "文件请求参数"
// @Success      200  {string}  string  "HTML content"
// @Failure      404  {string}  string  "<h1>404 Not Found</h1>"
// @Router       /files/single-html [get]
func (h *Handler) ReadHtmlFile(c *gin.Context) {
	path := c.Query("path")

	// Determine the full file path
	var fullPath string
	if filepath.IsAbs(path) {
		fullPath = path
	} else {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusNotFound, "<h1>404 Not Found</h1>")
		return
	}

	// Read the file
	content, err := h.fileService.ReadFile(fullPath)
	if err != nil {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusNotFound, "<h1>500 Server Error</h1>")
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, content)
}
