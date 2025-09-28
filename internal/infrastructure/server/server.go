package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"unipile-connector/internal/adapter/handler"
	"unipile-connector/internal/adapter/middleware"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/internal/infrastructure/client"
	"unipile-connector/internal/usecase/account"
	"unipile-connector/internal/usecase/user"
)

// Server holds server dependencies
type Server struct {
	router         *gin.Engine
	httpServer     *http.Server
	authHandler    *handler.AuthHandler
	accountHandler *handler.AccountHandler
	authMiddleware *middleware.AuthMiddleware
	logger         *logrus.Logger
}

// NewServer creates a new server instance
func NewServer(
	userRepo repository.UserRepository,
	accountRepo repository.AccountRepository,
	unipileClient *client.UnipileClient,
	logger *logrus.Logger,
) *Server {
	// Initialize use cases
	userUsecase := user.NewUserUsecase(userRepo)
	accountUsecase := account.NewAccountUsecase(accountRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userUsecase, logger)
	accountHandler := handler.NewAccountHandler(accountUsecase, unipileClient, logger)
	authMiddleware := middleware.NewAuthMiddleware(userUsecase, logger)

	// Setup router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	server := &Server{
		router:         router,
		authHandler:    authHandler,
		accountHandler: accountHandler,
		authMiddleware: authMiddleware,
		logger:         logger,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
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
		// Public routes
		api.POST("/auth/register", s.authHandler.Register)
		api.POST("/auth/login", s.authHandler.Login)

		// Protected routes
		protected := api.Group("/")
		protected.Use(s.authMiddleware.SimpleAuthMiddleware())
		{
			protected.GET("/auth/me", s.authHandler.GetCurrentUser)
			protected.POST("/accounts/linkedin/connect", s.accountHandler.ConnectLinkedIn)
			protected.GET("/accounts", s.accountHandler.GetUserAccounts)
			protected.GET("/accounts/linkedin", s.accountHandler.GetLinkedInAccount)
			protected.DELETE("/accounts/linkedin", s.accountHandler.DisconnectLinkedIn)
		}
	}
}

// serveHomePage serves the home page
func (s *Server) serveHomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Unipile Connector",
	})
}

// serveLoginPage serves the login page
func (s *Server) serveLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - Unipile Connector",
	})
}

// serveRegisterPage serves the register page
func (s *Server) serveRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title": "Register - Unipile Connector",
	})
}

// serveDashboardPage serves the dashboard page
func (s *Server) serveDashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Dashboard - Unipile Connector",
	})
}

// Run starts the server
func (s *Server) Run(addr string) error {
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
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	s.logger.Info("Gracefully shutting down server...")
	return s.httpServer.Shutdown(ctx)
}
