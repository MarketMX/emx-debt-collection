package handlers

import (
	"net/http"
	"time"

	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/models"

	"github.com/labstack/echo/v4"
)

// WebhookHandler handles webhook events from external systems
type WebhookHandler struct {
	repo database.Repository
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(repo database.Repository) *WebhookHandler {
	return &WebhookHandler{
		repo: repo,
	}
}

// WebhookEvent represents a webhook event payload
type WebhookEvent struct {
	EventType   string                 `json:"event_type" validate:"required"`
	EventID     string                 `json:"event_id" validate:"required"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source" validate:"required"` // e.g., "django-admin", "mmx-dashboard"
	Data        map[string]interface{} `json:"data" validate:"required"`
	Version     string                 `json:"version,omitempty"`
	Signature   string                 `json:"signature,omitempty"` // For webhook verification
}

// WebhookResponse represents the webhook response
type WebhookResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	EventID   string `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
	ProcessedAt time.Time `json:"processed_at"`
}

// UserCreatedEvent represents user creation event data
type UserCreatedEvent struct {
	KeycloakID       string `json:"keycloak_id"`
	Email            string `json:"email"`
	FirstName        string `json:"first_name,omitempty"`
	LastName         string `json:"last_name,omitempty"`
	EngageMXClientID string `json:"engagemx_client_id"`
	IsActive         bool   `json:"is_active"`
}

// UserUpdatedEvent represents user update event data
type UserUpdatedEvent struct {
	KeycloakID       string `json:"keycloak_id"`
	Email            string `json:"email,omitempty"`
	FirstName        string `json:"first_name,omitempty"`
	LastName         string `json:"last_name,omitempty"`
	EngageMXClientID string `json:"engagemx_client_id,omitempty"`
	IsActive         *bool  `json:"is_active,omitempty"`
	Changes          []string `json:"changes,omitempty"` // List of changed fields
}

// UserDeactivatedEvent represents user deactivation event data
type UserDeactivatedEvent struct {
	KeycloakID string `json:"keycloak_id"`
	Email      string `json:"email"`
	Reason     string `json:"reason,omitempty"`
}

// ClientUpdatedEvent represents client update event data
type ClientUpdatedEvent struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name,omitempty"`
	IsActive bool   `json:"is_active"`
	Changes  []string `json:"changes,omitempty"`
}

// HandleWebhook processes incoming webhook events
func (h *WebhookHandler) HandleWebhook(c echo.Context) error {
	var event WebhookEvent
	if err := c.Bind(&event); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook payload: "+err.Error())
	}

	// Validate webhook event
	if err := c.Validate(event); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Webhook validation failed: "+err.Error())
	}

	// Set processing timestamp
	processedAt := time.Now()
	if event.Timestamp.IsZero() {
		event.Timestamp = processedAt
	}

	// Process event based on type
	var err error
	switch event.EventType {
	case "user.created":
		err = h.handleUserCreated(c, event)
	case "user.updated":
		err = h.handleUserUpdated(c, event)
	case "user.deactivated":
		err = h.handleUserDeactivated(c, event)
	case "client.updated":
		err = h.handleClientUpdated(c, event)
	case "user.bulk_created":
		err = h.handleBulkUserCreated(c, event)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown event type: "+event.EventType)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook: "+err.Error())
	}

	response := WebhookResponse{
		Success:     true,
		Message:     "Webhook processed successfully",
		EventID:     event.EventID,
		Timestamp:   event.Timestamp,
		ProcessedAt: processedAt,
	}

	return c.JSON(http.StatusOK, response)
}

