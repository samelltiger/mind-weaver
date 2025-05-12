package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
	"mind-weaver/internal/services"
	"mind-weaver/internal/third/assistantmessage"
	thirdPrompts "mind-weaver/internal/third/prompts"
	"mind-weaver/internal/third/prompts/sections"
	"mind-weaver/internal/third/tools"
	"mind-weaver/internal/utils"
	"mind-weaver/pkg/logger"
	"mind-weaver/pkg/prompts"
	"mind-weaver/pkg/util"
)

// OpenAICompatRequest matches OpenAI API request format
type OpenAICompatRequest struct {
	Model string `json:"model"`
	// Messages    []ChatMessage `json:"messages"`
	Stream      bool    `json:"stream"`
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`

	// Additional OpenAI parameters can be added as needed
	// ProjectPath  string   `json:"project_path"`
	ContextFiles []string                    `json:"context_files,omitempty"`
	SessionID    int64                       `json:"session_id" binding:"required"`
	Type         string                      `json:"type" binding:"required"`
	ProjectPath  string                      `json:"project_path" binding:"required"`
	Content      string                      `json:"content,omitempty"`
	ToolUse      assistantmessage.ToolUseReq `json:"tool_use,omitempty"`
}

// ChatMessage matches OpenAI message format
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAICompatResponse matches OpenAI API response format
type OpenAICompatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageInfo `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	FinishReason string       `json:"finish_reason"`
	Delta        *ChatMessage `json:"delta,omitempty"` // Used for streaming
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAICompatStreamHandler handles OpenAI-compatible streaming completions
// @Summary      OpenAI-compatible streaming completions
// @Description  Provides an OpenAI-compatible streaming API endpoint for chat completions
// @Tags         session
// @Accept       json
// @Produce      text/event-stream
// @Param        request body api.OpenAICompatRequest true "Completion request"
// @Success      200  {object}  object  "SSE stream of completion chunks"
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/{id}/completions [post]
func (h *Handler) OpenAICompatStreamHandler(c *gin.Context) {
	logger.Info("Starting OpenAI-compatible stream handler")

	var req OpenAICompatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid request body, bind error")
		return
	}

	if req.Model == "" {
		req.Model = "qt-claude37" // Default model
	}

	if req.Type == "" {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Invalid request body, type is required")
		return
	}

	var err error
	if (req.Type == "normal" && req.Content == "") || req.ProjectPath == "" {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, "Missing required parameters, type is required, project_path is required")
		return
	}

	logger.Infof("Content: %s\nProject Path: %s\nContext Files: %v\n", req.Content, req.ProjectPath, req.ContextFiles)

	/**
	1. 根据会话的不同模式做不同的处理
	2. 根据 typeStr 做不同的处理
	*/

	// 获取用户历史消息
	historyMessages := h.aiService.GetUserHistoryMsgs(req.SessionID)
	// 因为当前这个消息就是用户发送的，因为如果最后一条也是user发的，那么需要去掉
	if len(historyMessages) > 0 && historyMessages[len(historyMessages)-1].Role == services.MsgTypeUser {
		historyMessages = historyMessages[:len(historyMessages)-1]
	}

	// 组装systemPrompt
	systemtPrompt := ""
	userMsg := &services.MessageInfo{}

	sessionInfo, err := h.sessionService.GetSession(req.SessionID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to get session: %v", err))
		return
	}

	if req.Type != "retry" {
		switch sessionInfo.Mode {
		case services.SessionModeManual:
			systemtPrompt, userMsg, err = h.switchManualType(c, req, historyMessages)

		case services.SessionModeAuto:
			if req.Type == "tool_use" {
				systemtPrompt, userMsg, err = h.AutoToolUse(c, req, historyMessages)
			} else {
				systemtPrompt, userMsg, err = h.switchAutoType(c, req, historyMessages)
			}

		case services.SessionModeSingleHtml:
			systemtPrompt, userMsg, err = h.switchSingleHtmlSystemPrompt(c, req, historyMessages)
		}
	} else {
		// 如果是重试，那么需要对历史消息进行重新组装
		systemtPrompt, userMsg, historyMessages, err = h.switchRetry(c, req, historyMessages)
	}
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to generate prompt: %v", err))
		return
	}

	logger.Infof("\n\n\nUser Message ID: %d, model name: %v, Session Mode: %v， type: %v\n\n\n", userMsg.ID, req.Model, sessionInfo.Mode, req.Type)

	// Handle streaming response
	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Transfer-Encoding", "chunked")

		clientGone := c.Writer.CloseNotify()
		doneCh := make(chan bool)

		go func() {
			defer close(doneCh)

			switch sessionInfo.Mode {
			case services.SessionModeManual:
				h.streamManual(c, req, historyMessages, systemtPrompt, userMsg)
			case services.SessionModeAuto,
				services.SessionModeSingleHtml:
				h.streamByLine(c, req, historyMessages, systemtPrompt, userMsg, sessionInfo)
			}
		}()

		select {
		case <-clientGone:
			return
		case <-doneCh:
			return
		}
	} else {
		resContent, err := h.aiService.Chat(systemtPrompt, req.Content, historyMessages, req.Model)
		if err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError,
				fmt.Sprintf("Failed to get AI response: %v", err))
			return
		}

		// Save the response to the database
		h.database.AddMessage(req.SessionID, services.MsgTypeAssistant, resContent)

		// Return OpenAI compatible response
		response := OpenAICompatResponse{
			ID:      fmt.Sprintf("chatcmpl-%d", req.SessionID),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []Choice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: resContent,
					},
					FinishReason: "stop",
				},
			},
			Usage: UsageInfo{
				PromptTokens:     0, // We're not tracking tokens
				CompletionTokens: 0,
				TotalTokens:      0,
			},
		}

		c.JSON(http.StatusOK, response)
	}
}

