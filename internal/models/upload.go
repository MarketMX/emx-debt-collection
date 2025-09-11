package models

import (
	"time"

	"github.com/google/uuid"
)

// UploadStatus represents the possible states of an upload
type UploadStatus string

const (
	UploadStatusPending    UploadStatus = "pending"
	UploadStatusProcessing UploadStatus = "processing"
	UploadStatusCompleted  UploadStatus = "completed"
	UploadStatusFailed     UploadStatus = "failed"
)

// Upload represents a file upload and its processing status
type Upload struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	UserID                 uuid.UUID  `json:"user_id" db:"user_id"`
	Filename               string     `json:"filename" db:"filename"`
	OriginalFilename       string     `json:"original_filename" db:"original_filename"`
	FilePath               *string    `json:"file_path,omitempty" db:"file_path"`
	FileSize               *int64     `json:"file_size,omitempty" db:"file_size"`
	MimeType               *string    `json:"mime_type,omitempty" db:"mime_type"`
	Status                 string     `json:"status" db:"status"`
	ProcessingStartedAt    *time.Time `json:"processing_started_at,omitempty" db:"processing_started_at"`
	ProcessingCompletedAt  *time.Time `json:"processing_completed_at,omitempty" db:"processing_completed_at"`
	ErrorMessage           *string    `json:"error_message,omitempty" db:"error_message"`
	RecordsProcessed       int        `json:"records_processed" db:"records_processed"`
	RecordsFailed          int        `json:"records_failed" db:"records_failed"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateUploadRequest represents the request payload for creating an upload
type CreateUploadRequest struct {
	OriginalFilename string  `json:"original_filename" validate:"required"`
	FileSize         *int64  `json:"file_size"`
	MimeType         *string `json:"mime_type"`
}

// UpdateUploadRequest represents the request payload for updating an upload
type UpdateUploadRequest struct {
	Status                 *string    `json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	ProcessingStartedAt    *time.Time `json:"processing_started_at"`
	ProcessingCompletedAt  *time.Time `json:"processing_completed_at"`
	ErrorMessage           *string    `json:"error_message"`
	RecordsProcessed       *int       `json:"records_processed"`
	RecordsFailed          *int       `json:"records_failed"`
}

// UploadResponse represents the response payload for upload data
type UploadResponse struct {
	ID                     uuid.UUID  `json:"id"`
	UserID                 uuid.UUID  `json:"user_id"`
	Filename               string     `json:"filename"`
	OriginalFilename       string     `json:"original_filename"`
	FileSize               *int64     `json:"file_size,omitempty"`
	MimeType               *string    `json:"mime_type,omitempty"`
	Status                 string     `json:"status"`
	ProcessingStartedAt    *time.Time `json:"processing_started_at,omitempty"`
	ProcessingCompletedAt  *time.Time `json:"processing_completed_at,omitempty"`
	ErrorMessage           *string    `json:"error_message,omitempty"`
	RecordsProcessed       int        `json:"records_processed"`
	RecordsFailed          int        `json:"records_failed"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// UploadSummary represents the upload summary view with account statistics
type UploadSummary struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	Filename             string    `json:"filename" db:"filename"`
	OriginalFilename     string    `json:"original_filename" db:"original_filename"`
	Status               string    `json:"status" db:"status"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UserID               uuid.UUID `json:"user_id" db:"user_id"`
	UserEmail            *string   `json:"user_email" db:"user_email"`
	EngageMXClientID     string    `json:"engagemx_client_id" db:"engagemx_client_id"`
	TotalAccounts        int       `json:"total_accounts" db:"total_accounts"`
	SelectedAccounts     int       `json:"selected_accounts" db:"selected_accounts"`
	TotalBalanceSum      *float64  `json:"total_balance_sum" db:"total_balance_sum"`
	SelectedBalanceSum   *float64  `json:"selected_balance_sum" db:"selected_balance_sum"`
}

// ToResponse converts an Upload to UploadResponse
func (u *Upload) ToResponse() UploadResponse {
	return UploadResponse{
		ID:                     u.ID,
		UserID:                 u.UserID,
		Filename:               u.Filename,
		OriginalFilename:       u.OriginalFilename,
		FileSize:               u.FileSize,
		MimeType:               u.MimeType,
		Status:                 u.Status,
		ProcessingStartedAt:    u.ProcessingStartedAt,
		ProcessingCompletedAt:  u.ProcessingCompletedAt,
		ErrorMessage:           u.ErrorMessage,
		RecordsProcessed:       u.RecordsProcessed,
		RecordsFailed:          u.RecordsFailed,
		CreatedAt:              u.CreatedAt,
		UpdatedAt:              u.UpdatedAt,
	}
}

// IsValid checks if the upload status is valid
func (s UploadStatus) IsValid() bool {
	switch s {
	case UploadStatusPending, UploadStatusProcessing, UploadStatusCompleted, UploadStatusFailed:
		return true
	default:
		return false
	}
}

// ProcessingDuration returns the duration of processing if completed
func (u *Upload) ProcessingDuration() *time.Duration {
	if u.ProcessingStartedAt != nil && u.ProcessingCompletedAt != nil {
		duration := u.ProcessingCompletedAt.Sub(*u.ProcessingStartedAt)
		return &duration
	}
	return nil
}
