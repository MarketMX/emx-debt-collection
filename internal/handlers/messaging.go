package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/models"
	"emx-debt-collection/internal/services"
	"emx-debt-collection/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MessagingHandler handles messaging operations for debt collection
type MessagingHandler struct {
	repo              database.Repository
	messagingService  *services.MessagingService
	accountService    *services.AccountService
}

// NewMessagingHandler creates a new messaging handler
func NewMessagingHandler(repo database.Repository, messagingService *services.MessagingService, accountService *services.AccountService) *MessagingHandler {
	return &MessagingHandler{
		repo:             repo,
		messagingService: messagingService,
		accountService:   accountService,
	}
}

// SendMessages handles bulk message sending to selected accounts
func (h *MessagingHandler) SendMessages(c echo.Context) error {
	// Get user from context
	userInterface := c.Get("db_user")
	if userInterface == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid user type in context")
	}

	// Parse request body
	var req models.BulkMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Validate request
	if len(req.AccountIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Account IDs are required")
	}

	if req.MessageTemplate == "" {
		// Use default template if none provided
		defaultTemplate := models.DefaultMessageTemplate()
		req.MessageTemplate = defaultTemplate.Template
	}

	// Get account details for validation
	accounts := make([]models.Account, 0, len(req.AccountIDs))
	uploadIDs := make(map[uuid.UUID]bool)

	for _, accountID := range req.AccountIDs {
		account, err := h.repo.GetAccountByID(c.Request().Context(), accountID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Account %s not found", accountID))
		}

		// Verify user has access to this account via upload ownership
		upload, err := h.repo.GetUploadByID(c.Request().Context(), account.UploadID)
		if err != nil {
			return echo.NewHTTPError(http.StatusForbidden, "Access denied to account")
		}

		if upload.UserID != user.ID {
			return echo.NewHTTPError(http.StatusForbidden, "Access denied to account")
		}

		accounts = append(accounts, *account)
		uploadIDs[account.UploadID] = true
	}

	// Validate that all accounts belong to the same upload
	if len(uploadIDs) > 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "All accounts must belong to the same upload")
	}

	// Get the upload ID (we know there's only one)
	var uploadID uuid.UUID
	for id := range uploadIDs {
		uploadID = id
		break
	}

	// Start async messaging process
	go h.processMessagesAsync(accounts, uploadID, user.ID, req.MessageTemplate, req.MaxRetries)

	// Return immediate response
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message":         "Messaging job started successfully",
		"account_count":   len(accounts),
		"upload_id":       uploadID,
		"template_used":   req.MessageTemplate,
		"estimated_time":  fmt.Sprintf("%d seconds", len(accounts)*2), // Estimate 2 seconds per message
	})
}

// processMessagesAsync processes messages asynchronously
func (h *MessagingHandler) processMessagesAsync(accounts []models.Account, uploadID, userID uuid.UUID, messageTemplate string, maxRetries *int) {
	ctx := context.Background()

	// Create message logs for all accounts
	messageLogs := make([]models.CreateMessageLogRequest, 0, len(accounts))
	for _, account := range accounts {
		// Generate message content from template
		messageContent := h.messagingService.GenerateMessageContent(messageTemplate, account)

		logReq := models.CreateMessageLogRequest{
			AccountID:          account.ID,
			UploadID:           uploadID,
			UserID:             userID,
			MessageTemplate:    messageTemplate,
			MessageContent:     messageContent,
			RecipientTelephone: account.Telephone,
			MaxRetries:         maxRetries,
		}
		messageLogs = append(messageLogs, logReq)
	}

	// Save all message logs to database
	err := h.repo.CreateMessageLogsBatch(ctx, messageLogs)
	if err != nil {
		fmt.Printf("Failed to create message logs: %v\n", err)
		return
	}

	// Send messages one by one
	for _, account := range accounts {
		messageContent := h.messagingService.GenerateMessageContent(messageTemplate, account)
		
		result := h.messagingService.SendMessage(ctx, services.MessageRequest{
			AccountID:         account.ID,
			RecipientPhone:    account.Telephone,
			MessageContent:    messageContent,
			CustomerName:      account.CustomerName,
			AccountCode:       account.AccountCode,
		})

		// Update message log with result
		h.updateMessageLogFromResult(ctx, account.ID, uploadID, result)

		// Small delay between messages to avoid overwhelming the service
		time.Sleep(1 * time.Second)
	}
}

// updateMessageLogFromResult updates the message log based on sending result
func (h *MessagingHandler) updateMessageLogFromResult(ctx context.Context, accountID, uploadID uuid.UUID, result services.MessageResult) {
	// Find the message log (we'll need to query by account_id and upload_id since we don't store the log ID)
	// This is a simplified approach - in production you might want to return the log IDs from batch creation
	
	now := time.Now()
	var updateReq models.UpdateMessageLogRequest
	
	if result.Success {
		updateReq = models.UpdateMessageLogRequest{
			Status:              utils.StringPtr(string(models.MessageStatusSent)),
			SentAt:              &now,
			ExternalMessageID:   &result.MessageID,
			ResponseFromService: &result.Response,
		}
	} else {
		updateReq = models.UpdateMessageLogRequest{
			Status:              utils.StringPtr(string(models.MessageStatusFailed)),
			FailedAt:            &now,
			ErrorMessage:        &result.Error,
			ResponseFromService: &result.Response,
		}
	}

	// This is a simplified update - you'd need to implement a method to find log by account_id/upload_id
	// For now, we'll skip the update as it requires additional repository methods
	fmt.Printf("Message result for account %s: success=%t, error=%s, status=%v\n", 
		accountID, result.Success, result.Error, updateReq.Status)
}

