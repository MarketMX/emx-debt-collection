package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"emx-debt-collection/internal/models"
	"github.com/google/uuid"
)

// MessagingService handles external messaging service integration
type MessagingService struct {
	apiURL      string
	apiKey      string
	timeout     time.Duration
	rateLimiter *RateLimiter
}

// NewMessagingService creates a new messaging service instance
func NewMessagingService(apiURL, apiKey string) *MessagingService {
	return &MessagingService{
		apiURL:      apiURL,
		apiKey:      apiKey,
		timeout:     30 * time.Second,
		rateLimiter: NewRateLimiter(5, time.Second), // 5 messages per second
	}
}

// MessageRequest represents a message sending request
type MessageRequest struct {
	AccountID      uuid.UUID `json:"account_id"`
	RecipientPhone string    `json:"recipient_phone"`
	MessageContent string    `json:"message_content"`
	CustomerName   string    `json:"customer_name"`
	AccountCode    string    `json:"account_code"`
}

// MessageResult represents the result of a message sending attempt
type MessageResult struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
	Response  string `json:"response,omitempty"`
}

// ExternalMessageRequest represents the request format for the external messaging service
type ExternalMessageRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
	From    string `json:"from,omitempty"`
}

// ExternalMessageResponse represents the response from the external messaging service
type ExternalMessageResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

// SendMessage sends a message to the external messaging service
func (s *MessagingService) SendMessage(ctx context.Context, req MessageRequest) MessageResult {
	// Apply rate limiting
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return MessageResult{
			Success: false,
			Error:   "Rate limit exceeded: " + err.Error(),
		}
	}

	// Validate phone number
	cleanPhone := s.cleanPhoneNumber(req.RecipientPhone)
	if !s.isValidPhoneNumber(cleanPhone) {
		return MessageResult{
			Success: false,
			Error:   "Invalid phone number format: " + req.RecipientPhone,
		}
	}

	// For now, simulate the external API call
	// In production, you would replace this with actual HTTP calls to your messaging service
	return s.simulateExternalAPI(ctx, req, cleanPhone)
}

// simulateExternalAPI simulates calling an external messaging API
// In production, replace this with actual HTTP client calls
func (s *MessagingService) simulateExternalAPI(ctx context.Context, req MessageRequest, cleanPhone string) MessageResult {
	// Simulate network delay
	time.Sleep(500 * time.Millisecond)

	// Simulate different outcomes based on phone number patterns (for demo purposes)
	lastDigit := cleanPhone[len(cleanPhone)-1:]
	
	switch lastDigit {
	case "1", "2", "3", "4", "5", "6", "7": // 70% success rate
		return MessageResult{
			Success:   true,
			MessageID: fmt.Sprintf("msg_%s_%d", req.AccountID.String()[:8], time.Now().Unix()),
			Response:  "Message sent successfully",
		}
	case "8": // Simulate temporary failure
		return MessageResult{
			Success:  false,
			Error:    "Temporary service unavailable, retry later",
			Response: "HTTP 503 Service Unavailable",
		}
	case "9": // Simulate permanent failure
		return MessageResult{
			Success:  false,
			Error:    "Invalid phone number or recipient blocked",
			Response: "HTTP 400 Bad Request",
		}
	default: // "0" - simulate network timeout
		return MessageResult{
			Success:  false,
			Error:    "Network timeout",
			Response: "Request timeout after 30 seconds",
		}
	}
}

// TODO: Replace simulateExternalAPI with this actual implementation
func (s *MessagingService) callExternalAPI(ctx context.Context, req MessageRequest, cleanPhone string) MessageResult {
	// This is the actual implementation you would use in production
	/*
	import "net/http"
	import "bytes"
	import "encoding/json"
	
	externalReq := ExternalMessageRequest{
		To:      cleanPhone,
		Message: req.MessageContent,
		From:    "DebtCollection", // Your service name
	}

	reqBody, err := json.Marshal(externalReq)
	if err != nil {
		return MessageResult{
			Success: false,
			Error:   "Failed to marshal request: " + err.Error(),
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.apiURL+"/send", bytes.NewBuffer(reqBody))
	if err != nil {
		return MessageResult{
			Success: false,
			Error:   "Failed to create request: " + err.Error(),
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{Timeout: s.timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return MessageResult{
			Success: false,
			Error:   "HTTP request failed: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	var externalResp ExternalMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&externalResp); err != nil {
		return MessageResult{
			Success: false,
			Error:   "Failed to decode response: " + err.Error(),
			Response: fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	return MessageResult{
		Success:   externalResp.Success,
		MessageID: externalResp.MessageID,
		Error:     externalResp.Error,
		Response:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, externalResp.Status),
	}
	*/
	
	return MessageResult{
		Success: false,
		Error:   "External API not implemented yet",
	}
}

