package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mind-weaver/config"
	"mind-weaver/internal/db"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/third/tools"
	"mind-weaver/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	MsgTypeUser      = "user"
	MsgTypeAssistant = "assistant"
	MsgTypeSystem    = "system"

	// 流式返回结束标志，这个标记用于在结束时做相关处理
	StreamMsgEndTag = "<__STREAM_END__>"
)

type AIService struct {
	apiKey    string
	model     string
	maxTokens int
	cfg       config.Config
	database  *db.Database
}

type Message struct {
	Id      int64  `json:"-"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func NewAIService(database *db.Database, cfg *config.Config) *AIService {
	return &AIService{
		database:  database,
		apiKey:    cfg.LLM.APIKey,
		model:     cfg.LLM.Model,
		maxTokens: cfg.LLM.MaxTokens,
		cfg:       *cfg,
	}
}

func (s *AIService) GenerateCompletion(prompt string, contextFiles []*FileContext) (string, error) {
	// Prepare system message with context
	systemMsg := "You are MindWeaver AI, an intelligent coding assistant. "
	systemMsg += "Help the user with their coding tasks based on the following context:\n\n"

	// Add file contexts
	for _, file := range contextFiles {
		systemMsg += fmt.Sprintf("File: %s (%s)\n```%s\n%s\n```\n\n",
			file.FilePath, file.Language, file.Language, file.Content)
	}

	systemMsg += "Provide concise, working code solutions. Explain your approach briefly if needed."

	// Prepare request
	messages := []Message{
		{Role: MsgTypeSystem, Content: systemMsg},
		{Role: MsgTypeUser, Content: prompt},
	}

	reqBody := CompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   s.maxTokens,
		Stream:      false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Make request to OpenAI
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no completion choices returned")
	}

	return result.Choices[0].Message.Content, nil
}

func (s *AIService) GenerateCompletionStream(prompt string, contextFiles []*FileContext, writer io.Writer) error {
	// Prepare system message with context
	systemMsg := "You are MindWeaver AI, an intelligent coding assistant. "
	systemMsg += "Help the user with their coding tasks based on the following context:\n\n"

	// Add file contexts
	for _, file := range contextFiles {
		systemMsg += fmt.Sprintf("File: %s (%s)\n```%s\n%s\n```\n\n",
			file.FilePath, file.Language, file.Language, file.Content)
	}

	systemMsg += "Provide concise, working code solutions. Explain your approach briefly if needed."

	// Prepare request
	messages := []Message{
		{Role: MsgTypeSystem, Content: systemMsg},
		{Role: MsgTypeUser, Content: prompt},
	}

	reqBody := CompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   s.maxTokens,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Make request to OpenAI
	req, err := http.NewRequest("POST", "http://192.168.0.113:8020/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// SSE format: each message starts with "data: "
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove the "data: " prefix
		line = strings.TrimPrefix(line, "data: ")

		// Check for end of stream
		if line == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue // Skip malformed chunks
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				// Write the content to the response writer
				if _, err := writer.Write([]byte(content)); err != nil {
					return err
				}
				// Flush the writer if it's a flusher (like http.ResponseWriter)
				if flusher, ok := writer.(http.Flusher); ok {
					flusher.Flush()
				}
			}
		}
	}

	return nil
}

// Optimize the prompt for better code generation
func (s *AIService) OptimizePrompt(userPrompt string, language string) string {
	// Add language-specific instructions
	optimizedPrompt := fmt.Sprintf("I need to write %s code that: %s\n", language, userPrompt)

	// Add quality guidelines
	optimizedPrompt += "\nPlease provide:\n"
	optimizedPrompt += "1. Clean, efficient, and well-structured code\n"
	optimizedPrompt += "2. Brief explanations of non-obvious parts\n"
	optimizedPrompt += "3. Proper error handling where appropriate\n"

	return optimizedPrompt
}

// buildChatRequest creates a CompletionRequest with the provided parameters
func (s *AIService) buildChatRequest(sysPrompt string, prompt string, historyMsgs []*Message, modelName string, stream bool) CompletionRequest {
	// Use specified model or fall back to default
	model := s.model
	if modelName != "" {
		model = modelName
	}

	// Start with system message if provided
	messages := []Message{}
	if sysPrompt != "" {
		messages = append(messages, Message{Role: MsgTypeSystem, Content: sysPrompt})
	}

	// Add history messages
	if len(historyMsgs) > 0 {
		for _, msg := range historyMsgs {
			messages = append(messages, Message{Role: msg.Role, Content: msg.Content})
		}
	}

	// Add current user prompt
	if prompt != "" {
		messages = append(messages, Message{Role: MsgTypeUser, Content: prompt})
	}

	// Get config values for temperature and max tokens
	temperature := s.cfg.LLM.Temperature
	maxTokens := s.maxTokens
	if maxTokens == 0 {
		maxTokens = s.cfg.LLM.MaxTokens
	}
	// 如果不是默认模型，则使用配置文件中的参数
	if modelName != s.cfg.LLM.Model {
		for _, modelInfo := range s.cfg.LLM.Models {
			if modelInfo.Name == modelName {
				maxTokens = modelInfo.MaxTokens
				break
			}
		}
	}

	return CompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      stream,
	}
}

// Chat sends a chat request with custom parameters and returns a non-streaming response
func (s *AIService) Chat(sysPrompt string, prompt string, historyMsgs []*Message, modelName string) (string, error) {
	reqBody := s.buildChatRequest(sysPrompt, prompt, historyMsgs, modelName, false)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Errorf("Chatson.Marshal(reqBody) error: %v", reqBody)
		return "", err
	}
	logger.Infof("Chat  request data: %v", string(jsonData))

	// Use base URL from config if available
	baseURL := "https://api.openai.com/v1/chat/completions"
	if s.cfg.LLM.BaseURL != "" {
		baseURL = s.cfg.LLM.BaseURL + "/v1/chat/completions"
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Chat http.NewRequest error: %v", err.Error())
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Chat do request error: %v", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("ChatStreamByLine resp.StatusCode is not: %v, but is %v", 200, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	logger.Infof("Chat response: %v", string(bodyBytes))

	var result CompletionResponse
	if err := json.Unmarshal([]byte(bodyBytes), &result); err != nil {
		logger.Errorf("ChatStreamByLine json.NewDecoder error: %v,  response body: %v", err.Error(), resp.Body)
		return "", err
	}

	if len(result.Choices) == 0 {
		logger.Errorf("ChatStreamByLine len(result.Choices) == 0")
		return "", errors.New("no completion choices returned")
	}

	return result.Choices[0].Message.Content, nil
}

// ChatStream sends a chat request with custom parameters and streams the response
func (s *AIService) ChatStream(sysPrompt string, prompt string, historyMsgs []*Message, modelName string, writer io.Writer) error {
	reqBody := s.buildChatRequest(sysPrompt, prompt, historyMsgs, modelName, true)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Use base URL from config if available
	baseURL := "https://api.openai.com/v1/chat/completions"
	if s.cfg.LLM.BaseURL != "" {
		baseURL = s.cfg.LLM.BaseURL + "/v1/chat/completions"
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	client := &http.Client{
		Timeout: time.Duration(s.cfg.LLM.Timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// SSE format: each message starts with "data: "
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove the "data: " prefix
		line = strings.TrimPrefix(line, "data: ")

		// Check for end of stream
		if line == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue // Skip malformed chunks
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				// Write the content to the response writer
				if _, err := writer.Write([]byte(content)); err != nil {
					return err
				}
				// Flush the writer if it's a flusher (like http.ResponseWriter)
				if flusher, ok := writer.(http.Flusher); ok {
					flusher.Flush()
				}
			}
		}
	}

	return nil
}

// ChatStreamByLine sends a chat request and streams the response line by line
func (s *AIService) ChatStreamByLine(sysPrompt string, prompt string, historyMsgs []*Message, modelName string, writer io.Writer, mode string, isContinue bool) error {
	reqBody := s.buildChatRequest(sysPrompt, prompt, historyMsgs, modelName, true)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Use base URL from config if available
	baseURL := "https://api.openai.com/v1/chat/completions"
	if s.cfg.LLM.BaseURL != "" {
		baseURL = s.cfg.LLM.BaseURL + "/v1/chat/completions"
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("ChatStreamByLine http.NewRequest error: %v", err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	client := &http.Client{
		Timeout: time.Duration(s.cfg.LLM.Timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("ChatStreamByLine do request error: %v", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("ChatStreamByLine resp.StatusCode is not: %v, but is %v", 200, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	var lineBuffer strings.Builder
	var fullResponse strings.Builder
	lineNum := 1

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Send any remaining content in the buffer at the end
				if lineBuffer.Len() > 0 {
					// 判断模式是否为single-html，这个模式需要将开头的```这种文本去掉
					if isContinue && lineNum < 5 && assistantmessage.StartsWithCodeBlock(lineBuffer.String()) {
						logger.Infof("ChatStreamByLine line string: %v, line number: %v", lineBuffer.String(), lineNum)
						continue
					}

					if _, err := writer.Write([]byte(lineBuffer.String())); err != nil {
						return err
					}
					if flusher, ok := writer.(http.Flusher); ok {
						flusher.Flush()
					}
					lineNum += 1
				}
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// SSE format: each message starts with "data: "
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove the "data: " prefix
		line = strings.TrimPrefix(line, "data: ")

		// Check for end of stream
		if line == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue // Skip malformed chunks
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content == "" {
				if mode != SessionModeSingleHtml {
					// SessionModeSingleHtml模式如果最后没有数据，那么就丢弃最后一行数据
					content = "\n"
				} else if mode == SessionModeSingleHtml && strings.Contains(lineBuffer.String(), "</write_to_file>") {
					// 判断最后一行里面是否有 write_to_file 闭合标签
					content = "\n"
				}
			}
			if content != "" {
				// Add to full response
				fullResponse.WriteString(content)

				// Add to line buffer
				lineBuffer.WriteString(content)

				// If we have a complete line, send it
				if strings.Contains(content, "\n") {
					lines := strings.Split(lineBuffer.String(), "\n")
					for i := 0; i < len(lines)-1; i++ {
						// 判断模式是否为single-html，这个模式需要将开头的```这种文本去掉
						if isContinue && lineNum < 5 && assistantmessage.StartsWithCodeBlock(lines[i]) {
							logger.Infof("ChatStreamByLine line string: %v, line number: %v", lines[i], lineNum)
							continue
						}
						if _, err := writer.Write([]byte(lines[i] + "\n")); err != nil {
							return err
						}
						if flusher, ok := writer.(http.Flusher); ok {
							flusher.Flush()
						}
						lineNum += 1
					}

					// Keep the remainder in the buffer
					lineBuffer.Reset()
					lineBuffer.WriteString(lines[len(lines)-1])
				}
			}
		}
	}

	return nil
}

