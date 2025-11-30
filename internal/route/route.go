package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin-mongo-aws/internal/config"
	"gin-mongo-aws/internal/database"
	"gin-mongo-aws/internal/handlers"
	"gin-mongo-aws/internal/middleware"
	"gin-mongo-aws/internal/repository"
	"gin-mongo-aws/internal/service"
	"gin-mongo-aws/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Run() {
	// Initialize Logger
	logger.InitLogger(s.cfg.Server.Mode)

	// Connect to MongoDB
	if err := database.ConnectMongoDB(s.cfg.MongoDB.URI); err != nil {
		logger.Log.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer database.DisconnectMongoDB()

	// Connect to Redis
	if err := database.ConnectRedis(s.cfg.Redis.Addr, s.cfg.Redis.Password, s.cfg.Redis.DB); err != nil {
		logger.Log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer database.CloseRedis()

	// Initialize Gin
	if s.cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Middleware
	r.Use(middleware.ZapLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimiter())

	// Dependencies
	userRepo := repository.NewUserRepository(s.cfg.MongoDB.Database)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Routes
	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    ":" + s.cfg.Server.Port,
		Handler: r,
	}

	// Graceful Shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("listen: %s\n", zap.Error(err))
		}
	}()

	logger.Log.Info("Server started", zap.String("port", s.cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}