// handleUserCreated processes user creation events
func (h *WebhookHandler) handleUserCreated(ctx echo.Context, event WebhookEvent) error {
	var userData UserCreatedEvent
	if err := mapEventData(event.Data, &userData); err != nil {
		return err
	}

	// Create user in database
	createReq := models.CreateUserRequest{
		KeycloakID:       userData.KeycloakID,
		Email:            userData.Email,
		FirstName:        stringPtr(userData.FirstName),
		LastName:         stringPtr(userData.LastName),
		EngageMXClientID: userData.EngageMXClientID,
	}

	_, err := h.repo.CreateUser(ctx.Request().Context(), createReq)
	if err != nil {
		// If user already exists, try to update instead
		existingUser, getErr := h.repo.GetUserByKeycloakID(ctx.Request().Context(), userData.KeycloakID)
		if getErr == nil {
			updateReq := models.UpdateUserRequest{
				Email:            &userData.Email,
				FirstName:        stringPtr(userData.FirstName),
				LastName:         stringPtr(userData.LastName),
				EngageMXClientID: &userData.EngageMXClientID,
				IsActive:         &userData.IsActive,
			}
			_, err = h.repo.UpdateUser(ctx.Request().Context(), existingUser.ID, updateReq)
		}
	}

	return err
}

// handleUserUpdated processes user update events
func (h *WebhookHandler) handleUserUpdated(ctx echo.Context, event WebhookEvent) error {
	var userData UserUpdatedEvent
	if err := mapEventData(event.Data, &userData); err != nil {
		return err
	}

	// Get existing user
	existingUser, err := h.repo.GetUserByKeycloakID(ctx.Request().Context(), userData.KeycloakID)
	if err != nil {
		return err
	}

	// Build update request with only changed fields
	updateReq := models.UpdateUserRequest{}
	if userData.Email != "" {
		updateReq.Email = &userData.Email
	}
	if userData.FirstName != "" {
		updateReq.FirstName = &userData.FirstName
	}
	if userData.LastName != "" {
		updateReq.LastName = &userData.LastName
	}
	if userData.EngageMXClientID != "" {
		updateReq.EngageMXClientID = &userData.EngageMXClientID
	}
	if userData.IsActive != nil {
		updateReq.IsActive = userData.IsActive
	}

	_, err = h.repo.UpdateUser(ctx.Request().Context(), existingUser.ID, updateReq)
	return err
}

// handleUserDeactivated processes user deactivation events
func (h *WebhookHandler) handleUserDeactivated(ctx echo.Context, event WebhookEvent) error {
	var userData UserDeactivatedEvent
	if err := mapEventData(event.Data, &userData); err != nil {
		return err
	}

	// Get existing user
	existingUser, err := h.repo.GetUserByKeycloakID(ctx.Request().Context(), userData.KeycloakID)
	if err != nil {
		return err
	}

	// Deactivate user
	isActive := false
	updateReq := models.UpdateUserRequest{
		IsActive: &isActive,
	}

	_, err = h.repo.UpdateUser(ctx.Request().Context(), existingUser.ID, updateReq)
	return err
}

// handleClientUpdated processes client update events
func (h *WebhookHandler) handleClientUpdated(ctx echo.Context, event WebhookEvent) error {
	var clientData ClientUpdatedEvent
	if err := mapEventData(event.Data, &clientData); err != nil {
		return err
	}

	// For now, we just log this event since we don't have a clients table
	// In the future, you might want to add client management
	ctx.Logger().Info("Client updated: ", clientData.ClientID, " - ", clientData.Name)
	
	return nil
}

