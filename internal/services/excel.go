package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"emx-debt-collection/internal/models"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ExcelService handles Excel file parsing operations
type ExcelService struct{}

// NewExcelService creates a new Excel service instance
func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// ParseProgress tracks the progress of Excel parsing
type ParseProgress struct {
	TotalRows     int `json:"total_rows"`
	ProcessedRows int `json:"processed_rows"`
	SuccessRows   int `json:"success_rows"`
	FailedRows    int `json:"failed_rows"`
	IsComplete    bool `json:"is_complete"`
}

// ExcelParseResult contains the results of parsing an Excel file
type ExcelParseResult struct {
	Accounts       []models.CreateAccountRequest `json:"accounts"`
	TotalRows      int                           `json:"total_rows"`
	ProcessedRows  int                           `json:"processed_rows"`
	SuccessRows    int                           `json:"success_rows"`
	FailedRows     int                           `json:"failed_rows"`
	Errors         []string                      `json:"errors,omitempty"`
	ValidationLogs []string                      `json:"validation_logs,omitempty"`
}

// ExcelColumnMapping defines the expected column headers and their variations
type ExcelColumnMapping struct {
	Account      []string
	Name         []string
	Contact      []string
	Telephone    []string
	Current      []string
	Days30       []string
	Days60       []string
	Days90       []string
	Days120      []string
	TotalBalance []string
}

// getDefaultColumnMapping returns the default column mappings with common variations
func getDefaultColumnMapping() ExcelColumnMapping {
	return ExcelColumnMapping{
		Account: []string{"Acc. no.", "Account Number", "Account", "Account Code", "Acc Code", "AccCode", "Account #"},
		Name: []string{"Account holder", "Company Name", "Name", "Customer Name", "Customer", "Client Name", "Client", "Company"},
		Contact: []string{"Contact Person", "Contact", "Contact Name", "Rep", "Representative", "Account Manager"},
		Telephone: []string{"Contact number", "Telephone Number", "Telephone", "Phone", "Phone Number", "Mobile", "Cell", "Contact Number", "Tel"},
		Current: []string{"Current", "Current Balance", "0-30", "0 Days", "Current Amount"},
		Days30: []string{"30 days", "30 Days", "30-60", "30d", "31-60 Days", "30-59 Days", "30-60 Days"},
		Days60: []string{"60 days", "60 Days", "60-90", "60d", "61-90 Days", "60-89 Days", "60-90 Days"},
		Days90: []string{"90 days", "90 Days", "90-120", "90d", "91-120 Days", "90-119 Days", "90-120 Days"},
		Days120: []string{"120 days", "150+ days", "Over 120 Days", "120 Days", "120+", "120d", "120+ Days", "Over 120", "120 Plus", "120 Days+"},
		TotalBalance: []string{"Total", "Over All Balance", "Total Balance", "Balance", "Amount", "Outstanding", "Amount Outstanding", "Total Outstanding"},
	}
}

