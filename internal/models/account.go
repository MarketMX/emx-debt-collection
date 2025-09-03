package models

import (
	"time"

	"github.com/google/uuid"
)

// Account represents a parsed account from an age analysis report
type Account struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UploadID       uuid.UUID `json:"upload_id" db:"upload_id"`
	AccountCode    string    `json:"account_code" db:"account_code"`
	CustomerName   string    `json:"customer_name" db:"customer_name"`
	ContactPerson  *string   `json:"contact_person" db:"contact_person"`
	Telephone      string    `json:"telephone" db:"telephone"`
	AmountCurrent  float64   `json:"amount_current" db:"amount_current"`
	Amount30d      float64   `json:"amount_30d" db:"amount_30d"`
	Amount60d      float64   `json:"amount_60d" db:"amount_60d"`
	Amount90d      float64   `json:"amount_90d" db:"amount_90d"`
	Amount120d     float64   `json:"amount_120d" db:"amount_120d"`
	TotalBalance   float64   `json:"total_balance" db:"total_balance"`
	IsSelected     bool      `json:"is_selected" db:"is_selected"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// CreateAccountRequest represents the request payload for creating an account
type CreateAccountRequest struct {
	UploadID      uuid.UUID `json:"upload_id" validate:"required"`
	AccountCode   string    `json:"account_code" validate:"required"`
	CustomerName  string    `json:"customer_name" validate:"required"`
	ContactPerson *string   `json:"contact_person"`
	Telephone     string    `json:"telephone" validate:"required"`
	AmountCurrent float64   `json:"amount_current"`
	Amount30d     float64   `json:"amount_30d"`
	Amount60d     float64   `json:"amount_60d"`
	Amount90d     float64   `json:"amount_90d"`
	Amount120d    float64   `json:"amount_120d"`
	TotalBalance  float64   `json:"total_balance" validate:"required"`
}

// UpdateAccountRequest represents the request payload for updating an account
type UpdateAccountRequest struct {
	AccountCode   *string  `json:"account_code"`
	CustomerName  *string  `json:"customer_name"`
	ContactPerson *string  `json:"contact_person"`
	Telephone     *string  `json:"telephone"`
	AmountCurrent *float64 `json:"amount_current"`
	Amount30d     *float64 `json:"amount_30d"`
	Amount60d     *float64 `json:"amount_60d"`
	Amount90d     *float64 `json:"amount_90d"`
	Amount120d    *float64 `json:"amount_120d"`
	TotalBalance  *float64 `json:"total_balance"`
	IsSelected    *bool    `json:"is_selected"`
}

// BulkUpdateSelectionRequest represents the request for bulk updating selection status
type BulkUpdateSelectionRequest struct {
	AccountIDs []uuid.UUID `json:"account_ids" validate:"required,min=1"`
	IsSelected bool        `json:"is_selected"`
}

// AccountResponse represents the response payload for account data
type AccountResponse struct {
	ID             uuid.UUID `json:"id"`
	UploadID       uuid.UUID `json:"upload_id"`
	AccountCode    string    `json:"account_code"`
	CustomerName   string    `json:"customer_name"`
	ContactPerson  *string   `json:"contact_person"`
	Telephone      string    `json:"telephone"`
	AmountCurrent  float64   `json:"amount_current"`
	Amount30d      float64   `json:"amount_30d"`
	Amount60d      float64   `json:"amount_60d"`
	Amount90d      float64   `json:"amount_90d"`
	Amount120d     float64   `json:"amount_120d"`
	TotalBalance   float64   `json:"total_balance"`
	IsSelected     bool      `json:"is_selected"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AccountListResponse represents a paginated list of accounts with metadata
type AccountListResponse struct {
	Accounts   []AccountResponse `json:"accounts"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
	HasMore    bool              `json:"has_more"`
}

// AccountSummary represents summary statistics for accounts
type AccountSummary struct {
	TotalAccounts        int     `json:"total_accounts"`
	SelectedAccounts     int     `json:"selected_accounts"`
	TotalBalance         float64 `json:"total_balance"`
	SelectedBalance      float64 `json:"selected_balance"`
	OverdueAccounts      int     `json:"overdue_accounts"`      // Accounts with balance > current
	SelectedOverdue      int     `json:"selected_overdue"`
	AverageBalance       float64 `json:"average_balance"`
	SelectedAvgBalance   float64 `json:"selected_avg_balance"`
}

// ToResponse converts an Account to AccountResponse
func (a *Account) ToResponse() AccountResponse {
	return AccountResponse{
		ID:             a.ID,
		UploadID:       a.UploadID,
		AccountCode:    a.AccountCode,
		CustomerName:   a.CustomerName,
		ContactPerson:  a.ContactPerson,
		Telephone:      a.Telephone,
		AmountCurrent:  a.AmountCurrent,
		Amount30d:      a.Amount30d,
		Amount60d:      a.Amount60d,
		Amount90d:      a.Amount90d,
		Amount120d:     a.Amount120d,
		TotalBalance:   a.TotalBalance,
		IsSelected:     a.IsSelected,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

// IsOverdue returns true if the account has overdue amounts (anything beyond current)
func (a *Account) IsOverdue() bool {
	return a.Amount30d > 0 || a.Amount60d > 0 || a.Amount90d > 0 || a.Amount120d > 0
}

// OverdueAmount returns the total overdue amount (excluding current)
func (a *Account) OverdueAmount() float64 {
	return a.Amount30d + a.Amount60d + a.Amount90d + a.Amount120d
}

// GetOldestOverdueAmount returns the oldest overdue amount and its age bracket
func (a *Account) GetOldestOverdueAmount() (amount float64, ageBracket string) {
	if a.Amount120d > 0 {
		return a.Amount120d, "120+ days"
	}
	if a.Amount90d > 0 {
		return a.Amount90d, "90-119 days"
	}
	if a.Amount60d > 0 {
		return a.Amount60d, "60-89 days"
	}
	if a.Amount30d > 0 {
		return a.Amount30d, "30-59 days"
	}
	return 0, "current"
}

// ExcelRow represents the structure expected from Excel parsing
type ExcelRow struct {
	Account      string  `excel:"Account"`
	Name         string  `excel:"Name"`
	Contact      *string `excel:"Contact"`
	Telephone    string  `excel:"Telephone"`
	Current      float64 `excel:"Current"`
	Days30       float64 `excel:"30 Days"`
	Days60       float64 `excel:"60 Days"`
	Days90       float64 `excel:"90 Days"`
	Days120      float64 `excel:"120 Days"`
	TotalBalance float64 `excel:"Total Balance"`
}

// ToAccount converts ExcelRow to Account
func (e *ExcelRow) ToAccount(uploadID uuid.UUID) Account {
	return Account{
		ID:            uuid.New(),
		UploadID:      uploadID,
		AccountCode:   e.Account,
		CustomerName:  e.Name,
		ContactPerson: e.Contact,
		Telephone:     e.Telephone,
		AmountCurrent: e.Current,
		Amount30d:     e.Days30,
		Amount60d:     e.Days60,
		Amount90d:     e.Days90,
		Amount120d:    e.Days120,
		TotalBalance:  e.TotalBalance,
		IsSelected:    false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
