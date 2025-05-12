package api

import (
	"mind-weaver/config"
	"mind-weaver/internal/db"
	"mind-weaver/internal/services"
)

type Handler struct {
	fileService    *services.FileService
	contextService *services.ContextService
	sessionService *services.SessionService
	aiService      *services.AIService
	database       *db.Database
	cfg            config.Config

	// New command services
	commandService *services.CommandService
	swaggerService *services.SwaggerService
}

func NewHandler(
	fileService *services.FileService,
	contextService *services.ContextService,
	sessionService *services.SessionService,
	aiService *services.AIService,
	commandService *services.CommandService,
	swaggerService *services.SwaggerService,
	database *db.Database,
	cfg *config.Config,
) *Handler {
	return &Handler{
		fileService:    fileService,
		contextService: contextService,
		sessionService: sessionService,
		aiService:      aiService,
		database:       database,
		cfg:            *cfg,
		commandService: commandService,
	}
}
