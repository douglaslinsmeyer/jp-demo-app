package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/douglasl/tokyo-commute-optimizer/internal/models"
)

// OptimizerService handles route optimization logic
type OptimizerService struct {
	googleMaps *GoogleMapsService
	odpt       *ODPTService
	weather    *WeatherService
	crowding   *CrowdingService
}

// NewOptimizerService creates a new optimizer service
func NewOptimizerService() *OptimizerService {
	return &OptimizerService{
		googleMaps: NewGoogleMapsService(),
		odpt:       NewODPTService(),
		weather:    NewWeatherService(),
		crowding:   NewCrowdingService(),
	}
}

// OptimizeRoutes calculates and optimizes routes based on all factors
func (s *OptimizerService) OptimizeRoutes(origin, destination string, departureTime *time.Time) (*models.RouteResponse, error) {
	// Use current time if not specified
	if departureTime == nil {
		now := time.Now()
		departureTime = &now
	}

	// Fetch all data in parallel
	routesChan := make(chan []models.Route, 1)
	delaysChan := make(chan []models.DelayInfo, 1)
	weatherChan := make(chan *models.WeatherInfo, 1)
	errChan := make(chan error, 3)

	// Fetch routes
	go func() {
		routes, err := s.googleMaps.GetRoutes(origin, destination, departureTime)
		if err != nil {
			errChan <- fmt.Errorf("failed to get routes: %w", err)
			return
		}
		routesChan <- routes
	}()

	// Fetch delays
	go func() {
		delays, err := s.odpt.GetDelays()
		if err != nil {
			errChan <- fmt.Errorf("failed to get delays: %w", err)
			return
		}
		delaysChan <- delays
	}()

	// Fetch weather
	go func() {
		weather, err := s.weather.GetWeather()
		if err != nil {
			errChan <- fmt.Errorf("failed to get weather: %w", err)
			return
		}
		weatherChan <- weather
	}()

	// Collect results
	var routes []models.Route
	var delays []models.DelayInfo
	var weather *models.WeatherInfo

	for i := 0; i < 3; i++ {
		select {
		case r := <-routesChan:
			routes = r
		case d := <-delaysChan:
			delays = d
		case w := <-weatherChan:
			weather = w
		case err := <-errChan:
			return nil, err
		}
	}

	// Optimize each route
	for i := range routes {
		s.optimizeRoute(&routes[i], delays, weather, *departureTime)
	}

	// Find optimal departure time by testing multiple times
	optimalTime := s.findOptimalDepartureTime(origin, destination, *departureTime, delays, weather)

	// Sort routes by score
	s.sortRoutesByScore(routes)

	// Build response
	response := &models.RouteResponse{
		Routes:               routes,
		OptimalDepartureTime: optimalTime,
		Factors: models.FactorInfo{
			Delays:           delays,
			Weather:          *weather,
			CrowdingEstimate: s.estimateOverallCrowding(*departureTime),
		},
	}

	return response, nil
}

// optimizeRoute applies all optimization factors to a single route
func (s *OptimizerService) optimizeRoute(route *models.Route, delays []models.DelayInfo, weather *models.WeatherInfo, departureTime time.Time) {
	// Calculate delay penalty
	delayPenalty := s.calculateDelayPenalty(route, delays)
	route.Breakdown.DelayPenalty = delayPenalty

	// Calculate weather penalty
	weatherPenalty := s.calculateWeatherPenalty(route, weather)
	route.Breakdown.WeatherPenalty = weatherPenalty

	// Calculate crowding penalty
	crowdingPenalty := s.calculateCrowdingPenalty(route, departureTime)
	route.Breakdown.CrowdingPenalty = crowdingPenalty

	// Calculate transfer penalty
	transferPenalty := s.calculateTransferPenalty(route)

	// Update total time
	route.TotalTime = route.Breakdown.TransitTime +
		route.Breakdown.WalkingTime +
		route.Breakdown.WaitingTime +
		delayPenalty +
		weatherPenalty +
		crowdingPenalty +
		transferPenalty

	// Calculate score (lower total time = higher score)
	// Score is 100 - (penalty percentage)
	baseTime := route.Breakdown.TransitTime + route.Breakdown.WalkingTime + route.Breakdown.WaitingTime
	if baseTime > 0 {
		totalPenalty := delayPenalty + weatherPenalty + crowdingPenalty + transferPenalty
		penaltyPercent := (float64(totalPenalty) / float64(baseTime)) * 100
		route.Score = 100 - penaltyPercent
		if route.Score < 0 {
			route.Score = 0
		}
	}
}