// GenerateMessageContent generates message content from template and account data
func (s *MessagingService) GenerateMessageContent(template string, account models.Account) string {
	content := template

	// Define replacement map
	replacements := map[string]string{
		"[CustomerName]": account.CustomerName,
		"[AccountCode]":  account.AccountCode,
		"[TotalBalance]": fmt.Sprintf("%.2f", account.TotalBalance),
		"[Current]":      fmt.Sprintf("%.2f", account.AmountCurrent),
		"[30Days]":       fmt.Sprintf("%.2f", account.Amount30d),
		"[60Days]":       fmt.Sprintf("%.2f", account.Amount60d),
		"[90Days]":       fmt.Sprintf("%.2f", account.Amount90d),
		"[120Days]":      fmt.Sprintf("%.2f", account.Amount120d),
		"[ContactPerson]": s.getContactPersonOrDefault(account.ContactPerson),
		"[Telephone]":    account.Telephone,
	}

	// Add derived values
	overdueAmount := account.OverdueAmount()
	replacements["[OverdueAmount]"] = fmt.Sprintf("%.2f", overdueAmount)
	
	_, ageBracket := account.GetOldestOverdueAmount()
	replacements["[OldestAgeBracket]"] = ageBracket

	// Perform replacements
	for placeholder, value := range replacements {
		content = strings.ReplaceAll(content, placeholder, value)
	}

	return content
}

// getContactPersonOrDefault returns contact person or customer name as fallback
func (s *MessagingService) getContactPersonOrDefault(contactPerson *string) string {
	if contactPerson != nil && *contactPerson != "" {
		return *contactPerson
	}
	return "valued customer"
}

// cleanPhoneNumber cleans and formats phone number for messaging
func (s *MessagingService) cleanPhoneNumber(phone string) string {
	// Remove common formatting characters
	cleaned := strings.TrimSpace(phone)
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	
	// Remove leading + if present
	if strings.HasPrefix(cleaned, "+") {
		cleaned = cleaned[1:]
	}
	
	// Add country code if missing (assuming South African numbers)
	if len(cleaned) == 10 && strings.HasPrefix(cleaned, "0") {
		cleaned = "27" + cleaned[1:] // Replace leading 0 with 27
	} else if len(cleaned) == 9 && !strings.HasPrefix(cleaned, "27") {
		cleaned = "27" + cleaned // Add 27 prefix
	}
	
	return cleaned
}

// isValidPhoneNumber validates phone number format
func (s *MessagingService) isValidPhoneNumber(phone string) bool {
	if phone == "" {
		return false
	}

	// Check length (South African numbers: 11 digits with country code)
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}

	// Check that it contains only digits
	matched, _ := regexp.MatchString(`^\d+$`, phone)
	return matched
}

// GetMessageTemplates returns available message templates
func (s *MessagingService) GetMessageTemplates() []models.MessageTemplate {
	return []models.MessageTemplate{
		models.DefaultMessageTemplate(),
		{
			Name:     "Friendly Reminder",
			Template: "Hi [CustomerName], hope you're well! This is a friendly reminder about your account [AccountCode] with an outstanding balance of R[TotalBalance]. Please get in touch when convenient.",
		},
		{
			Name:     "Professional Notice",
			Template: "Dear [CustomerName], please note that account [AccountCode] has an outstanding balance of R[TotalBalance]. We kindly request payment at your earliest convenience. Contact us for payment arrangements.",
		},
		{
			Name:     "Overdue Notice",
			Template: "[CustomerName], your account [AccountCode] has overdue amounts totaling R[OverdueAmount] from [OldestAgeBracket]. Total outstanding: R[TotalBalance]. Please contact us urgently.",
		},
		{
			Name:     "Payment Arrangement",
			Template: "Hi [CustomerName], we're here to help with account [AccountCode] (R[TotalBalance]). Please call us to discuss flexible payment options that work for you.",
		},
		{
			Name:     "Final Notice",
			Template: "FINAL NOTICE: [CustomerName], immediate action required for account [AccountCode] (R[TotalBalance]). Please contact us within 48 hours to avoid further collection action.",
		},
	}
}

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	rate     int
	interval time.Duration
	tokens   chan struct{}
	ticker   *time.Ticker
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:     rate,
		interval: interval,
		tokens:   make(chan struct{}, rate),
		ticker:   time.NewTicker(interval),
	}

	// Fill initial tokens
	for i := 0; i < rate; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token replenishment
	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Token bucket is full
			}
		}
	}()

	return rl
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
	close(rl.tokens)
}

// ValidateTemplate validates a message template
func (s *MessagingService) ValidateTemplate(template string) []string {
	var errors []string

	if template == "" {
		errors = append(errors, "Template cannot be empty")
		return errors
	}

	if len(template) > 1600 { // SMS limit
		errors = append(errors, "Template too long (max 1600 characters for SMS)")
	}

	// Check for required placeholders
	requiredPlaceholders := []string{"[CustomerName]", "[AccountCode]", "[TotalBalance]"}
	for _, placeholder := range requiredPlaceholders {
		if !strings.Contains(template, placeholder) {
			errors = append(errors, fmt.Sprintf("Template should contain %s placeholder", placeholder))
		}
	}

	// Check for unknown placeholders
	validPlaceholders := []string{
		"[CustomerName]", "[AccountCode]", "[TotalBalance]", "[Current]",
		"[30Days]", "[60Days]", "[90Days]", "[120Days]", "[ContactPerson]",
		"[Telephone]", "[OverdueAmount]", "[OldestAgeBracket]",
	}

	// Find all placeholders in template
	placeholderRegex := regexp.MustCompile(`\[([^\]]+)\]`)
	matches := placeholderRegex.FindAllString(template, -1)

	for _, match := range matches {
		found := false
		for _, valid := range validPlaceholders {
			if match == valid {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, fmt.Sprintf("Unknown placeholder: %s", match))
		}
	}

	return errors
}