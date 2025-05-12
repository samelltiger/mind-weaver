// @title           mind-weaver
// @version         1.0
// @description     思维编织者，将想法转化为代码。微信公众号：思维编织者
// @termsOfService  http://swagger.io/terms/

// @contact.name   ct
// @contact.url    http://www.example.com/support
// @contact.email  xxxxxxx@xxx.com

// @host      192.168.0.106:14010
// @BasePath  /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"mind-weaver/config"
	"mind-weaver/internal/api"
	"mind-weaver/internal/db"
	"mind-weaver/internal/middleware"
	"mind-weaver/internal/services"
	"mind-weaver/pkg/logger"
	"mind-weaver/pkg/prompts"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Println(cfg.LLM.BaseURL)

	// 初始化日志
	logger.Setup(cfg.Logger)

	// 初始化提示语
	prompts.Init(false)
	// code, err := prompts.GetPrompt("code_analysis")
	// if err != nil {
	// 	logger.Errorf("Failed to get prompt: %v", err)
	// 	return
	// }
	// logger.Infof(code)

	// Setup database
	database, err := db.InitDB(cfg.Sqliter.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create service instances
	fileService := services.NewFileService()
	contextService := services.NewContextService(fileService)
	aiService := services.NewAIService(database, cfg)
	sessionService := services.NewSessionService(database, fileService, contextService, aiService)
	commandService := services.NewCommandService()
	swaggerService := services.NewSwaggerService()

	// Create API handler
	handler := api.NewHandler(
		fileService,
		contextService,
		sessionService,
		aiService,
		commandService,
		swaggerService,
		database,
		cfg,
	)

	// Setup router
	router := gin.Default()
	gin.SetMode(cfg.Server.Mode)

	// Register middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// 添加 Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register routes
	api.RegisterRoutes(router, handler)

	// Serve static files for frontend
	frontendDir := os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		frontendDir = "./frontend" // Default frontend directory
	}

	// Check if frontend directory exists
	if _, err := os.Stat(frontendDir); !os.IsNotExist(err) {
		// Serve static files
		router.Static("/web", frontendDir)

		// Serve static files from subdirectories
		router.Static("/css", filepath.Join(frontendDir, "css"))
		router.Static("/js", filepath.Join(frontendDir, "js"))

		// Serve index.html for any unmatched routes (SPA support)
		router.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(frontendDir, "index.html"))
		})
	} else {
		log.Printf("Warning: Frontend directory %s not found. Serving API only.", frontendDir)
	}

	// Start server
	port := cfg.Server.Port
	fmt.Printf("Starting server on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
