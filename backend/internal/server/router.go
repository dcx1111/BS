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
}

func New(db *gorm.DB, cfg config.Config) *Server {
	tagService := services.NewTagService(db)
	imageService := services.NewImageService(db, cfg, tagService)
	authService := services.NewAuthService(db, cfg.JWTSecret)

	s := &Server{
		cfg:          cfg,
		engine:       gin.New(),
		authHandler:  handlers.NewAuthHandler(authService),
		imageHandler: handlers.NewImageHandler(imageService, tagService),
		tagHandler:   handlers.NewTagHandler(tagService),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.engine.Use(gin.Logger())
	s.engine.Use(gin.Recovery())

	corsCfg := cors.Config{
		AllowOrigins:     s.cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Content-Length", "X-Requested-With", "Accept", "Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowCredentials(s.cfg.CORSOrigins),
		MaxAge:           12 * time.Hour,
	}

	s.engine.Use(cors.New(corsCfg))
}

func allowCredentials(origins []string) bool {
	return !(len(origins) == 1 && origins[0] == "*")
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

	api.GET("/images/:id/thumbnail", s.imageHandler.Thumbnail)
	api.GET("/images/:id/original", s.imageHandler.Original)

	protected.POST("/images/:id/tags", s.tagHandler.Assign)
	protected.DELETE("/images/:id/tags/:tagId", s.tagHandler.Remove)
	protected.POST("/images/:id/tags/add", s.tagHandler.AddImageTag)
	protected.PUT("/images/:id/tags/update", s.tagHandler.UpdateImageTag)
	protected.GET("/tags", s.tagHandler.List)
	protected.POST("/tags", s.tagHandler.Create)
	protected.PUT("/tags/:id/color", s.tagHandler.UpdateColor)
}

func (s *Server) Run() error {
	address := fmt.Sprintf(":%s", s.cfg.ServerPort)
	return s.engine.Run(address)
}
