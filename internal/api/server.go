package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewEchoServer creates a new Echo instance configured with the Server handlers.
func NewEchoServer(server *Server, withRateLimit bool) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(middleware.Logger())  // Echo's built-in request logger
	e.Use(middleware.Recover()) // Panic recovery

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{"Content-Type"},
	}))

	// Health endpoint (no rate limit)
	e.GET("/health", server.HandleHealth)

	// API routes
	apiGroup := e.Group("/api/v1")

	// Read-only endpoints (no rate limit - lightweight, frequently polled)
	apiGroup.GET("/environments", server.HandleGetEnvironments)
	apiGroup.GET("/workers/stats", server.HandleGetWorkerStats)
	apiGroup.GET("/compile/:job_id", server.HandleGetJob)

	// Compilation endpoint (with optional rate limiting)
	// This is resource-intensive and should be rate-limited
	if withRateLimit {
		rateLimiter := NewRateLimiter(10, time.Minute)
		compileGroup := apiGroup.Group("")
		compileGroup.Use(RateLimitMiddleware(rateLimiter))
		compileGroup.POST("/compile", server.HandleCompile)
	} else {
		apiGroup.POST("/compile", server.HandleCompile)
	}

	return e
}
