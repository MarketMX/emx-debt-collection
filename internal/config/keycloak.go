package config

import (
	"fmt"

	"emx-debt-collection/internal/utils"
)

// KeycloakConfig holds the configuration for Keycloak integration
type KeycloakConfig struct {
	RealmURL        string
	ClientID        string
	ClientSecret    string
	Realm           string
	ServerURL       string    // Internal backend-to-Keycloak URL
	FrontendURL     string    // Frontend-to-Keycloak URL (for browser access)
	FrontendRealmURL string   // Frontend realm URL
}

// LoadKeycloakConfig loads Keycloak configuration from environment variables
func LoadKeycloakConfig() *KeycloakConfig {
	serverURL := utils.GetEnv("KEYCLOAK_SERVER_URL", "http://localhost:8081")
	frontendURL := utils.GetEnv("KEYCLOAK_FRONTEND_URL", "http://localhost:8081") // Default to localhost for frontend
	realm := utils.GetEnv("KEYCLOAK_REALM", "debt-collection")
	clientID := utils.GetEnv("KEYCLOAK_CLIENT_ID", "debt-collection-backend")
	clientSecret := utils.GetEnv("KEYCLOAK_CLIENT_SECRET", "")

	realmURL := fmt.Sprintf("%s/realms/%s", serverURL, realm)
	frontendRealmURL := fmt.Sprintf("%s/realms/%s", frontendURL, realm)

	return &KeycloakConfig{
		RealmURL:         realmURL,
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		Realm:            realm,
		ServerURL:        serverURL,
		FrontendURL:      frontendURL,
		FrontendRealmURL: frontendRealmURL,
	}
}

// JWKSEndpoint returns the JWKS endpoint URL for token verification
func (k *KeycloakConfig) JWKSEndpoint() string {
	return fmt.Sprintf("%s/protocol/openid-connect/certs", k.RealmURL)
}

// UserInfoEndpoint returns the user info endpoint URL
func (k *KeycloakConfig) UserInfoEndpoint() string {
	return fmt.Sprintf("%s/protocol/openid-connect/userinfo", k.RealmURL)
}

// TokenEndpoint returns the token endpoint URL
func (k *KeycloakConfig) TokenEndpoint() string {
	return fmt.Sprintf("%s/protocol/openid-connect/token", k.RealmURL)
}

// Frontend endpoints for browser access
func (k *KeycloakConfig) FrontendTokenEndpoint() string {
	return fmt.Sprintf("%s/protocol/openid-connect/token", k.FrontendRealmURL)
}

func (k *KeycloakConfig) FrontendUserInfoEndpoint() string {
	return fmt.Sprintf("%s/protocol/openid-connect/userinfo", k.FrontendRealmURL)
}

