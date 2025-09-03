package server

import (
	"net/http"

	"emx-debt-collection/internal/handlers"
	"emx-debt-collection/internal/middleware"
	"emx-debt-collection/internal/services"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	// Enhanced CORS configuration for frontend integration
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173",    // Vite dev server
			"http://localhost:3000",    // Common React dev port
			"https://localhost:5173",   // HTTPS dev server
			"https://localhost:3000",   // HTTPS React dev
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize services
	excelService := services.NewExcelService()
	accountService := services.NewAccountService(s.db.Repository())
	messagingService := services.NewMessagingService("https://api.messaging.example.com", "your-api-key-here")
	progressService := services.NewProgressService()
	
	// Initialize handlers
	uploadHandler := handlers.NewUploadHandler(s.db.Repository(), excelService, accountService, progressService)
	messagingHandler := handlers.NewMessagingHandler(s.db.Repository(), messagingService, accountService)
	adminHandler := handlers.NewAdminHandler(s.db.Repository())
	reportsHandler := handlers.NewReportsHandler(s.db.Repository())
	systemHandler := handlers.NewSystemHandler(s.db)

	// Public routes (no authentication required)
	public := e.Group("")
	public.GET("/", s.HelloWorldHandler)
	public.GET("/health", systemHandler.HealthCheck)
	public.GET("/health/ready", systemHandler.ReadinessCheck)
	public.GET("/health/live", systemHandler.LivenessCheck)
	public.GET("/api/info", systemHandler.GetAPIInfo)
	public.GET("/auth/config", s.authConfigHandler)

	// Protected routes (authentication required)
	protected := e.Group("/api")
	protected.Use(s.jwtMiddleware.RequireAuth())
	protected.Use(middleware.EnsureUserExists(s))

	// User routes
	protected.GET("/profile", s.profileHandler)
	protected.GET("/user/me", s.currentUserHandler)

	// Upload routes
	uploads := protected.Group("/uploads")
	uploads.POST("", uploadHandler.UploadFile)                              // POST /api/uploads
	uploads.GET("", uploadHandler.GetUploads)                               // GET /api/uploads
	uploads.GET("/:id", uploadHandler.GetUpload)                            // GET /api/uploads/{id}
	uploads.GET("/:id/accounts", uploadHandler.GetUploadAccounts)           // GET /api/uploads/{id}/accounts
	uploads.GET("/:id/summary", uploadHandler.GetUploadSummary)             // GET /api/uploads/{id}/summary
	uploads.GET("/:id/progress", uploadHandler.GetUploadProgress)           // GET /api/uploads/{id}/progress
	uploads.PUT("/:id/selection", uploadHandler.UpdateAccountSelection)     // PUT /api/uploads/{id}/selection
	
	// Messaging routes
	messaging := protected.Group("/messaging")
	messaging.POST("/send", messagingHandler.SendMessages)                  // POST /api/messaging/send
	messaging.GET("/templates", messagingHandler.GetMessageTemplates)       // GET /api/messaging/templates
	messaging.GET("/logs/:id", messagingHandler.GetMessageLogs)             // GET /api/messaging/logs/{upload_id}
	messaging.GET("/logs/:id/summary", messagingHandler.GetMessageLogSummary) // GET /api/messaging/logs/{upload_id}/summary

	// Global messaging logs (matches spec: GET /api/v1/logs/messaging)
	protected.GET("/logs/messaging", messagingHandler.GetAllMessageLogs)    // GET /api/logs/messaging

	// Reports routes
	reports := protected.Group("/reports")
	reports.GET("/upload/:id", reportsHandler.GetUploadReport)              // GET /api/reports/upload/{id}
	reports.GET("/user/activity", reportsHandler.GetUserActivityReport)     // GET /api/reports/user/activity
	reports.GET("/messaging", reportsHandler.GetMessagingReport)            // GET /api/reports/messaging

	// System information routes (protected)
	system := protected.Group("/system")
	system.GET("/info", systemHandler.GetSystemInfo)                        // GET /api/system/info
	system.GET("/metrics", systemHandler.GetSystemMetrics)                  // GET /api/system/metrics

	// Admin routes (requires admin role)
	admin := protected.Group("/admin")
	admin.Use(s.jwtMiddleware.RequireRole("admin"))
	admin.GET("/users", adminHandler.ListUsers)                             // GET /api/admin/users
	admin.GET("/users/:id", adminHandler.GetUser)                           // GET /api/admin/users/{id}
	admin.PUT("/users/:id", adminHandler.UpdateUser)                        // PUT /api/admin/users/{id}
	admin.GET("/users/:id/uploads", adminHandler.GetUserUploads)            // GET /api/admin/users/{id}/uploads
	admin.GET("/system/stats", adminHandler.GetSystemStats)                 // GET /api/admin/system/stats
	admin.GET("/reports/system", reportsHandler.GetSystemReport)            // GET /api/admin/reports/system

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}


// authConfigHandler provides Keycloak configuration for frontend
func (s *Server) authConfigHandler(c echo.Context) error {
	config := map[string]interface{}{
		"realm":      s.keycloakConfig.Realm,
		"serverUrl":  s.keycloakConfig.ServerURL,
		"clientId":   s.keycloakConfig.ClientID,
		"realmUrl":   s.keycloakConfig.RealmURL,
		"tokenUrl":   s.keycloakConfig.TokenEndpoint(),
		"userInfoUrl": s.keycloakConfig.UserInfoEndpoint(),
	}

	return c.JSON(http.StatusOK, config)
}

// profileHandler returns the current user's profile information
func (s *Server) profileHandler(c echo.Context) error {
	authUser, dbUser, err := s.userHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	profile := map[string]interface{}{
		"id":               authUser.Sub,
		"email":            authUser.Email,
		"username":         authUser.PreferredUsername,
		"name":             authUser.Name,
		"given_name":       authUser.GivenName,
		"family_name":      authUser.FamilyName,
		"email_verified":   authUser.EmailVerified,
		"realm_roles":      authUser.RealmAccess.Roles,
		"resource_access":  authUser.ResourceAccess,
	}

	// Include database user info if available
	if dbUser != nil {
		profile["db_user"] = dbUser.ToResponse()
	}

	return c.JSON(http.StatusOK, profile)
}

// currentUserHandler returns simplified current user information
func (s *Server) currentUserHandler(c echo.Context) error {
	authUser, dbUser, err := s.userHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	response := map[string]interface{}{
		"id":       authUser.Sub,
		"email":    authUser.Email,
		"username": authUser.PreferredUsername,
		"name":     authUser.Name,
		"roles":    authUser.RealmAccess.Roles,
	}

	// Include database user info if available
	if dbUser != nil {
		response["database_id"] = dbUser.ID
		response["is_active"] = dbUser.IsActive
		response["created_at"] = dbUser.CreatedAt
	}

	return c.JSON(http.StatusOK, response)
}

