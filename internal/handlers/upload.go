package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/models"
	"emx-debt-collection/internal/services"
	"emx-debt-collection/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	MaxFileSize    = 50 << 20 // 50MB
	UploadDir      = "uploads"
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	repo           database.Repository
	excelService   *services.ExcelService
	accountService *services.AccountService
	progressService *services.ProgressService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(repo database.Repository, excelService *services.ExcelService, accountService *services.AccountService, progressService *services.ProgressService) *UploadHandler {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}

	return &UploadHandler{
		repo:           repo,
		excelService:   excelService,
		accountService: accountService,
		progressService: progressService,
	}
}

// UploadFile handles file upload requests
func (h *UploadHandler) UploadFile(c echo.Context) error {
	// Get user from context
	userInterface := c.Get("db_user")
	if userInterface == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid user type in context")
	}

	// Parse multipart form
	err := c.Request().ParseMultipartForm(MaxFileSize)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
	}

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided or invalid file")
	}
	defer file.Close()

	// Validate file size
	if header.Size > MaxFileSize {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("File too large. Maximum size is %d MB", MaxFileSize>>20))
	}

	// Validate file type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		// Try to determine MIME type from file extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		switch ext {
		case ".xlsx":
			mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case ".xls":
			mimeType = "application/vnd.ms-excel"
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "Unsupported file format. Only Excel files (.xlsx, .xls) are supported")
		}
	}

	if !h.excelService.IsSupportedFormat(mimeType) {
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported file format. Only Excel files are supported")
	}

	// Validate Excel file structure
	// Reset file pointer for validation
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate file")
	}

	if err := h.excelService.ValidateExcelFile(file); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Excel file structure: "+err.Error())
	}

	// Reset file pointer for processing
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process file")
	}

	// Create upload record
	fileSize := header.Size
	createReq := models.CreateUploadRequest{
		OriginalFilename: header.Filename,
		FileSize:         &fileSize,
		MimeType:         &mimeType,
	}

	upload, err := h.repo.CreateUpload(c.Request().Context(), user.ID, createReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create upload record: "+err.Error())
	}

	// Save file to disk
	filePath := filepath.Join(UploadDir, upload.Filename)
	destFile, err := os.Create(filePath)
	if err != nil {
		// Clean up upload record on file creation failure
		h.updateUploadStatus(c.Request().Context(), upload.ID, models.UploadStatusFailed, fmt.Sprintf("Failed to save file: %s", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save uploaded file")
	}
	defer destFile.Close()

	// Copy file data
	if _, err := io.Copy(destFile, file); err != nil {
		os.Remove(filePath) // Clean up file
		h.updateUploadStatus(c.Request().Context(), upload.ID, models.UploadStatusFailed, fmt.Sprintf("Failed to save file: %s", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save uploaded file")
	}

	// Update upload record with file path
	upload.FilePath = &filePath

	// Start async processing
	go h.processUploadAsync(upload, filePath)

	// Return upload info immediately
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "File uploaded successfully. Processing started.",
		"upload":  upload.ToResponse(),
	})
}

// processUploadAsync processes the uploaded file asynchronously
func (h *UploadHandler) processUploadAsync(upload *models.Upload, filePath string) {
	ctx := context.Background()

	// Update status to processing
	h.updateUploadStatusWithTime(ctx, upload.ID, models.UploadStatusProcessing, "", time.Now())

	// Open file for processing
	file, err := os.Open(filePath)
	if err != nil {
		h.updateUploadStatus(ctx, upload.ID, models.UploadStatusFailed, fmt.Sprintf("Failed to open file: %s", err.Error()))
		return
	}
	defer file.Close()

	// Parse Excel file
	result, err := h.excelService.ParseExcelFile(ctx, file, upload.ID, nil)
	if err != nil {
		h.updateUploadStatus(ctx, upload.ID, models.UploadStatusFailed, fmt.Sprintf("Failed to parse Excel file: %s", err.Error()))
		return
	}

	// Save accounts to database
	if len(result.Accounts) > 0 {
		err = h.repo.CreateAccountsBatch(ctx, result.Accounts)
		if err != nil {
			h.updateUploadStatus(ctx, upload.ID, models.UploadStatusFailed, fmt.Sprintf("Failed to save accounts: %s", err.Error()))
			return
		}
	}

	// Update upload status to completed
	now := time.Now()
	updateReq := models.UpdateUploadRequest{
		Status:                 utils.StringPtr(string(models.UploadStatusCompleted)),
		ProcessingCompletedAt:  &now,
		RecordsProcessed:       &result.ProcessedRows,
		RecordsFailed:          &result.FailedRows,
	}

	if len(result.Errors) > 0 {
		errorMsg := strings.Join(result.Errors, "; ")
		updateReq.ErrorMessage = &errorMsg
	}

	_, err = h.repo.UpdateUpload(ctx, upload.ID, updateReq)
	if err != nil {
		// Log error but don't fail the processing
		fmt.Printf("Failed to update upload status: %v\n", err)
	}
}

// GetUploads returns a list of uploads for the current user
func (h *UploadHandler) GetUploads(c echo.Context) error {
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

	// Get uploads from database
	uploads, total, err := h.repo.GetUploadsByUserID(c.Request().Context(), user.ID, perPage, offset)
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
		"uploads":     responses,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
		"has_more":    hasMore,
	})
}

