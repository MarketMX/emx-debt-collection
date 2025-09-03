package handlers

import (
	"net/http"
	"strconv"

	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdminHandler handles administrative operations
type AdminHandler struct {
	repo database.Repository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(repo database.Repository) *AdminHandler {
	return &AdminHandler{
		repo: repo,
	}
}

// ListUsers returns a paginated list of all users (admin only)
func (h *AdminHandler) ListUsers(c echo.Context) error {
	// Parse pagination parameters
	page := 1
	perPage := DefaultPerPage

	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= MaxPerPage {
			perPage = pp
		}
	}

	// Parse filter parameters
	activeOnly := c.QueryParam("active_only") == "true"
	
	offset := (page - 1) * perPage

	// Get users from database
	users, total, err := h.repo.ListUsers(c.Request().Context(), perPage, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve users: "+err.Error())
	}

	// Filter active users if requested
	if activeOnly {
		filteredUsers := make([]models.User, 0)
		for _, user := range users {
			if user.IsActive {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
		// Note: This is a simplified implementation. In production, filtering should be done at the database level.
	}

	// Convert to response format
	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	hasMore := page < totalPages

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users":       responses,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
		"has_more":    hasMore,
	})
}

// GetUser returns details for a specific user (admin only)
func (h *AdminHandler) GetUser(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	user, err := h.repo.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateUser updates user details (admin only)
func (h *AdminHandler) UpdateUser(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Parse request body
	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Update user in database
	user, err := h.repo.UpdateUser(c.Request().Context(), userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "User updated successfully",
		"user":    user.ToResponse(),
	})
}

// GetUserUploads returns uploads for a specific user (admin only)
func (h *AdminHandler) GetUserUploads(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Verify user exists
	_, err = h.repo.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	// Parse pagination parameters
	page := 1
	perPage := DefaultPerPage

	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= MaxPerPage {
			perPage = pp
		}
	}

	offset := (page - 1) * perPage

	// Get uploads from database
	uploads, total, err := h.repo.GetUploadsByUserID(c.Request().Context(), userID, perPage, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve uploads: "+err.Error())
	}

	// Convert to response format
	responses := make([]models.UploadResponse, len(uploads))
	for i, upload := range uploads {
		responses[i] = upload.ToResponse()
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	hasMore := page < totalPages

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":     userID,
		"uploads":     responses,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
		"has_more":    hasMore,
	})
}

// GetSystemStats returns overall system statistics (admin only)
func (h *AdminHandler) GetSystemStats(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user count
	_, totalUsers, err := h.repo.ListUsers(ctx, 1, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user count")
	}

	// Get upload summaries for system-wide stats
	summaries, totalUploads, err := h.repo.ListUploadSummaries(ctx, nil, 1000, 0) // Get recent uploads
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get upload summaries")
	}

	// Calculate totals
	var totalAccounts int64
	var totalSelectedAccounts int64
	var totalBalance float64
	var totalSelectedBalance float64
	completedUploads := 0
	failedUploads := 0

	for _, summary := range summaries {
		totalAccounts += int64(summary.TotalAccounts)
		totalSelectedAccounts += int64(summary.SelectedAccounts)
		if summary.TotalBalanceSum != nil {
			totalBalance += *summary.TotalBalanceSum
		}
		if summary.SelectedBalanceSum != nil {
			totalSelectedBalance += *summary.SelectedBalanceSum
		}

		switch summary.Status {
		case string(models.UploadStatusCompleted):
			completedUploads++
		case string(models.UploadStatusFailed):
			failedUploads++
		}
	}

	stats := map[string]interface{}{
		"users": map[string]interface{}{
			"total": totalUsers,
		},
		"uploads": map[string]interface{}{
			"total":     totalUploads,
			"completed": completedUploads,
			"failed":    failedUploads,
		},
		"accounts": map[string]interface{}{
			"total":          totalAccounts,
			"total_selected": totalSelectedAccounts,
		},
		"balances": map[string]interface{}{
			"total_balance":    totalBalance,
			"selected_balance": totalSelectedBalance,
		},
	}

	return c.JSON(http.StatusOK, stats)
}