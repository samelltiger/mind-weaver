package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
)

// CreateProject 创建项目
// @Summary      创建项目
// @Description  创建项目
// @Tags         project
// @Accept       json
// @Produce      json
// @Success      200  {object}  base.Response{data=db.Project}
// @Failure      401  {object}  base.Response
// @Failure      404  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Path     string `json:"path" binding:"required"`
		Language string `json:"language"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// Check if project already exists in database
	existingProject, err := h.database.GetProjectByPath(req.Path)
	if err == nil && existingProject != nil {
		base.ErrorResponse(c, http.StatusConflict, base.ErrCodeInvalidParams, "Project with this path already exists")
		return
	}

	// Check if path exists in filesystem using standard library
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		// Create the directory with 0755 permissions (rwxr-xr-x)
		if err := os.MkdirAll(req.Path, 0755); err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError,
				fmt.Sprintf("Failed to create project directory: %v", err))
			return
		}
	} else if err != nil {
		// Handle other stat errors (e.g., permission issues)
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError,
			fmt.Sprintf("Failed to check project directory: %v", err))
		return
	}

	// Create the project in database
	projectID, err := h.database.CreateProject(req.Name, req.Path, req.Language)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError,
			fmt.Sprintf("Failed to create project: %v", err))
		return
	}

	// Get the created project
	project, err := h.database.GetProject(projectID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError,
			fmt.Sprintf("Project created but failed to retrieve: %v", err))
		return
	}

	base.SuccessResponse(c, project)
}

type UpdateProjectReq struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Language string `json:"language"`
}

// UpdateProject 更新项目信息
// @Summary      更新项目
// @Description  更新项目名称、路径或语言
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "项目ID"
// @Param        body body      UpdateProjectReq  true  "项目信息（如果没有修改，那么传入原来的值）"
// @Success      200  {object}  base.Response{data=db.Project}
// @Failure      400  {object}  base.Response
// @Failure      404  {object}  base.Response
// @Failure      409  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /projects/{id} [put]
func (h *Handler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid project ID")
		return
	}

	var req UpdateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// Validate the project path if it's being updated
	if req.Path != "" {
		if err := h.fileService.ValidatePath(req.Path); err != nil {
			base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, fmt.Sprintf("Invalid project path: %v", err))
			return
		}

		// Check if new path already exists
		existingProject, err := h.database.GetProjectByPath(req.Path)
		if err == nil && existingProject != nil && existingProject.ID != id {
			base.ErrorResponse(c, http.StatusConflict, base.ErrCodeInvalidParams, "Project with this path already exists")
			return
		}
	}

	// Update the project
	if err := h.database.UpdateProject(id, req.Name, req.Path, req.Language); err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to update project: %v", err))
		return
	}

	// Get the updated project
	project, err := h.database.GetProject(id)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Project updated but failed to retrieve: %v", err))
		return
	}

	base.SuccessResponse(c, project)
}

// GetProjects 获取项目列表
// @Summary      获取所有项目
// @Description  获取所有项目列表，按最后打开时间降序排列
// @Tags         project
// @Accept       json
// @Produce      json
// @Success      200  {object}  base.Response{data=[]db.Project}
// @Failure      401  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /projects [get]
func (h *Handler) GetProjects(c *gin.Context) {
	projects, err := h.database.ListProjects()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to list projects: %v", err))
		return
	}

	base.SuccessResponse(c, projects)
}

// GetProject 获取单个项目详情
// @Summary      获取项目详情
// @Description  根据ID获取单个项目详情，并更新最后打开时间
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "项目ID"
// @Success      200  {object}  base.Response{data=db.Project}
// @Failure      400  {object}  base.Response
// @Failure      404  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /projects/{id} [get]
func (h *Handler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid project ID")
		return
	}

	project, err := h.database.GetProject(id)
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, base.ErrCodeNotFound, "Project not found")
		return
	}

	// Update last opened timestamp
	h.database.UpdateProjectLastOpened(id)

	base.SuccessResponse(c, project)
}

// GetProjectFiles 获取项目文件树
// @Summary      获取项目文件结构
// @Description  获取项目的文件树结构，可指定最大深度
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        id        path      int     true   "项目ID"
// @Param        maxDepth  query     int     false  "最大深度，默认为3"  minimum(1)
// @Success      200       {object}  base.Response{data=services.FileNode}
// @Failure      400       {object}  base.Response
// @Failure      404       {object}  base.Response
// @Failure      500       {object}  base.Response
// @Router       /projects/{id}/files [get]
func (h *Handler) GetProjectFiles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid project ID")
		return
	}

	// Get max depth from query param, default to 3
	maxDepthStr := c.DefaultQuery("maxDepth", "3")
	maxDepth, err := strconv.Atoi(maxDepthStr)
	if err != nil || maxDepth < 1 {
		maxDepth = 3
	}

	// Get the project
	project, err := h.database.GetProject(id)
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, base.ErrCodeNotFound, "Project not found")
		return
	}

	// Get the file tree
	fileTree, err := h.fileService.GetProjectFiles(project.Path, maxDepth)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to get project files: %v", err))
		return
	}

	base.SuccessResponse(c, fileTree)
}