// ValidateExcelFile validates the Excel file format and structure
func (s *ExcelService) ValidateExcelFile(reader io.Reader) error {
	// Try to open the Excel file
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first worksheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return errors.New("Excel file contains no worksheets")
	}

	// Get rows from the first sheet
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return fmt.Errorf("failed to read rows from worksheet: %w", err)
	}

	if len(rows) < 2 {
		return errors.New("Excel file must contain at least a header row and one data row")
	}

	// Validate that we can find required columns
	mapping := getDefaultColumnMapping()
	headerRow := rows[0]
	
	// Log the actual headers found for debugging
	fmt.Printf("[Excel] Found headers in file: %v\n", headerRow)
	
	requiredMappings := map[string][]string{
		"Account":      mapping.Account,
		"Name":         mapping.Name,
		"Telephone":    mapping.Telephone,
		"TotalBalance": mapping.TotalBalance,
	}

	for fieldName, variations := range requiredMappings {
		found := false
		for _, header := range headerRow {
			cleanHeader := strings.TrimSpace(header)
			for _, variation := range variations {
				if strings.EqualFold(cleanHeader, variation) {
					found = true
					fmt.Printf("[Excel] Matched '%s' -> '%s' (field: %s)\n", cleanHeader, variation, fieldName)
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("required column '%s' not found in Excel file. Headers found: %v. Expected one of: %v", fieldName, headerRow, variations)
		}
	}

	return nil
}

// ParseExcelFile parses an Excel file and returns account data
func (s *ExcelService) ParseExcelFile(ctx context.Context, reader io.Reader, uploadID uuid.UUID, progressCallback func(ParseProgress)) (*ExcelParseResult, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first worksheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("Excel file contains no worksheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from worksheet: %w", err)
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel file must contain at least a header row and one data row")
	}

	result := &ExcelParseResult{
		Accounts:       []models.CreateAccountRequest{},
		TotalRows:      len(rows) - 1, // Excluding header
		ProcessedRows:  0,
		SuccessRows:    0,
		FailedRows:     0,
		Errors:         []string{},
		ValidationLogs: []string{},
	}

	// Parse header row to map columns
	headerRow := rows[0]
	fmt.Printf("[Excel Parse] Found headers in file: %v\n", headerRow)
	columnIndexes, err := s.mapColumns(headerRow)
	if err != nil {
		return nil, fmt.Errorf("failed to map columns: %w", err)
	}
	fmt.Printf("[Excel Parse] Column mappings: %v\n", columnIndexes)

	// Send initial progress
	if progressCallback != nil {
		progressCallback(ParseProgress{
			TotalRows:     result.TotalRows,
			ProcessedRows: 0,
			SuccessRows:   0,
			FailedRows:    0,
			IsComplete:    false,
		})
	}

	// Process data rows
	for i, row := range rows[1:] {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result.ProcessedRows++
		rowNumber := i + 2 // +2 because we skip header and arrays are 0-indexed

		account, validationErrors := s.parseRowToAccount(row, columnIndexes, uploadID, rowNumber)
		
		if len(validationErrors) > 0 {
			result.FailedRows++
			for _, validationError := range validationErrors {
				errorMsg := fmt.Sprintf("Row %d: %s", rowNumber, validationError)
				result.Errors = append(result.Errors, errorMsg)
				result.ValidationLogs = append(result.ValidationLogs, errorMsg)
			}
		} else {
			result.SuccessRows++
			result.Accounts = append(result.Accounts, account)
		}

		// Send progress updates every 10 rows or on last row
		if progressCallback != nil && (result.ProcessedRows%10 == 0 || result.ProcessedRows == result.TotalRows) {
			progressCallback(ParseProgress{
				TotalRows:     result.TotalRows,
				ProcessedRows: result.ProcessedRows,
				SuccessRows:   result.SuccessRows,
				FailedRows:    result.FailedRows,
				IsComplete:    result.ProcessedRows == result.TotalRows,
			})
		}
	}

	return result, nil
}

