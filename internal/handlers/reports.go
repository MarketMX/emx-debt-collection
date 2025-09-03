package handlers

import (
	"net/http"
	"time"

	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ReportsHandler handles report generation and analytics
type ReportsHandler struct {
	repo database.Repository
}

// NewReportsHandler creates a new reports handler
func NewReportsHandler(repo database.Repository) *ReportsHandler {
	return &ReportsHandler{
		repo: repo,
	}
}

// GetUploadReport generates a detailed report for a specific upload
func (h *ReportsHandler) GetUploadReport(c echo.Context) error {
	// Parse upload ID
	uploadIDStr := c.Param("id")
	uploadID, err := uuid.Parse(uploadIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid upload ID")
	}

	// Get user from context
	userInterface := c.Get("db_user")
	if userInterface == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid user type in context")
	}

	// Verify user owns this upload
	upload, err := h.repo.GetUploadByID(c.Request().Context(), uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Upload not found")
	}

	if upload.UserID != user.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	// Get account summary
	accountSummary, err := h.repo.GetAccountSummary(c.Request().Context(), uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get account summary")
	}

	// Get message log summary
	messagingSummary, err := h.repo.GetMessageLogSummary(c.Request().Context(), uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get messaging summary")
	}

	// Get processing time
	var processingTime *time.Duration
	if upload.ProcessingStartedAt != nil && upload.ProcessingCompletedAt != nil {
		duration := upload.ProcessingCompletedAt.Sub(*upload.ProcessingStartedAt)
		processingTime = &duration
	}

	report := map[string]interface{}{
		"upload": map[string]interface{}{
			"id":               upload.ID,
			"filename":         upload.OriginalFilename,
			"status":           upload.Status,
			"created_at":       upload.CreatedAt,
			"processing_time":  processingTime,
			"records_processed": upload.RecordsProcessed,
			"records_failed":   upload.RecordsFailed,
			"file_size":        upload.FileSize,
		},
		"accounts": map[string]interface{}{
			"summary":         accountSummary,
			"processing_rate": calculateProcessingRate(upload.RecordsProcessed, upload.RecordsFailed),
		},
		"messaging": messagingSummary,
		"generated_at": time.Now(),
	}

	return c.JSON(http.StatusOK, report)
}

// GetUserActivityReport generates an activity report for the current user
func (h *ReportsHandler) GetUserActivityReport(c echo.Context) error {
	// Get user from context
	userInterface := c.Get("db_user")
	if userInterface == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid user type in context")
	}

	// Parse date range parameters
	fromStr := c.QueryParam("from")
	toStr := c.QueryParam("to")

	// Default to last 30 days if no range provided
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	if fromStr != "" {
		if parsedFrom, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = parsedFrom
		}
	}

	if toStr != "" {
		if parsedTo, err := time.Parse("2006-01-02", toStr); err == nil {
			to = parsedTo
		}
	}

	// Get user's uploads
	uploads, _, err := h.repo.GetUploadsByUserID(c.Request().Context(), user.ID, 1000, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user uploads")
	}

	// Filter uploads by date range and calculate stats
	var uploadsInRange []models.Upload
	var totalAccounts int
	var totalMessages int
	var successfulUploads int
	var failedUploads int

	for _, upload := range uploads {
		if upload.CreatedAt.After(from) && upload.CreatedAt.Before(to) {
			uploadsInRange = append(uploadsInRange, upload)
			totalAccounts += upload.RecordsProcessed

			// Get message count for this upload
			_, messageCount, err := h.repo.GetMessageLogsByUploadID(c.Request().Context(), upload.ID, 1, 0)
			if err == nil {
				totalMessages += int(messageCount)
			}

			if upload.Status == string(models.UploadStatusCompleted) {
				successfulUploads++
			} else if upload.Status == string(models.UploadStatusFailed) {
				failedUploads++
			}
		}
	}

	report := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"period": map[string]interface{}{
			"from": from,
			"to":   to,
		},
		"activity": map[string]interface{}{
			"uploads_total":     len(uploadsInRange),
			"uploads_successful": successfulUploads,
			"uploads_failed":    failedUploads,
			"accounts_processed": totalAccounts,
			"messages_sent":     totalMessages,
			"success_rate":      calculateSuccessRate(successfulUploads, len(uploadsInRange)),
		},
		"recent_uploads": limitUploads(uploadsInRange, 10),
		"generated_at":   time.Now(),
	}

	return c.JSON(http.StatusOK, report)
}

