package services

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ProgressService tracks upload and processing progress
type ProgressService struct {
	mu        sync.RWMutex
	progress  map[uuid.UUID]*UploadProgress
	callbacks map[uuid.UUID][]ProgressCallback
}

// UploadProgress tracks the progress of an upload processing
type UploadProgress struct {
	UploadID      uuid.UUID `json:"upload_id"`
	Stage         string    `json:"stage"`         // "uploading", "parsing", "saving", "completed", "failed"
	TotalRows     int       `json:"total_rows"`
	ProcessedRows int       `json:"processed_rows"`
	SuccessRows   int       `json:"success_rows"`
	FailedRows    int       `json:"failed_rows"`
	CurrentRow    int       `json:"current_row"`
	Progress      float64   `json:"progress"`      // 0.0 to 1.0
	Message       string    `json:"message"`
	StartTime     time.Time `json:"start_time"`
	UpdateTime    time.Time `json:"update_time"`
	IsComplete    bool      `json:"is_complete"`
	HasError      bool      `json:"has_error"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// ProgressCallback is called when progress is updated
type ProgressCallback func(progress *UploadProgress)

// NewProgressService creates a new progress tracking service
func NewProgressService() *ProgressService {
	return &ProgressService{
		progress:  make(map[uuid.UUID]*UploadProgress),
		callbacks: make(map[uuid.UUID][]ProgressCallback),
	}
}

// StartTracking starts tracking progress for an upload
func (ps *ProgressService) StartTracking(uploadID uuid.UUID, totalRows int) *UploadProgress {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	progress := &UploadProgress{
		UploadID:      uploadID,
		Stage:         "starting",
		TotalRows:     totalRows,
		ProcessedRows: 0,
		SuccessRows:   0,
		FailedRows:    0,
		CurrentRow:    0,
		Progress:      0.0,
		Message:       "Starting upload processing...",
		StartTime:     time.Now(),
		UpdateTime:    time.Now(),
		IsComplete:    false,
		HasError:      false,
	}

	ps.progress[uploadID] = progress
	return progress
}

// UpdateProgress updates the progress for an upload
func (ps *ProgressService) UpdateProgress(uploadID uuid.UUID, update func(*UploadProgress)) {
	ps.mu.Lock()
	progress, exists := ps.progress[uploadID]
	if !exists {
		ps.mu.Unlock()
		return
	}

	// Apply the update
	update(progress)
	progress.UpdateTime = time.Now()

	// Calculate percentage
	if progress.TotalRows > 0 {
		progress.Progress = float64(progress.ProcessedRows) / float64(progress.TotalRows)
	}

	// Get callbacks while holding lock
	callbacks := append([]ProgressCallback(nil), ps.callbacks[uploadID]...)
	ps.mu.Unlock()

	// Call callbacks outside of lock to avoid deadlocks
	for _, callback := range callbacks {
		callback(progress)
	}
}

// SetStage sets the current processing stage
func (ps *ProgressService) SetStage(uploadID uuid.UUID, stage, message string) {
	ps.UpdateProgress(uploadID, func(p *UploadProgress) {
		p.Stage = stage
		p.Message = message
	})
}

// SetError sets an error state
func (ps *ProgressService) SetError(uploadID uuid.UUID, errorMessage string) {
	ps.UpdateProgress(uploadID, func(p *UploadProgress) {
		p.Stage = "failed"
		p.HasError = true
		p.ErrorMessage = errorMessage
		p.IsComplete = true
		p.Message = "Processing failed: " + errorMessage
	})
}

// SetComplete marks the upload as completed
func (ps *ProgressService) SetComplete(uploadID uuid.UUID, message string) {
	ps.UpdateProgress(uploadID, func(p *UploadProgress) {
		p.Stage = "completed"
		p.IsComplete = true
		p.Progress = 1.0
		if message != "" {
			p.Message = message
		} else {
			p.Message = "Processing completed successfully"
		}
	})
}

// IncrementRow increments the processed row count
func (ps *ProgressService) IncrementRow(uploadID uuid.UUID, success bool) {
	ps.UpdateProgress(uploadID, func(p *UploadProgress) {
		p.ProcessedRows++
		p.CurrentRow++
		if success {
			p.SuccessRows++
		} else {
			p.FailedRows++
		}
		p.Message = formatProgressMessage(p)
	})
}

// GetProgress gets the current progress for an upload
func (ps *ProgressService) GetProgress(uploadID uuid.UUID) (*UploadProgress, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	progress, exists := ps.progress[uploadID]
	if !exists {
		return nil, false
	}
	
	// Return a copy to avoid race conditions
	progressCopy := *progress
	return &progressCopy, true
}

// AddCallback adds a progress callback for an upload
func (ps *ProgressService) AddCallback(uploadID uuid.UUID, callback ProgressCallback) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	ps.callbacks[uploadID] = append(ps.callbacks[uploadID], callback)
}

// RemoveTracking removes tracking data for an upload (call after completion)
func (ps *ProgressService) RemoveTracking(uploadID uuid.UUID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	delete(ps.progress, uploadID)
	delete(ps.callbacks, uploadID)
}

// CleanupOldTracking removes tracking data older than the specified duration
func (ps *ProgressService) CleanupOldTracking(maxAge time.Duration) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	
	for uploadID, progress := range ps.progress {
		if progress.IsComplete && progress.UpdateTime.Before(cutoff) {
			delete(ps.progress, uploadID)
			delete(ps.callbacks, uploadID)
		}
	}
}

// GetAllProgress returns all current progress items (for admin/monitoring)
func (ps *ProgressService) GetAllProgress() map[uuid.UUID]*UploadProgress {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	result := make(map[uuid.UUID]*UploadProgress, len(ps.progress))
	for id, progress := range ps.progress {
		progressCopy := *progress
		result[id] = &progressCopy
	}
	
	return result
}

// formatProgressMessage creates a user-friendly progress message
func formatProgressMessage(p *UploadProgress) string {
	switch p.Stage {
	case "parsing":
		if p.TotalRows > 0 {
			return "Parsing Excel file... " + formatRowProgress(p.ProcessedRows, p.TotalRows)
		}
		return "Parsing Excel file..."
	case "saving":
		return "Saving accounts to database..."
	case "completed":
		return "Processing completed successfully"
	case "failed":
		return "Processing failed"
	default:
		return p.Message
	}
}

// formatRowProgress formats row progress display
func formatRowProgress(processed, total int) string {
	if total > 0 {
		percentage := int((float64(processed) / float64(total)) * 100)
		return "(" + string(rune(processed)) + "/" + string(rune(total)) + " rows, " + string(rune(percentage)) + "%)"
	}
	return "(" + string(rune(processed)) + " rows processed)"
}

// Duration returns how long the upload has been processing
func (p *UploadProgress) Duration() time.Duration {
	return p.UpdateTime.Sub(p.StartTime)
}

// EstimatedTimeRemaining estimates remaining processing time
func (p *UploadProgress) EstimatedTimeRemaining() time.Duration {
	if p.ProcessedRows <= 0 || p.TotalRows <= 0 || p.IsComplete {
		return 0
	}
	
	elapsed := p.Duration()
	if elapsed <= 0 {
		return 0
	}
	
	rowsRemaining := p.TotalRows - p.ProcessedRows
	timePerRow := elapsed / time.Duration(p.ProcessedRows)
	
	return timePerRow * time.Duration(rowsRemaining)
}