// mapColumns maps Excel column headers to their indexes
func (s *ExcelService) mapColumns(headerRow []string) (map[string]int, error) {
	mapping := getDefaultColumnMapping()
	columnIndexes := make(map[string]int)

	// Required fields
	requiredFields := map[string][]string{
		"Account":      mapping.Account,
		"Name":         mapping.Name,
		"Telephone":    mapping.Telephone,
		"TotalBalance": mapping.TotalBalance,
	}

	// Optional fields
	optionalFields := map[string][]string{
		"Contact":  mapping.Contact,
		"Current":  mapping.Current,
		"Days30":   mapping.Days30,
		"Days60":   mapping.Days60,
		"Days90":   mapping.Days90,
		"Days120":  mapping.Days120,
	}

	// Map required fields
	for fieldName, variations := range requiredFields {
		found := false
		for colIndex, header := range headerRow {
			for _, variation := range variations {
				if strings.EqualFold(strings.TrimSpace(header), variation) {
					columnIndexes[fieldName] = colIndex
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("required column '%s' not found. Expected one of: %v", fieldName, variations)
		}
	}

	// Map optional fields
	for fieldName, variations := range optionalFields {
		for colIndex, header := range headerRow {
			for _, variation := range variations {
				if strings.EqualFold(strings.TrimSpace(header), variation) {
					columnIndexes[fieldName] = colIndex
					break
				}
			}
		}
	}

	return columnIndexes, nil
}

// parseRowToAccount parses a single row into an Account model
func (s *ExcelService) parseRowToAccount(row []string, columnIndexes map[string]int, uploadID uuid.UUID, rowNumber int) (models.CreateAccountRequest, []string) {
	var errors []string
	
	account := models.CreateAccountRequest{
		UploadID: uploadID,
	}

	// Parse required fields
	if idx, exists := columnIndexes["Account"]; exists && idx < len(row) {
		account.AccountCode = strings.TrimSpace(row[idx])
		if account.AccountCode == "" {
			errors = append(errors, "Account code is required")
		}
	} else {
		errors = append(errors, "Account code column not found or empty")
	}

	if idx, exists := columnIndexes["Name"]; exists && idx < len(row) {
		account.CustomerName = strings.TrimSpace(row[idx])
		if account.CustomerName == "" {
			errors = append(errors, "Customer name is required")
		}
	} else {
		errors = append(errors, "Customer name column not found or empty")
	}

	if idx, exists := columnIndexes["Telephone"]; exists && idx < len(row) {
		telephone := strings.TrimSpace(row[idx])
		if telephone == "" {
			errors = append(errors, "Telephone is required")
		} else {
			account.Telephone = s.cleanPhoneNumber(telephone)
		}
	} else {
		errors = append(errors, "Telephone column not found or empty")
	}

	if idx, exists := columnIndexes["TotalBalance"]; exists && idx < len(row) {
		totalBalance, err := s.parseAmount(row[idx])
		if err != nil {
			errors = append(errors, fmt.Sprintf("Invalid total balance: %s", err.Error()))
		} else {
			account.TotalBalance = totalBalance
		}
	} else {
		errors = append(errors, "Total balance column not found or empty")
	}

	// Parse optional fields
	if idx, exists := columnIndexes["Contact"]; exists && idx < len(row) {
		contact := strings.TrimSpace(row[idx])
		if contact != "" {
			account.ContactPerson = &contact
		}
	}

	// Parse aging buckets (default to 0 if not found or invalid)
	account.AmountCurrent = s.parseAmountWithDefault(row, columnIndexes, "Current")
	account.Amount30d = s.parseAmountWithDefault(row, columnIndexes, "Days30")
	account.Amount60d = s.parseAmountWithDefault(row, columnIndexes, "Days60")
	account.Amount90d = s.parseAmountWithDefault(row, columnIndexes, "Days90")
	account.Amount120d = s.parseAmountWithDefault(row, columnIndexes, "Days120")

	// Validate that aging buckets sum approximately equals total balance (allow 1% variance)
	calculatedTotal := account.AmountCurrent + account.Amount30d + account.Amount60d + account.Amount90d + account.Amount120d
	if account.TotalBalance != 0 {
		variance := ((calculatedTotal - account.TotalBalance) / account.TotalBalance) * 100
		if variance < -1 || variance > 1 { // Allow 1% variance
			errors = append(errors, fmt.Sprintf("Aging buckets sum (%.2f) doesn't match total balance (%.2f)", calculatedTotal, account.TotalBalance))
		}
	}

	return account, errors
}

// parseAmountWithDefault parses an amount from a row with a default value of 0
func (s *ExcelService) parseAmountWithDefault(row []string, columnIndexes map[string]int, fieldName string) float64 {
	if idx, exists := columnIndexes[fieldName]; exists && idx < len(row) {
		amount, err := s.parseAmount(row[idx])
		if err == nil {
			return amount
		}
	}
	return 0.0
}

// parseAmount parses a string to a float64, handling common currency formats
func (s *ExcelService) parseAmount(value string) (float64, error) {
	if value == "" {
		return 0.0, nil
	}

	// Clean the string - remove currency symbols, commas, and whitespace
	cleaned := strings.TrimSpace(value)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "$", "")
	cleaned = strings.ReplaceAll(cleaned, "€", "")
	cleaned = strings.ReplaceAll(cleaned, "£", "")
	cleaned = strings.ReplaceAll(cleaned, "₹", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	// Handle parentheses as negative numbers (common in accounting)
	isNegative := false
	if strings.HasPrefix(cleaned, "(") && strings.HasSuffix(cleaned, ")") {
		isNegative = true
		cleaned = strings.TrimPrefix(cleaned, "(")
		cleaned = strings.TrimSuffix(cleaned, ")")
	}

	// Handle minus sign
	if strings.HasPrefix(cleaned, "-") {
		isNegative = true
		cleaned = strings.TrimPrefix(cleaned, "-")
	}

	if cleaned == "" {
		return 0.0, nil
	}

	amount, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0.0, fmt.Errorf("invalid number format: %s", value)
	}

	if isNegative {
		amount = -amount
	}

	return amount, nil
}

// cleanPhoneNumber cleans and validates a phone number
func (s *ExcelService) cleanPhoneNumber(phone string) string {
	// Remove common formatting characters
	cleaned := strings.TrimSpace(phone)
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	
	// Remove leading + or country codes if present
	if strings.HasPrefix(cleaned, "+") {
		cleaned = cleaned[1:]
	}
	
	return cleaned
}

// GetSupportedFormats returns the supported Excel file formats
func (s *ExcelService) GetSupportedFormats() []string {
	return []string{
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // .xlsx
		"application/vnd.ms-excel", // .xls
	}
}

// IsSupportedFormat checks if the given MIME type is supported
func (s *ExcelService) IsSupportedFormat(mimeType string) bool {
	supported := s.GetSupportedFormats()
	for _, format := range supported {
		if strings.EqualFold(mimeType, format) {
			return true
		}
	}
	return false
}