func (h *Handler) switchManualType(c *gin.Context, req OpenAICompatRequest, historyMessages []*services.Message) (string, *services.MessageInfo, error) {
	var err error
	systemtPrompt := ""
	userMsg := &services.MessageInfo{}
	switch req.Type {
	case "normal":
		// 如果消息长度是0，那么会把当前session携带的代码文件读取到prompt中
		if len(historyMessages) == 0 {
			systemtPrompt = h.buildSystemPrompt(req.Model, req.SessionID, true)

			// 组装完系统prompt之后，将其添加到历史消息中
			h.sessionService.AddSysMessage(req.SessionID, systemtPrompt)
		}
		logger.Infof("History Messages: %v\n", systemtPrompt)

		// 添加用户消息
		userMsg, err = h.sessionService.AddUserMessage(req.SessionID, req.Content)
		if err != nil {
			return systemtPrompt, userMsg, err
		}
	case "explain":
		// 拼接上下文文件，生成prompt
		// 解释代码不携带上下文文件
		codeFileContent := h.buildSystemPrompt(req.Model, req.SessionID, false)
		sysPrompt, err := prompts.GetPrompt("code_analysis")
		if err != nil {
			logger.Errorf("Failed to get prompt: %v", err)
			return systemtPrompt, userMsg, err
		}
		systemtPrompt = sysPrompt

		// 组装完系统prompt之后，将其添加到历史消息中
		h.sessionService.AddSysMessage(req.SessionID, systemtPrompt)
		// 添加用户消息
		userMsg, err = h.sessionService.AddUserMessage(req.SessionID, codeFileContent)
		if err != nil {
			return systemtPrompt, userMsg, err
		}
	}

	return systemtPrompt, userMsg, nil
}

func (h *Handler) switchAutoType(c *gin.Context, req OpenAICompatRequest, historyMessages []*services.Message) (string, *services.MessageInfo, error) {
	var err error
	systemtPrompt := ""
	userMsg := &services.MessageInfo{}

	if len(historyMessages) == 0 {
		// Build args for BuildSystemPrompt
		args := thirdPrompts.BuildSystemPromptArgs{
			EnvCtx: thirdPrompts.EnvironmentContext{
				Cwd:                 req.ProjectPath,
				SupportsComputerUse: false,
				BrowserViewportSize: "",
				Language:            "zh-cn",
			},
			Mode:               sections.ModeSlug("code"),
			CustomModeConfigs:  nil,
			GlobalInstructions: "",
			// Note: DiffStrategy and RooIgnoreController are nil here
		}

		// Generate the system prompt
		systemtPrompt, err = thirdPrompts.BuildSystemPrompt(args)
		if err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to add system message: %v", err))
			return systemtPrompt, userMsg, err
		}

		// 组装完系统prompt之后，将其添加到历史消息中
		h.sessionService.AddSysMessage(req.SessionID, systemtPrompt)
	}

	// 添加用户消息
	userMsg, err = h.sessionService.AddUserMessage(req.SessionID, req.Content)
	if err != nil {
		return systemtPrompt, userMsg, err
	}

	return systemtPrompt, userMsg, nil
}