// handleBulkUserCreated processes bulk user creation events
func (h *WebhookHandler) handleBulkUserCreated(ctx echo.Context, event WebhookEvent) error {
	// Extract users array from event data
	usersData, ok := event.Data["users"].([]interface{})
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid bulk user data format")
	}

	successCount := 0
	errorCount := 0

	for _, userData := range usersData {
		userMap, ok := userData.(map[string]interface{})
		if !ok {
			errorCount++
			continue
		}

		var user UserCreatedEvent
		if err := mapToStruct(userMap, &user); err != nil {
			errorCount++
			continue
		}

		// Create individual user
		createReq := models.CreateUserRequest{
			KeycloakID:       user.KeycloakID,
			Email:            user.Email,
			FirstName:        stringPtr(user.FirstName),
			LastName:         stringPtr(user.LastName),
			EngageMXClientID: user.EngageMXClientID,
		}

		_, err := h.repo.CreateUser(ctx.Request().Context(), createReq)
		if err != nil {
			// Try update if user exists
			existingUser, getErr := h.repo.GetUserByKeycloakID(ctx.Request().Context(), user.KeycloakID)
			if getErr == nil {
				updateReq := models.UpdateUserRequest{
					Email:            &user.Email,
					FirstName:        stringPtr(user.FirstName),
					LastName:         stringPtr(user.LastName),
					EngageMXClientID: &user.EngageMXClientID,
					IsActive:         &user.IsActive,
				}
				_, err = h.repo.UpdateUser(ctx.Request().Context(), existingUser.ID, updateReq)
			}
		}

		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	ctx.Logger().Info("Bulk user creation completed: ", successCount, " success, ", errorCount, " errors")
	return nil
}

// GetWebhookStatus returns webhook system status
func (h *WebhookHandler) GetWebhookStatus(c echo.Context) error {
	status := map[string]interface{}{
		"webhook_system": "active",
		"supported_events": []string{
			"user.created",
			"user.updated", 
			"user.deactivated",
			"user.bulk_created",
			"client.updated",
		},
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	return c.JSON(http.StatusOK, status)
}

// Helper functions

// mapEventData maps event data to a specific struct
func mapEventData(data map[string]interface{}, target interface{}) error {
	// This is a simplified mapper - in production you'd use a proper JSON marshaler
	return mapToStruct(data, target)
}

// mapToStruct maps a map to a struct (simplified implementation)
func mapToStruct(source map[string]interface{}, target interface{}) error {
	// For simplicity, we'll just handle the basic cases
	// In production, you'd use json.Marshal/Unmarshal or a proper mapper library
	
	switch v := target.(type) {
	case *UserCreatedEvent:
		if keycloakID, ok := source["keycloak_id"].(string); ok {
			v.KeycloakID = keycloakID
		}
		if email, ok := source["email"].(string); ok {
			v.Email = email
		}
		if firstName, ok := source["first_name"].(string); ok {
			v.FirstName = firstName
		}
		if lastName, ok := source["last_name"].(string); ok {
			v.LastName = lastName
		}
		if clientID, ok := source["engagemx_client_id"].(string); ok {
			v.EngageMXClientID = clientID
		}
		if isActive, ok := source["is_active"].(bool); ok {
			v.IsActive = isActive
		}
	case *UserUpdatedEvent:
		if keycloakID, ok := source["keycloak_id"].(string); ok {
			v.KeycloakID = keycloakID
		}
		if email, ok := source["email"].(string); ok {
			v.Email = email
		}
		if firstName, ok := source["first_name"].(string); ok {
			v.FirstName = firstName
		}
		if lastName, ok := source["last_name"].(string); ok {
			v.LastName = lastName
		}
		if clientID, ok := source["engagemx_client_id"].(string); ok {
			v.EngageMXClientID = clientID
		}
		if isActive, ok := source["is_active"].(bool); ok {
			v.IsActive = &isActive
		}
	case *UserDeactivatedEvent:
		if keycloakID, ok := source["keycloak_id"].(string); ok {
			v.KeycloakID = keycloakID
		}
		if email, ok := source["email"].(string); ok {
			v.Email = email
		}
		if reason, ok := source["reason"].(string); ok {
			v.Reason = reason
		}
	case *ClientUpdatedEvent:
		if clientID, ok := source["client_id"].(string); ok {
			v.ClientID = clientID
		}
		if name, ok := source["name"].(string); ok {
			v.Name = name
		}
		if isActive, ok := source["is_active"].(bool); ok {
			v.IsActive = isActive
		}
	}

	return nil
}

// stringPtr returns a pointer to the string value, or nil if empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}