package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"mind-weaver/internal/db"
	"mind-weaver/internal/third/assistantmessage"
	"mind-weaver/internal/utils"
)

const (
	SessionModeManual     = "manual"
	SessionModeAuto       = "auto"
	SessionModeSingleHtml = "single-html"
)

type SessionService struct {
	database       *db.Database
	fileService    *FileService
	contextService *ContextService
	aiService      *AIService
}

type SessionInfo struct {
	ID              int64         `json:"id"`
	ProjectID       int64         `json:"project_id"`
	Name            string        `json:"name"`
	Mode            string        `json:"mode"`
	ExcludePatterns []db.FileInfo `json:"exclude_patterns,omitempty"`
	IncludePatterns []db.FileInfo `json:"include_patterns,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Messages        []MessageInfo `json:"messages,omitempty"`
}

type MessageInfo struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ContextInfo struct {
	Files          []string `json:"files"`
	CurrentFile    string   `json:"current_file,omitempty"`
	CursorPosition int      `json:"cursor_position,omitempty"`
	SelectedCode   string   `json:"selected_code,omitempty"`
}

func NewSessionService(
	database *db.Database,
	fileService *FileService,
	contextService *ContextService,
	aiService *AIService,
) *SessionService {
	return &SessionService{
		database:       database,
		fileService:    fileService,
		contextService: contextService,
		aiService:      aiService,
	}
}

func (s *SessionService) CreateSession(projectID int64, name string, mode string, excludePatterns []db.FileInfo, includePatterns []db.FileInfo) (*SessionInfo, error) {
	// Create session in database
	excludePatternsJSON, err := json.Marshal(excludePatterns)
	if err != nil {
		return nil, err
	}

	// Create session in database
	includePatternsJSON, err := json.Marshal(includePatterns)
	if err != nil {
		return nil, err
	}

	contextInfo := ContextInfo{Files: []string{}}
	contextStr, _ := json.Marshal(contextInfo)

	// Create session in database
	sessionID, err := s.database.CreateSession(projectID, name, mode, string(excludePatternsJSON), string(includePatternsJSON), string(contextStr))
	if err != nil {
		return nil, err
	}

	// Get the created session
	session, err := s.database.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Parse exclude patterns
	var patterns []db.FileInfo
	if session.ExcludePatterns != "" {
		if err := json.Unmarshal([]byte(session.ExcludePatterns), &patterns); err != nil {
			return nil, err
		}
	}

	// Parse exclude patterns
	var incloudePatterns []db.FileInfo
	if session.IncludePatterns != "" {
		if err := json.Unmarshal([]byte(session.IncludePatterns), &patterns); err != nil {
			return nil, err
		}
	}

	return &SessionInfo{
		ID:              session.ID,
		ProjectID:       session.ProjectID,
		Name:            session.Name,
		Mode:            session.Mode,
		ExcludePatterns: patterns,
		IncludePatterns: incloudePatterns,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
	}, nil
}

func (s *SessionService) GetSession(sessionID int64) (*SessionInfo, error) {
	// Get session from database
	session, err := s.database.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Parse exclude patterns
	var excludePatterns []db.FileInfo
	if session.ExcludePatterns != "" {
		if err := json.Unmarshal([]byte(session.ExcludePatterns), &excludePatterns); err != nil {
			return nil, fmt.Errorf("failed to unmarshal exclude patterns: %w", err)
		}
	}

	// Parse include patterns
	var includePatterns []db.FileInfo
	if session.IncludePatterns != "" {
		if err := json.Unmarshal([]byte(session.IncludePatterns), &includePatterns); err != nil {
			return nil, fmt.Errorf("failed to unmarshal include patterns: %w", err)
		}
	}

	// Parse context
	var contextInfo ContextInfo
	if session.Context != "" {
		if err := json.Unmarshal([]byte(session.Context), &contextInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}
	}

	// Get messages for the session
	messages, err := s.database.GetSessionMessages(sessionID)
	if err != nil {
		return nil, err
	}

	// Convert to message info
	messageInfos := make([]MessageInfo, len(messages))
	for i, msg := range messages {
		content := msg.Content

		if msg.Role == MsgTypeSystem {
			// 跳过system消息
			continue
		}
		if msg.Role == MsgTypeSystem && session.Mode == SessionModeManual {
			content = s.FormatSystemMessage(session, includePatterns, msg)
		}
		// 对于有工具调用的 assistant message，需要解析出内容，并生成markdown格式的
		if msg.Role == MsgTypeAssistant && session.Mode == SessionModeAuto ||
			(msg.Role == MsgTypeAssistant && session.Mode == SessionModeSingleHtml) {
			parsed := assistantmessage.ParseAssistantMessage(content)
			content = assistantmessage.GenerateMarkdown(parsed)
		}

		messageInfos[i] = MessageInfo{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   content,
			Timestamp: msg.Timestamp,
		}
	}

	return &SessionInfo{
		ID:              session.ID,
		ProjectID:       session.ProjectID,
		Name:            session.Name,
		Mode:            session.Mode,
		ExcludePatterns: excludePatterns,
		IncludePatterns: includePatterns,
		// Context:         contextInfo,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		Messages:  messageInfos,
	}, nil
}

// func (s *SessionService) FormatSystemMessage(sessionInfo *db.Session, msg *db.Message) string {
// 	//
// }

func (s *SessionService) FormatSystemMessage(sessionInfo *db.Session, includePatterns []db.FileInfo, msg *db.Message) string {
	var builder strings.Builder
	builder.WriteString("系统消息中包含以下文件：\n\n")

	for _, file := range includePatterns {
		if file.IsDir {
			builder.WriteString(fmt.Sprintf("- 📁 `%s`\n", file.Path))
		} else {
			// 获取文件行数
			lineCount, err := utils.CountFileLines(file.Path)
			if err != nil {
				builder.WriteString(fmt.Sprintf("- 📄 `%s` (无法读取行数: %v)\n", file.Path, err))
			} else {
				builder.WriteString(fmt.Sprintf("- 📄 `%s` (%d 行)\n", file.Path, lineCount))
			}
		}
	}

	return builder.String()
}

func (s *SessionService) AddUserMessage(sessionID int64, content string) (*MessageInfo, error) {
	return s.AddMessage(sessionID, content, MsgTypeUser)
}

func (s *SessionService) AddSysMessage(sessionID int64, content string) (*MessageInfo, error) {
	return s.AddMessage(sessionID, content, MsgTypeSystem)
}

func (s *SessionService) AddMessage(sessionID int64, content string, role string) (*MessageInfo, error) {
	// Add message to database
	msgID, err := s.database.AddMessage(sessionID, role, content)
	if err != nil {
		return nil, err
	}

	// Get the message
	messages, err := s.database.GetSessionMessages(sessionID)
	if err != nil {
		return nil, err
	}

	// Find the added message
	var addedMsg *db.Message
	for _, msg := range messages {
		if msg.ID == msgID {
			addedMsg = msg
			break
		}
	}

	if addedMsg == nil {
		return nil, errors.New("added message not found")
	}

	return &MessageInfo{
		ID:        addedMsg.ID,
		Role:      addedMsg.Role,
		Content:   addedMsg.Content,
		Timestamp: addedMsg.Timestamp,
	}, nil
}

func (s *SessionService) GenerateAIResponse(sessionID int64, projectPath string, contextFiles []string, userPrompt string) (*MessageInfo, error) {
	// Get project and session info
	session, err := s.database.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Load file contexts
	fileContexts := []*FileContext{}
	for _, filePath := range contextFiles {
		fullPath := filePath
		if !isAbsolutePath(filePath) {
			fullPath = fmt.Sprintf("%s/%s", projectPath, filePath)
		}

		fileContext, err := s.contextService.GetFileContext(fullPath)
		if err != nil {
			// Skip files that can't be loaded
			continue
		}
		fileContexts = append(fileContexts, fileContext)
	}

	// Generate AI response
	aiResponse, err := s.aiService.GenerateCompletion(userPrompt, fileContexts)
	if err != nil {
		return nil, err
	}

	// Save the AI response to the database
	msgID, err := s.database.AddMessage(session.ID, "ai", aiResponse)
	if err != nil {
		return nil, err
	}

	// Return the message info
	return &MessageInfo{
		ID:        msgID,
		Role:      "ai",
		Content:   aiResponse,
		Timestamp: time.Now(),
	}, nil
}

func (s *SessionService) UpdateSessionContext(sessionID int64, contextInfo ContextInfo) error {
	// Get the session
	session, err := s.database.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Parse existing context
	var existingContext ContextInfo
	if session.Context != "" {
		if err := json.Unmarshal([]byte(session.Context), &existingContext); err != nil {
			return fmt.Errorf("failed to parse existing context: %w", err)
		}
	}

	// Merge with new context, only updating non-empty/non-zero values
	if len(contextInfo.Files) > 0 {
		existingContext.Files = contextInfo.Files
	}
	if contextInfo.CurrentFile != "" {
		existingContext.CurrentFile = contextInfo.CurrentFile
	}
	if contextInfo.CursorPosition != 0 {
		existingContext.CursorPosition = contextInfo.CursorPosition
	}
	if contextInfo.SelectedCode != "" {
		existingContext.SelectedCode = contextInfo.SelectedCode
	}

	// Marshal the merged context
	contextJSON, err := json.Marshal(existingContext)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	// Update in database
	_, err = s.database.Exec(
		"UPDATE sessions SET context = ?, updated_at = ? WHERE id = ?",
		string(contextJSON), time.Now(), session.ID,
	)

	return err
}

func (s *SessionService) GetSessionContext(sessionID int64) (*ContextInfo, error) {
	// Get the session
	session, err := s.database.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Parse the context
	var contextInfo ContextInfo
	if err := json.Unmarshal([]byte(session.Context), &contextInfo); err != nil {
		return nil, err
	}

	return &contextInfo, nil
}

// Helper function to check if a path is absolute
func isAbsolutePath(path string) bool {
	// This is a simplified check that works for both Unix and Windows
	if len(path) == 0 {
		return false
	}

	// Unix-style absolute path
	if path[0] == '/' {
		return true
	}

	// Windows-style absolute path (e.g., C:\path\to\file)
	if len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
		return true
	}

	return false
}

func (s *SessionService) UpdateSession(sessionID int64, name string, mode string, excludePatterns []db.FileInfo, includePatterns []db.FileInfo) (*SessionInfo, error) {
	// Validate mode if provided
	if !s.IsAllowedMode(mode) {
		return nil, errors.New("invalid mode: must be 'auto', 'manual', or 'all'")
	}

	// Prepare the data for database update
	var excludePatternsJSON, includePatternsJSON string
	var err error

	excludePatternsBytes, err := json.Marshal(excludePatterns)
	if err != nil {
		return nil, err
	}
	excludePatternsJSON = string(excludePatternsBytes)

	includePatternsBytes, err := json.Marshal(includePatterns)
	if err != nil {
		return nil, err
	}
	includePatternsJSON = string(includePatternsBytes)

	// Update session in database
	err = s.database.UpdateSession(sessionID, name, mode, excludePatternsJSON, includePatternsJSON)
	if err != nil {
		return nil, err
	}

	// Get the updated session
	return s.GetSession(sessionID)
}

func (s *SessionService) DeleteSession(sessionID int64) error {
	return s.database.DeleteSession(sessionID)
}

func (s *SessionService) DeleteMessage(sessionID, msgId int64) error {
	if msgId == 0 {
		// 为0表示删除会话下所有消息
		return s.database.DeleteAllMessage(sessionID)
	}

	return s.database.DeleteMessage(msgId)
}

func (s *SessionService) IsAllowedMode(mode string) bool {
	return mode == SessionModeAuto || mode == SessionModeManual || mode == SessionModeSingleHtml
}