// GetUpload returns details for a specific upload
func (h *UploadHandler) GetUpload(c echo.Context) error {
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

	// Get upload from database
	upload, err := h.repo.GetUploadByID(c.Request().Context(), uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Upload not found")
	}

	// Check if user owns this upload
	if upload.UserID != user.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	return c.JSON(http.StatusOK, upload.ToResponse())
}

// GetUploadAccounts returns accounts for a specific upload
func (h *UploadHandler) GetUploadAccounts(c echo.Context) error {
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

	// Get accounts from database
	accounts, total, err := h.repo.GetAccountsByUploadID(c.Request().Context(), uploadID, perPage, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve accounts: "+err.Error())
	}

	// Convert to response format
	responses := make([]models.AccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = account.ToResponse()
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	hasMore := page < totalPages

	return c.JSON(http.StatusOK, models.AccountListResponse{
		Accounts:   responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasMore:    hasMore,
	})
}

// GetUploadSummary returns summary statistics for a specific upload
func (h *UploadHandler) GetUploadSummary(c echo.Context) error {
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
	summary, err := h.repo.GetAccountSummary(c.Request().Context(), uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve summary: "+err.Error())
	}

	// Get age analysis from account service
	ageAnalysis, err := h.accountService.GetAgeAnalysis(c.Request().Context(), uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve age analysis: "+err.Error())
	}

	response := map[string]interface{}{
		"upload_id":     uploadID,
		"upload":        upload.ToResponse(),
		"summary":       summary,
		"age_analysis":  ageAnalysis,
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateAccountSelection updates the selection status of accounts
func (h *UploadHandler) UpdateAccountSelection(c echo.Context) error {
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

	// Parse request body
	var req models.BulkUpdateSelectionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if len(req.AccountIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Account IDs are required")
	}

	// Update account selection
	err = h.repo.BulkUpdateAccountSelection(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update account selection: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Updated selection for %d accounts", len(req.AccountIDs)),
		"updated_count": len(req.AccountIDs),
	})
}

// GetUploadProgress returns the current processing progress for an upload
func (h *UploadHandler) GetUploadProgress(c echo.Context) error {
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

	// Get progress from service
	progress, exists := h.progressService.GetProgress(uploadID)
	if !exists {
		// If no progress tracking, return upload status
		return c.JSON(http.StatusOK, map[string]interface{}{
			"upload_id":   uploadID,
			"stage":       upload.Status,
			"is_complete": upload.Status == string(models.UploadStatusCompleted) || upload.Status == string(models.UploadStatusFailed),
			"has_error":   upload.Status == string(models.UploadStatusFailed),
			"message":     getStatusMessage(upload.Status),
		})
	}

	return c.JSON(http.StatusOK, progress)
}

// getStatusMessage returns a user-friendly message for upload status
func getStatusMessage(status string) string {
	switch status {
	case string(models.UploadStatusPending):
		return "Upload pending"
	case string(models.UploadStatusProcessing):
		return "Processing file..."
	case string(models.UploadStatusCompleted):
		return "Processing completed successfully"
	case string(models.UploadStatusFailed):
		return "Processing failed"
	default:
		return "Unknown status"
	}
}

// Helper functions

func (h *UploadHandler) updateUploadStatus(ctx context.Context, uploadID uuid.UUID, status models.UploadStatus, errorMessage string) {
	req := models.UpdateUploadRequest{
		Status: utils.StringPtr(string(status)),
	}

	if errorMessage != "" {
		req.ErrorMessage = &errorMessage
	}

	_, err := h.repo.UpdateUpload(ctx, uploadID, req)
	if err != nil {
		fmt.Printf("Failed to update upload status: %v\n", err)
	}
}

func (h *UploadHandler) updateUploadStatusWithTime(ctx context.Context, uploadID uuid.UUID, status models.UploadStatus, errorMessage string, startTime time.Time) {
	req := models.UpdateUploadRequest{
		Status: utils.StringPtr(string(status)),
	}

	if status == models.UploadStatusProcessing {
		req.ProcessingStartedAt = &startTime
	}

	if errorMessage != "" {
		req.ErrorMessage = &errorMessage
	}

	_, err := h.repo.UpdateUpload(ctx, uploadID, req)
	if err != nil {
		fmt.Printf("Failed to update upload status: %v\n", err)
	}
}

