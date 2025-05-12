package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
	"mind-weaver/internal/db"
	"mind-weaver/internal/services"
	"mind-weaver/internal/utils"
	"mind-weaver/pkg/logger"
)

// Session handlers
type CreateSessionReq struct {
	ProjectID       int64         `json:"project_id" binding:"required"`
	Name            string        `json:"name" binding:"required"`
	Mode            string        `json:"mode" binding:"required,oneof=auto manual single-html product-design all"`
	ExcludePatterns []db.FileInfo `json:"exclude_patterns,omitempty"`
	IncludePatterns []db.FileInfo `json:"include_patterns,omitempty"`
}

// CreateSession 创建会话
// @Summary      创建新会话
// @Description  为指定项目创建新的代码会话
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        request  body  CreateSessionReq  true  "会话创建参数"
// @Success      200  {object}  base.Response{data=services.SessionInfo}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions [post]
func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateSessionReq

	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// Create the session
	session, err := h.sessionService.CreateSession(req.ProjectID, req.Name, req.Mode, req.ExcludePatterns, req.IncludePatterns)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to create session: %v", err))
		return
	}

	base.SuccessResponse(c, session)
}

// GetSessions 获取项目会话列表
// @Summary      获取项目所有会话
// @Description  获取指定项目的所有会话列表，按更新时间降序排列
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        projectId  path      int  true  "项目ID"
// @Success      200  {object}  base.Response{data=[]services.SessionInfo}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/project/{projectId} [get]
func (h *Handler) GetSessions(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid project ID")
		return
	}

	sessions, err := h.database.ListProjectSessions(projectID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to list sessions: %v", err))
		return
	}

	base.SuccessResponse(c, sessions)
}

// GetSession 获取会话详情
// @Summary      获取会话详情
// @Description  获取指定会话的详细信息，包括消息历史
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "会话ID"
// @Success      200  {object}  base.Response{data=services.SessionInfo}
// @Failure      400  {object}  base.Response
// @Failure      404  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}

	session, err := h.sessionService.GetSession(id)
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, base.ErrCodeNotFound, "Session not found")
		return
	}

	base.SuccessResponse(c, session)
}

type SendMessageReq struct {
	Content      string   `json:"content" binding:"required"`
	ProjectPath  string   `json:"project_path" binding:"required"`
	ContextFiles []string `json:"context_files"`
}

type SendMessageResp struct {
	UserMessage string `json:"user_message"`
	AiMessage   string `json:"ai_message"`
}

// SendMessage 发送消息
// @Summary      发送消息并获取AI响应
// @Description  向会话发送消息并同步获取AI响应
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "会话ID"
// @Param        request  body  SendMessageReq  true  "消息内容"
// @Success      200  {object}  base.Response{data=SendMessageResp}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id}/message [post]
func (h *Handler) SendMessage(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}

	var req SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// Add user message
	userMsg, err := h.sessionService.AddUserMessage(sessionID, req.Content)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to add user message: %v", err))
		return
	}

	// Generate AI response
	aiMsg, err := h.sessionService.GenerateAIResponse(sessionID, req.ProjectPath, req.ContextFiles, req.Content)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to generate AI response: %v", err))
		return
	}

	base.SuccessResponse(c, SendMessageResp{
		UserMessage: userMsg.Content,
		AiMessage:   aiMsg.Content,
	})
}

type UpdateContextResp struct {
	Status string `json:"status"`
}

// UpdateContext 更新会话上下文
// @Summary      更新会话上下文
// @Description  更新会话的上下文信息(当前文件、光标位置等)
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "会话ID"
// @Param        request  body      services.ContextInfo  true  "上下文信息"
// @Success      200  {object}  base.Response{data=api.UpdateContextResp}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id}/context [put]
func (h *Handler) UpdateContext(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}

	var contextInfo services.ContextInfo
	if err := c.ShouldBindJSON(&contextInfo); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// 使用防抖动机制处理频繁的上下文更新
	updateNow := processContextUpdate(sessionID, &contextInfo)

	if !updateNow {
		// 如果不需要立即更新，返回成功但不执行数据库操作
		base.SuccessResponse(c, UpdateContextResp{
			Status: "queued",
		})
		return
	}

	// 执行实际的上下文更新
	err = h.sessionService.UpdateSessionContext(sessionID, contextInfo)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to update context: %v", err))
		return
	}

	base.SuccessResponse(c, UpdateContextResp{
		Status: "ok",
	})
}

