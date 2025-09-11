package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emx-debt-collection/internal/models"
	"github.com/google/uuid"
)

// Repository interface defines all database operations
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error)
	CreateOrUpdateUser(ctx context.Context, keycloakID, email, firstName, lastName string) (*models.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error)

	// Upload operations
	CreateUpload(ctx context.Context, userID uuid.UUID, req models.CreateUploadRequest) (*models.Upload, error)
	GetUploadByID(ctx context.Context, id uuid.UUID) (*models.Upload, error)
	GetUploadsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Upload, int64, error)
	UpdateUpload(ctx context.Context, id uuid.UUID, req models.UpdateUploadRequest) (*models.Upload, error)
	GetUploadSummary(ctx context.Context, id uuid.UUID) (*models.UploadSummary, error)
	ListUploadSummaries(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]models.UploadSummary, int64, error)

	// Account operations
	CreateAccount(ctx context.Context, req models.CreateAccountRequest) (*models.Account, error)
	CreateAccountsBatch(ctx context.Context, accounts []models.CreateAccountRequest) error
	GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	GetAccountsByUploadID(ctx context.Context, uploadID uuid.UUID, limit, offset int) ([]models.Account, int64, error)
	UpdateAccount(ctx context.Context, id uuid.UUID, req models.UpdateAccountRequest) (*models.Account, error)
	BulkUpdateAccountSelection(ctx context.Context, req models.BulkUpdateSelectionRequest) error
	GetAccountSummary(ctx context.Context, uploadID uuid.UUID) (*models.AccountSummary, error)
	GetSelectedAccounts(ctx context.Context, uploadID uuid.UUID) ([]models.Account, error)

	// Message log operations
	CreateMessageLog(ctx context.Context, req models.CreateMessageLogRequest) (*models.MessageLog, error)
	CreateMessageLogsBatch(ctx context.Context, logs []models.CreateMessageLogRequest) error
	GetMessageLogByID(ctx context.Context, id uuid.UUID) (*models.MessageLog, error)
	GetMessageLogsByUploadID(ctx context.Context, uploadID uuid.UUID, limit, offset int) ([]models.MessageLogResponse, int64, error)
	GetMessageLogsByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]models.MessageLogResponse, int64, error)
	UpdateMessageLog(ctx context.Context, id uuid.UUID, req models.UpdateMessageLogRequest) (*models.MessageLog, error)
	GetMessageLogSummary(ctx context.Context, uploadID uuid.UUID) (*models.MessageLogSummary, error)
	GetRetryableMessages(ctx context.Context, limit int) ([]models.MessageLog, error)
}

// repository implements the Repository interface
type repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(s Service) Repository {
	// Get the underlying *sql.DB from the service
	svc, ok := s.(*service)
	if !ok {
		panic("invalid service type")
	}
	return &repository{db: svc.db}
}

// User operations

