package services

import (
	"context"
	"fmt"
	"sort"

	"emx-debt-collection/internal/database"
	"emx-debt-collection/internal/models"

	"github.com/google/uuid"
)

// AccountService handles account data processing and analysis
type AccountService struct {
	repo database.Repository
}

// NewAccountService creates a new account service instance
func NewAccountService(repo database.Repository) *AccountService {
	return &AccountService{
		repo: repo,
	}
}

// AgeAnalysis represents age analysis data
type AgeAnalysis struct {
	TotalBalance    float64              `json:"total_balance"`
	SelectedBalance float64              `json:"selected_balance"`
	Buckets         []AgeBucket          `json:"buckets"`
	TopAccounts     []TopAccountAnalysis `json:"top_accounts"`
	Statistics      AnalysisStatistics   `json:"statistics"`
}

// AgeBucket represents an aging bucket analysis
type AgeBucket struct {
	Name            string  `json:"name"`
	DaysRange       string  `json:"days_range"`
	TotalAmount     float64 `json:"total_amount"`
	SelectedAmount  float64 `json:"selected_amount"`
	AccountCount    int     `json:"account_count"`
	SelectedCount   int     `json:"selected_count"`
	Percentage      float64 `json:"percentage"`
	SelectedPercentage float64 `json:"selected_percentage"`
}

// TopAccountAnalysis represents top account by balance
type TopAccountAnalysis struct {
	ID            uuid.UUID `json:"id"`
	AccountCode   string    `json:"account_code"`
	CustomerName  string    `json:"customer_name"`
	TotalBalance  float64   `json:"total_balance"`
	OverdueAmount float64   `json:"overdue_amount"`
	OldestBucket  string    `json:"oldest_bucket"`
	IsSelected    bool      `json:"is_selected"`
}

// AnalysisStatistics represents statistical analysis
type AnalysisStatistics struct {
	TotalAccounts         int     `json:"total_accounts"`
	SelectedAccounts      int     `json:"selected_accounts"`
	OverdueAccounts       int     `json:"overdue_accounts"`
	SelectedOverdue       int     `json:"selected_overdue"`
	AverageBalance        float64 `json:"average_balance"`
	SelectedAverageBalance float64 `json:"selected_average_balance"`
	MedianBalance         float64 `json:"median_balance"`
	LargestBalance        float64 `json:"largest_balance"`
	SmallestBalance       float64 `json:"smallest_balance"`
	OverduePercentage     float64 `json:"overdue_percentage"`
	SelectionRate         float64 `json:"selection_rate"`
}

// GetAgeAnalysis calculates comprehensive age analysis for an upload
func (s *AccountService) GetAgeAnalysis(ctx context.Context, uploadID uuid.UUID) (*AgeAnalysis, error) {
	// Get all accounts for the upload (without pagination for analysis)
	accounts, _, err := s.repo.GetAccountsByUploadID(ctx, uploadID, 10000, 0) // Large limit for analysis
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	if len(accounts) == 0 {
		return &AgeAnalysis{
			Buckets:     []AgeBucket{},
			TopAccounts: []TopAccountAnalysis{},
			Statistics:  AnalysisStatistics{},
		}, nil
	}

	// Calculate basic totals
	var totalBalance, selectedBalance float64
	var selectedCount, overdueCount, selectedOverdueCount int
	balances := make([]float64, 0, len(accounts))

	for _, account := range accounts {
		totalBalance += account.TotalBalance
		balances = append(balances, account.TotalBalance)
		
		if account.IsSelected {
			selectedCount++
			selectedBalance += account.TotalBalance
		}
		
		if account.IsOverdue() {
			overdueCount++
			if account.IsSelected {
				selectedOverdueCount++
			}
		}
	}

	// Calculate age buckets
	buckets := s.calculateAgeBuckets(accounts, totalBalance, selectedBalance)

	// Get top accounts by balance
	topAccounts := s.getTopAccounts(accounts, 10)

	// Calculate statistics
	statistics := s.calculateStatistics(accounts, balances, totalBalance, selectedBalance, selectedCount, overdueCount, selectedOverdueCount)

	return &AgeAnalysis{
		TotalBalance:    totalBalance,
		SelectedBalance: selectedBalance,
		Buckets:         buckets,
		TopAccounts:     topAccounts,
		Statistics:      statistics,
	}, nil
}

