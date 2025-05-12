package util

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func FormatString(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func QueryIntWithDefault(c *gin.Context, key string, defaultValue int) (int, error) {
	str := c.Query(key)
	if str == "" {
		return defaultValue, nil
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", key)
	}

	return val, nil
}

func QueryInt(c *gin.Context, key string) (int, error) {
	str := c.Query(key)
	if str == "" {
		return 0, fmt.Errorf("missing %s parameter", key)
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", key)
	}

	return val, nil
}

func ReadFileToString(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// Base64Tool 对文本进行Base64编码或解码
// text: 要处理的文本
// operation: "encode"表示编码，"decode"表示解码
// 返回处理后的字符串和错误信息
func Base64Tool(text string, operation string) (string, error) {
	switch operation {
	case "encode":
		// 编码
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		return encoded, nil
	case "decode":
		// 解码
		decoded, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return "", fmt.Errorf("解码失败: %v", err)
		}
		return string(decoded), nil
	default:
		return "", fmt.Errorf("无效操作: %s (必须是 'encode' 或 'decode')", operation)
	}
}
func Base64ToolNoError(text string, operation string) string {
	txt, _ := Base64Tool(text, operation)
	return txt
}
func Base64ToolEncode(text string) string {
	return Base64ToolNoError(text, "encode")
}
func Base64ToolDecode(text string) string {
	return Base64ToolNoError(text, "decode")
}
