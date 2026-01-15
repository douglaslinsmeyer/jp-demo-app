package services

import (
	"strings"
	"time"

	"github.com/douglasl/tokyo-commute-optimizer/internal/models"
)

// CrowdingService handles crowding estimation
type CrowdingService struct{}

// NewCrowdingService creates a new crowding service
func NewCrowdingService() *CrowdingService {
	return &CrowdingService{}
}

// EstimateCrowding estimates crowding level based on time, day, and route
func (s *CrowdingService) EstimateCrowding(departureTime time.Time, lineName string, direction string) models.CrowdingInfo {
	level := s.calculateCrowdingLevel(departureTime, lineName, direction)

	return models.CrowdingInfo{
		Level:       level,
		Description: s.getCrowdingDescription(level),
	}
}

// calculateCrowdingLevel calculates the crowding level
func (s *CrowdingService) calculateCrowdingLevel(departureTime time.Time, lineName string, direction string) string {
	score := 0.0

	// Factor 1: Time of day (most important)
	hour := departureTime.Hour()
	minute := departureTime.Minute()
	timeScore := s.getTimeScore(hour, minute)
	score += timeScore * 0.5 // 50% weight

	// Factor 2: Day of week
	dayScore := s.getDayScore(departureTime.Weekday())
	score += dayScore * 0.2 // 20% weight

	// Factor 3: Line popularity
	lineScore := s.getLineScore(lineName)
	score += lineScore * 0.2 // 20% weight

	// Factor 4: Direction (inbound vs outbound)
	directionScore := s.getDirectionScore(direction, hour)
	score += directionScore * 0.1 // 10% weight

	// Convert score to level
	switch {
	case score < 0.25:
		return "LOW"
	case score < 0.5:
		return "MODERATE"
	case score < 0.75:
		return "HIGH"
	default:
		return "VERY_HIGH"
	}
}

// getTimeScore returns a crowding score based on time of day (0-1)
func (s *CrowdingService) getTimeScore(hour, minute int) float64 {
	timeInMinutes := hour*60 + minute

	// Morning rush hour: 7:00-9:30 (peak at 8:00)
	if timeInMinutes >= 420 && timeInMinutes <= 570 { // 7:00-9:30
		// Peak at 8:00 (480 minutes)
		if timeInMinutes >= 450 && timeInMinutes <= 510 { // 7:30-8:30
			return 1.0 // Maximum crowding
		} else if timeInMinutes >= 420 && timeInMinutes < 450 { // 7:00-7:30
			return 0.7 + (float64(timeInMinutes-420) / 30.0 * 0.3) // Ramp up
		} else { // 8:30-9:30
			return 1.0 - (float64(timeInMinutes-510) / 60.0 * 0.5) // Ramp down
		}
	}

	// Evening rush hour: 17:00-20:00 (peak at 18:30)
	if timeInMinutes >= 1020 && timeInMinutes <= 1200 { // 17:00-20:00
		// Peak at 18:30 (1110 minutes)
		if timeInMinutes >= 1080 && timeInMinutes <= 1140 { // 18:00-19:00
			return 0.9 // High crowding (slightly less than morning)
		} else if timeInMinutes >= 1020 && timeInMinutes < 1080 { // 17:00-18:00
			return 0.5 + (float64(timeInMinutes-1020) / 60.0 * 0.4) // Ramp up
		} else { // 19:00-20:00
			return 0.9 - (float64(timeInMinutes-1140) / 60.0 * 0.5) // Ramp down
		}
	}

	// Lunch time: 12:00-13:30
	if timeInMinutes >= 720 && timeInMinutes <= 810 { // 12:00-13:30
		return 0.4 // Moderate crowding
	}

	// Late night: 22:00-24:00
	if timeInMinutes >= 1320 { // After 22:00
		return 0.2 // Low crowding
	}

	// Early morning: 5:00-7:00
	if timeInMinutes >= 300 && timeInMinutes < 420 { // 5:00-7:00
		return 0.2 // Low crowding
	}

	// Off-peak hours
	return 0.3
}