// calculateAgeBuckets calculates age bucket analysis
func (s *AccountService) calculateAgeBuckets(accounts []models.Account, totalBalance, selectedBalance float64) []AgeBucket {
	buckets := []AgeBucket{
		{Name: "Current", DaysRange: "0 days", TotalAmount: 0, SelectedAmount: 0, AccountCount: 0, SelectedCount: 0},
		{Name: "30 Days", DaysRange: "30-59 days", TotalAmount: 0, SelectedAmount: 0, AccountCount: 0, SelectedCount: 0},
		{Name: "60 Days", DaysRange: "60-89 days", TotalAmount: 0, SelectedAmount: 0, AccountCount: 0, SelectedCount: 0},
		{Name: "90 Days", DaysRange: "90-119 days", TotalAmount: 0, SelectedAmount: 0, AccountCount: 0, SelectedCount: 0},
		{Name: "120+ Days", DaysRange: "120+ days", TotalAmount: 0, SelectedAmount: 0, AccountCount: 0, SelectedCount: 0},
	}

	for _, account := range accounts {
		// Current bucket
		if account.AmountCurrent > 0 {
			buckets[0].TotalAmount += account.AmountCurrent
			buckets[0].AccountCount++
			if account.IsSelected {
				buckets[0].SelectedAmount += account.AmountCurrent
				buckets[0].SelectedCount++
			}
		}

		// 30 days bucket
		if account.Amount30d > 0 {
			buckets[1].TotalAmount += account.Amount30d
			buckets[1].AccountCount++
			if account.IsSelected {
				buckets[1].SelectedAmount += account.Amount30d
				buckets[1].SelectedCount++
			}
		}

		// 60 days bucket
		if account.Amount60d > 0 {
			buckets[2].TotalAmount += account.Amount60d
			buckets[2].AccountCount++
			if account.IsSelected {
				buckets[2].SelectedAmount += account.Amount60d
				buckets[2].SelectedCount++
			}
		}

		// 90 days bucket
		if account.Amount90d > 0 {
			buckets[3].TotalAmount += account.Amount90d
			buckets[3].AccountCount++
			if account.IsSelected {
				buckets[3].SelectedAmount += account.Amount90d
				buckets[3].SelectedCount++
			}
		}

		// 120+ days bucket
		if account.Amount120d > 0 {
			buckets[4].TotalAmount += account.Amount120d
			buckets[4].AccountCount++
			if account.IsSelected {
				buckets[4].SelectedAmount += account.Amount120d
				buckets[4].SelectedCount++
			}
		}
	}

	// Calculate percentages
	for i := range buckets {
		if totalBalance > 0 {
			buckets[i].Percentage = (buckets[i].TotalAmount / totalBalance) * 100
		}
		if selectedBalance > 0 {
			buckets[i].SelectedPercentage = (buckets[i].SelectedAmount / selectedBalance) * 100
		}
	}

	return buckets
}

// getTopAccounts returns the top accounts by balance
func (s *AccountService) getTopAccounts(accounts []models.Account, limit int) []TopAccountAnalysis {
	// Sort accounts by total balance descending
	sortedAccounts := make([]models.Account, len(accounts))
	copy(sortedAccounts, accounts)
	
	sort.Slice(sortedAccounts, func(i, j int) bool {
		return sortedAccounts[i].TotalBalance > sortedAccounts[j].TotalBalance
	})

	// Take top accounts up to limit
	if limit > len(sortedAccounts) {
		limit = len(sortedAccounts)
	}

	topAccounts := make([]TopAccountAnalysis, limit)
	for i := 0; i < limit; i++ {
		account := sortedAccounts[i]
		_, oldestBucket := account.GetOldestOverdueAmount()
		
		topAccounts[i] = TopAccountAnalysis{
			ID:            account.ID,
			AccountCode:   account.AccountCode,
			CustomerName:  account.CustomerName,
			TotalBalance:  account.TotalBalance,
			OverdueAmount: account.OverdueAmount(),
			OldestBucket:  oldestBucket,
			IsSelected:    account.IsSelected,
		}
	}

	return topAccounts
}