func (h *Handler) switchSingleHtmlSystemPrompt(c *gin.Context, req OpenAICompatRequest, historyMessages []*services.Message) (string, *services.MessageInfo, error) {
	var err error
	systemtPrompt := ""
	userMsg := &services.MessageInfo{}

	if len(historyMessages) == 0 {
		// Generate the system prompt
		systemtPrompt, err = prompts.GetPrompt("single_html_mode_system")
		if err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, base.ErrCodeInternalError, fmt.Sprintf("Failed to add system message: %v", err))
			return systemtPrompt, userMsg, err
		}

		// 组装完系统prompt之后，将其添加到历史消息中
		h.sessionService.AddSysMessage(req.SessionID, systemtPrompt)
	}

	// 添加用户消息
	userMsg, err = h.sessionService.AddUserMessage(req.SessionID, req.Content)
	if err != nil {
		return systemtPrompt, userMsg, err
	}

	return systemtPrompt, userMsg, nil
}

func (h *Handler) AutoToolUse(c *gin.Context, req OpenAICompatRequest, historyMessages []*services.Message) (string, *services.MessageInfo, error) {
	var err error
	systemtPrompt := ""
	userMsg := &services.MessageInfo{}

	// 执行工具
	executeParams := tools.ExecutorInput{
		ToolUse:             req.ToolUse.ToolUse,
		Ctx:                 c.Request.Context(),
		Cwd:                 req.ProjectPath,
		DiffStrategy:        nil,
		RooIgnoreController: nil,
		Confirmed:           req.ToolUse.Confirmed,
	}

	var executeRes *tools.ExecutorResult
	// 检查用户是否同意使用工具
	if !req.ToolUse.Confirmed {
		errText := fmt.Sprintf("Tool '%s' not approved by user.", req.ToolUse.ToolUse.Name)
		executeRes = &tools.ExecutorResult{Result: thirdPrompts.FormatToolError(errText), IsError: true}
	} else {
		executeParams.RooIgnoreController = nil
		executeRes, err = tools.ExecuteTool(executeParams)
	}
	if err != nil {
		userMsg, err = h.sessionService.AddUserMessage(req.SessionID, err.Error())
		if err != nil {
			return systemtPrompt, userMsg, err
		}
	} else {
		// 添加用户消息
		content := fmt.Sprintf("%s。一次回复只能使用一个工具", executeRes.Result)
		userMsg, err = h.sessionService.AddUserMessage(req.SessionID, content)
		if err != nil {
			return systemtPrompt, userMsg, err
		}
	}

	return systemtPrompt, userMsg, nil
}

