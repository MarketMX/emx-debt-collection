package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageStatus represents the possible states of a message
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusFailed    MessageStatus = "failed"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// MessageLog represents a log entry for a message sent to an account
type MessageLog struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	AccountID           uuid.UUID  `json:"account_id" db:"account_id"`
	UploadID            uuid.UUID  `json:"upload_id" db:"upload_id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	MessageTemplate     string     `json:"message_template" db:"message_template"`
	MessageContent      string     `json:"message_content" db:"message_content"`
	RecipientTelephone  string     `json:"recipient_telephone" db:"recipient_telephone"`
	Status              string     `json:"status" db:"status"`
	ExternalMessageID   *string    `json:"external_message_id,omitempty" db:"external_message_id"`
	SentAt              *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt         *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	FailedAt            *time.Time `json:"failed_at,omitempty" db:"failed_at"`
	RetryCount          int        `json:"retry_count" db:"retry_count"`
	MaxRetries          int        `json:"max_retries" db:"max_retries"`
	ResponseFromService *string    `json:"response_from_service,omitempty" db:"response_from_service"`
	ErrorMessage        *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateMessageLogRequest represents the request payload for creating a message log
type CreateMessageLogRequest struct {
	AccountID          uuid.UUID `json:"account_id" validate:"required"`
	UploadID           uuid.UUID `json:"upload_id" validate:"required"`
	UserID             uuid.UUID `json:"user_id" validate:"required"`
	MessageTemplate    string    `json:"message_template" validate:"required"`
	MessageContent     string    `json:"message_content" validate:"required"`
	RecipientTelephone string    `json:"recipient_telephone" validate:"required"`
	MaxRetries         *int      `json:"max_retries"`
}

// UpdateMessageLogRequest represents the request payload for updating a message log
type UpdateMessageLogRequest struct {
	Status              *string    `json:"status" validate:"omitempty,oneof=pending sent failed delivered read"`
	ExternalMessageID   *string    `json:"external_message_id"`
	SentAt              *time.Time `json:"sent_at"`
	DeliveredAt         *time.Time `json:"delivered_at"`
	FailedAt            *time.Time `json:"failed_at"`
	RetryCount          *int       `json:"retry_count"`
	ResponseFromService *string    `json:"response_from_service"`
	ErrorMessage        *string    `json:"error_message"`
}

// MessageLogResponse represents the response payload for message log data
type MessageLogResponse struct {
	ID                  uuid.UUID  `json:"id"`
	AccountID           uuid.UUID  `json:"account_id"`
	UploadID            uuid.UUID  `json:"upload_id"`
	UserID              uuid.UUID  `json:"user_id"`
	MessageTemplate     string     `json:"message_template"`
	MessageContent      string     `json:"message_content"`
	RecipientTelephone  string     `json:"recipient_telephone"`
	Status              string     `json:"status"`
	ExternalMessageID   *string    `json:"external_message_id,omitempty"`
	SentAt              *time.Time `json:"sent_at,omitempty"`
	DeliveredAt         *time.Time `json:"delivered_at,omitempty"`
	FailedAt            *time.Time `json:"failed_at,omitempty"`
	RetryCount          int        `json:"retry_count"`
	MaxRetries          int        `json:"max_retries"`
	ResponseFromService *string    `json:"response_from_service,omitempty"`
	ErrorMessage        *string    `json:"error_message,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// MessageLogWithAccount represents a message log with associated account information
type MessageLogWithAccount struct {
	MessageLog
	AccountCode    string  `json:"account_code" db:"account_code"`
	CustomerName   string  `json:"customer_name" db:"customer_name"`
	ContactPerson  *string `json:"contact_person" db:"contact_person"`
	TotalBalance   float64 `json:"total_balance" db:"total_balance"`
}

