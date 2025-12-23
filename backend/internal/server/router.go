package server

import (
	"fmt"
	"net/http"
	"time"

	"image-manager/internal/config"
	"image-manager/internal/handlers"
	"image-manager/internal/middleware"
	"image-manager/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	cfg          config.Config
	engine       *gin.Engine
	authHandler  *handlers.AuthHandler
	imageHandler *handlers.ImageHandler
	tagHandler   *handlers.TagHandler
	mcpHandler   *handlers.MCPHandler
}

func New(db *gorm.DB, cfg config.Config) *Server {
	tagService := services.NewTagService(db)
	aiService := services.NewAIService(cfg)
	imageService := services.NewImageService(db, cfg, tagService, aiService)
	authService := services.NewAuthService(db, cfg.JWTSecret)

	s := &Server{
		cfg:          cfg,
		engine:       gin.New(),
		authHandler:  handlers.NewAuthHandler(authService),
		imageHandler: handlers.NewImageHandler(imageService, tagService, authService),
		tagHandler:   handlers.NewTagHandler(tagService),
		mcpHandler:   handlers.NewMCPHandler(imageService, aiService, tagService),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.engine.Use(gin.Logger())
	s.engine.Use(gin.Recovery())

	corsCfg := cors.Config{
		AllowOrigins:     []string{"*"},  // 允许所有来源
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Content-Length", "X-Requested-With", "Accept", "Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,  // 当AllowOrigins为"*"时，必须设置为false
		MaxAge:           12 * time.Hour,
	}

	s.engine.Use(cors.New(corsCfg))
}

func (s *Server) setupRoutes() {
	api := s.engine.Group("/api/v1")

	api.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api.POST("/auth/register", s.authHandler.Register)
	api.POST("/auth/login", s.authHandler.Login)

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))

	protected.GET("/images", s.imageHandler.List)
	protected.POST("/images/upload", s.imageHandler.Upload)
	protected.GET("/images/:id", s.imageHandler.Detail)
	protected.PUT("/images/:id", s.imageHandler.Update)
	protected.DELETE("/images/:id", s.imageHandler.Delete)
	protected.POST("/images/:id/crop", s.imageHandler.Crop)
	protected.POST("/images/:id/adjust", s.imageHandler.Adjust)
	protected.POST("/images/import/verify", s.imageHandler.ImportVerify)
	protected.POST("/images/import", s.imageHandler.Import)

	api.GET("/images/:id/thumbnail", s.imageHandler.Thumbnail)
	api.GET("/images/:id/original", s.imageHandler.Original)

	protected.POST("/images/:id/tags", s.tagHandler.Assign)
	protected.DELETE("/images/:id/tags/:tagId", s.tagHandler.Remove)
	protected.POST("/images/:id/tags/add", s.tagHandler.AddImageTag)
	protected.PUT("/images/:id/tags/update", s.tagHandler.UpdateImageTag)
	protected.GET("/tags", s.tagHandler.List)
	protected.POST("/tags", s.tagHandler.Create)
	protected.PUT("/tags/:id/color", s.tagHandler.UpdateColor)
	protected.DELETE("/tags/:id", s.tagHandler.Delete)

	// MCP对话式图片检索接口
	protected.POST("/mcp/search", s.mcpHandler.Search)
}

func (s *Server) Run() error {
	address := fmt.Sprintf(":%s", s.cfg.ServerPort)
	return s.engine.Run(address)
}
