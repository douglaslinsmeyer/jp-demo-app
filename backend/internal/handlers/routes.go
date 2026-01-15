package handlers

import (
	"net/http"

	"github.com/douglasl/tokyo-commute-optimizer/internal/models"
	"github.com/douglasl/tokyo-commute-optimizer/internal/services"
	"github.com/gin-gonic/gin"
)

var (
	optimizerService *services.OptimizerService
	odptService      *services.ODPTService
	weatherService   *services.WeatherService
)

// InitServices initializes all service instances
func InitServices() {
	optimizerService = services.NewOptimizerService()
	odptService = services.NewODPTService()
	weatherService = services.NewWeatherService()
}

// HealthCheck godoc
// @Summary Health check
// @Description Returns the health status of the API
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// CalculateRoute godoc
// @Summary Calculate optimized routes
// @Description Calculates the best routes and departure times based on multiple factors including delays, weather, and crowding. Optionally specify the earliest time you can leave.
// @Tags routes
// @Accept json
// @Produce json
// @Param request body models.RouteRequest true "Route calculation request with optional earliest departure time"
// @Success 200 {object} models.RouteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/calculate-route [post]
func CalculateRoute(c *gin.Context) {
	var req models.RouteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Calculate optimized routes
	response, err := optimizerService.OptimizeRoutes(req.Origin, req.Destination, req.DepartureTime, req.EarliestDepartureTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to calculate routes",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDelays godoc
// @Summary Get current train delays
// @Description Returns current delay information for Tokyo trains from ODPT API
// @Tags transit
// @Produce json
// @Success 200 {object} map[string]interface{} "delays"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/delays [get]
func GetDelays(c *gin.Context) {
	delays, err := odptService.GetDelays()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to fetch delays",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"delays": delays,
	})
}

// GetWeather godoc
// @Summary Get current weather
// @Description Returns current weather conditions in Tokyo from Open-Meteo JMA API
// @Tags weather
// @Produce json
// @Success 200 {object} models.WeatherInfo
// @Failure 500 {object} models.ErrorResponse
// @Router /api/weather [get]
func GetWeather(c *gin.Context) {
	weather, err := weatherService.GetWeather()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to fetch weather",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, weather)
}