// GetSystemReport generates a system-wide analytics report (admin only)
func (h *ReportsHandler) GetSystemReport(c echo.Context) error {
	// Parse date range parameters
	fromStr := c.QueryParam("from")
	toStr := c.QueryParam("to")

	// Default to last 30 days if no range provided
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	if fromStr != "" {
		if parsedFrom, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = parsedFrom
		}
	}

	if toStr != "" {
		if parsedTo, err := time.Parse("2006-01-02", toStr); err == nil {
			to = parsedTo
		}
	}

	// Get all upload summaries
	summaries, _, err := h.repo.ListUploadSummaries(c.Request().Context(), nil, 1000, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get upload summaries")
	}

	// Calculate system-wide metrics
	var totalUploads int
	var successfulUploads int
	var failedUploads int
	var totalAccounts int64
	var totalSelectedAccounts int64
	var totalBalance float64
	var totalSelectedBalance float64
	var totalUsers int

	// Get user count
	_, userCount, err := h.repo.ListUsers(c.Request().Context(), 1, 0)
	if err == nil {
		totalUsers = int(userCount)
	}

	// Process summaries
	userActivityMap := make(map[uuid.UUID]int)
	for _, summary := range summaries {
		if summary.CreatedAt.After(from) && summary.CreatedAt.Before(to) {
			totalUploads++
			totalAccounts += int64(summary.TotalAccounts)
			totalSelectedAccounts += int64(summary.SelectedAccounts)
			if summary.TotalBalanceSum != nil {
				totalBalance += *summary.TotalBalanceSum
			}
			if summary.SelectedBalanceSum != nil {
				totalSelectedBalance += *summary.SelectedBalanceSum
			}

			if summary.Status == string(models.UploadStatusCompleted) {
				successfulUploads++
			} else if summary.Status == string(models.UploadStatusFailed) {
				failedUploads++
			}

			// Track user activity
			userActivityMap[summary.UserID]++
		}
	}

	// Calculate metrics
	successRate := calculateSuccessRate(successfulUploads, totalUploads)
	avgAccountsPerUpload := float64(0)
	if totalUploads > 0 {
		avgAccountsPerUpload = float64(totalAccounts) / float64(totalUploads)
	}

	activeUsers := len(userActivityMap)

	report := map[string]interface{}{
		"period": map[string]interface{}{
			"from": from,
			"to":   to,
		},
		"system_overview": map[string]interface{}{
			"total_users":   totalUsers,
			"active_users":  activeUsers,
			"total_uploads": totalUploads,
			"success_rate":  successRate,
		},
		"uploads": map[string]interface{}{
			"total":      totalUploads,
			"successful": successfulUploads,
			"failed":     failedUploads,
		},
		"accounts": map[string]interface{}{
			"total_processed":    totalAccounts,
			"total_selected":     totalSelectedAccounts,
			"avg_per_upload":     avgAccountsPerUpload,
			"selection_rate":     calculateSelectionRate(int(totalSelectedAccounts), int(totalAccounts)),
		},
		"financial": map[string]interface{}{
			"total_balance":    totalBalance,
			"selected_balance": totalSelectedBalance,
			"avg_balance":      calculateAvgBalance(totalBalance, int(totalAccounts)),
		},
		"generated_at": time.Now(),
	}

	return c.JSON(http.StatusOK, report)
}

// GetMessagingReport generates a messaging analytics report
func (h *ReportsHandler) GetMessagingReport(c echo.Context) error {
	// Parse upload ID if provided
	uploadIDStr := c.QueryParam("upload_id")
	var uploadID *uuid.UUID
	if uploadIDStr != "" {
		if parsed, err := uuid.Parse(uploadIDStr); err == nil {
			uploadID = &parsed

			// If upload_id is provided, verify user has access
			userInterface := c.Get("db_user")
			if userInterface != nil {
				if user, ok := userInterface.(*models.User); ok {
					upload, err := h.repo.GetUploadByID(c.Request().Context(), *uploadID)
					if err != nil || upload.UserID != user.ID {
						return echo.NewHTTPError(http.StatusForbidden, "Access denied to upload")
					}
				}
			}
		}
	}

	// Get messaging summaries
	if uploadID != nil {
		// Get summary for specific upload
		summary, err := h.repo.GetMessageLogSummary(c.Request().Context(), *uploadID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get messaging summary")
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"upload_id": uploadID,
			"summary":   summary,
			"generated_at": time.Now(),
		})
	}

	// Return error if no specific upload and no admin privileges
	// In a production system, you'd implement system-wide messaging analytics here
	return echo.NewHTTPError(http.StatusBadRequest, "upload_id parameter is required")
}

// Helper functions

func calculateProcessingRate(processed, failed int) float64 {
	total := processed + failed
	if total == 0 {
		return 0
	}
	return (float64(processed) / float64(total)) * 100
}

func calculateSuccessRate(successful, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(successful) / float64(total)) * 100
}

func calculateSelectionRate(selected, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(selected) / float64(total)) * 100
}

func calculateAvgBalance(totalBalance float64, totalAccounts int) float64 {
	if totalAccounts == 0 {
		return 0
	}
	return totalBalance / float64(totalAccounts)
}

func limitUploads(uploads []models.Upload, limit int) []models.UploadResponse {
	if len(uploads) > limit {
		uploads = uploads[:limit]
	}
	
	responses := make([]models.UploadResponse, len(uploads))
	for i, upload := range uploads {
		responses[i] = upload.ToResponse()
	}
	
	return responses
}