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
	"unipile-connector/pkg/jwt"
	jwtMiddleware "unipile-connector/pkg/middleware"
)

// Server holds server dependencies
type Server struct {
	router         *gin.Engine
	httpServer     *http.Server
	authHandler    *handler.AuthHandler
	accountHandler *handler.AccountHandler
	authMiddleware *middleware.AuthMiddleware
	jwtMiddleware  *jwtMiddleware.JWTMiddleware
	logger         *logrus.Logger
}

// NewServer creates a new server instance
func NewServer(
	repos repository.Repositories,
	unipileClient *client.UnipileClient,
	logger *logrus.Logger,
	jwtSecretKey string,
	jwtIssuer string,
) *Server {
	// Initialize use cases
	userUsecase := user.NewUserUsecase(repos.User)
	accountUsecase := account.NewAccountUsecase(repos.Account, repos.Tx, unipileClient, logger)

	// Initialize JWT service
	jwtService := jwt.NewJWTService(jwtSecretKey, jwtIssuer)
	jwtMiddleware := jwtMiddleware.NewJWTMiddleware(jwtService)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userUsecase, jwtService, logger)
	accountHandler := handler.NewAccountHandler(accountUsecase)
	authMiddleware := middleware.NewAuthMiddleware(userUsecase, logger)

	// Setup router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	server := &Server{
		router:         router,
		authHandler:    authHandler,
		accountHandler: accountHandler,
		authMiddleware: authMiddleware,
		jwtMiddleware:  jwtMiddleware,
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
		// Public routes (Auth routes)
		api.POST("/auth/register", s.authHandler.Register)
		api.POST("/auth/login", s.authHandler.Login)
		api.POST("/auth/logout", s.authHandler.Logout)
		api.POST("/auth/refresh", s.authHandler.RefreshToken)

		// Protected routes
		protected := api.Group("/")
		protected.Use(s.jwtMiddleware.AuthMiddleware())
		{
			// Auth routes
			protected.GET("/auth/me", s.authHandler.GetCurrentUser)
			// Account routes
			protected.GET("/accounts", s.accountHandler.ListUserAccounts)
			protected.POST("/accounts/linkedin/connect", s.accountHandler.ConnectLinkedIn)
			protected.POST("/accounts/linkedin/checkpoint", s.accountHandler.SolveCheckpoint)
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
