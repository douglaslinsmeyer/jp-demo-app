package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/douglasl/tokyo-commute-optimizer/internal/cache"
	"github.com/douglasl/tokyo-commute-optimizer/internal/models"
)

const (
	googleMapsBaseURL = "https://routes.googleapis.com/directions/v2:computeRoutes"
	routeCacheTTL     = 5 * time.Minute
)

// GoogleMapsService handles Google Maps API operations
type GoogleMapsService struct {
	apiKey     string
	httpClient *http.Client
}

// GoogleMapsRequest represents the request to Google Routes API
type GoogleMapsRequest struct {
	Origin struct {
		Address string `json:"address"`
	} `json:"origin"`
	Destination struct {
		Address string `json:"address"`
	} `json:"destination"`
	TravelMode         string    `json:"travelMode"`
	RoutingPreference  string    `json:"routingPreference,omitempty"`
	DepartureTime      time.Time `json:"departureTime,omitempty"`
	ComputeAlternativeRoutes bool `json:"computeAlternativeRoutes"`
	LanguageCode       string    `json:"languageCode"`
	Units              string    `json:"units"`
}

// GoogleMapsResponse represents the response from Google Routes API
type GoogleMapsResponse struct {
	Routes []struct {
		Duration      string `json:"duration"`
		DistanceMeters int    `json:"distanceMeters"`
		Polyline      struct {
			EncodedPolyline string `json:"encodedPolyline"`
		} `json:"polyline"`
		Legs []struct {
			Duration      string `json:"duration"`
			DistanceMeters int    `json:"distanceMeters"`
			StartLocation struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			} `json:"startLocation"`
			EndLocation struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			} `json:"endLocation"`
			Steps []struct {
				Duration      string `json:"duration"`
				DistanceMeters int    `json:"distanceMeters"`
				NavigationInstruction struct {
					Instructions string `json:"instructions"`
				} `json:"navigationInstruction"`
				TravelMode string `json:"travelMode"`
				TransitDetails *struct {
					StopDetails struct {
						DepartureStop struct {
							Name string `json:"name"`
						} `json:"departureStop"`
						ArrivalStop struct {
							Name string `json:"name"`
						} `json:"arrivalStop"`
						DepartureTime time.Time `json:"departureTime"`
						ArrivalTime   time.Time `json:"arrivalTime"`
					} `json:"stopDetails"`
					TransitLine struct {
						Name      string `json:"name"`
						NameShort string `json:"nameShort"`
						Color     string `json:"color"`
						Vehicle   struct {
							Name string `json:"name"`
							Type string `json:"type"`
						} `json:"vehicle"`
					} `json:"transitLine"`
				} `json:"transitDetails,omitempty"`
			} `json:"steps"`
		} `json:"legs"`
	} `json:"routes"`
}

// NewGoogleMapsService creates a new Google Maps service
func NewGoogleMapsService() *GoogleMapsService {
	return &GoogleMapsService{
		apiKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetRoutes fetches route options from Google Maps API
func (s *GoogleMapsService) GetRoutes(origin, destination string, departureTime *time.Time) ([]models.Route, error) {
	// If no API key is configured, return mock data
	if s.apiKey == "" {
		return s.getMockRoutes(origin, destination), nil
	}

	// Generate cache key
	cacheKey := s.generateCacheKey(origin, destination, departureTime)

	// Try to get from cache first
	var cachedRoutes []models.Route
	err := cache.Get(cacheKey, &cachedRoutes)
	if err == nil && len(cachedRoutes) > 0 {
		return cachedRoutes, nil
	}

	// If not in cache, fetch from API
	routes, err := s.fetchRoutesFromAPI(origin, destination, departureTime)
	if err != nil {
		// If API fails, return mock data instead of failing
		fmt.Printf("Google Maps API error (using mock data): %v\n", err)
		return s.getMockRoutes(origin, destination), nil
	}

	// Cache the result
	if err := cache.Set(cacheKey, routes, routeCacheTTL); err != nil {
		fmt.Printf("Failed to cache route data: %v\n", err)
	}

	return routes, nil
}

// fetchRoutesFromAPI fetches routes from Google Maps Routes API
func (s *GoogleMapsService) fetchRoutesFromAPI(origin, destination string, departureTime *time.Time) ([]models.Route, error) {
	// Prepare request
	reqBody := GoogleMapsRequest{
		TravelMode:               "TRANSIT",
		ComputeAlternativeRoutes: true,
		LanguageCode:             "en",
		Units:                    "METRIC",
	}
	reqBody.Origin.Address = origin
	reqBody.Destination.Address = destination

	if departureTime != nil {
		reqBody.DepartureTime = *departureTime
	} else {
		reqBody.DepartureTime = time.Now()
	}

	// Marshal request
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", googleMapsBaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", s.apiKey)
	req.Header.Set("X-Goog-FieldMask", "routes.duration,routes.distanceMeters,routes.polyline,routes.legs")

	// Set body
	req.Body = io.NopCloser(strings.NewReader(string(jsonData)))

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google Maps API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp GoogleMapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Google Maps response: %w", err)
	}

	// Convert to our model
	routes := make([]models.Route, 0, len(apiResp.Routes))
	for _, googleRoute := range apiResp.Routes {
		route := s.convertToRoute(googleRoute, origin, destination)
		routes = append(routes, route)
	}

	return routes, nil
}

