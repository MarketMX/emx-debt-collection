package config

import (
	"fmt"
	"time"

	"emx-debt-collection/internal/utils"
)

// MessagingConfig contains configuration for the messaging service
type MessagingConfig struct {
	APIUrl           string        `json:"api_url"`
	APIKey           string        `json:"-"` // Don't expose in JSON
	Timeout          time.Duration `json:"timeout"`
	RateLimit        int           `json:"rate_limit"`          // Messages per second
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	DefaultFrom      string        `json:"default_from"`
	EnableSimulation bool          `json:"enable_simulation"`   // For testing/development
}

// LoadMessagingConfig loads messaging configuration from environment variables
func LoadMessagingConfig() *MessagingConfig {
	config := &MessagingConfig{
		APIUrl:           utils.GetEnv("MESSAGING_API_URL", "https://api.messaging.example.com"),
		APIKey:           utils.GetEnv("MESSAGING_API_KEY", ""),
		Timeout:          utils.GetDurationEnv("MESSAGING_TIMEOUT", 30*time.Second),
		RateLimit:        utils.GetIntEnv("MESSAGING_RATE_LIMIT", 5), // 5 messages per second
		MaxRetries:       utils.GetIntEnv("MESSAGING_MAX_RETRIES", 3),
		RetryDelay:       utils.GetDurationEnv("MESSAGING_RETRY_DELAY", 5*time.Second),
		DefaultFrom:      utils.GetEnv("MESSAGING_FROM", "DebtCollection"),
		EnableSimulation: utils.GetBoolEnv("MESSAGING_SIMULATION", true), // Default to simulation for development
	}
	
	return config
}

// Validate validates the messaging configuration
func (c *MessagingConfig) Validate() error {
	if c.APIUrl == "" {
		return fmt.Errorf("MESSAGING_API_URL is required")
	}
	
	if !c.EnableSimulation && c.APIKey == "" {
		return fmt.Errorf("MESSAGING_API_KEY is required when simulation is disabled")
	}
	
	if c.RateLimit <= 0 {
		return fmt.Errorf("MESSAGING_RATE_LIMIT must be positive")
	}
	
	if c.MaxRetries < 0 {
		return fmt.Errorf("MESSAGING_MAX_RETRIES must be non-negative")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("MESSAGING_TIMEOUT must be positive")
	}
	
	return nil
}

