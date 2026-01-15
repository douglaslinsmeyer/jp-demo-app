package models

import "time"

// RouteRequest represents the request payload for route calculation
type RouteRequest struct {
	Origin        string     `json:"origin" binding:"required"`
	Destination   string     `json:"destination" binding:"required"`
	DepartureTime *time.Time `json:"departureTime,omitempty"`
}

// RouteResponse represents the response containing route options
type RouteResponse struct {
	Routes               []Route       `json:"routes"`
	OptimalDepartureTime time.Time     `json:"optimalDepartureTime"`
	Factors              FactorInfo    `json:"factors"`
}

// Route represents a single route option
type Route struct {
	Summary   string        `json:"summary"`
	TotalTime int           `json:"totalTime"` // in seconds
	Breakdown TimeBreakdown `json:"breakdown"`
	Steps     []RouteStep   `json:"steps"`
	Score     float64       `json:"score"`
}

// TimeBreakdown shows the breakdown of time spent
type TimeBreakdown struct {
	TransitTime     int `json:"transitTime"`     // in seconds
	WalkingTime     int `json:"walkingTime"`     // in seconds
	WaitingTime     int `json:"waitingTime"`     // in seconds
	DelayPenalty    int `json:"delayPenalty"`    // in seconds
	WeatherPenalty  int `json:"weatherPenalty"`  // in seconds
	CrowdingPenalty int `json:"crowdingPenalty"` // in seconds
}

// RouteStep represents a step in the journey
type RouteStep struct {
	Type         string     `json:"type"` // "WALK", "TRANSIT", "WAIT"
	Instructions string     `json:"instructions"`
	Distance     int        `json:"distance,omitempty"`     // in meters
	Duration     int        `json:"duration"`               // in seconds
	Line         *LineInfo  `json:"line,omitempty"`         // for TRANSIT steps
	DepartAt     *time.Time `json:"departAt,omitempty"`     // for TRANSIT steps
	ArriveAt     *time.Time `json:"arriveAt,omitempty"`     // for TRANSIT steps
}

// LineInfo represents transit line information
type LineInfo struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Color     string `json:"color,omitempty"`
	Vehicle   string `json:"vehicle"` // "SUBWAY", "TRAIN", "BUS"
}

// FactorInfo contains information about factors affecting the route
type FactorInfo struct {
	Delays           []DelayInfo    `json:"delays"`
	Weather          WeatherInfo    `json:"weather"`
	CrowdingEstimate CrowdingInfo   `json:"crowdingEstimate"`
}

// DelayInfo represents delay information for a transit line
type DelayInfo struct {
	LineName     string `json:"lineName"`
	DelayMinutes int    `json:"delayMinutes"`
	Status       string `json:"status"` // "NORMAL", "DELAYED", "SUSPENDED"
	Message      string `json:"message,omitempty"`
}

// WeatherInfo represents weather conditions
type WeatherInfo struct {
	Temperature    float64 `json:"temperature"`    // in Celsius
	Precipitation  float64 `json:"precipitation"`  // in mm/hour
	Condition      string  `json:"condition"`      // "CLEAR", "RAIN", "SNOW", etc.
	LastUpdated    time.Time `json:"lastUpdated"`
}

// CrowdingInfo represents estimated crowding levels
type CrowdingInfo struct {
	Level       string `json:"level"`       // "LOW", "MODERATE", "HIGH", "VERY_HIGH"
	Description string `json:"description"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