// convertToRoute converts Google Maps route to our route model
func (s *GoogleMapsService) convertToRoute(googleRoute interface{}, origin, destination string) models.Route {
	// This is a simplified conversion - you would implement full conversion logic here
	route := models.Route{
		Summary:   fmt.Sprintf("Route from %s to %s", origin, destination),
		TotalTime: 1800, // Placeholder
		Breakdown: models.TimeBreakdown{
			TransitTime: 1200,
			WalkingTime: 300,
			WaitingTime: 180,
		},
		Steps: []models.RouteStep{},
		Score: 0, // Will be calculated by optimizer
	}

	return route
}

// getMockRoutes returns mock route data for development/testing
func (s *GoogleMapsService) getMockRoutes(origin, destination string) []models.Route {
	now := time.Now()

	return []models.Route{
		{
			Summary:   fmt.Sprintf("Via Yamanote Line from %s to %s", origin, destination),
			TotalTime: 1800, // 30 minutes
			Breakdown: models.TimeBreakdown{
				TransitTime: 1200, // 20 minutes
				WalkingTime: 360,  // 6 minutes
				WaitingTime: 240,  // 4 minutes
			},
			Steps: []models.RouteStep{
				{
					Type:         "WALK",
					Instructions: "Walk to nearest station",
					Distance:     350,
					Duration:     300,
				},
				{
					Type:         "TRANSIT",
					Instructions: "Take Yamanote Line",
					Duration:     1200,
					Line: &models.LineInfo{
						Name:      "Yamanote Line",
						ShortName: "JY",
						Color:     "#9ACD32",
						Vehicle:   "TRAIN",
					},
					DepartAt: &now,
					ArriveAt: func() *time.Time { t := now.Add(20 * time.Minute); return &t }(),
				},
				{
					Type:         "WALK",
					Instructions: "Walk to destination",
					Distance:     200,
					Duration:     180,
				},
			},
			Score: 0,
		},
		{
			Summary:   fmt.Sprintf("Via Chuo Line from %s to %s", origin, destination),
			TotalTime: 1680, // 28 minutes
			Breakdown: models.TimeBreakdown{
				TransitTime: 1080, // 18 minutes
				WalkingTime: 420,  // 7 minutes
				WaitingTime: 180,  // 3 minutes
			},
			Steps: []models.RouteStep{
				{
					Type:         "WALK",
					Instructions: "Walk to Chuo Line station",
					Distance:     400,
					Duration:     360,
				},
				{
					Type:         "TRANSIT",
					Instructions: "Take Chuo Line (Rapid)",
					Duration:     1080,
					Line: &models.LineInfo{
						Name:      "Chuo Line",
						ShortName: "JC",
						Color:     "#F15A22",
						Vehicle:   "TRAIN",
					},
					DepartAt: &now,
					ArriveAt: func() *time.Time { t := now.Add(18 * time.Minute); return &t }(),
				},
				{
					Type:         "WALK",
					Instructions: "Walk to destination",
					Distance:     150,
					Duration:     180,
				},
			},
			Score: 0,
		},
	}
}

// generateCacheKey generates a cache key for route requests
func (s *GoogleMapsService) generateCacheKey(origin, destination string, departureTime *time.Time) string {
	// Round departure time to nearest 5 minutes for better cache hits
	var timeStr string
	if departureTime != nil {
		rounded := departureTime.Round(5 * time.Minute)
		timeStr = rounded.Format("2006-01-02T15:04")
	} else {
		timeStr = "now"
	}

	// URL encode the addresses
	originEnc := url.QueryEscape(origin)
	destEnc := url.QueryEscape(destination)

	return fmt.Sprintf("route:%s:%s:%s", originEnc, destEnc, timeStr)
}
