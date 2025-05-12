package api

import (
	_ "mind-weaver/internal/api/docs" // 导入生成的 docs

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, handler *Handler) {
	// API group
	api := router.Group("/api")
	{
		// Project routes
		projects := api.Group("/projects")
		{
			projects.GET("", handler.GetProjects)
			projects.POST("", handler.CreateProject)
			projects.PUT("/:id", handler.UpdateProject)
			projects.GET("/:id", handler.GetProject)
			projects.GET("/:id/files", handler.GetProjectFiles)
		}

		// File routes
		api.POST("/files/read", handler.ReadFile)
		api.GET("/files/single-html", handler.ReadHtmlFile)

		// Session routes
		sessions := api.Group("/sessions")
		{
			sessions.POST("", handler.CreateSession)
			sessions.GET("/project/:projectId", handler.GetSessions)
			sessions.GET("/:id", handler.GetSession)
			sessions.PUT("/:id", handler.UpdateSession)
			sessions.DELETE("/:id", handler.DeleteSession)
			// 大模型相关接口
			sessions.POST("/:id/message", handler.SendMessage)                   // 消息列表
			sessions.DELETE("/:id/messages/:msgId", handler.DeleteMessage)       // 删除消息，msgId为消息id，当msgId为0时，删除所有消息
			sessions.POST("/:id/completions", handler.OpenAICompatStreamHandler) // 流式响应
			sessions.POST("/parse/ai-res", handler.ParseAiContent)               // 解析ai响应文本

			// 上下文信息相关接口
			sessions.PUT("/:id/context", handler.UpdateContext)
			sessions.GET("/:id/context", handler.GetContext)
		}

		api.GET("/models", handler.GetModels)
		api.POST("/prompts/test", handler.TestPrompt)

		// Command execution routes
		commands := api.Group("/commands")
		{
			commands.POST("/execute", handler.ExecuteCommand)
			commands.POST("/execute-code", handler.ExecuteCode)
		}

		swaggerGroup := api.Group("/swaggers")
		{
			swaggerGroup.POST("/list", handler.ListInterfaces)
			swaggerGroup.POST("/doc", handler.GenerateDoc)
		}

		openai := router.Group("/v1")
		{
			openai.POST("/chat/completions", handler.OpenAICompatStreamHandler)
		}

		// 工具接口
		toolsApi := api.Group("/tools")
		{
			toolsApi.GET("/jsinspector", handler.JsInspector)
			toolsApi.GET("/handle-llm-response", handler.HandleLlmResponseError)
		}

	}
}