// calculateStatistics calculates comprehensive statistics
func (s *AccountService) calculateStatistics(accounts []models.Account, balances []float64, totalBalance, selectedBalance float64, selectedCount, overdueCount, selectedOverdueCount int) AnalysisStatistics {
	totalAccounts := len(accounts)
	
	var averageBalance, selectedAverageBalance float64
	if totalAccounts > 0 {
		averageBalance = totalBalance / float64(totalAccounts)
	}
	if selectedCount > 0 {
		selectedAverageBalance = selectedBalance / float64(selectedCount)
	}

	// Calculate median balance
	sortedBalances := make([]float64, len(balances))
	copy(sortedBalances, balances)
	sort.Float64s(sortedBalances)
	
	var medianBalance float64
	if len(sortedBalances) > 0 {
		if len(sortedBalances)%2 == 0 {
			mid := len(sortedBalances) / 2
			medianBalance = (sortedBalances[mid-1] + sortedBalances[mid]) / 2
		} else {
			medianBalance = sortedBalances[len(sortedBalances)/2]
		}
	}

	var largestBalance, smallestBalance float64
	if len(sortedBalances) > 0 {
		largestBalance = sortedBalances[len(sortedBalances)-1]
		smallestBalance = sortedBalances[0]
	}

	var overduePercentage, selectionRate float64
	if totalAccounts > 0 {
		overduePercentage = (float64(overdueCount) / float64(totalAccounts)) * 100
		selectionRate = (float64(selectedCount) / float64(totalAccounts)) * 100
	}

	return AnalysisStatistics{
		TotalAccounts:          totalAccounts,
		SelectedAccounts:       selectedCount,
		OverdueAccounts:        overdueCount,
		SelectedOverdue:        selectedOverdueCount,
		AverageBalance:         averageBalance,
		SelectedAverageBalance: selectedAverageBalance,
		MedianBalance:          medianBalance,
		LargestBalance:         largestBalance,
		SmallestBalance:        smallestBalance,
		OverduePercentage:      overduePercentage,
		SelectionRate:          selectionRate,
	}
}

// GetAccountSummary retrieves account summary for an upload
func (s *AccountService) GetAccountSummary(ctx context.Context, uploadID uuid.UUID) (*models.AccountSummary, error) {
	return s.repo.GetAccountSummary(ctx, uploadID)
}

// ProcessAccounts processes and validates account data
func (s *AccountService) ProcessAccounts(ctx context.Context, uploadID uuid.UUID, accounts []models.CreateAccountRequest) (*AccountProcessingResult, error) {
	result := &AccountProcessingResult{
		TotalAccounts:    len(accounts),
		ProcessedAccounts: 0,
		FailedAccounts:   0,
		ValidationErrors: []string{},
	}

	if len(accounts) == 0 {
		return result, nil
	}

	// Validate accounts
	validAccounts := []models.CreateAccountRequest{}
	for i, account := range accounts {
		errors := s.validateAccount(account, i+1)
		if len(errors) > 0 {
			result.FailedAccounts++
			result.ValidationErrors = append(result.ValidationErrors, errors...)
		} else {
			validAccounts = append(validAccounts, account)
		}
	}

	// Save valid accounts
	if len(validAccounts) > 0 {
		err := s.repo.CreateAccountsBatch(ctx, validAccounts)
		if err != nil {
			return nil, fmt.Errorf("failed to save accounts: %w", err)
		}
		result.ProcessedAccounts = len(validAccounts)
	}

	return result, nil
}

// AccountProcessingResult represents the result of account processing
type AccountProcessingResult struct {
	TotalAccounts     int      `json:"total_accounts"`
	ProcessedAccounts int      `json:"processed_accounts"`
	FailedAccounts    int      `json:"failed_accounts"`
	ValidationErrors  []string `json:"validation_errors,omitempty"`
}

