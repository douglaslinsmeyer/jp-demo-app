package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/douglasl/tokyo-commute-optimizer/internal/cache"
	"github.com/douglasl/tokyo-commute-optimizer/internal/models"
)

const (
	openMeteoBaseURL = "https://api.open-meteo.com/v1/jma"
	weatherCacheTTL  = 15 * time.Minute
	weatherCacheKey  = "weather:tokyo"
)

// OpenMeteoResponse represents the response from Open-Meteo API
type OpenMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Current   struct {
		Time          string  `json:"time"`
		Temperature2m float64 `json:"temperature_2m"`
		Precipitation float64 `json:"precipitation"`
		WeatherCode   int     `json:"weather_code"`
	} `json:"current"`
}

// WeatherService handles weather data operations
type WeatherService struct {
	httpClient *http.Client
}

// NewWeatherService creates a new weather service
func NewWeatherService() *WeatherService {
	return &WeatherService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetWeather fetches weather information for Tokyo
func (s *WeatherService) GetWeather() (*models.WeatherInfo, error) {
	// Try to get from cache first
	var cachedWeather models.WeatherInfo
	err := cache.Get(weatherCacheKey, &cachedWeather)
	if err == nil && cachedWeather.Temperature != 0 {
		return &cachedWeather, nil
	}

	// If not in cache, fetch from API
	weather, err := s.fetchWeatherFromAPI()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := cache.Set(weatherCacheKey, weather, weatherCacheTTL); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to cache weather data: %v\n", err)
	}

	return weather, nil
}

// fetchWeatherFromAPI fetches weather data from Open-Meteo API
func (s *WeatherService) fetchWeatherFromAPI() (*models.WeatherInfo, error) {
	// Tokyo coordinates
	lat := 35.6762
	lon := 139.6503

	// Build URL with parameters
	url := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f&current=temperature_2m,precipitation,weather_code",
		openMeteoBaseURL, lat, lon)

	// Make request
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("weather API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	// Convert to our model
	weather := &models.WeatherInfo{
		Temperature:   apiResp.Current.Temperature2m,
		Precipitation: apiResp.Current.Precipitation,
		Condition:     getWeatherCondition(apiResp.Current.WeatherCode),
		LastUpdated:   time.Now(),
	}

	return weather, nil
}

// getWeatherCondition converts weather code to human-readable condition
// Based on WMO Weather interpretation codes
func getWeatherCondition(code int) string {
	switch {
	case code == 0:
		return "CLEAR"
	case code >= 1 && code <= 3:
		return "PARTLY_CLOUDY"
	case code >= 45 && code <= 48:
		return "FOG"
	case code >= 51 && code <= 57:
		return "DRIZZLE"
	case code >= 61 && code <= 67:
		return "RAIN"
	case code >= 71 && code <= 77:
		return "SNOW"
	case code >= 80 && code <= 82:
		return "RAIN_SHOWERS"
	case code >= 85 && code <= 86:
		return "SNOW_SHOWERS"
	case code == 95:
		return "THUNDERSTORM"
	case code >= 96 && code <= 99:
		return "THUNDERSTORM_WITH_HAIL"
	default:
		return "UNKNOWN"
	}
}

// GetWeatherMultiplier calculates the walking time multiplier based on weather
func GetWeatherMultiplier(weather *models.WeatherInfo) float64 {
	// Base multiplier
	multiplier := 1.0

	// Precipitation impact
	switch {
	case weather.Precipitation == 0:
		multiplier = 1.0
	case weather.Precipitation < 2.5:
		multiplier = 1.2 // Light rain
	case weather.Precipitation < 10:
		multiplier = 1.5 // Moderate rain
	default:
		multiplier = 2.0 // Heavy rain
	}

	// Temperature extremes
	if weather.Temperature < 0 || weather.Temperature > 35 {
		multiplier += 0.1 // Additional penalty for extreme temperatures
	}

	return multiplier
}