func (r *repository) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		ID:               uuid.New(),
		KeycloakID:       req.KeycloakID,
		Email:            req.Email,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		IsActive:         true,
		EngageMXClientID: req.EngageMXClientID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	query := `
		INSERT INTO users (id, keycloak_id, email, first_name, last_name, is_active, engagemx_client_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.KeycloakID, user.Email, user.FirstName, user.LastName,
		user.IsActive, user.EngageMXClientID, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, keycloak_id, email, first_name, last_name, is_active, engagemx_client_id, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.FirstName, &user.LastName,
		&user.IsActive, &user.EngageMXClientID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *repository) GetUserByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, keycloak_id, email, first_name, last_name, is_active, engagemx_client_id, created_at, updated_at
		FROM users WHERE keycloak_id = $1`

	err := r.db.QueryRowContext(ctx, query, keycloakID).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.FirstName, &user.LastName,
		&user.IsActive, &user.EngageMXClientID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, keycloak_id, email, first_name, last_name, is_active, engagemx_client_id, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.FirstName, &user.LastName,
		&user.IsActive, &user.EngageMXClientID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *repository) UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *req.Email)
		argIndex++
	}

	if req.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, req.FirstName)
		argIndex++
	}

	if req.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, req.LastName)
		argIndex++
	}

	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if req.EngageMXClientID != nil {
		setParts = append(setParts, fmt.Sprintf("engagemx_client_id = $%d", argIndex))
		args = append(args, *req.EngageMXClientID)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetUserByID(ctx, id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, keycloak_id, email, first_name, last_name, is_active, engagemx_client_id, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.FirstName, &user.LastName,
		&user.IsActive, &user.EngageMXClientID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// CreateOrUpdateUser creates a new user or updates an existing one based on Keycloak ID
func (r *repository) CreateOrUpdateUser(ctx context.Context, keycloakID, email, firstName, lastName string) (*models.User, error) {
	// First try to get existing user
	existingUser, err := r.GetUserByKeycloakID(ctx, keycloakID)
	if err == nil {
		// User exists, update if needed
		updateReq := models.UpdateUserRequest{}
		needsUpdate := false
		
		if existingUser.Email != email {
			updateReq.Email = &email
			needsUpdate = true
		}
		
		if (firstName != "" && (existingUser.FirstName == nil || *existingUser.FirstName != firstName)) ||
		   (firstName == "" && existingUser.FirstName != nil) {
			if firstName == "" {
				updateReq.FirstName = nil
			} else {
				updateReq.FirstName = &firstName
			}
			needsUpdate = true
		}
		
		if (lastName != "" && (existingUser.LastName == nil || *existingUser.LastName != lastName)) ||
		   (lastName == "" && existingUser.LastName != nil) {
			if lastName == "" {
				updateReq.LastName = nil
			} else {
				updateReq.LastName = &lastName
			}
			needsUpdate = true
		}
		
		if needsUpdate {
			return r.UpdateUser(ctx, existingUser.ID, updateReq)
		}
		
		return existingUser, nil
	}
	
	// User doesn't exist, create new one
	createReq := models.CreateUserRequest{
		KeycloakID:       keycloakID,
		Email:            email,
		EngageMXClientID: "default", // Default client ID for existing users
	}
	
	if firstName != "" {
		createReq.FirstName = &firstName
	}
	if lastName != "" {
		createReq.LastName = &lastName
	}
	
	return r.CreateUser(ctx, createReq)
}

func (r *repository) ListUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get users
	query := `
		SELECT id, keycloak_id, email, first_name, last_name, is_active, engagemx_client_id, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.KeycloakID, &user.Email, &user.FirstName, &user.LastName,
			&user.IsActive, &user.EngageMXClientID, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

// Upload operations

func (r *repository) CreateUpload(ctx context.Context, userID uuid.UUID, req models.CreateUploadRequest) (*models.Upload, error) {
	id := uuid.New()
	upload := &models.Upload{
		ID:               id,
		UserID:           userID,
		Filename:         fmt.Sprintf("%s_%s", id.String(), req.OriginalFilename),
		OriginalFilename: req.OriginalFilename,
		FileSize:         req.FileSize,
		MimeType:         req.MimeType,
		Status:           string(models.UploadStatusPending),
		RecordsProcessed: 0,
		RecordsFailed:    0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	query := `
		INSERT INTO uploads (id, user_id, filename, original_filename, file_size, mime_type, status, records_processed, records_failed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		upload.ID, upload.UserID, upload.Filename, upload.OriginalFilename, upload.FileSize,
		upload.MimeType, upload.Status, upload.RecordsProcessed, upload.RecordsFailed,
		upload.CreatedAt, upload.UpdatedAt,
	).Scan(&upload.ID, &upload.CreatedAt, &upload.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create upload: %w", err)
	}

	return upload, nil
}

func (r *repository) GetUploadByID(ctx context.Context, id uuid.UUID) (*models.Upload, error) {
	upload := &models.Upload{}
	query := `
		SELECT id, user_id, filename, original_filename, file_path, file_size, mime_type, status,
		       processing_started_at, processing_completed_at, error_message, records_processed,
		       records_failed, created_at, updated_at
		FROM uploads WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&upload.ID, &upload.UserID, &upload.Filename, &upload.OriginalFilename, &upload.FilePath,
		&upload.FileSize, &upload.MimeType, &upload.Status, &upload.ProcessingStartedAt,
		&upload.ProcessingCompletedAt, &upload.ErrorMessage, &upload.RecordsProcessed,
		&upload.RecordsFailed, &upload.CreatedAt, &upload.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("upload not found")
		}
		return nil, fmt.Errorf("failed to get upload: %w", err)
	}

	return upload, nil
}

func (r *repository) GetUploadsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Upload, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM uploads WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get upload count: %w", err)
	}

	// Get uploads
	query := `
		SELECT id, user_id, filename, original_filename, file_path, file_size, mime_type, status,
		       processing_started_at, processing_completed_at, error_message, records_processed,
		       records_failed, created_at, updated_at
		FROM uploads
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list uploads: %w", err)
	}
	defer rows.Close()

	var uploads []models.Upload
	for rows.Next() {
		var upload models.Upload
		err := rows.Scan(
			&upload.ID, &upload.UserID, &upload.Filename, &upload.OriginalFilename, &upload.FilePath,
			&upload.FileSize, &upload.MimeType, &upload.Status, &upload.ProcessingStartedAt,
			&upload.ProcessingCompletedAt, &upload.ErrorMessage, &upload.RecordsProcessed,
			&upload.RecordsFailed, &upload.CreatedAt, &upload.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan upload: %w", err)
		}
		uploads = append(uploads, upload)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate uploads: %w", err)
	}

	return uploads, total, nil
}