// validateAccount validates a single account
func (s *AccountService) validateAccount(account models.CreateAccountRequest, rowNumber int) []string {
	var errors []string

	// Required field validations
	if account.AccountCode == "" {
		errors = append(errors, fmt.Sprintf("Row %d: Account code is required", rowNumber))
	}
	if account.CustomerName == "" {
		errors = append(errors, fmt.Sprintf("Row %d: Customer name is required", rowNumber))
	}
	if account.Telephone == "" {
		errors = append(errors, fmt.Sprintf("Row %d: Telephone is required", rowNumber))
	}

	// Business logic validations
	if account.TotalBalance < 0 {
		errors = append(errors, fmt.Sprintf("Row %d: Total balance cannot be negative", rowNumber))
	}

	// Validate aging bucket consistency
	calculatedTotal := account.AmountCurrent + account.Amount30d + account.Amount60d + account.Amount90d + account.Amount120d
	if account.TotalBalance > 0 {
		variance := ((calculatedTotal - account.TotalBalance) / account.TotalBalance) * 100
		if variance < -5 || variance > 5 { // Allow 5% variance for validation
			errors = append(errors, fmt.Sprintf("Row %d: Aging buckets sum (%.2f) doesn't match total balance (%.2f)", rowNumber, calculatedTotal, account.TotalBalance))
		}
	}

	// Validate individual bucket amounts
	if account.AmountCurrent < 0 {
		errors = append(errors, fmt.Sprintf("Row %d: Current amount cannot be negative", rowNumber))
	}
	if account.Amount30d < 0 {
		errors = append(errors, fmt.Sprintf("Row %d: 30 days amount cannot be negative", rowNumber))
	}
	if account.Amount60d < 0 {
		errors = append(errors, fmt.Sprintf("Row %d: 60 days amount cannot be negative", rowNumber))
	}
	if account.Amount90d < 0 {
		errors = append(errors, fmt.Sprintf("Row %d: 90 days amount cannot be negative", rowNumber))
	}
	if account.Amount120d < 0 {
		errors = append(errors, fmt.Sprintf("Row %d: 120+ days amount cannot be negative", rowNumber))
	}

	return errors
}

// GetSelectedAccountsForMessaging retrieves selected accounts ready for messaging
func (s *AccountService) GetSelectedAccountsForMessaging(ctx context.Context, uploadID uuid.UUID) ([]models.Account, error) {
	accounts, err := s.repo.GetSelectedAccounts(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get selected accounts: %w", err)
	}

	// Filter out accounts without valid phone numbers
	validAccounts := []models.Account{}
	for _, account := range accounts {
		if account.Telephone != "" && len(account.Telephone) >= 10 {
			validAccounts = append(validAccounts, account)
		}
	}

	return validAccounts, nil
}

// GetAccountDetails retrieves detailed information for a specific account
func (s *AccountService) GetAccountDetails(ctx context.Context, accountID uuid.UUID) (*models.Account, error) {
	return s.repo.GetAccountByID(ctx, accountID)
}

// BulkUpdateAccountSelection updates selection status for multiple accounts
func (s *AccountService) BulkUpdateAccountSelection(ctx context.Context, req models.BulkUpdateSelectionRequest) error {
	return s.repo.BulkUpdateAccountSelection(ctx, req)
}

