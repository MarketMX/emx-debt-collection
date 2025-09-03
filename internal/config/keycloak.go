package config

import (
	"fmt"

	"emx-debt-collection/internal/utils"
)

// KeycloakConfig holds the configuration for Keycloak integration
type KeycloakConfig struct {
	RealmURL     string
	ClientID     string
	ClientSecret string
	Realm        string
	ServerURL    string
}

// LoadKeycloakConfig loads Keycloak configuration from environment variables
func LoadKeycloakConfig() *KeycloakConfig {
	serverURL := utils.GetEnv("KEYCLOAK_SERVER_URL", "http://localhost:8081")
	realm := utils.GetEnv("KEYCLOAK_REALM", "debt-collection")
	clientID := utils.GetEnv("KEYCLOAK_CLIENT_ID", "debt-collection-backend")
	clientSecret := utils.GetEnv("KEYCLOAK_CLIENT_SECRET", "")

	realmURL := fmt.Sprintf("%s/realms/%s", serverURL, realm)

	return &KeycloakConfig{
		RealmURL:     realmURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Realm:        realm,
		ServerURL:    serverURL,
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

