package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"emx-debt-collection/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// UserContextKey is the key used to store user data in the request context
const UserContextKey = "user"

// AuthenticatedUser represents the authenticated user data extracted from JWT
type AuthenticatedUser struct {
	Sub               string   `json:"sub"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	Exp int64 `json:"exp"`
	Iat int64 `json:"iat"`
}

// HasRole checks if the user has a specific realm role
func (u *AuthenticatedUser) HasRole(role string) bool {
	for _, r := range u.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasClientRole checks if the user has a specific client role
func (u *AuthenticatedUser) HasClientRole(clientID, role string) bool {
	if clientAccess, exists := u.ResourceAccess[clientID]; exists {
		for _, r := range clientAccess.Roles {
			if r == role {
				return true
			}
		}
	}
	return false
}

// JWTMiddleware provides JWT authentication middleware
type JWTMiddleware struct {
	keycloakConfig *config.KeycloakConfig
	publicKeys     map[string]*rsa.PublicKey
	keysMutex      sync.RWMutex
	lastKeyFetch   time.Time
	keysCacheTTL   time.Duration
}

// NewJWTMiddleware creates a new JWT middleware instance
func NewJWTMiddleware(keycloakConfig *config.KeycloakConfig) *JWTMiddleware {
	return &JWTMiddleware{
		keycloakConfig: keycloakConfig,
		publicKeys:     make(map[string]*rsa.PublicKey),
		keysCacheTTL:   5 * time.Minute, // Cache keys for 5 minutes
	}
}

// RequireAuth is a middleware that requires valid JWT authentication
func (jm *JWTMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Check for Bearer token format
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := tokenParts[1]

			// Parse and validate the JWT token
			user, err := jm.validateToken(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("invalid token: %v", err))
			}

			// Check if token is expired
			if time.Now().Unix() > user.Exp {
				return echo.NewHTTPError(http.StatusUnauthorized, "token expired")
			}

			// Store user data in context
			c.Set(UserContextKey, user)

			return next(c)
		}
	}
}

// OptionalAuth is a middleware that extracts user info if a valid JWT is present
func (jm *JWTMiddleware) OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				tokenParts := strings.Split(authHeader, " ")
				if len(tokenParts) == 2 && strings.ToLower(tokenParts[0]) == "bearer" {
					tokenString := tokenParts[1]
					
					// Try to validate the token, but don't fail if it's invalid
					if user, err := jm.validateToken(tokenString); err == nil {
						if time.Now().Unix() <= user.Exp {
							c.Set(UserContextKey, user)
						}
					}
				}
			}

			return next(c)
		}
	}
}

// RequireRole is a middleware that requires the user to have a specific realm role
func (jm *JWTMiddleware) RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get(UserContextKey).(*AuthenticatedUser)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
			}

			if !user.HasRole(role) {
				return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("requires role: %s", role))
			}

			return next(c)
		}
	}
}

// RequireClientRole is a middleware that requires the user to have a specific client role
func (jm *JWTMiddleware) RequireClientRole(clientID, role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get(UserContextKey).(*AuthenticatedUser)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
			}

			if !user.HasClientRole(clientID, role) {
				return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("requires client role: %s in %s", role, clientID))
			}

			return next(c)
		}
	}
}

// validateToken validates a JWT token against Keycloak's public keys
func (jm *JWTMiddleware) validateToken(tokenString string) (*AuthenticatedUser, error) {
	// Parse the token without verification first to get the kid (key ID)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Get the public key for this kid
		publicKey, err := jm.getPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %v", err)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims and convert to our user structure
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Marshal claims to JSON and then unmarshal to our struct for type safety
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claims: %v", err)
	}

	var user AuthenticatedUser
	if err := json.Unmarshal(claimsJSON, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user claims: %v", err)
	}

	return &user, nil
}

// getPublicKey retrieves the public key for the given key ID from Keycloak
func (jm *JWTMiddleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check if we have a cached key and it's still valid
	jm.keysMutex.RLock()
	if key, exists := jm.publicKeys[kid]; exists && time.Since(jm.lastKeyFetch) < jm.keysCacheTTL {
		jm.keysMutex.RUnlock()
		log.Printf("[JWT] Using cached key for kid: %s", kid)
		return key, nil
	}
	jm.keysMutex.RUnlock()

	log.Printf("[JWT] Key not found in cache for kid: %s, fetching from Keycloak...", kid)
	// Fetch keys from Keycloak if cache is expired or key doesn't exist
	if err := jm.fetchPublicKeys(); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}

	// Try to get the key again after fetching
	jm.keysMutex.RLock()
	defer jm.keysMutex.RUnlock()

	// List all available keys for debugging
	log.Printf("[JWT] Available keys after fetch:")
	for k := range jm.publicKeys {
		log.Printf("[JWT]   - kid: %s", k)
	}

	key, exists := jm.publicKeys[kid]
	if !exists {
		return nil, fmt.Errorf("public key not found for kid: %s (have %d keys cached)", kid, len(jm.publicKeys))
	}

	return key, nil
}


// fetchPublicKeys fetches public keys from Keycloak's JWKS endpoint
func (jm *JWTMiddleware) fetchPublicKeys() error {
	jwksURL := jm.keycloakConfig.JWKSEndpoint()
	log.Printf("[JWT] Fetching JWKS from: %s", jwksURL)
	
	// Fetch the JWK set from Keycloak
	keySet, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS from %s: %v", jwksURL, err)
	}

	// Parse and cache the public keys
	jm.keysMutex.Lock()
	defer jm.keysMutex.Unlock()

	// Clear existing keys
	jm.publicKeys = make(map[string]*rsa.PublicKey)

	// Iterate through the key set
	keyCount := 0
	for it := keySet.Keys(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)
		keyCount++
		
		log.Printf("[JWT] Processing key %d: Type=%s, Use=%s, Kid=%s", keyCount, key.KeyType(), key.KeyUsage(), key.KeyID())

		// Only process RSA keys for signature verification
		if key.KeyType() == "RSA" {
			keyID := key.KeyID()
			
			// Skip keys without a key ID
			if keyID == "" {
				log.Printf("[JWT] Skipping key without kid")
				continue
			}
			
			// Convert JWK to RSA public key
			var rawKey interface{}
			if err := key.Raw(&rawKey); err != nil {
				log.Printf("[JWT] Failed to extract raw key %s: %v", keyID, err)
				continue // Skip invalid keys
			}
			
			publicKey, ok := rawKey.(*rsa.PublicKey)
			if !ok {
				log.Printf("[JWT] Key %s is not an RSA public key", keyID)
				continue
			}
			
			jm.publicKeys[keyID] = publicKey
			log.Printf("[JWT] Successfully cached public key with kid: %s", keyID)
		} else {
			log.Printf("[JWT] Skipping non-RSA key: %s", key.KeyType())
		}
	}

	log.Printf("[JWT] Total keys cached: %d", len(jm.publicKeys))
	jm.lastKeyFetch = time.Now()
	return nil
}

// GetUserFromContext extracts the authenticated user from the Echo context
func GetUserFromContext(c echo.Context) (*AuthenticatedUser, bool) {
	user, ok := c.Get(UserContextKey).(*AuthenticatedUser)
	return user, ok
}

// GetUserFromGoContext extracts the authenticated user from a standard Go context
func GetUserFromGoContext(ctx context.Context) (*AuthenticatedUser, bool) {
	user, ok := ctx.Value(UserContextKey).(*AuthenticatedUser)
	return user, ok
}