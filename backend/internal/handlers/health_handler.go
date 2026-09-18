package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"live-polling-app/backend/internal/redisclient"
)

func HealthCheck(mongoClient *mongo.Client, redisClient *redisclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		status := gin.H{"status": "ok"}
		httpStatus := http.StatusOK

		if err := mongoClient.Ping(ctx, nil); err != nil {
			status["mongo"] = "unreachable"
			status["status"] = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else {
			status["mongo"] = "ok"
		}

		if err := redisClient.Ping(ctx); err != nil {
			status["redis"] = "unreachable"
			status["status"] = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else {
			status["redis"] = "ok"
		}

		c.JSON(httpStatus, status)
	}
}
