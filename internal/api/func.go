package api

import (
	"mind-weaver/internal/services"
	"sync"
	"time"
)

// 用于保存各个session的最近上下文更新时间和缓存的更新
type contextUpdateCache struct {
	sync.Mutex
	lastUpdate       map[int64]time.Time
	pendingUpdates   map[int64]*services.ContextInfo
	lastCursorUpdate map[int64]time.Time
	lastFileUpdate   map[int64]time.Time
}

// 用于存储预备的流式消息
type preparedStreamMessage struct {
	SessionID    int64
	Type         string
	Content      string
	ProjectPath  string
	ContextFiles []string
	ExpiresAt    time.Time
	Model        string
}

var (
	contextCache = &contextUpdateCache{
		lastUpdate:       make(map[int64]time.Time),
		pendingUpdates:   make(map[int64]*services.ContextInfo),
		lastCursorUpdate: make(map[int64]time.Time),
		lastFileUpdate:   make(map[int64]time.Time),
	}

	// 存储预备的流式消息，键为UUID
	preparedMessages      = make(map[string]preparedStreamMessage)
	preparedMessagesMutex sync.Mutex
	// 过期时间，15分钟
	messageExpiryDuration = 15 * time.Minute
)

// 清理过期的预备消息
func cleanupExpiredMessages() {
	preparedMessagesMutex.Lock()
	defer preparedMessagesMutex.Unlock()

	now := time.Now()
	for id, msg := range preparedMessages {
		if now.After(msg.ExpiresAt) {
			delete(preparedMessages, id)
		}
	}
}

// 处理上下文更新的防抖动逻辑
func processContextUpdate(sessionID int64, info *services.ContextInfo) bool {
	const (
		debounceInterval = 500 * time.Millisecond // 防抖动时间间隔
	)

	now := time.Now()
	contextCache.Lock()
	defer contextCache.Unlock()

	// 检查上次更新时间
	lastUpdate, exists := contextCache.lastUpdate[sessionID]

	// 检测是什么类型的更新
	isCursorUpdate := info.CurrentFile != "" && info.CursorPosition != 0 && len(info.SelectedCode) == 0
	isFileUpdate := info.CurrentFile != "" && len(info.SelectedCode) > 0

	// 更新对应类型的最后更新时间
	if isCursorUpdate {
		contextCache.lastCursorUpdate[sessionID] = now
	} else if isFileUpdate {
		contextCache.lastFileUpdate[sessionID] = now
	}

	// 存储或更新挂起的上下文
	if !exists {
		// 首次更新，保存并立即处理
		contextCache.lastUpdate[sessionID] = now
		contextCache.pendingUpdates[sessionID] = info
		return true
	}

	// 如果距离上次更新不到防抖动间隔，队列化更新
	if now.Sub(lastUpdate) < debounceInterval {
		// 如果已有挂起的更新，合并新的上下文
		existing := contextCache.pendingUpdates[sessionID]
		if existing != nil {
			// 仅更新非空字段
			if info.CurrentFile != "" {
				existing.CurrentFile = info.CurrentFile
			}
			if info.CursorPosition != 0 {
				existing.CursorPosition = info.CursorPosition
			}
			if len(info.Files) > 0 {
				existing.Files = info.Files
			}
			if len(info.SelectedCode) > 0 {
				existing.SelectedCode = info.SelectedCode
			}
		} else {
			contextCache.pendingUpdates[sessionID] = info
		}

		// 不立即更新
		return false
	}

	// 如果已过防抖动间隔，允许更新
	contextCache.lastUpdate[sessionID] = now

	// 使用合并后的上下文（如果有）
	mergedContext := contextCache.pendingUpdates[sessionID]
	if mergedContext != nil {
		// 用合并的上下文替换当前上下文
		*info = *mergedContext
		// 清除挂起的更新
		delete(contextCache.pendingUpdates, sessionID)
	}

	return true
}
