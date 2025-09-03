package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"emx-debt-collection/internal/config"
	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/middleware"
	"emx-debt-collection/internal/models"
)

type Server struct {
	port           int
	db             database.Service
	jwtMiddleware  *middleware.JWTMiddleware
	userHelper     *middleware.UserContextHelper
	keycloakConfig *config.KeycloakConfig
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	
	// Load Keycloak configuration
	keycloakConfig := config.LoadKeycloakConfig()
	
	// Initialize middleware components
	jwtMiddleware := middleware.NewJWTMiddleware(keycloakConfig)
	userHelper := middleware.NewUserContextHelper()
	
	NewServer := &Server{
		port:           port,
		db:             database.New(),
		jwtMiddleware:  jwtMiddleware,
		userHelper:     userHelper,
		keycloakConfig: keycloakConfig,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

// UserService implementation for middleware
func (s *Server) GetUserByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error) {
	return s.db.Repository().GetUserByKeycloakID(ctx, keycloakID)
}

func (s *Server) CreateOrUpdateUser(ctx context.Context, keycloakID, email, firstName, lastName string) (*models.User, error) {
	return s.db.Repository().CreateOrUpdateUser(ctx, keycloakID, email, firstName, lastName)
}