func (h *Handler) switchRetry(c *gin.Context, req OpenAICompatRequest, historyMessages []*services.Message) (string, *services.MessageInfo, []*services.Message, error) {
	systemPrompt := ""
	userMsg := &services.MessageInfo{}

	// 标准化消息序列
	normalizedMessages := normalizeMessageSequence(historyMessages)

	// 处理重试类型的请求
	if len(normalizedMessages) > 0 {
		lastMessage := normalizedMessages[len(normalizedMessages)-1]

		// 情况1: 最后一条消息是assistant，表示上次已生成完毕，需删除并重新生成
		if lastMessage.Role == "assistant" {
			// 从数据库中删除最后一条assistant消息
			err := h.sessionService.DeleteMessage(req.SessionID, lastMessage.Id)
			if err != nil {
				return "", nil, normalizedMessages, fmt.Errorf("删除assistant消息失败: %v", err)
			}

			// 从历史消息数组中删除最后一条assistant消息
			normalizedMessages = normalizedMessages[:len(normalizedMessages)-1]
		}

		// 找到最后一条user消息，并保存到userMsg中
		for i := len(normalizedMessages) - 1; i >= 0; i-- {
			if normalizedMessages[i].Role == "user" {
				userMsg = &services.MessageInfo{
					ID:        normalizedMessages[i].Id,
					Role:      normalizedMessages[i].Role,
					Content:   normalizedMessages[i].Content,
					Timestamp: time.Now(),
				}

				// 从数组中删除最后一条user消息（不从数据库删除）
				normalizedMessages = append(normalizedMessages[:i], normalizedMessages[i+1:]...)
				break
			}
		}
	}

	// 如果没有找到user消息，返回错误
	if userMsg.ID == 0 {
		return "", nil, normalizedMessages, fmt.Errorf("未找到用户消息")
	}

	return systemPrompt, userMsg, normalizedMessages, nil
}

func normalizeMessageSequence(messages []*services.Message) []*services.Message {
	if len(messages) == 0 {
		return messages
	}

	normalized := make([]*services.Message, 0, len(messages))

	// 处理第一条消息
	if messages[0].Role == "system" {
		// 保留第一条system消息
		normalized = append(normalized, messages[0])

		// 移除后续所有system消息
		for i := 1; i < len(messages); i++ {
			if messages[i].Role != "system" {
				normalized = append(normalized, messages[i])
			}
		}
	} else if messages[0].Role == "user" {
		// 如果第一条是user，直接开始
		normalized = append(normalized, messages[0])
	} else {
		// 如果第一条既不是system也不是user，强制转换为user
		firstMsg := *messages[0]
		firstMsg.Role = "user"
		normalized = append(normalized, &firstMsg)
	}

	// 处理后续消息，确保user和assistant交替出现
	result := make([]*services.Message, 0, len(normalized))
	result = append(result, normalized[0])

	for i := 1; i < len(normalized); i++ {
		current := normalized[i]
		last := result[len(result)-1]

		// 跳过与上一条相同role的消息
		if current.Role == last.Role {
			continue
		}

		// 确保user和assistant交替出现
		if last.Role == "user" && current.Role != "assistant" {
			// 如果不是assistant，插入一个虚拟的assistant消息
			result = append(result, &services.Message{
				Role:    "assistant",
				Content: "[系统自动补充的过渡消息]",
			})
		} else if last.Role == "assistant" && current.Role != "user" {
			// 如果不是user，插入一个虚拟的user消息
			result = append(result, &services.Message{
				Role:    "user",
				Content: "[系统自动补充的过渡消息]",
			})
		}

		result = append(result, current)
	}

	return result
}

func (h *Handler) streamManual(c *gin.Context, req OpenAICompatRequest, historyMessages []*services.Message, systemtPrompt string, userMsg *services.MessageInfo) {
	// 创建缓冲区以收集完整响应
	var responseBuffer strings.Builder

	// 自定义写入器，直接发送数据到客户端
	writer := &utils.SseWriter{
		ResponseWriter: c.Writer,
		Buffer:         &responseBuffer,
		UserMsgId:      userMsg.ID,
	}

	// 生成流式响应
	err := h.aiService.ChatStream(systemtPrompt, userMsg.Content, historyMessages, req.Model, writer)
	if err != nil {
		// 如果出现错误，我们仍要保存已获得的内容
		fmt.Fprintf(c.Writer, "data: {\"error\":\"%v\"}\n\n", err)
	}

	// 将完整响应保存到数据库
	aiRes := responseBuffer.String()
	if responseBuffer.Len() > 0 {
		msgId, err := h.database.AddMessage(req.SessionID, services.MsgTypeAssistant, aiRes)
		fmt.Printf("msgId: %v\n", msgId)
		if err != nil {
			logger.Errorf("Failed to add message to database: %v", err)
		} else {
			// 如果消息ID不为0，则将消息ID保存到预备消息中
			if msgId != 0 {
				writer.MsgId = msgId
				writer.Write([]byte("\n\n"))
			}
		}
	}

	// 发送完成事件
	fmt.Fprintf(c.Writer, "data: [DONE]\ndata: {\"status\":\"complete\"}\n\n")
	c.Writer.Flush()
}

