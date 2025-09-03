package middleware

import (
	"context"

	"emx-debt-collection/internal/models"

	"github.com/labstack/echo/v4"
)

// UserService defines the interface for user operations
type UserService interface {
	GetUserByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error)
	CreateOrUpdateUser(ctx context.Context, keycloakID, email, firstName, lastName string) (*models.User, error)
}

// EnsureUserExists middleware ensures that the authenticated user exists in the database
// If the user doesn't exist, it creates a new user record based on JWT claims
func EnsureUserExists(userService UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get authenticated user from context
			authUser, ok := GetUserFromContext(c)
			if !ok {
				// No authenticated user, skip middleware
				return next(c)
			}

			ctx := c.Request().Context()

			// Check if user exists in database
			dbUser, err := userService.GetUserByKeycloakID(ctx, authUser.Sub)
			if err != nil {
				// User doesn't exist, create a new one
				firstName := authUser.GivenName
				lastName := authUser.FamilyName
				
				dbUser, err = userService.CreateOrUpdateUser(
					ctx,
					authUser.Sub,
					authUser.Email,
					firstName,
					lastName,
				)
				if err != nil {
					return echo.NewHTTPError(500, "failed to create user record")
				}
			}

			// Store database user in context for later use
			c.Set("db_user", dbUser)

			return next(c)
		}
	}
}

// GetDatabaseUserFromContext extracts the database user from the Echo context
func GetDatabaseUserFromContext(c echo.Context) (*models.User, bool) {
	user, ok := c.Get("db_user").(*models.User)
	return user, ok
}

// UserContextHelper provides helper methods for working with user context
type UserContextHelper struct{}

// NewUserContextHelper creates a new UserContextHelper instance
func NewUserContextHelper() *UserContextHelper {
	return &UserContextHelper{}
}

// GetAuthenticatedUser returns both the JWT user and database user from context
func (h *UserContextHelper) GetAuthenticatedUser(c echo.Context) (*AuthenticatedUser, *models.User, bool) {
	authUser, hasAuth := GetUserFromContext(c)
	dbUser, hasDB := GetDatabaseUserFromContext(c)
	
	if !hasAuth {
		return nil, nil, false
	}
	
	return authUser, dbUser, hasDB
}

// RequireAuthenticatedUser returns the authenticated users or returns an unauthorized error
func (h *UserContextHelper) RequireAuthenticatedUser(c echo.Context) (*AuthenticatedUser, *models.User, error) {
	authUser, dbUser, hasAuth := h.GetAuthenticatedUser(c)
	if !hasAuth {
		return nil, nil, echo.NewHTTPError(401, "user not authenticated")
	}
	
	return authUser, dbUser, nil
}

// GetUserID returns the Keycloak ID of the authenticated user
func (h *UserContextHelper) GetUserID(c echo.Context) (string, bool) {
	authUser, ok := GetUserFromContext(c)
	if !ok {
		return "", false
	}
	return authUser.Sub, true
}

// GetUserEmail returns the email of the authenticated user
func (h *UserContextHelper) GetUserEmail(c echo.Context) (string, bool) {
	authUser, ok := GetUserFromContext(c)
	if !ok {
		return "", false
	}
	return authUser.Email, true
}

// GetUserRoles returns all realm roles of the authenticated user
func (h *UserContextHelper) GetUserRoles(c echo.Context) ([]string, bool) {
	authUser, ok := GetUserFromContext(c)
	if !ok {
		return nil, false
	}
	return authUser.RealmAccess.Roles, true
}

// HasRole checks if the authenticated user has a specific realm role
func (h *UserContextHelper) HasRole(c echo.Context, role string) bool {
	authUser, ok := GetUserFromContext(c)
	if !ok {
		return false
	}
	return authUser.HasRole(role)
}

// HasClientRole checks if the authenticated user has a specific client role
func (h *UserContextHelper) HasClientRole(c echo.Context, clientID, role string) bool {
	authUser, ok := GetUserFromContext(c)
	if !ok {
		return false
	}
	return authUser.HasClientRole(clientID, role)
}

// IsUserActive checks if the user is active in the database
func (h *UserContextHelper) IsUserActive(c echo.Context) bool {
	_, dbUser, hasDB := h.GetAuthenticatedUser(c)
	if !hasDB {
		return false
	}
	return dbUser.IsActive
}