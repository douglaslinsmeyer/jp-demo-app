package main

import (
	"log"
	"os"
	"time"

	"github.com/douglasl/tokyo-commute-optimizer/internal/cache"
	"github.com/douglasl/tokyo-commute-optimizer/internal/handlers"
	"github.com/douglasl/tokyo-commute-optimizer/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/douglasl/tokyo-commute-optimizer/docs"
)

// @title Tokyo Commute Optimizer API
// @version 1.0
// @description API for optimizing Tokyo commute times based on real-time transit, weather, and crowding data
// @contact.name API Support
// @host localhost:3000
// @BasePath /
// @schemes http
func main() {
	// Initialize Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	if err := cache.InitRedis(redisURL); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v (continuing without cache)", err)
	}
	defer cache.Close()

	// Initialize services
	handlers.InitServices()

	// Set Gin mode based on environment
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS configuration
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))

	// Error handling middleware
	router.Use(middleware.ErrorHandler())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		api.POST("/calculate-route", handlers.CalculateRoute)
		api.GET("/delays", handlers.GetDelays)
		api.GET("/weather", handlers.GetWeather)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s...", port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