// SSE 写入器，格式化为 SSE 事件
type SseLineWriter struct {
	ResponseWriter gin.ResponseWriter
	Buffer         *strings.Builder
	LastSendText   string
	Mode           string
	Cwd            string
	LastSaveAt     *time.Time
	MsgId          int64
	UserMsgId      int64
}
type StreamLineChunk struct {
	Filename   string                                     `json:"filename"`
	Content    string                                     `json:"content"`
	Id         *int64                                     `json:"id"`
	UserMsgId  *int64                                     `json:"user_msg_id"`
	ParsedLine []assistantmessage.AssistantMessageContent `json:"parsed_line"`
}

func (w *SseLineWriter) Write(p []byte) (n int, err error) {
	switch w.Mode {

	case SessionModeSingleHtml:
		return w.WriteModeSingleHtml(p)
	case SessionModeAuto:
		return w.WriteModeAuto(p)
	default:
		return w.WriteModeAuto(p)
	}
}

func (w *SseLineWriter) GetBuffer() string {
	return w.Buffer.String()
}

func (w *SseLineWriter) WriteModeAuto(p []byte) (n int, err error) {
	// 结束标记
	isEnd := false
	if StreamMsgEndTag == string(p) {
		isEnd = true
	}

	// 将数据写入缓冲区
	if isEnd {
		w.Buffer.Write([]byte("\n"))
	} else {
		w.Buffer.Write(p)
	}

	// 格式化为SSE事件并发送到客户端
	parsed := assistantmessage.ParseAssistantMessage(w.GetBuffer())
	content := assistantmessage.GenerateMarkdown(parsed)

	if isEnd {
		parsed = assistantmessage.RemoveTextNode(parsed)
	} else {
		parsed = nil
	}

	// 为了确保流式是正常的增量数据，那么这里把已发送的文本去掉
	newStreamText := content
	if w.LastSendText != "" {
		newStreamText = strings.Replace(newStreamText, w.LastSendText, "", 1)
		// 最后一次需要发送工具调用到前端，所以这里要判断以下工具调用
		if newStreamText == "" && parsed == nil {
			return len(p), nil
		}
	}

	w.LastSendText = content

	chunk := StreamLineChunk{
		Content:    newStreamText,
		ParsedLine: parsed,
		UserMsgId:  &w.UserMsgId,
	}
	if w.MsgId != 0 {
		chunk.Id = &w.MsgId
	}
	jsData, _ := json.Marshal(chunk)
	data := fmt.Sprintf("data: %v\n\n", string(jsData))
	_, err = w.ResponseWriter.Write([]byte(data))
	w.ResponseWriter.Flush()
	return len(p), err
}
func (w *SseLineWriter) WriteModeSingleHtml(p []byte) (n int, err error) {
	// 结束标记
	isEnd := false
	if StreamMsgEndTag == string(p) {
		isEnd = true
	}

	// 将数据写入缓冲区
	if isEnd {
		w.Buffer.Write([]byte("\n"))
	} else {
		w.Buffer.Write(p)
	}

	// 格式化为SSE事件并发送到客户端
	parsed := assistantmessage.ParseAssistantMessage(w.GetBuffer())
	content := assistantmessage.GenerateMarkdown(parsed)

	filename := ""

	for _, tool := range parsed {
		switch c := tool.(type) {
		case *assistantmessage.ToolUse:
			// 用于判断代码是否编写完成，特别是非常长的代码逻辑
			if c.Name == assistantmessage.WriteToFile {
				now := time.Now()
				if w.LastSaveAt == nil {
					w.LastSaveAt = &now
				}
				if now.Before((*w.LastSaveAt).Add(3*time.Second)) && !isEnd {
					break
				}

				if _, ok := c.Params["line_count"]; !ok {
					c.Params["line_count"] = "10"
				}

				// 执行工具
				executeParams := tools.ExecutorInput{
					ToolUse:             *c,
					Ctx:                 context.Background(),
					Cwd:                 w.Cwd,
					DiffStrategy:        nil,
					RooIgnoreController: nil,
				}
				executeRes, err := tools.ExecuteTool(executeParams)
				if err != nil {
					logger.Infof("WriteModeSingleHtml execute: %v, error: %v", executeRes, err.Error())
				}
				logger.Infof("WriteModeSingleHtml execute success....., executeRes: %v", executeRes)

				// 前端需要返回path数据
				if v, ok := c.Params["path"]; ok {
					filename = v
					w.LastSaveAt = &now
				}
			}
		}
	}

	// fmt.Println("===========================================================")
	// pJs, _ := json.Marshal(parsed)
	// fmt.Println(string(pJs))
	// fmt.Println("===========================================================")

	if isEnd {
		parsed = assistantmessage.RemoveTextNode(parsed)
	} else {
		parsed = nil
	}

	// 为了确保流式是正常的增量数据，那么这里把已发送的文本去掉
	newStreamText := content
	if w.LastSendText != "" {
		newStreamText = strings.Replace(newStreamText, w.LastSendText, "", 1)
		// 最后一次需要发送工具调用到前端，所以这里要判断以下工具调用
		if newStreamText == "" && parsed == nil {
			return len(p), nil
		}
	}

	w.LastSendText = content

	chunk := StreamLineChunk{
		Filename:   filename,
		Content:    newStreamText,
		ParsedLine: parsed,
		UserMsgId:  &w.UserMsgId,
	}
	if w.MsgId != 0 {
		chunk.Id = &w.MsgId
	}

	jsData, _ := json.Marshal(chunk)
	data := fmt.Sprintf("data: %v\n\n", string(jsData))
	_, err = w.ResponseWriter.Write([]byte(data))
	w.ResponseWriter.Flush()
	return len(p), err
}