// GetAccountRiskAnalysis provides risk analysis for accounts
func (s *AccountService) GetAccountRiskAnalysis(ctx context.Context, uploadID uuid.UUID) (*RiskAnalysis, error) {
	accounts, _, err := s.repo.GetAccountsByUploadID(ctx, uploadID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	riskAnalysis := &RiskAnalysis{
		HighRisk:   []RiskAccount{},
		MediumRisk: []RiskAccount{},
		LowRisk:    []RiskAccount{},
		Summary:    RiskSummary{},
	}

	var highRiskCount, mediumRiskCount, lowRiskCount int
	var totalHighRiskAmount, totalMediumRiskAmount, totalLowRiskAmount float64

	for _, account := range accounts {
		riskLevel, riskScore := s.calculateRiskLevel(account)
		
		riskAccount := RiskAccount{
			ID:            account.ID,
			AccountCode:   account.AccountCode,
			CustomerName:  account.CustomerName,
			TotalBalance:  account.TotalBalance,
			OverdueAmount: account.OverdueAmount(),
			RiskLevel:     riskLevel,
			RiskScore:     riskScore,
			IsSelected:    account.IsSelected,
		}

		switch riskLevel {
		case "high":
			riskAnalysis.HighRisk = append(riskAnalysis.HighRisk, riskAccount)
			highRiskCount++
			totalHighRiskAmount += account.TotalBalance
		case "medium":
			riskAnalysis.MediumRisk = append(riskAnalysis.MediumRisk, riskAccount)
			mediumRiskCount++
			totalMediumRiskAmount += account.TotalBalance
		case "low":
			riskAnalysis.LowRisk = append(riskAnalysis.LowRisk, riskAccount)
			lowRiskCount++
			totalLowRiskAmount += account.TotalBalance
		}
	}

	riskAnalysis.Summary = RiskSummary{
		TotalAccounts:        len(accounts),
		HighRiskCount:        highRiskCount,
		MediumRiskCount:      mediumRiskCount,
		LowRiskCount:         lowRiskCount,
		HighRiskAmount:       totalHighRiskAmount,
		MediumRiskAmount:     totalMediumRiskAmount,
		LowRiskAmount:        totalLowRiskAmount,
	}

	return riskAnalysis, nil
}

// RiskAnalysis represents risk analysis data
type RiskAnalysis struct {
	HighRisk   []RiskAccount `json:"high_risk"`
	MediumRisk []RiskAccount `json:"medium_risk"`
	LowRisk    []RiskAccount `json:"low_risk"`
	Summary    RiskSummary   `json:"summary"`
}

// RiskAccount represents an account with risk information
type RiskAccount struct {
	ID            uuid.UUID `json:"id"`
	AccountCode   string    `json:"account_code"`
	CustomerName  string    `json:"customer_name"`
	TotalBalance  float64   `json:"total_balance"`
	OverdueAmount float64   `json:"overdue_amount"`
	RiskLevel     string    `json:"risk_level"`
	RiskScore     float64   `json:"risk_score"`
	IsSelected    bool      `json:"is_selected"`
}

// RiskSummary represents risk analysis summary
type RiskSummary struct {
	TotalAccounts    int     `json:"total_accounts"`
	HighRiskCount    int     `json:"high_risk_count"`
	MediumRiskCount  int     `json:"medium_risk_count"`
	LowRiskCount     int     `json:"low_risk_count"`
	HighRiskAmount   float64 `json:"high_risk_amount"`
	MediumRiskAmount float64 `json:"medium_risk_amount"`
	LowRiskAmount    float64 `json:"low_risk_amount"`
}

// calculateRiskLevel calculates risk level and score for an account
func (s *AccountService) calculateRiskLevel(account models.Account) (string, float64) {
	score := 0.0

	// Base score from total balance (normalized to 0-30)
	if account.TotalBalance > 100000 {
		score += 30
	} else if account.TotalBalance > 50000 {
		score += 20
	} else if account.TotalBalance > 10000 {
		score += 10
	}

	// Age score (0-50 based on oldest overdue amount)
	if account.Amount120d > 0 {
		score += 50
	} else if account.Amount90d > 0 {
		score += 40
	} else if account.Amount60d > 0 {
		score += 30
	} else if account.Amount30d > 0 {
		score += 20
	}

	// Overdue ratio score (0-20)
	if account.TotalBalance > 0 {
		overdueRatio := account.OverdueAmount() / account.TotalBalance
		score += overdueRatio * 20
	}

	// Determine risk level
	if score >= 70 {
		return "high", score
	} else if score >= 40 {
		return "medium", score
	} else {
		return "low", score
	}
}