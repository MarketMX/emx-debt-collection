package handlers

import (
	"net/http"

	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ProvisioningHandler handles user provisioning from external systems (Django Admin)
type ProvisioningHandler struct {
	repo database.Repository
}

// NewProvisioningHandler creates a new provisioning handler
func NewProvisioningHandler(repo database.Repository) *ProvisioningHandler {
	return &ProvisioningHandler{
		repo: repo,
	}
}

// ProvisionUserRequest represents a user provisioning request from Django Admin
type ProvisionUserRequest struct {
	KeycloakID       string  `json:"keycloak_id" validate:"required"`
	Email            string  `json:"email" validate:"required,email"`
	FirstName        *string `json:"first_name"`
	LastName         *string `json:"last_name"`
	EngageMXClientID string  `json:"engagemx_client_id" validate:"required"`
	IsActive         *bool   `json:"is_active"`
}

// ProvisionUserResponse represents the response for user provisioning
type ProvisionUserResponse struct {
	ID               uuid.UUID `json:"id"`
	KeycloakID       string    `json:"keycloak_id"`
	Email            string    `json:"email"`
	FirstName        *string   `json:"first_name"`
	LastName         *string   `json:"last_name"`
	EngageMXClientID string    `json:"engagemx_client_id"`
	IsActive         bool      `json:"is_active"`
	Action           string    `json:"action"` // "created" or "updated"
	Message          string    `json:"message"`
}

// ProvisionUser creates or updates a user from Django Admin
func (h *ProvisioningHandler) ProvisionUser(c echo.Context) error {
	var req ProvisionUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// Check if user already exists
	existingUser, err := h.repo.GetUserByKeycloakID(c.Request().Context(), req.KeycloakID)
	
	var user *models.User
	var action string

	if err != nil {
		// User doesn't exist, create new one
		createReq := models.CreateUserRequest{
			KeycloakID:       req.KeycloakID,
			Email:            req.Email,
			FirstName:        req.FirstName,
			LastName:         req.LastName,
			EngageMXClientID: req.EngageMXClientID,
		}

		user, err = h.repo.CreateUser(c.Request().Context(), createReq)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user: "+err.Error())
		}
		action = "created"
	} else {
		// User exists, update if needed
		updateReq := models.UpdateUserRequest{
			Email:            &req.Email,
			FirstName:        req.FirstName,
			LastName:         req.LastName,
			EngageMXClientID: &req.EngageMXClientID,
		}

		if req.IsActive != nil {
			updateReq.IsActive = req.IsActive
		}

		user, err = h.repo.UpdateUser(c.Request().Context(), existingUser.ID, updateReq)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user: "+err.Error())
		}
		action = "updated"
	}

	response := ProvisionUserResponse{
		ID:               user.ID,
		KeycloakID:       user.KeycloakID,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		EngageMXClientID: user.EngageMXClientID,
		IsActive:         user.IsActive,
		Action:           action,
		Message:          "User successfully " + action,
	}

	return c.JSON(http.StatusOK, response)
}

// BulkProvisionUsers handles bulk user provisioning
func (h *ProvisioningHandler) BulkProvisionUsers(c echo.Context) error {
	var requests []ProvisionUserRequest
	if err := c.Bind(&requests); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if len(requests) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No users provided for provisioning")
	}

	if len(requests) > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "Maximum 100 users can be provisioned at once")
	}

	var responses []ProvisionUserResponse
	var errors []string

	for i, req := range requests {
		// Validate individual request
		if req.KeycloakID == "" || req.Email == "" || req.EngageMXClientID == "" {
			errors = append(errors, map[string]string{
				"index": string(rune(i)),
				"error": "Missing required fields: keycloak_id, email, and engagemx_client_id are required",
			}[""]) // This is a simplified error handling - in production you'd want structured error responses
			continue
		}

		// Process individual user
		existingUser, err := h.repo.GetUserByKeycloakID(c.Request().Context(), req.KeycloakID)
		
		var user *models.User
		var action string

		if err != nil {
			// Create new user
			createReq := models.CreateUserRequest{
				KeycloakID:       req.KeycloakID,
				Email:            req.Email,
				FirstName:        req.FirstName,
				LastName:         req.LastName,
				EngageMXClientID: req.EngageMXClientID,
			}

			user, err = h.repo.CreateUser(c.Request().Context(), createReq)
			if err != nil {
				errors = append(errors, "User "+req.Email+": "+err.Error())
				continue
			}
			action = "created"
		} else {
			// Update existing user
			updateReq := models.UpdateUserRequest{
				Email:            &req.Email,
				FirstName:        req.FirstName,
				LastName:         req.LastName,
				EngageMXClientID: &req.EngageMXClientID,
			}

			if req.IsActive != nil {
				updateReq.IsActive = req.IsActive
			}

			user, err = h.repo.UpdateUser(c.Request().Context(), existingUser.ID, updateReq)
			if err != nil {
				errors = append(errors, "User "+req.Email+": "+err.Error())
				continue
			}
			action = "updated"
		}

		responses = append(responses, ProvisionUserResponse{
			ID:               user.ID,
			KeycloakID:       user.KeycloakID,
			Email:            user.Email,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			EngageMXClientID: user.EngageMXClientID,
			IsActive:         user.IsActive,
			Action:           action,
			Message:          "User successfully " + action,
		})
	}

	result := map[string]interface{}{
		"success_count": len(responses),
		"error_count":   len(errors),
		"users":         responses,
	}

	if len(errors) > 0 {
		result["errors"] = errors
	}

	statusCode := http.StatusOK
	if len(errors) > 0 && len(responses) == 0 {
		statusCode = http.StatusBadRequest
	} else if len(errors) > 0 {
		statusCode = http.StatusPartialContent
	}

	return c.JSON(statusCode, result)
}

// DeactivateUser deactivates a user (soft delete)
func (h *ProvisioningHandler) DeactivateUser(c echo.Context) error {
	keycloakID := c.Param("keycloak_id")
	if keycloakID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Keycloak ID is required")
	}

	user, err := h.repo.GetUserByKeycloakID(c.Request().Context(), keycloakID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	isActive := false
	updateReq := models.UpdateUserRequest{
		IsActive: &isActive,
	}

	updatedUser, err := h.repo.UpdateUser(c.Request().Context(), user.ID, updateReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to deactivate user: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "User successfully deactivated",
		"user":       updatedUser.ToResponse(),
		"action":     "deactivated",
		"timestamp":  updatedUser.UpdatedAt,
	})
}

// GetUserByKeycloakID retrieves user information for Django Admin
func (h *ProvisioningHandler) GetUserByKeycloakID(c echo.Context) error {
	keycloakID := c.Param("keycloak_id")
	if keycloakID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Keycloak ID is required")
	}

	user, err := h.repo.GetUserByKeycloakID(c.Request().Context(), keycloakID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, user.ToResponse())
}

// ListUsersByClient lists all users for a specific client
func (h *ProvisioningHandler) ListUsersByClient(c echo.Context) error {
	clientID := c.QueryParam("client_id")
	if clientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "client_id query parameter is required")
	}

	// This would need a new repository method to filter by client_id
	// For now, we'll get all users and filter (not efficient for large datasets)
	users, _, err := h.repo.ListUsers(c.Request().Context(), 1000, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users: "+err.Error())
	}

	var clientUsers []models.UserResponse
	for _, user := range users {
		if user.EngageMXClientID == clientID {
			clientUsers = append(clientUsers, user.ToResponse())
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"client_id": clientID,
		"count":     len(clientUsers),
		"users":     clientUsers,
	})
}