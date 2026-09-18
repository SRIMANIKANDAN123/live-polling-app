package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"live-polling-app/backend/internal/config"
	"live-polling-app/backend/internal/handlers"
	"live-polling-app/backend/internal/middleware"
	"live-polling-app/backend/internal/redisclient"
	"live-polling-app/backend/internal/repositories"
	"live-polling-app/backend/internal/services"
	appws "live-polling-app/backend/internal/websocket"
)

func main() {
	_ = godotenv.Load() // no-op if .env is absent (e.g. in production)
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}
	db := mongoClient.Database(cfg.MongoDatabase)
	log.Println("connected to MongoDB")

	redisClient, err := redisclient.New(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	log.Println("connected to Redis")

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	pollRepo := repositories.NewPollRepository(db)
	voteRepo := repositories.NewVoteRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	pollService := services.NewPollService(pollRepo)
	voteService := services.NewVoteService(pollRepo, voteRepo, redisClient)

	// Realtime hub bridges Redis Pub/Sub -> WebSocket clients
	hub := appws.NewHub(redisClient)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	pollHandler := handlers.NewPollHandler(pollService, voteService)
	voteHandler := handlers.NewVoteHandler(voteService, hub)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS(cfg.CORSOrigin))
	router.Use(middleware.SecurityHeaders())

	router.GET("/api/health", func() gin.HandlerFunc {
		return handlers.HealthCheck(mongoClient, redisClient)
	}())

	authLimiter := middleware.RateLimit(20, time.Minute)
	voteLimiter := middleware.RateLimit(30, time.Minute)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authLimiter, authHandler.Register)
			auth.POST("/login", authLimiter, authHandler.Login)
			auth.GET("/me", middleware.RequireAuth(cfg.JWTSecret), authHandler.Me)
		}

		polls := api.Group("/polls")
		{
			polls.POST("", middleware.RequireAuth(cfg.JWTSecret), pollHandler.Create)
			polls.GET("/mine", middleware.RequireAuth(cfg.JWTSecret), pollHandler.ListMine)
			polls.GET("/:id", pollHandler.Get)
			polls.GET("/:id/results", pollHandler.Results)
			polls.POST("/:id/vote", voteLimiter, middleware.OptionalAuth(cfg.JWTSecret), voteHandler.Vote)
			polls.POST("/:id/close", middleware.RequireAuth(cfg.JWTSecret), pollHandler.Close)
			polls.DELETE("/:id", middleware.RequireAuth(cfg.JWTSecret), pollHandler.Delete)
		}
	}

	// WebSocket endpoint: wss://<host>/ws/polls/:id
	router.GET("/ws/polls/:id", voteHandler.ServeWS)

	log.Printf("live-polling-app backend listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
