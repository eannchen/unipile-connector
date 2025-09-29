package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/adapter/handler"
	"unipile-connector/internal/adapter/middleware"
)

// Server holds server dependencies
type Server interface {
	Run(addr string) error
	Shutdown(ctx context.Context) error
}

// Impl holds server dependencies
type Impl struct {
	router      *gin.Engine
	httpServer  *http.Server
	middlewares *middleware.Middlewares
	handlers    *handler.Handlers
}

// NewServer creates a new server instance
func NewServer(
	middlewares *middleware.Middlewares,
	handlers *handler.Handlers,
) Server {

	// Setup router
	router := gin.Default()
	router.Use(middlewares.CORSMiddleware, middlewares.RateLimitMiddleware)

	server := &Impl{
		router:      router,
		middlewares: middlewares,
		handlers:    handlers,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all routes
func (s *Impl) setupRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Serve static files
	s.router.Static("/static", "./web/static")
	s.router.LoadHTMLGlob("web/templates/*")

	// Frontend routes
	s.router.GET("/", s.serveHomePage)
	s.router.GET("/login", s.serveLoginPage)
	s.router.GET("/register", s.serveRegisterPage)
	s.router.GET("/dashboard", s.serveDashboardPage)

	// API routes
	api := s.router.Group("/api/v1")
	{
		// Public routes (Auth routes)
		api.POST("/auth/register", s.handlers.AuthHandler.Register)
		api.POST("/auth/login", s.handlers.AuthHandler.Login)
		api.POST("/auth/logout", s.handlers.AuthHandler.Logout)
		api.POST("/auth/refresh", s.handlers.AuthHandler.RefreshToken)

		// Protected routes
		protected := api.Group("/")
		protected.Use(s.middlewares.JWTMiddleware)
		{
			// Auth routes
			protected.GET("/auth/me", s.handlers.AuthHandler.GetCurrentUser)
			// Account routes
			protected.GET("/accounts", s.handlers.AccountHandler.ListUserAccounts)
			protected.POST("/accounts/linkedin/connect", s.handlers.AccountHandler.ConnectLinkedIn)
			protected.POST("/accounts/linkedin/checkpoint", s.handlers.AccountHandler.SolveCheckpoint)
			protected.DELETE("/accounts/linkedin", s.handlers.AccountHandler.DisconnectLinkedIn)
		}
	}
}

// serveHomePage serves the home page
func (s *Impl) serveHomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Unipile Connector",
	})
}

// serveLoginPage serves the login page
func (s *Impl) serveLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - Unipile Connector",
	})
}

// serveRegisterPage serves the register page
func (s *Impl) serveRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title": "Register - Unipile Connector",
	})
}

// serveDashboardPage serves the dashboard page
func (s *Impl) serveDashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Dashboard - Unipile Connector",
	})
}

// Run starts the server
func (s *Impl) Run(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Impl) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