// GetContext 获取会话上下文
// @Summary      获取会话上下文
// @Description  获取会话的上下文信息
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "会话ID"
// @Success      200  {object}  base.Response{data=services.ContextInfo}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id}/context [get]
func (h *Handler) GetContext(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}

	contextInfo, err := h.sessionService.GetSessionContext(sessionID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to get context: %v", err))
		return
	}

	base.SuccessResponse(c, contextInfo)
}

type UpdateSessionReq struct {
	Name            string        `json:"name,omitempty"`
	Mode            string        `json:"mode,omitempty"`
	ExcludePatterns []db.FileInfo `json:"exclude_patterns,omitempty"`
	IncludePatterns []db.FileInfo `json:"include_patterns,omitempty"`
}

// UpdateSession 更新会话
// @Summary      更新会话信息
// @Description  更新会话的名称、模式或文件模式
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "会话ID"
// @Param        request  body  UpdateSessionReq  true  "会话更新参数"
// @Success      200  {object}  base.Response{data=services.SessionInfo}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id} [put]
func (h *Handler) UpdateSession(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}

	var req UpdateSessionReq

	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	// Update the session
	session, err := h.sessionService.UpdateSession(sessionID, req.Name, req.Mode, req.ExcludePatterns, req.IncludePatterns)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to update session: %v", err))
		return
	}

	base.SuccessResponse(c, session)
}

// DeleteSession 删除会话
// @Summary      删除会话
// @Description  删除指定会话及其所有相关数据
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "会话ID"
// @Success      200  {object}  base.Response{data=object{status=string}}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id} [delete]
func (h *Handler) DeleteSession(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}

	err = h.sessionService.DeleteSession(sessionID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to delete session: %v", err))
		return
	}

	base.SuccessResponse(c, gin.H{"status": "ok"})
}

// DeleteSession 删除消息
// @Summary      删除消息
// @Description  删除指定消息及其所有相关数据
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "会话ID"
// @Param        msgId  path      int  true  "消息ID，0表示删除会话下所有消息"
// @Success      200  {object}  base.Response{data=object{status=string}}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id}/messages/{msgId} [delete]
func (h *Handler) DeleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid session ID")
		return
	}
	msgIdStr := c.Param("msgId")
	msgId, err := strconv.ParseInt(msgIdStr, 10, 64)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid message ID")
		return
	}

	err = h.sessionService.DeleteMessage(sessionID, msgId)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to delete session: %v", err))
		return
	}

	base.SuccessResponse(c, gin.H{"status": "ok"})
}

type PrepareStreamMessageReq struct {
	Type         string   `json:"type" binding:"required"`
	Content      string   `json:"content"`
	ProjectPath  string   `json:"project_path" binding:"required"`
	ContextFiles []string `json:"context_files"`
	Model        string   `json:"model"`
}

type PrepareStreamMessageResp struct {
	MessageID string    `json:"message_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// 构建系统提示（通用会话）
func (h *Handler) buildSystemPrompt(modelName string, sessionID int64, addSysPrompt bool) string {
	codePrompt := ""
	currentModel := h.cfg.LLM.GetCurrentLLMInfo(modelName)
	codeContextBuilder := utils.NewPromptBuilder(currentModel.MaxContext)
	sessionInfo, _ := h.sessionService.GetSession(sessionID)
	var files []string
	for _, fileContext := range sessionInfo.IncludePatterns {
		if fileContext.IsDir {
			// 递归读取文件夹，使用默认代码文件过滤器
			dirFiles, err := utils.GetFilesInDirectory(fileContext.Path, utils.DefaultCodeFilter(), 0)
			if err != nil {
				logger.Infof("Failed to read directory %s: %v", fileContext.Path, err)
				continue
			}
			files = append(files, dirFiles...)
		} else {
			files = append(files, fileContext.Path)
		}
	}

	// 过滤重复的文件
	hasAdd := map[string]bool{}
	for _, filePathStr := range files {
		if _, ok := hasAdd[filePathStr]; ok {
			// 跳过已经添加过的文件
			continue
		}
		codeContextBuilder.AddCodeFile(filePathStr)
		hasAdd[filePathStr] = true
	}
	codePrompt = codeContextBuilder.BuildSystemPrompt(addSysPrompt)

	return codePrompt
}