func (r *repository) UpdateUpload(ctx context.Context, id uuid.UUID, req models.UpdateUploadRequest) (*models.Upload, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.ProcessingStartedAt != nil {
		setParts = append(setParts, fmt.Sprintf("processing_started_at = $%d", argIndex))
		args = append(args, req.ProcessingStartedAt)
		argIndex++
	}

	if req.ProcessingCompletedAt != nil {
		setParts = append(setParts, fmt.Sprintf("processing_completed_at = $%d", argIndex))
		args = append(args, req.ProcessingCompletedAt)
		argIndex++
	}

	if req.ErrorMessage != nil {
		setParts = append(setParts, fmt.Sprintf("error_message = $%d", argIndex))
		args = append(args, req.ErrorMessage)
		argIndex++
	}

	if req.RecordsProcessed != nil {
		setParts = append(setParts, fmt.Sprintf("records_processed = $%d", argIndex))
		args = append(args, *req.RecordsProcessed)
		argIndex++
	}

	if req.RecordsFailed != nil {
		setParts = append(setParts, fmt.Sprintf("records_failed = $%d", argIndex))
		args = append(args, *req.RecordsFailed)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetUploadByID(ctx, id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE uploads SET %s
		WHERE id = $%d
		RETURNING id, user_id, filename, original_filename, file_path, file_size, mime_type, status,
		         processing_started_at, processing_completed_at, error_message, records_processed,
		         records_failed, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	upload := &models.Upload{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&upload.ID, &upload.UserID, &upload.Filename, &upload.OriginalFilename, &upload.FilePath,
		&upload.FileSize, &upload.MimeType, &upload.Status, &upload.ProcessingStartedAt,
		&upload.ProcessingCompletedAt, &upload.ErrorMessage, &upload.RecordsProcessed,
		&upload.RecordsFailed, &upload.CreatedAt, &upload.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("upload not found")
		}
		return nil, fmt.Errorf("failed to update upload: %w", err)
	}

	return upload, nil
}

func (r *repository) GetUploadSummary(ctx context.Context, id uuid.UUID) (*models.UploadSummary, error) {
	summary := &models.UploadSummary{}
	query := `SELECT * FROM upload_summary WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&summary.ID, &summary.Filename, &summary.OriginalFilename, &summary.Status,
		&summary.CreatedAt, &summary.UserID, &summary.UserEmail, &summary.EngageMXClientID,
		&summary.TotalAccounts, &summary.SelectedAccounts, &summary.TotalBalanceSum, &summary.SelectedBalanceSum,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("upload summary not found")
		}
		return nil, fmt.Errorf("failed to get upload summary: %w", err)
	}

	return summary, nil
}

func (r *repository) ListUploadSummaries(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]models.UploadSummary, int64, error) {
	var args []interface{}
	var whereClause string
	argIndex := 1

	if userID != nil {
		whereClause = fmt.Sprintf(" WHERE user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM upload_summary%s", whereClause)
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get upload summary count: %w", err)
	}

	// Get summaries
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, filename, original_filename, status, created_at, user_id, user_email,
		       engagemx_client_id, total_accounts, selected_accounts, total_balance_sum, selected_balance_sum
		FROM upload_summary%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list upload summaries: %w", err)
	}
	defer rows.Close()

	var summaries []models.UploadSummary
	for rows.Next() {
		var summary models.UploadSummary
		err := rows.Scan(
			&summary.ID, &summary.Filename, &summary.OriginalFilename, &summary.Status,
			&summary.CreatedAt, &summary.UserID, &summary.UserEmail, &summary.EngageMXClientID,
			&summary.TotalAccounts, &summary.SelectedAccounts, &summary.TotalBalanceSum, &summary.SelectedBalanceSum,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan upload summary: %w", err)
		}
		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate upload summaries: %w", err)
	}

	return summaries, total, nil
}

// Account operations

func (r *repository) CreateAccount(ctx context.Context, req models.CreateAccountRequest) (*models.Account, error) {
	account := &models.Account{
		ID:            uuid.New(),
		UploadID:      req.UploadID,
		AccountCode:   req.AccountCode,
		CustomerName:  req.CustomerName,
		ContactPerson: req.ContactPerson,
		Telephone:     req.Telephone,
		AmountCurrent: req.AmountCurrent,
		Amount30d:     req.Amount30d,
		Amount60d:     req.Amount60d,
		Amount90d:     req.Amount90d,
		Amount120d:    req.Amount120d,
		TotalBalance:  req.TotalBalance,
		IsSelected:    false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO accounts (id, upload_id, account_code, customer_name, contact_person, telephone,
		                     amount_current, amount_30d, amount_60d, amount_90d, amount_120d,
		                     total_balance, is_selected, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		account.ID, account.UploadID, account.AccountCode, account.CustomerName,
		account.ContactPerson, account.Telephone, account.AmountCurrent,
		account.Amount30d, account.Amount60d, account.Amount90d, account.Amount120d,
		account.TotalBalance, account.IsSelected, account.CreatedAt, account.UpdatedAt,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

func (r *repository) CreateAccountsBatch(ctx context.Context, accounts []models.CreateAccountRequest) error {
	if len(accounts) == 0 {
		return nil
	}

	// Use a transaction for batch insert
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO accounts (id, upload_id, account_code, customer_name, contact_person, telephone,
		                     amount_current, amount_30d, amount_60d, amount_90d, amount_120d,
		                     total_balance, is_selected, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, req := range accounts {
		_, err = stmt.ExecContext(ctx,
			uuid.New(), req.UploadID, req.AccountCode, req.CustomerName,
			req.ContactPerson, req.Telephone, req.AmountCurrent,
			req.Amount30d, req.Amount60d, req.Amount90d, req.Amount120d,
			req.TotalBalance, false, now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert account: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *repository) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	account := &models.Account{}
	query := `
		SELECT id, upload_id, account_code, customer_name, contact_person, telephone,
		       amount_current, amount_30d, amount_60d, amount_90d, amount_120d,
		       total_balance, is_selected, created_at, updated_at
		FROM accounts WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.UploadID, &account.AccountCode, &account.CustomerName,
		&account.ContactPerson, &account.Telephone, &account.AmountCurrent,
		&account.Amount30d, &account.Amount60d, &account.Amount90d, &account.Amount120d,
		&account.TotalBalance, &account.IsSelected, &account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

func (r *repository) GetAccountsByUploadID(ctx context.Context, uploadID uuid.UUID, limit, offset int) ([]models.Account, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM accounts WHERE upload_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, uploadID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get account count: %w", err)
	}

	// Get accounts
	query := `
		SELECT id, upload_id, account_code, customer_name, contact_person, telephone,
		       amount_current, amount_30d, amount_60d, amount_90d, amount_120d,
		       total_balance, is_selected, created_at, updated_at
		FROM accounts
		WHERE upload_id = $1
		ORDER BY customer_name ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, uploadID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID, &account.UploadID, &account.AccountCode, &account.CustomerName,
			&account.ContactPerson, &account.Telephone, &account.AmountCurrent,
			&account.Amount30d, &account.Amount60d, &account.Amount90d, &account.Amount120d,
			&account.TotalBalance, &account.IsSelected, &account.CreatedAt, &account.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate accounts: %w", err)
	}

	return accounts, total, nil
}

func (r *repository) UpdateAccount(ctx context.Context, id uuid.UUID, req models.UpdateAccountRequest) (*models.Account, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.AccountCode != nil {
		setParts = append(setParts, fmt.Sprintf("account_code = $%d", argIndex))
		args = append(args, *req.AccountCode)
		argIndex++
	}

	if req.CustomerName != nil {
		setParts = append(setParts, fmt.Sprintf("customer_name = $%d", argIndex))
		args = append(args, *req.CustomerName)
		argIndex++
	}

	if req.ContactPerson != nil {
		setParts = append(setParts, fmt.Sprintf("contact_person = $%d", argIndex))
		args = append(args, req.ContactPerson)
		argIndex++
	}

	if req.Telephone != nil {
		setParts = append(setParts, fmt.Sprintf("telephone = $%d", argIndex))
		args = append(args, *req.Telephone)
		argIndex++
	}

	if req.AmountCurrent != nil {
		setParts = append(setParts, fmt.Sprintf("amount_current = $%d", argIndex))
		args = append(args, *req.AmountCurrent)
		argIndex++
	}

	if req.Amount30d != nil {
		setParts = append(setParts, fmt.Sprintf("amount_30d = $%d", argIndex))
		args = append(args, *req.Amount30d)
		argIndex++
	}

	if req.Amount60d != nil {
		setParts = append(setParts, fmt.Sprintf("amount_60d = $%d", argIndex))
		args = append(args, *req.Amount60d)
		argIndex++
	}

	if req.Amount90d != nil {
		setParts = append(setParts, fmt.Sprintf("amount_90d = $%d", argIndex))
		args = append(args, *req.Amount90d)
		argIndex++
	}

	if req.Amount120d != nil {
		setParts = append(setParts, fmt.Sprintf("amount_120d = $%d", argIndex))
		args = append(args, *req.Amount120d)
		argIndex++
	}

	if req.TotalBalance != nil {
		setParts = append(setParts, fmt.Sprintf("total_balance = $%d", argIndex))
		args = append(args, *req.TotalBalance)
		argIndex++
	}

	if req.IsSelected != nil {
		setParts = append(setParts, fmt.Sprintf("is_selected = $%d", argIndex))
		args = append(args, *req.IsSelected)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetAccountByID(ctx, id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE accounts SET %s
		WHERE id = $%d
		RETURNING id, upload_id, account_code, customer_name, contact_person, telephone,
		         amount_current, amount_30d, amount_60d, amount_90d, amount_120d,
		         total_balance, is_selected, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	account := &models.Account{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&account.ID, &account.UploadID, &account.AccountCode, &account.CustomerName,
		&account.ContactPerson, &account.Telephone, &account.AmountCurrent,
		&account.Amount30d, &account.Amount60d, &account.Amount90d, &account.Amount120d,
		&account.TotalBalance, &account.IsSelected, &account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

func (r *repository) BulkUpdateAccountSelection(ctx context.Context, req models.BulkUpdateSelectionRequest) error {
	if len(req.AccountIDs) == 0 {
		return fmt.Errorf("no account IDs provided")
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(req.AccountIDs))
	args := make([]interface{}, len(req.AccountIDs)+2)
	
	for i, id := range req.AccountIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	
	args[len(req.AccountIDs)] = req.IsSelected
	args[len(req.AccountIDs)+1] = time.Now()

	query := fmt.Sprintf(`
		UPDATE accounts 
		SET is_selected = $%d, updated_at = $%d
		WHERE id IN (%s)`,
		len(req.AccountIDs)+1, len(req.AccountIDs)+2, strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk update account selection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected != int64(len(req.AccountIDs)) {
		return fmt.Errorf("expected to update %d accounts, but updated %d", len(req.AccountIDs), rowsAffected)
	}

	return nil
}

func (r *repository) GetAccountSummary(ctx context.Context, uploadID uuid.UUID) (*models.AccountSummary, error) {
	summary := &models.AccountSummary{}
	query := `
		SELECT 
			COUNT(*) as total_accounts,
			COUNT(CASE WHEN is_selected THEN 1 END) as selected_accounts,
			COALESCE(SUM(total_balance), 0) as total_balance,
			COALESCE(SUM(CASE WHEN is_selected THEN total_balance ELSE 0 END), 0) as selected_balance,
			COUNT(CASE WHEN (amount_30d + amount_60d + amount_90d + amount_120d) > 0 THEN 1 END) as overdue_accounts,
			COUNT(CASE WHEN is_selected AND (amount_30d + amount_60d + amount_90d + amount_120d) > 0 THEN 1 END) as selected_overdue,
			CASE WHEN COUNT(*) > 0 THEN COALESCE(AVG(total_balance), 0) ELSE 0 END as average_balance,
			CASE WHEN COUNT(CASE WHEN is_selected THEN 1 END) > 0 THEN 
				COALESCE(AVG(CASE WHEN is_selected THEN total_balance END), 0) 
			ELSE 0 END as selected_avg_balance
		FROM accounts 
		WHERE upload_id = $1`

	err := r.db.QueryRowContext(ctx, query, uploadID).Scan(
		&summary.TotalAccounts, &summary.SelectedAccounts, &summary.TotalBalance,
		&summary.SelectedBalance, &summary.OverdueAccounts, &summary.SelectedOverdue,
		&summary.AverageBalance, &summary.SelectedAvgBalance,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get account summary: %w", err)
	}

	return summary, nil
}

func (r *repository) GetSelectedAccounts(ctx context.Context, uploadID uuid.UUID) ([]models.Account, error) {
	query := `
		SELECT id, upload_id, account_code, customer_name, contact_person, telephone,
		       amount_current, amount_30d, amount_60d, amount_90d, amount_120d,
		       total_balance, is_selected, created_at, updated_at
		FROM accounts
		WHERE upload_id = $1 AND is_selected = true
		ORDER BY customer_name ASC`

	rows, err := r.db.QueryContext(ctx, query, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get selected accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID, &account.UploadID, &account.AccountCode, &account.CustomerName,
			&account.ContactPerson, &account.Telephone, &account.AmountCurrent,
			&account.Amount30d, &account.Amount60d, &account.Amount90d, &account.Amount120d,
			&account.TotalBalance, &account.IsSelected, &account.CreatedAt, &account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate selected accounts: %w", err)
	}

	return accounts, nil
}

// Message log operations

func (r *repository) CreateMessageLog(ctx context.Context, req models.CreateMessageLogRequest) (*models.MessageLog, error) {
	messageLog := &models.MessageLog{
		ID:                 uuid.New(),
		AccountID:          req.AccountID,
		UploadID:           req.UploadID,
		UserID:             req.UserID,
		MessageTemplate:    req.MessageTemplate,
		MessageContent:     req.MessageContent,
		RecipientTelephone: req.RecipientTelephone,
		Status:             string(models.MessageStatusPending),
		RetryCount:         0,
		MaxRetries:         3,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if req.MaxRetries != nil {
		messageLog.MaxRetries = *req.MaxRetries
	}

	query := `
		INSERT INTO message_logs (id, account_id, upload_id, user_id, message_template, message_content,
		                         recipient_telephone, status, retry_count, max_retries, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		messageLog.ID, messageLog.AccountID, messageLog.UploadID, messageLog.UserID,
		messageLog.MessageTemplate, messageLog.MessageContent, messageLog.RecipientTelephone,
		messageLog.Status, messageLog.RetryCount, messageLog.MaxRetries,
		messageLog.CreatedAt, messageLog.UpdatedAt,
	).Scan(&messageLog.ID, &messageLog.CreatedAt, &messageLog.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create message log: %w", err)
	}

	return messageLog, nil
}

func (r *repository) CreateMessageLogsBatch(ctx context.Context, logs []models.CreateMessageLogRequest) error {
	if len(logs) == 0 {
		return nil
	}

	// Use a transaction for batch insert
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO message_logs (id, account_id, upload_id, user_id, message_template, message_content,
		                         recipient_telephone, status, retry_count, max_retries, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, req := range logs {
		maxRetries := 3
		if req.MaxRetries != nil {
			maxRetries = *req.MaxRetries
		}

		_, err = stmt.ExecContext(ctx,
			uuid.New(), req.AccountID, req.UploadID, req.UserID,
			req.MessageTemplate, req.MessageContent, req.RecipientTelephone,
			string(models.MessageStatusPending), 0, maxRetries, now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert message log: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *repository) GetMessageLogByID(ctx context.Context, id uuid.UUID) (*models.MessageLog, error) {
	messageLog := &models.MessageLog{}
	query := `
		SELECT id, account_id, upload_id, user_id, message_template, message_content,
		       recipient_telephone, status, external_message_id, sent_at, delivered_at,
		       failed_at, retry_count, max_retries, response_from_service, error_message,
		       created_at, updated_at
		FROM message_logs WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&messageLog.ID, &messageLog.AccountID, &messageLog.UploadID, &messageLog.UserID,
		&messageLog.MessageTemplate, &messageLog.MessageContent, &messageLog.RecipientTelephone,
		&messageLog.Status, &messageLog.ExternalMessageID, &messageLog.SentAt,
		&messageLog.DeliveredAt, &messageLog.FailedAt, &messageLog.RetryCount,
		&messageLog.MaxRetries, &messageLog.ResponseFromService, &messageLog.ErrorMessage,
		&messageLog.CreatedAt, &messageLog.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message log not found")
		}
		return nil, fmt.Errorf("failed to get message log: %w", err)
	}

	return messageLog, nil
}

func (r *repository) GetMessageLogsByUploadID(ctx context.Context, uploadID uuid.UUID, limit, offset int) ([]models.MessageLogResponse, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM message_logs WHERE upload_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, uploadID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get message log count: %w", err)
	}

	// Get message logs
	query := `
		SELECT ml.id, ml.account_id, ml.upload_id, ml.user_id, ml.message_template, ml.message_content,
		       ml.recipient_telephone, ml.status, ml.external_message_id, ml.sent_at, ml.delivered_at,
		       ml.failed_at, ml.retry_count, ml.max_retries, ml.response_from_service, ml.error_message,
		       ml.created_at, ml.updated_at
		FROM message_logs ml
		WHERE ml.upload_id = $1
		ORDER BY ml.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, uploadID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list message logs: %w", err)
	}
	defer rows.Close()

	var messageLogs []models.MessageLogResponse
	for rows.Next() {
		var messageLog models.MessageLog
		err := rows.Scan(
			&messageLog.ID, &messageLog.AccountID, &messageLog.UploadID, &messageLog.UserID,
			&messageLog.MessageTemplate, &messageLog.MessageContent, &messageLog.RecipientTelephone,
			&messageLog.Status, &messageLog.ExternalMessageID, &messageLog.SentAt,
			&messageLog.DeliveredAt, &messageLog.FailedAt, &messageLog.RetryCount,
			&messageLog.MaxRetries, &messageLog.ResponseFromService, &messageLog.ErrorMessage,
			&messageLog.CreatedAt, &messageLog.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message log: %w", err)
		}
		messageLogs = append(messageLogs, messageLog.ToResponse())
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate message logs: %w", err)
	}

	return messageLogs, total, nil
}

func (r *repository) GetMessageLogsByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]models.MessageLogResponse, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM message_logs WHERE account_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, accountID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get message log count: %w", err)
	}

	// Get message logs
	query := `
		SELECT id, account_id, upload_id, user_id, message_template, message_content,
		       recipient_telephone, status, external_message_id, sent_at, delivered_at,
		       failed_at, retry_count, max_retries, response_from_service, error_message,
		       created_at, updated_at
		FROM message_logs
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list message logs by account: %w", err)
	}
	defer rows.Close()

	var messageLogs []models.MessageLogResponse
	for rows.Next() {
		var messageLog models.MessageLog
		err := rows.Scan(
			&messageLog.ID, &messageLog.AccountID, &messageLog.UploadID, &messageLog.UserID,
			&messageLog.MessageTemplate, &messageLog.MessageContent, &messageLog.RecipientTelephone,
			&messageLog.Status, &messageLog.ExternalMessageID, &messageLog.SentAt,
			&messageLog.DeliveredAt, &messageLog.FailedAt, &messageLog.RetryCount,
			&messageLog.MaxRetries, &messageLog.ResponseFromService, &messageLog.ErrorMessage,
			&messageLog.CreatedAt, &messageLog.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message log: %w", err)
		}
		messageLogs = append(messageLogs, messageLog.ToResponse())
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate message logs: %w", err)
	}

	return messageLogs, total, nil
}

func (r *repository) UpdateMessageLog(ctx context.Context, id uuid.UUID, req models.UpdateMessageLogRequest) (*models.MessageLog, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.ExternalMessageID != nil {
		setParts = append(setParts, fmt.Sprintf("external_message_id = $%d", argIndex))
		args = append(args, req.ExternalMessageID)
		argIndex++
	}

	if req.SentAt != nil {
		setParts = append(setParts, fmt.Sprintf("sent_at = $%d", argIndex))
		args = append(args, req.SentAt)
		argIndex++
	}

	if req.DeliveredAt != nil {
		setParts = append(setParts, fmt.Sprintf("delivered_at = $%d", argIndex))
		args = append(args, req.DeliveredAt)
		argIndex++
	}

	if req.FailedAt != nil {
		setParts = append(setParts, fmt.Sprintf("failed_at = $%d", argIndex))
		args = append(args, req.FailedAt)
		argIndex++
	}

	if req.RetryCount != nil {
		setParts = append(setParts, fmt.Sprintf("retry_count = $%d", argIndex))
		args = append(args, *req.RetryCount)
		argIndex++
	}

	if req.ResponseFromService != nil {
		setParts = append(setParts, fmt.Sprintf("response_from_service = $%d", argIndex))
		args = append(args, req.ResponseFromService)
		argIndex++
	}

	if req.ErrorMessage != nil {
		setParts = append(setParts, fmt.Sprintf("error_message = $%d", argIndex))
		args = append(args, req.ErrorMessage)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetMessageLogByID(ctx, id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE message_logs SET %s
		WHERE id = $%d
		RETURNING id, account_id, upload_id, user_id, message_template, message_content,
		         recipient_telephone, status, external_message_id, sent_at, delivered_at,
		         failed_at, retry_count, max_retries, response_from_service, error_message,
		         created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	messageLog := &models.MessageLog{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&messageLog.ID, &messageLog.AccountID, &messageLog.UploadID, &messageLog.UserID,
		&messageLog.MessageTemplate, &messageLog.MessageContent, &messageLog.RecipientTelephone,
		&messageLog.Status, &messageLog.ExternalMessageID, &messageLog.SentAt,
		&messageLog.DeliveredAt, &messageLog.FailedAt, &messageLog.RetryCount,
		&messageLog.MaxRetries, &messageLog.ResponseFromService, &messageLog.ErrorMessage,
		&messageLog.CreatedAt, &messageLog.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message log not found")
		}
		return nil, fmt.Errorf("failed to update message log: %w", err)
	}

	return messageLog, nil
}

func (r *repository) GetMessageLogSummary(ctx context.Context, uploadID uuid.UUID) (*models.MessageLogSummary, error) {
	summary := &models.MessageLogSummary{}
	query := `
		SELECT upload_id, engagemx_client_id, total_messages, sent_messages, failed_messages, 
		       delivered_messages, first_sent_at, last_sent_at
		FROM message_log_summary 
		WHERE upload_id = $1`

	err := r.db.QueryRowContext(ctx, query, uploadID).Scan(
		&summary.UploadID, &summary.EngageMXClientID, &summary.TotalMessages, &summary.SentMessages,
		&summary.FailedMessages, &summary.DeliveredMessages, &summary.FirstSentAt,
		&summary.LastSentAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty summary if no messages exist yet
			return &models.MessageLogSummary{
				UploadID:          uploadID,
				EngageMXClientID:  "default", // Default client ID when no messages exist
				TotalMessages:     0,
				SentMessages:      0,
				FailedMessages:    0,
				DeliveredMessages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get message log summary: %w", err)
	}

	return summary, nil
}

func (r *repository) GetRetryableMessages(ctx context.Context, limit int) ([]models.MessageLog, error) {
	query := `
		SELECT id, account_id, upload_id, user_id, message_template, message_content,
		       recipient_telephone, status, external_message_id, sent_at, delivered_at,
		       failed_at, retry_count, max_retries, response_from_service, error_message,
		       created_at, updated_at
		FROM message_logs
		WHERE status = 'failed' AND retry_count < max_retries
		ORDER BY failed_at ASC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get retryable messages: %w", err)
	}
	defer rows.Close()

	var messageLogs []models.MessageLog
	for rows.Next() {
		var messageLog models.MessageLog
		err := rows.Scan(
			&messageLog.ID, &messageLog.AccountID, &messageLog.UploadID, &messageLog.UserID,
			&messageLog.MessageTemplate, &messageLog.MessageContent, &messageLog.RecipientTelephone,
			&messageLog.Status, &messageLog.ExternalMessageID, &messageLog.SentAt,
			&messageLog.DeliveredAt, &messageLog.FailedAt, &messageLog.RetryCount,
			&messageLog.MaxRetries, &messageLog.ResponseFromService, &messageLog.ErrorMessage,
			&messageLog.CreatedAt, &messageLog.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retryable message: %w", err)
		}
		messageLogs = append(messageLogs, messageLog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate retryable messages: %w", err)
	}

	return messageLogs, nil
}