// GetMessageLogs returns message logs for an upload
func (h *MessagingHandler) GetMessageLogs(c echo.Context) error {
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

	// Get message logs from database
	messageLogs, total, err := h.repo.GetMessageLogsByUploadID(c.Request().Context(), uploadID, perPage, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve message logs: "+err.Error())
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	hasMore := page < totalPages

	return c.JSON(http.StatusOK, models.MessageLogListResponse{
		MessageLogs: messageLogs,
		Total:       total,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		HasMore:     hasMore,
	})
}

// GetMessageLogSummary returns summary of messaging activity for an upload
func (h *MessagingHandler) GetMessageLogSummary(c echo.Context) error {
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

	// Get message log summary
	summary, err := h.repo.GetMessageLogSummary(c.Request().Context(), uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve message summary: "+err.Error())
	}

	return c.JSON(http.StatusOK, summary)
}

// GetMessageTemplates returns available message templates
func (h *MessagingHandler) GetMessageTemplates(c echo.Context) error {
	templates := []models.MessageTemplate{
		models.DefaultMessageTemplate(),
		{
			Name:     "Friendly Reminder",
			Template: "Hi [CustomerName], hope you're well! This is a friendly reminder about your account [AccountCode] with an outstanding balance of R[TotalBalance]. Please get in touch when convenient.",
		},
		{
			Name:     "Urgent Payment Required",
			Template: "URGENT: [CustomerName], your account [AccountCode] has an overdue balance of R[TotalBalance]. Please contact us immediately to avoid further action.",
		},
		{
			Name:     "Final Notice",
			Template: "FINAL NOTICE: [CustomerName], this is your final reminder for account [AccountCode] (R[TotalBalance]). Please contact us within 48 hours to resolve this matter.",
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

// GetAllMessageLogs returns message logs across all uploads for a user or system-wide (admin)
func (h *MessagingHandler) GetAllMessageLogs(c echo.Context) error {
	// Get user from context
	userInterface := c.Get("db_user")
	if userInterface == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid user type in context")
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

	// Parse filter parameters
	uploadIDStr := c.QueryParam("upload_id")
	statusFilter := c.QueryParam("status")

	// Get user's uploads first to ensure security
	userUploads, _, err := h.repo.GetUploadsByUserID(c.Request().Context(), user.ID, 1000, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user uploads")
	}

	// If upload_id filter is specified, verify user owns it
	var uploadID *uuid.UUID
	if uploadIDStr != "" {
		parsed, err := uuid.Parse(uploadIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid upload ID format")
		}

		// Verify ownership
		found := false
		for _, upload := range userUploads {
			if upload.ID == parsed {
				found = true
				break
			}
		}

		if !found {
			return echo.NewHTTPError(http.StatusForbidden, "Access denied to upload")
		}

		uploadID = &parsed
	}

	// If specific upload requested, get logs for that upload
	if uploadID != nil {
		messageLogs, total, err := h.repo.GetMessageLogsByUploadID(c.Request().Context(), *uploadID, perPage, offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve message logs: "+err.Error())
		}

		// Filter by status if requested
		if statusFilter != "" {
			filteredLogs := make([]models.MessageLogResponse, 0)
			for _, log := range messageLogs {
				if log.Status == statusFilter {
					filteredLogs = append(filteredLogs, log)
				}
			}
			messageLogs = filteredLogs
		}

		totalPages := int((total + int64(perPage) - 1) / int64(perPage))
		hasMore := page < totalPages

		return c.JSON(http.StatusOK, models.MessageLogListResponse{
			MessageLogs: messageLogs,
			Total:       total,
			Page:        page,
			PerPage:     perPage,
			TotalPages:  totalPages,
			HasMore:     hasMore,
		})
	}

	// Otherwise, get logs across all user's uploads
	var allMessageLogs []models.MessageLogResponse
	var totalCount int64

	// This is a simplified approach - in production you'd want a more efficient query
	for _, upload := range userUploads {
		logs, count, err := h.repo.GetMessageLogsByUploadID(c.Request().Context(), upload.ID, 1000, 0)
		if err != nil {
			continue // Skip uploads with errors
		}
		
		// Filter by status if requested
		if statusFilter != "" {
			for _, log := range logs {
				if log.Status == statusFilter {
					allMessageLogs = append(allMessageLogs, log)
				}
			}
		} else {
			allMessageLogs = append(allMessageLogs, logs...)
		}
		
		totalCount += count
	}

	// Apply pagination to the combined results
	startIndex := offset
	endIndex := offset + perPage
	
	if startIndex >= len(allMessageLogs) {
		allMessageLogs = []models.MessageLogResponse{}
	} else {
		if endIndex > len(allMessageLogs) {
			endIndex = len(allMessageLogs)
		}
		allMessageLogs = allMessageLogs[startIndex:endIndex]
	}

	totalPages := int((totalCount + int64(perPage) - 1) / int64(perPage))
	hasMore := page < totalPages

	return c.JSON(http.StatusOK, models.MessageLogListResponse{
		MessageLogs: allMessageLogs,
		Total:       totalCount,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		HasMore:     hasMore,
	})
}

