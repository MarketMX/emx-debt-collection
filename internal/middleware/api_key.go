package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// RequireAPIKey middleware validates API key for provisioning endpoints
func RequireAPIKey() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get API key from environment
			expectedAPIKey := os.Getenv("PROVISIONING_API_KEY")
			if expectedAPIKey == "" {
				// If no API key is set, log warning but allow access (development mode)
				c.Logger().Warn("PROVISIONING_API_KEY not set - API key validation disabled")
				return next(c)
			}

			// Check Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "API key required")
			}

			// Support both "Bearer" and "ApiKey" prefix
			var providedKey string
			if strings.HasPrefix(authHeader, "Bearer ") {
				providedKey = strings.TrimPrefix(authHeader, "Bearer ")
			} else if strings.HasPrefix(authHeader, "ApiKey ") {
				providedKey = strings.TrimPrefix(authHeader, "ApiKey ")
			} else {
				// Also check X-API-Key header as alternative
				providedKey = c.Request().Header.Get("X-API-Key")
				if providedKey == "" {
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization format. Use 'Bearer <key>', 'ApiKey <key>', or 'X-API-Key' header")
				}
			}

			// Validate API key
			if providedKey != expectedAPIKey {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}

			return next(c)
		}
	}
}

// RequireAPIKeyOrJWT allows either API key or JWT authentication
func RequireAPIKeyOrJWT(jwtMiddleware echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if API key is provided
			authHeader := c.Request().Header.Get("Authorization")
			apiKey := c.Request().Header.Get("X-API-Key")
			
			if apiKey != "" || strings.HasPrefix(authHeader, "ApiKey ") {
				// Use API key validation
				return RequireAPIKey()(next)(c)
			}
			
			// Use JWT validation
			return jwtMiddleware(next)(c)
		}
	}
}