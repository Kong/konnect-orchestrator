package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/Kong/konnect-orchestrator/internal/config"
	services "github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/server/handlers"
	"github.com/Kong/konnect-orchestrator/internal/server/middleware"
)

func RunServer(platformGitConfig manifest.GitConfig, applyHealth chan string, version, commit, date string) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up services
	authService := services.NewAuthService(cfg)
	githubService := services.NewGitHubService(authService)

	// Set up handlers
	healthHandler := handlers.NewHealthHandler(applyHealth, version, commit, date)
	authHandler := handlers.NewAuthHandler(authService, cfg)
	userHandler := handlers.NewUserHandler(githubService)
	repoHandler := handlers.NewRepoHandler(githubService)
	orgHandler := handlers.NewOrgHandler(githubService)
	platformHandler := handlers.NewPlatformHandler(githubService, platformGitConfig)

	// Set up middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Set up router
	router := setupRouter(cfg, healthHandler, authHandler, userHandler, repoHandler, orgHandler, platformHandler, authMiddleware)

	// Set up server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
	// TODO: Move this to the outer cmd package
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	// log.Println("Shutting down server...")

	// // Create a deadline for the shutdown
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// // Shut down the server
	// if err := server.Shutdown(ctx); err != nil {
	// 	log.Fatal("Server forced to shutdown:", err)
	// }

	// log.Println("Server exited")
}

// setupRouter sets up the router
func setupRouter(
	cfg *config.Config,
	healthHandler *handlers.HealthHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	repoHandler *handlers.RepoHandler,
	orgHandler *handlers.OrgHandler,
	platformHandler *handlers.PlatformHandler,
	authMiddleware *middleware.AuthMiddleware,
) *gin.Engine {
	// Set mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Load HTML templates
	//router.LoadHTMLGlob("templates/*")

	// Set up CSRF protection with cookie sessions
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(24 * time.Hour.Seconds()),
		Secure:   cfg.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	router.Use(sessions.Sessions("github_session", store))

	// Set up security headers
	router.Use(securityHeaders())

	// Set up CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{cfg.FrontendURL} // Only allow your frontend
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"}
	corsConfig.ExposeHeaders = []string{"X-Refresh-Token"}
	corsConfig.MaxAge = 12 * 60 * 60 // 12 hours
	router.Use(cors.New(corsConfig))

	router.GET("/health", healthHandler.HealthCheck)

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.GET("/github", authHandler.Login)
		auth.GET("/github/callback", authHandler.Callback)
		auth.GET("/verify", authHandler.VerifyCode) // Add this new route
		auth.POST("/logout", csrfProtected(), authHandler.Logout)
		auth.POST("/refresh", csrfProtected(), authMiddleware.RequireAuth(), authHandler.RefreshToken)
	}

	// API routes
	api := router.Group("/api")
	api.Use(authMiddleware.RequireAuth(), authMiddleware.RefreshToken())
	{
		// User routes
		api.GET("/user", userHandler.GetProfile)

		// Organization routes
		api.GET("/orgs", orgHandler.ListOrganizations)

		// Repository routes
		api.GET("/repos", repoHandler.ListRepositories)
		api.GET("/users/:username/repos", repoHandler.ListUserRepositories)
		api.GET("/orgs/:org/repos", repoHandler.ListOrganizationRepositories)
		api.GET("/repos/:owner/:repo/contents/*path", repoHandler.GetRepositoryContent)
		api.GET("/enterprise/:server/orgs", repoHandler.ListEnterpriseOrganizations)
		api.GET("/platform/pulls", platformHandler.GetRepositoryPullRequests)
		api.GET("/platform/service", platformHandler.GetExistingServices)
		api.POST("/platform/service", platformHandler.AddServiceRepo)

		// Any POST, PUT, DELETE or PATCH requests need CSRF protection
		apiWrite := api.Group("/")
		apiWrite.Use(csrfProtected())
		{
			// Add any write operations here (POST, PUT, DELETE, PATCH)
			// Example: apiWrite.POST("/repos/:owner/:repo/contents/*path", repoHandler.CreateContent)
		}
	}

	return router
}

// CSRF protection middleware
func csrfProtected() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for GET requests (as they should be idempotent)
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// Get the session
		session := sessions.Default(c)

		// Get the token from the header
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
			c.Abort()
			return
		}

		// Get the token from the session
		sessionToken := session.Get("csrf_token")
		if sessionToken == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "No CSRF token in session"})
			c.Abort()
			return
		}

		// Validate the token
		if token != sessionToken.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token invalid"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; connect-src 'self' https://api.github.com; img-src 'self' https://avatars.githubusercontent.com data:;")

		// Other security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