func (h *Handler) streamByLine(c *gin.Context, req OpenAICompatRequest, historyMessages []*services.Message, systemtPrompt string, userMsg *services.MessageInfo, sessionInfo *services.SessionInfo) {
	var err error
	// 创建缓冲区以收集完整响应
	var responseBuffer strings.Builder

	// 自定义写入器，按行发送数据到客户端
	writer := &services.SseLineWriter{
		ResponseWriter: c.Writer,
		Buffer:         &responseBuffer,
		Mode:           sessionInfo.Mode,
		Cwd:            req.ProjectPath,
		UserMsgId:      userMsg.ID,
	}

	switch sessionInfo.Mode {
	case services.SessionModeAuto:
		err = h.aiService.ChatStreamByLine(systemtPrompt, userMsg.Content, historyMessages, req.Model, writer, sessionInfo.Mode, false)
	case services.SessionModeSingleHtml:
		err = h.streamModeSingleHtml(c, systemtPrompt, userMsg.Content, historyMessages, &req, writer, &responseBuffer, sessionInfo.Mode)
	}
	// 生成按行流式响应
	if err != nil {
		// 如果出现错误，我们仍要保存已获得的内容
		fmt.Fprintf(c.Writer, "data: {\"error\":\"%v\"}\n\n", err)
	}

	// 将完整响应保存到数据库
	aiRes := responseBuffer.String()
	if responseBuffer.Len() > 0 {
		msgId, err := h.database.AddMessage(req.SessionID, services.MsgTypeAssistant, aiRes)
		if err != nil {
			logger.Errorf("Failed to add message to database: %v", err)
		} else {
			// 如果消息ID不为0，则将消息ID保存到预备消息中
			if msgId != 0 {
				writer.MsgId = msgId
				writer.Write([]byte("\n\n"))
			}
		}
	}

	// 发送结束标记
	if _, err := writer.Write([]byte(services.StreamMsgEndTag)); err != nil {
		logger.Infof("Write error: %v", err)
	}

	// 发送完成事件
	fmt.Fprintf(c.Writer, "data: [DONE]\ndata: {\"status\":\"complete\"}\n\n")
	c.Writer.Flush()
}

