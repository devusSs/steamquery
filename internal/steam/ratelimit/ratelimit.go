package ratelimit

import (
	"encoding/json"
	"os"
	"time"
)

var requestsPerMinute = 15
var resetInterval = time.Minute
var rateLimitFileName = "./.s_ratelimit.json"

type RateLimitState struct {
	LastRequestTime time.Time `json:"lastRequestTime"`
	RequestCount    int       `json:"requestCount"`
}

func SetRateLimitConfig(dir string, requests int, reset time.Duration) {
	requestsPerMinute = requests
	resetInterval = reset
	rateLimitFileName = dir + "/" + ".s_ratelimit.json"
}

func WithinRateLimit(state *RateLimitState, potential int) bool {
	now := time.Now()
	if now.Sub(state.LastRequestTime) > resetInterval {
		state.RequestCount = 0
	}

	return state.RequestCount+potential < requestsPerMinute
}

func LoadRateLimitState() (*RateLimitState, error) {
	state := RateLimitState{}
	if _, err := os.Stat(rateLimitFileName); os.IsNotExist(err) {
		return &state, nil
	}

	file, err := os.ReadFile(rateLimitFileName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func SaveRateLimitState(state *RateLimitState) error {
	file, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(rateLimitFileName, file, 0644)
	if err != nil {
		return err
	}

	return nil
}
