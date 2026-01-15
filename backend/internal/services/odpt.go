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
	odptBaseURL     = "https://api-tokyochallenge.odpt.org/api/v4"
	delayCacheTTL   = 2 * time.Minute
	delayCacheKey   = "delays:all"
)

// ODPTService handles ODPT API operations
type ODPTService struct {
	apiKey     string
	httpClient *http.Client
}

// ODPTTrainInformation represents train information from ODPT API
type ODPTTrainInformation struct {
	ID              string    `json:"@id"`
	Type            string    `json:"@type"`
	DCDate          time.Time `json:"dc:date"`
	Valid           time.Time `json:"dct:valid"`
	Operator        string    `json:"odpt:operator"`
	Railway         string    `json:"odpt:railway"`
	TimeOfOrigin    time.Time `json:"odpt:timeOfOrigin"`
	TrainInformationText struct {
		En string `json:"en"`
		Ja string `json:"ja"`
	} `json:"odpt:trainInformationText"`
	TrainInformationStatus string `json:"odpt:trainInformationStatus,omitempty"`
}

// NewODPTService creates a new ODPT service
func NewODPTService() *ODPTService {
	return &ODPTService{
		apiKey: os.Getenv("ODPT_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDelays fetches current train delays from ODPT API
func (s *ODPTService) GetDelays() ([]models.DelayInfo, error) {
	// If no API key is configured, return mock data
	if s.apiKey == "" {
		return s.getMockDelays(), nil
	}

	// Try to get from cache first
	var cachedDelays []models.DelayInfo
	err := cache.Get(delayCacheKey, &cachedDelays)
	if err == nil && len(cachedDelays) > 0 {
		return cachedDelays, nil
	}

	// If not in cache, fetch from API
	delays, err := s.fetchDelaysFromAPI()
	if err != nil {
		// If API fails, return mock data instead of failing
		fmt.Printf("ODPT API error (using mock data): %v\n", err)
		return s.getMockDelays(), nil
	}

	// Cache the result
	if err := cache.Set(delayCacheKey, delays, delayCacheTTL); err != nil {
		fmt.Printf("Failed to cache delay data: %v\n", err)
	}

	return delays, nil
}

// fetchDelaysFromAPI fetches delay data from ODPT API
func (s *ODPTService) fetchDelaysFromAPI() ([]models.DelayInfo, error) {
	// Build URL
	apiURL := fmt.Sprintf("%s/odpt:TrainInformation", odptBaseURL)
	params := url.Values{}
	params.Add("acl:consumerKey", s.apiKey)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Make request
	resp, err := s.httpClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ODPT data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ODPT API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var trainInfos []ODPTTrainInformation
	if err := json.NewDecoder(resp.Body).Decode(&trainInfos); err != nil {
		return nil, fmt.Errorf("failed to decode ODPT response: %w", err)
	}

	// Convert to our model
	delays := make([]models.DelayInfo, 0)
	for _, info := range trainInfos {
		// Only include if there's actually a delay or issue
		if info.TrainInformationStatus != "" {
			delay := models.DelayInfo{
				LineName:     extractLineName(info.Railway),
				DelayMinutes: estimateDelayMinutes(info.TrainInformationText.En),
				Status:       mapStatus(info.TrainInformationStatus),
				Message:      info.TrainInformationText.En,
			}
			delays = append(delays, delay)
		}
	}

	return delays, nil
}

// extractLineName extracts a readable line name from ODPT railway identifier
func extractLineName(railway string) string {
	// Railway format is like "odpt.Railway:JR-East.ChuoRapid"
	parts := strings.Split(railway, ":")
	if len(parts) < 2 {
		return railway
	}

	// Get the last part and format it
	lastPart := parts[len(parts)-1]
	parts = strings.Split(lastPart, ".")
	if len(parts) < 2 {
		return lastPart
	}

	// Format: "JR-East.ChuoRapid" -> "Chuo Rapid Line"
	lineName := parts[len(parts)-1]

	// Add spaces before capitals
	result := ""
	for i, r := range lineName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += " "
		}
		result += string(r)
	}

	return result + " Line"
}

// estimateDelayMinutes tries to extract delay minutes from the message
func estimateDelayMinutes(message string) int {
	// Simple heuristic - look for common patterns
	message = strings.ToLower(message)

	if strings.Contains(message, "suspended") || strings.Contains(message, "stopped") {
		return 30 // Major delay
	}

	if strings.Contains(message, "major delay") {
		return 15
	}

	if strings.Contains(message, "delay") {
		return 5 // Minor delay
	}

	return 0
}

// mapStatus maps ODPT status to our status enum
func mapStatus(odptStatus string) string {
	switch odptStatus {
	case "運転見合わせ": // Service suspended
		return "SUSPENDED"
	case "遅延": // Delayed
		return "DELAYED"
	default:
		return "NORMAL"
	}
}

// getMockDelays returns mock delay data for development/testing
func (s *ODPTService) getMockDelays() []models.DelayInfo {
	return []models.DelayInfo{
		{
			LineName:     "Chuo Line",
			DelayMinutes: 3,
			Status:       "DELAYED",
			Message:      "Minor delays due to earlier signal problems",
		},
		{
			LineName:     "Yamanote Line",
			DelayMinutes: 0,
			Status:       "NORMAL",
			Message:      "Service operating normally",
		},
	}
}

// GetDelayForLine gets delay information for a specific line
func (s *ODPTService) GetDelayForLine(lineName string) (*models.DelayInfo, error) {
	delays, err := s.GetDelays()
	if err != nil {
		return nil, err
	}

	// Find matching line
	for _, delay := range delays {
		if strings.Contains(strings.ToLower(delay.LineName), strings.ToLower(lineName)) {
			return &delay, nil
		}
	}

	// No delay found, return normal status
	return &models.DelayInfo{
		LineName:     lineName,
		DelayMinutes: 0,
		Status:       "NORMAL",
		Message:      "Service operating normally",
	}, nil
}