// MessageLogListResponse represents a paginated list of message logs with metadata
type MessageLogListResponse struct {
	MessageLogs []MessageLogResponse `json:"message_logs"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PerPage     int                  `json:"per_page"`
	TotalPages  int                  `json:"total_pages"`
	HasMore     bool                 `json:"has_more"`
}

// MessageLogSummary represents the message log summary view
type MessageLogSummary struct {
	UploadID          uuid.UUID  `json:"upload_id" db:"upload_id"`
	EngageMXClientID  string     `json:"engagemx_client_id" db:"engagemx_client_id"`
	TotalMessages     int        `json:"total_messages" db:"total_messages"`
	SentMessages      int        `json:"sent_messages" db:"sent_messages"`
	FailedMessages    int        `json:"failed_messages" db:"failed_messages"`
	DeliveredMessages int        `json:"delivered_messages" db:"delivered_messages"`
	FirstSentAt       *time.Time `json:"first_sent_at" db:"first_sent_at"`
	LastSentAt        *time.Time `json:"last_sent_at" db:"last_sent_at"`
}

// BulkMessageRequest represents a request to send messages to multiple accounts
type BulkMessageRequest struct {
	AccountIDs      []uuid.UUID `json:"account_ids" validate:"required,min=1"`
	MessageTemplate string      `json:"message_template" validate:"required"`
	MaxRetries      *int        `json:"max_retries"`
}

// MessageTemplate represents a message template with placeholders
type MessageTemplate struct {
	Template string `json:"template"`
	Name     string `json:"name"`
}

// DefaultMessageTemplate returns the default message template
func DefaultMessageTemplate() MessageTemplate {
	return MessageTemplate{
		Name:     "Default Payment Reminder",
		Template: "Hi [CustomerName], this is a reminder regarding account [AccountCode] for an outstanding balance of R[TotalBalance]. Please contact us to arrange payment.",
	}
}

// ToResponse converts a MessageLog to MessageLogResponse
func (ml *MessageLog) ToResponse() MessageLogResponse {
	return MessageLogResponse{
		ID:                  ml.ID,
		AccountID:           ml.AccountID,
		UploadID:            ml.UploadID,
		UserID:              ml.UserID,
		MessageTemplate:     ml.MessageTemplate,
		MessageContent:      ml.MessageContent,
		RecipientTelephone:  ml.RecipientTelephone,
		Status:              ml.Status,
		ExternalMessageID:   ml.ExternalMessageID,
		SentAt:              ml.SentAt,
		DeliveredAt:         ml.DeliveredAt,
		FailedAt:            ml.FailedAt,
		RetryCount:          ml.RetryCount,
		MaxRetries:          ml.MaxRetries,
		ResponseFromService: ml.ResponseFromService,
		ErrorMessage:        ml.ErrorMessage,
		CreatedAt:           ml.CreatedAt,
		UpdatedAt:           ml.UpdatedAt,
	}
}

// IsValid checks if the message status is valid
func (s MessageStatus) IsValid() bool {
	switch s {
	case MessageStatusPending, MessageStatusSent, MessageStatusFailed, MessageStatusDelivered, MessageStatusRead:
		return true
	default:
		return false
	}
}

// CanRetry determines if a message can be retried based on its status and retry count
func (ml *MessageLog) CanRetry() bool {
	return ml.Status == string(MessageStatusFailed) && ml.RetryCount < ml.MaxRetries
}

// ShouldRetry determines if a message should be automatically retried
func (ml *MessageLog) ShouldRetry() bool {
	if !ml.CanRetry() {
		return false
	}
	
	// Add exponential backoff logic here if needed
	// For now, allow immediate retry
	return true
}

// Duration returns the time taken from creation to completion/failure
func (ml *MessageLog) Duration() *time.Duration {
	var endTime *time.Time
	
	switch ml.Status {
	case string(MessageStatusSent), string(MessageStatusDelivered), string(MessageStatusRead):
		endTime = ml.SentAt
		if ml.DeliveredAt != nil {
			endTime = ml.DeliveredAt
		}
	case string(MessageStatusFailed):
		endTime = ml.FailedAt
	default:
		return nil
	}
	
	if endTime != nil {
		duration := endTime.Sub(ml.CreatedAt)
		return &duration
	}
	
	return nil
}