func (h *Handler) streamModeSingleHtml(c *gin.Context, sysPrompt string, prompt string, historyMsgs []*services.Message, req *OpenAICompatRequest, writer io.Writer, responseBuffer *strings.Builder, mode string) error {
	var err error
	isFinish := false
	limit := 5
	path := ""
	isContinue := false
	aiResList := []string{}

	for !isFinish {
		err = h.aiService.ChatStreamByLine(sysPrompt, prompt, historyMsgs, req.Model, writer, mode, isContinue)
		if err != nil {
			// 如果出现错误，我们仍要保存已获得的内容
			fmt.Fprintf(c.Writer, "data: {\"error\":\"%v\"}\n\n", err)
			break
		}
		aiRes := responseBuffer.String()
		parsed := assistantmessage.RemoveTextNode(assistantmessage.ParseAssistantMessage(aiRes))
		for _, content := range parsed {
			switch c := content.(type) {
			case *assistantmessage.ToolUse:
				// 用于判断代码是否编写完成，特别是非常长的代码逻辑
				if c.Name == assistantmessage.WriteToFile {
					isFinish = !c.Partial
				}
				if fpath, ok := c.Params["path"]; ok {
					path = fpath
				}
			}
		}

		// 本次ai新增回复
		savedAiRes := strings.Join(aiResList, "")
		aiNewAddRes := aiRes
		if len(aiResList) != 0 {
			aiNewAddRes = strings.Replace(aiRes, savedAiRes, "", 1)
		}
		aiResList = append(aiResList, aiNewAddRes)

		// 根据是否编写完成，决定是否要再次请求
		if isFinish {
			absolutePath := fmt.Sprintf("aiNewAddRes-%v.txt", "last")
			os.WriteFile(absolutePath, []byte(aiNewAddRes), 0644)
			break
		}
		isContinue = true

		// 限制最长次数，避免无限循环
		limit -= 1
		if limit <= 0 {
			fmt.Fprintf(c.Writer, "data: {\"error\":\"生成长度达到最长限制，当前限制最长询问次数：limit=%v\"}\n\n", limit)
			break
		}

		if sysPrompt != "" && len(historyMsgs) == 0 {
			// 添加系统提示词到历史记录
			historyMsgs = append(historyMsgs, &services.Message{
				Role:    services.MsgTypeSystem,
				Content: sysPrompt,
			})
			sysPrompt = ""
		}

		// 添加用户提示词到历史记录
		historyMsgs = append(historyMsgs, &services.Message{
			Role:    services.MsgTypeUser,
			Content: prompt,
		})

		// 处理再次询问大模型的参数逻辑
		prompt = fmt.Sprintf("代码内容已保存到`%s`文件中。请继续。", path) +
			`上面回复的代码已实时保存到代码文件中，所以在你的接下来的回复中，你需要直接从上一次结束的位置开始继续编写代码，确保续写的内容能100%正确衔接上一次结束时的文本内容，从而确保不会出错。如果你想说明其他信息，你必须以注释的方式写在代码中。如果你认为你已经将该html的所有代码都编写完成了，那么一定要在最后结束的位置添加write_to_file的闭合标签。`
		// 将大模型返回的代码和用户内容添加到历史消息中
		historyMsgs = append(historyMsgs, &services.Message{
			Role:    services.MsgTypeAssistant,
			Content: aiNewAddRes,
		}) // 添加ai回复到历史记录中

		// 输出到文件
		historyMsgsJs, _ := json.Marshal(historyMsgs)
		absolutePath := fmt.Sprintf("historyMsgs-%v.txt", limit)
		os.WriteFile(absolutePath, historyMsgsJs, 0644)

		absolutePath = fmt.Sprintf("userMessage-%v.txt", limit)
		os.WriteFile(absolutePath, []byte(prompt), 0644)

		absolutePath = fmt.Sprintf("aiNewAddRes-%v.txt", limit)
		os.WriteFile(absolutePath, []byte(aiNewAddRes), 0644)
	}

	// 重新拼接文件内容，使用大模型接口拼接位置的冲突
	// 解决冲突的方式以及顺序：
	// 1. 将生成的代码文件使用无头浏览器打开，确认控制台是否报错
	// 2. 如果有报错，那么就调用大模型解决拼接冲突
	// 3. 再次验证文件，如果还是报错，那么将文件发给模型进行排查
	if len(aiResList) > 1 && h.cfg.Bin.Python != "" {
		fullpath := req.ProjectPath + "/" + path
		h.handleLlmResponseError(aiResList, fullpath)
	}

	return nil
}