// ChatStreamByLine sends a chat request and streams the response line by line
func (s *AIService) ChatStreamByLineTest(sysPrompt string, prompt string, historyMsgs []*Message, modelName string, writer io.Writer) error {
	// Simulate processing time before starting the stream
	time.Sleep(100 * time.Millisecond)

	testData := `好的，我将为您创建一个简单的 "Hello World" HTML 文件。

<tool_use>
<write_to_file>
<path>index.html</path>
<content>
<!DOCTYPE html>
<html>
<head>
    <title>Hello World</title>
</head>
<body>
    <h1>Hello World</h1>
</body>
</html>
</content>
<line_count>9</line_count>
</write_to_file>
`

	// Split the test data into lines
	lines := strings.Split(testData, "\n")

	// Simulate streaming each line with a small delay
	for _, line := range lines {
		// Write the line with newline character
		if _, err := writer.Write([]byte(line + "\n")); err != nil {
			return err
		}

		// Flush if the writer supports it (e.g., http.ResponseWriter)
		if flusher, ok := writer.(http.Flusher); ok {
			flusher.Flush()
		}

		// Simulate network/processing delay
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func (s *AIService) GetUserHistoryMsgs(sessionID int64) []*Message {
	var userHistoryMsgs []*Message
	allMessages, err := s.database.GetSessionMessages(sessionID)
	if err != nil {
		return userHistoryMsgs
	}

	for _, item := range allMessages {
		if err != nil {
			return userHistoryMsgs
		}
		userHistoryMsgs = append(userHistoryMsgs, &Message{
			Id:      item.ID,
			Role:    item.Role,
			Content: item.Content,
		})
	}
	return userHistoryMsgs

}
