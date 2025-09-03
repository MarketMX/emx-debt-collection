package handlers

import (
	"net/http"
	"runtime"
	"time"

	"emx-debt-collection/internal/database"

	"github.com/labstack/echo/v4"
)

// SystemHandler handles system-level operations and health checks
type SystemHandler struct {
	db database.Service
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(db database.Service) *SystemHandler {
	return &SystemHandler{
		db: db,
	}
}

// HealthCheck provides detailed health information
func (h *SystemHandler) HealthCheck(c echo.Context) error {
	health := h.db.Health()
	
	// Add additional system information
	systemInfo := map[string]interface{}{
		"timestamp":     time.Now(),
		"service":       "emx-debt-collection",
		"version":       "1.0.0", // You might want to make this configurable
		"environment":   "development", // You might want to make this configurable
		"database":      health,
		"system": map[string]interface{}{
			"go_version":      runtime.Version(),
			"num_goroutines":  runtime.NumGoroutine(),
			"memory_alloc":    getMemoryStats(),
		},
	}

	// Determine overall status
	status := "healthy"
	httpStatus := http.StatusOK
	
	if dbStatus := health["status"]; dbStatus != "up" {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	systemInfo["status"] = status

	return c.JSON(httpStatus, systemInfo)
}

// ReadinessCheck provides a readiness probe for Kubernetes/container orchestration
func (h *SystemHandler) ReadinessCheck(c echo.Context) error {
	health := h.db.Health()
	
	// Check if database is ready
	if dbStatus := health["status"]; dbStatus != "up" {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":    "not_ready",
			"timestamp": time.Now(),
			"checks": map[string]interface{}{
				"database": "failed",
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
		"checks": map[string]interface{}{
			"database": "passed",
		},
	})
}

// LivenessCheck provides a liveness probe for Kubernetes/container orchestration
func (h *SystemHandler) LivenessCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// GetSystemInfo returns general system information (admin only)
func (h *SystemHandler) GetSystemInfo(c echo.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := map[string]interface{}{
		"service": map[string]interface{}{
			"name":        "EMX Debt Collection API",
			"version":     "1.0.0",
			"description": "Age Analysis Messaging Application API",
			"build_time":  "2025-01-01", // You might want to set this during build
		},
		"runtime": map[string]interface{}{
			"go_version":     runtime.Version(),
			"compiler":       runtime.Compiler,
			"num_cpu":        runtime.NumCPU(),
			"num_goroutines": runtime.NumGoroutine(),
			"os":             runtime.GOOS,
			"arch":           runtime.GOARCH,
		},
		"memory": map[string]interface{}{
			"alloc_mb":        bToMb(m.Alloc),
			"total_alloc_mb":  bToMb(m.TotalAlloc),
			"sys_mb":          bToMb(m.Sys),
			"gc_runs":         m.NumGC,
		},
		"database": h.db.Health(),
		"features": map[string]interface{}{
			"excel_parsing":     true,
			"messaging":         true,
			"keycloak_auth":     true,
			"file_uploads":      true,
			"progress_tracking": true,
		},
		"timestamp": time.Now(),
	}

	return c.JSON(http.StatusOK, info)
}

// GetSystemMetrics returns system performance metrics (admin only)
func (h *SystemHandler) GetSystemMetrics(c echo.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]interface{}{
		"memory": map[string]interface{}{
			"heap_alloc":      m.HeapAlloc,
			"heap_sys":        m.HeapSys,
			"heap_idle":       m.HeapIdle,
			"heap_inuse":      m.HeapInuse,
			"heap_released":   m.HeapReleased,
			"heap_objects":    m.HeapObjects,
			"stack_inuse":     m.StackInuse,
			"stack_sys":       m.StackSys,
			"gc_runs":         m.NumGC,
			"gc_pause_total":  m.PauseTotalNs,
		},
		"runtime": map[string]interface{}{
			"goroutines":    runtime.NumGoroutine(),
			"cgo_calls":     runtime.NumCgoCall(),
		},
		"database": h.db.Health(),
		"timestamp": time.Now(),
	}

	return c.JSON(http.StatusOK, metrics)
}

// GetAPIInfo returns API documentation information
func (h *SystemHandler) GetAPIInfo(c echo.Context) error {
	info := map[string]interface{}{
		"name":        "EMX Debt Collection API",
		"version":     "v1",
		"description": "REST API for the Age Analysis Messaging Application",
		"documentation": map[string]interface{}{
			"openapi_version": "3.0.0",
			"spec_url":       "/api/docs/openapi.json", // Could be implemented later
		},
		"endpoints": map[string]interface{}{
			"authentication": []string{
				"GET /auth/config",
			},
			"users": []string{
				"GET /api/user/me",
				"GET /api/profile",
			},
			"uploads": []string{
				"POST /api/uploads",
				"GET /api/uploads",
				"GET /api/uploads/{id}",
				"GET /api/uploads/{id}/accounts",
				"GET /api/uploads/{id}/summary",
				"GET /api/uploads/{id}/progress",
				"PUT /api/uploads/{id}/selection",
			},
			"messaging": []string{
				"POST /api/messaging/send",
				"GET /api/messaging/templates",
				"GET /api/messaging/logs/{upload_id}",
				"GET /api/messaging/logs/{upload_id}/summary",
			},
			"reports": []string{
				"GET /api/reports/upload/{id}",
				"GET /api/reports/user/activity",
				"GET /api/reports/messaging",
			},
			"admin": []string{
				"GET /api/admin/users",
				"GET /api/admin/users/{id}",
				"PUT /api/admin/users/{id}",
				"GET /api/admin/users/{id}/uploads",
				"GET /api/admin/system/stats",
				"GET /api/admin/reports/system",
			},
			"system": []string{
				"GET /health",
				"GET /health/ready",
				"GET /health/live",
				"GET /api/system/info",
				"GET /api/system/metrics",
			},
		},
		"authentication": map[string]interface{}{
			"type":        "Bearer JWT",
			"description": "Keycloak-issued JWT tokens required for protected endpoints",
			"header":      "Authorization: Bearer <token>",
		},
		"rate_limits": map[string]interface{}{
			"upload":    "10 requests per minute",
			"messaging": "100 requests per minute",
			"default":   "1000 requests per minute",
		},
		"timestamp": time.Now(),
	}

	return c.JSON(http.StatusOK, info)
}

// Helper functions

func getMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"alloc_mb":       bToMb(m.Alloc),
		"total_alloc_mb": bToMb(m.TotalAlloc),
		"sys_mb":         bToMb(m.Sys),
		"num_gc":         m.NumGC,
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}