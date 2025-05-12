package api

import (
	"github.com/gin-gonic/gin"

	"mind-weaver/internal/api/base"
)

// ModelInfo represents the safe version of model information to expose via API
type ModelInfo struct {
	Name         string   `json:"name"`          // 模型名称
	Description  string   `json:"description"`   // 模型描述
	MaxContext   int      `json:"max_context"`   // 最大上下文长度
	MaxTokens    int      `json:"max_tokens"`    // 最大token数
	Capabilities []string `json:"capabilities"`  // 能力列表
	IsChatModel  bool     `json:"is_chat_model"` // 是否为聊天模型
	Temperature  float64  `json:"temperature"`   // 温度参数
}

// GetModels 获取可用的LLM模型列表
// @Summary      获取模型列表
// @Description  获取系统中所有可用的LLM模型信息
// @Tags         model
// @Accept       json
// @Produce      json
// @Success      200  {object}  base.Response{data=[]ModelInfo}
// @Router       /models [get]
func (h *Handler) GetModels(c *gin.Context) {
	// Get models from config
	models := h.cfg.LLM.Models

	// Convert to safe response format
	safeModels := make([]ModelInfo, len(models))
	for i, model := range models {
		safeModels[i] = ModelInfo{
			Name:         model.Name,
			Description:  model.Description,
			MaxContext:   model.MaxContext,
			MaxTokens:    model.MaxTokens,
			Capabilities: model.Capabilities,
			IsChatModel:  model.IsChatModel,
			Temperature:  model.Temperature,
		}
	}

	base.SuccessResponse(c, safeModels)
}