// calculateDelayPenalty calculates time penalty from train delays
func (s *OptimizerService) calculateDelayPenalty(route *models.Route, delays []models.DelayInfo) int {
	penalty := 0

	// Check each transit step in the route
	for _, step := range route.Steps {
		if step.Type == "TRANSIT" && step.Line != nil {
			// Find matching delay
			for _, delay := range delays {
				if s.lineNamesMatch(step.Line.Name, delay.LineName) {
					// Add delay in seconds
					penalty += delay.DelayMinutes * 60

					// If severely delayed or suspended, add extra penalty
					if delay.Status == "SUSPENDED" {
						penalty += 600 // Extra 10 minutes
					}
					break
				}
			}
		}
	}

	return penalty
}

// calculateWeatherPenalty calculates time penalty from weather conditions
func (s *OptimizerService) calculateWeatherPenalty(route *models.Route, weather *models.WeatherInfo) int {
	// Get weather multiplier
	multiplier := GetWeatherMultiplier(weather)

	// Apply to walking time only
	walkingTime := route.Breakdown.WalkingTime
	penaltyTime := int(float64(walkingTime) * (multiplier - 1.0))

	return penaltyTime
}

// calculateCrowdingPenalty calculates time penalty from crowding
func (s *OptimizerService) calculateCrowdingPenalty(route *models.Route, departureTime time.Time) int {
	penalty := 0

	// Check each transit step
	for _, step := range route.Steps {
		if step.Type == "TRANSIT" && step.Line != nil {
			// Estimate crowding for this line
			crowding := s.crowding.EstimateCrowding(departureTime, step.Line.Name, "")
			stepPenalty := s.crowding.GetCrowdingPenalty(crowding.Level)
			penalty += stepPenalty
		}
	}

	return penalty
}

// calculateTransferPenalty calculates penalty for transfers
func (s *OptimizerService) calculateTransferPenalty(route *models.Route) int {
	transfers := 0

	// Count transit segments (each transit after the first is a transfer)
	for _, step := range route.Steps {
		if step.Type == "TRANSIT" {
			transfers++
		}
	}

	// Subtract 1 because first transit is not a transfer
	if transfers > 0 {
		transfers--
	}

	// Apply penalty per transfer
	// Assuming average transfer time of 5 minutes
	return transfers * 300 // 300 seconds = 5 minutes per transfer
}

// findOptimalDepartureTime finds the best time to leave by testing multiple options
func (s *OptimizerService) findOptimalDepartureTime(origin, destination string, baseTime time.Time, delays []models.DelayInfo, weather *models.WeatherInfo) time.Time {
	// Test departure times: now, +15, +30, +45, +60 minutes
	testTimes := []time.Duration{0, 15, 30, 45, 60}

	bestTime := baseTime
	bestScore := 0.0

	for _, offset := range testTimes {
		testTime := baseTime.Add(offset * time.Minute)

		// Get routes for this time
		routes, err := s.googleMaps.GetRoutes(origin, destination, &testTime)
		if err != nil || len(routes) == 0 {
			continue
		}

		// Optimize the best route for this time
		route := routes[0] // Take first route
		s.optimizeRoute(&route, delays, weather, testTime)

		// Use score to determine best time
		if route.Score > bestScore {
			bestScore = route.Score
			bestTime = testTime
		}
	}

	return bestTime
}

// estimateOverallCrowding estimates overall crowding at departure time
func (s *OptimizerService) estimateOverallCrowding(departureTime time.Time) models.CrowdingInfo {
	// Use a generic major line for overall estimate
	return s.crowding.EstimateCrowding(departureTime, "Yamanote Line", "")
}

// sortRoutesByScore sorts routes by score in descending order (best first)
func (s *OptimizerService) sortRoutesByScore(routes []models.Route) {
	// Simple bubble sort (sufficient for small number of routes)
	for i := 0; i < len(routes); i++ {
		for j := i + 1; j < len(routes); j++ {
			if routes[j].Score > routes[i].Score {
				routes[i], routes[j] = routes[j], routes[i]
			}
		}
	}
}

// lineNamesMatch checks if two line names refer to the same line
func (s *OptimizerService) lineNamesMatch(name1, name2 string) bool {
	n1 := strings.ToLower(strings.TrimSpace(name1))
	n2 := strings.ToLower(strings.TrimSpace(name2))

	// Direct match
	if n1 == n2 {
		return true
	}

	// Check if one contains the other
	if strings.Contains(n1, n2) || strings.Contains(n2, n1) {
		return true
	}

	// Remove common suffixes
	n1 = strings.Replace(n1, " line", "", -1)
	n2 = strings.Replace(n2, " line", "", -1)

	return n1 == n2
}