func (h *Handler) handleLlmResponseError(aiResList []string, fullpath string) error {
	_, outputs, err := h.commandService.JsInspector(fullpath)
	if err != nil {
		logger.Errorf("handleLlmResponseError h.commandService.JsInspector return err: %v", err)
		return err
	}

	// 检查是否有错误
	hasError := true
	for _, output := range outputs {
		if output.Line == "没有发现JavaScript错误" && !output.IsError {
			hasError = false
			break
		}
	}

	if !hasError {
		logger.Infof("handleLlmResponseError hasError == false")
		// 没有错误，不进行任何处理
		return nil
	}

	// 读取原始文件内容
	originalContent, err := os.ReadFile(fullpath)
	if err != nil {
		logger.Infof("handleLlmResponseError 读取文件失败: %v", err)
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 这里表示有错误，需要处理拼接冲突
	type diffInfo struct {
		prevIndex    int    // 前一个内容的索引
		currentIndex int    // 当前内容的索引
		prevContent  string // 前一个内容的末尾部分
		currContent  string // 当前内容的开头部分
		combined     string // 组合后的内容（需要被替换的部分）
	}

	diffInfos := []diffInfo{}

	// 获取前一个内容的最后几行时，应从后往前找，如果发现JavaScript的函数定义的开始，那么应该停止，就从这个函数定义处截断
	for i := 1; i < len(aiResList); i++ {
		prevContent := aiResList[i-1]
		newContent := aiResList[i]

		reserveLine := h.cfg.DiffLine // 默认保留的行数

		// 获取前一个内容的最后几行
		prevLines := strings.Split(prevContent, "\n")
		lastLines := []string{}
		if len(prevLines) <= reserveLine {
			lastLines = prevLines
		} else {
			lastLines = prevLines[len(prevLines)-reserveLine:]
		}
		lastPart := strings.Join(lastLines, "\n")

		// 获取新内容的前几行
		newLines := strings.Split(newContent, "\n")
		firstLines := []string{}
		if len(newLines) <= reserveLine {
			firstLines = newLines
		} else {
			firstLines = newLines[:reserveLine]
		}
		firstPart := strings.Join(firstLines, "\n")

		// 组合需要替换的内容
		// combinedContent := lastPart + "\n" + firstPart
		combinedContent := lastPart + firstPart

		diffInfos = append(diffInfos, diffInfo{
			prevIndex:    i - 1,
			currentIndex: i,
			prevContent:  lastPart,
			currContent:  firstPart,
			combined:     combinedContent,
		})
	}

	// 准备修复后的内容
	fixedContent := string(originalContent)

	// 然后对diffInfos进行遍历，依次发送大模型请求
	for _, di := range diffInfos {
		prompt, err := prompts.GetPromptWithParams("repaire_code_bug", map[string]string{
			"old_content": di.prevContent,
			"new_content": di.currContent,
			// "error_info":  outputsToString(outputs),
		})
		if err != nil {
			return fmt.Errorf("获取提示模板失败: %w", err)
		}

		logger.Infof("prompt: %v, model: %v", prompt, h.cfg.DiffModel)
		response, err := h.aiService.Chat("", prompt, nil, h.cfg.DiffModel)
		if err != nil {
			logger.Infof("handleLlmResponseError AI服务调用失败: %v", err)
			return fmt.Errorf("AI服务调用失败: %w", err)
		}
		logger.Infof("llm response: %v", response)

		// 匹配处理里面的 ``` 中的代码
		fixedCode := utils.ExtractCodeFromMarkdown(response)
		if fixedCode != "" {
			// 更新代码内容，用修复后的代码替换掉组合内容
			// fixedContent = strings.Replace(fixedContent, di.combined, fixedCode, 1)
			// 添加代码差异比较功能
			os.WriteFile("di.combined.txt", []byte(di.combined), 0644)
			os.WriteFile("fixedCode.txt", []byte(fixedCode), 0644)
			diffCode, err := h.commandService.CodeDiff(di.combined, fixedCode)
			if err != nil {
				logger.Infof("handleLlmResponseError call h.commandService.CodeDiff err: %v", err.Error())
				// 没有错误，不进行任何处理
				return nil
			}
			os.WriteFile("diffCode.txt", []byte(diffCode), 0644)
			fixedContent = strings.Replace(fixedContent, di.combined, util.ProcessDiffText(diffCode), 1)
		}
	}

	// 用最新的代码写入到代码文件中
	err = os.WriteFile(fullpath, []byte(fixedContent), 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// 辅助函数：将输出转换为字符串
func outputsToString(outputs []util.CommandOutput) string {
	var result strings.Builder
	for _, output := range outputs {
		if output.IsError {
			result.WriteString("错误: ")
		}
		result.WriteString(output.Line)
		result.WriteString("\n")
	}
	return result.String()
}

type ParseAiContentReq struct {
	Content string `json:"content"`
}

// ParseAiContent 解析ai响应文本
// @Summary      解析ai响应文本
// @Description  解析ai响应文本，返回解析完的数据内容
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        request  body  ParseAiContentReq  true  "需要解析的内容"
// @Success      200  {object}  base.Response{data=assistantmessage.AssistantMessageContent}
// @Failure      400  {object}  base.Response
// @Failure      500  {object}  base.Response
// @Router       /sessions/parse/ai-res [post]
func (h *Handler) ParseAiContent(c *gin.Context) {
	var req ParseAiContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, base.ErrCodeInvalidParams, err.Error())
		return
	}

	parsedInfo := assistantmessage.ParseAssistantMessage(req.Content)

	base.SuccessResponse(c, parsedInfo)
}
