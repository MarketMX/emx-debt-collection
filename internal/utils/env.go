package utils

import (
	"os"
	"strconv"
	"time"
)

// GetEnv gets environment variable with fallback to default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetIntEnv gets integer environment variable with fallback to default value
func GetIntEnv(key string, defaultValue int) int {
	if str := os.Getenv(key); str != "" {
		if value, err := strconv.Atoi(str); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetBoolEnv gets boolean environment variable with fallback to default value
func GetBoolEnv(key string, defaultValue bool) bool {
	if str := os.Getenv(key); str != "" {
		if value, err := strconv.ParseBool(str); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetDurationEnv gets duration environment variable with fallback to default value
func GetDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if str := os.Getenv(key); str != "" {
		if value, err := time.ParseDuration(str); err == nil {
			return value
		}
	}
	return defaultValue
}