// getDayScore returns a crowding score based on day of week (0-1)
func (s *CrowdingService) getDayScore(weekday time.Weekday) float64 {
	switch weekday {
	case time.Monday:
		return 1.0 // Busiest day
	case time.Tuesday, time.Wednesday, time.Thursday:
		return 0.9 // Busy
	case time.Friday:
		return 0.85 // Slightly less busy
	case time.Saturday:
		return 0.4 // Weekend - much less crowded
	case time.Sunday:
		return 0.3 // Least crowded
	default:
		return 0.5
	}
}

// getLineScore returns a crowding score based on line popularity (0-1)
func (s *CrowdingService) getLineScore(lineName string) float64 {
	lineLower := strings.ToLower(lineName)

	// Very busy lines
	if strings.Contains(lineLower, "yamanote") {
		return 1.0 // Busiest line in Tokyo
	}
	if strings.Contains(lineLower, "chuo") && strings.Contains(lineLower, "rapid") {
		return 0.95 // Very busy
	}
	if strings.Contains(lineLower, "tozai") {
		return 0.95 // Tokyo Metro's busiest line
	}

	// Busy lines
	if strings.Contains(lineLower, "sobu") ||
		strings.Contains(lineLower, "keihin-tohoku") ||
		strings.Contains(lineLower, "hanzomon") ||
		strings.Contains(lineLower, "hibiya") {
		return 0.85
	}

	// Moderate lines
	if strings.Contains(lineLower, "ginza") ||
		strings.Contains(lineLower, "marunouchi") ||
		strings.Contains(lineLower, "odakyu") ||
		strings.Contains(lineLower, "keio") {
		return 0.7
	}

	// Less crowded lines
	if strings.Contains(lineLower, "namboku") ||
		strings.Contains(lineLower, "yurakucho") ||
		strings.Contains(lineLower, "fukutoshin") {
		return 0.5
	}

	// Default for unknown lines
	return 0.6
}

// getDirectionScore returns a crowding score based on direction and time (0-1)
func (s *CrowdingService) getDirectionScore(direction string, hour int) float64 {
	dirLower := strings.ToLower(direction)

	// Morning rush (7-10): Inbound trains are more crowded
	if hour >= 7 && hour < 10 {
		if strings.Contains(dirLower, "inbound") ||
			strings.Contains(dirLower, "toward") && strings.Contains(dirLower, "tokyo") {
			return 1.0 // Very crowded inbound
		}
		return 0.3 // Less crowded outbound
	}

	// Evening rush (17-20): Outbound trains are more crowded
	if hour >= 17 && hour < 20 {
		if strings.Contains(dirLower, "outbound") ||
			strings.Contains(dirLower, "from") && strings.Contains(dirLower, "tokyo") {
			return 1.0 // Very crowded outbound
		}
		return 0.4 // Less crowded inbound
	}

	// Off-peak: No significant difference
	return 0.5
}

// getCrowdingDescription returns a human-readable description
func (s *CrowdingService) getCrowdingDescription(level string) string {
	switch level {
	case "LOW":
		return "Seats likely available, comfortable travel expected"
	case "MODERATE":
		return "Some crowding expected, standing room available"
	case "HIGH":
		return "Crowded trains, limited standing room"
	case "VERY_HIGH":
		return "Very crowded, packed trains during rush hour"
	default:
		return "Crowding level unknown"
	}
}

// GetCrowdingPenalty returns time penalty in seconds based on crowding level
func (s *CrowdingService) GetCrowdingPenalty(level string) int {
	switch level {
	case "LOW":
		return 0
	case "MODERATE":
		return 60 // 1 minute penalty
	case "HIGH":
		return 180 // 3 minutes penalty
	case "VERY_HIGH":
		return 300 // 5 minutes penalty
	default:
		return 0
	}
}
