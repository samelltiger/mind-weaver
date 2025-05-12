package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// SSE 写入器，格式化为 SSE 事件
type SseWriter struct {
	ResponseWriter gin.ResponseWriter
	Buffer         *strings.Builder
	MsgId          int64
	UserMsgId      int64
}

func (w *SseWriter) Write(p []byte) (n int, err error) {
	// 将数据写入缓冲区
	w.Buffer.Write(p)

	data := fmt.Sprintf("data: {\"content\":%q, \"user_msg_id\":%d}\n\n", string(p), w.UserMsgId)
	// 格式化为SSE事件并发送到客户端
	if w.MsgId != 0 {
		data = fmt.Sprintf("data: {\"content\":%q, \"id\":%d, \"user_msg_id\":%d}\n\n", string(p), w.MsgId, w.UserMsgId)
	}
	_, err = w.ResponseWriter.Write([]byte(data))
	w.ResponseWriter.Flush()
	return len(p), err
}

func (w *SseWriter) GetBuffer() string {
	return w.Buffer.String()
}
