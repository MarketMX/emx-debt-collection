package utils

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// FileValidation contains file validation utilities
type FileValidation struct{}

// NewFileValidation creates a new file validation utility
func NewFileValidation() *FileValidation {
	return &FileValidation{}
}

// ValidateFileType checks if the file type is supported
func (fv *FileValidation) ValidateFileType(filename, mimeType string) error {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Supported Excel formats
	supportedExts := map[string][]string{
		".xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		".xls":  {"application/vnd.ms-excel"},
	}
	
	// Check if extension is supported
	supportedMimes, exists := supportedExts[ext]
	if !exists {
		return fmt.Errorf("unsupported file extension: %s. Supported formats: .xlsx, .xls", ext)
	}
	
	// If MIME type is provided, validate it matches the extension
	if mimeType != "" {
		mimeValid := false
		for _, supportedMime := range supportedMimes {
			if strings.EqualFold(mimeType, supportedMime) {
				mimeValid = true
				break
			}
		}
		if !mimeValid {
			return fmt.Errorf("MIME type %s doesn't match file extension %s", mimeType, ext)
		}
	}
	
	return nil
}

// GetMimeTypeFromExtension returns the MIME type for a file extension
func (fv *FileValidation) GetMimeTypeFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	default:
		return mime.TypeByExtension(ext)
	}
}

// ValidateFileSize checks if the file size is within limits
func (fv *FileValidation) ValidateFileSize(size int64, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("file too large: %d bytes. Maximum allowed: %d bytes (%d MB)", 
			size, maxSize, maxSize/(1024*1024))
	}
	if size <= 0 {
		return fmt.Errorf("file is empty")
	}
	return nil
}

// SanitizeFilename sanitizes a filename for safe storage
func (fv *FileValidation) SanitizeFilename(filename string) string {
	// Remove path separators and other potentially dangerous characters
	sanitized := strings.ReplaceAll(filename, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, "..", "_")
	
	// Remove or replace other special characters
	replacements := map[string]string{
		"<":  "_",
		">":  "_",
		":":  "_",
		"\"": "_",
		"|":  "_",
		"?":  "_",
		"*":  "_",
	}
	
	for old, new := range replacements {
		sanitized = strings.ReplaceAll(sanitized, old, new)
	}
	
	// Ensure filename is not empty after sanitization
	if sanitized == "" || sanitized == "." || sanitized == ".." {
		sanitized = "uploaded_file.xlsx"
	}
	
	return sanitized
}

// GetFileSizeHuman returns file size in human readable format
func (fv *FileValidation) GetFileSizeHuman(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}