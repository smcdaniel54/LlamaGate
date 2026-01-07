// Package main is the entry point for the LlamaGate server.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/api"
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/logger"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/proxy"
	"github.com/llamagate/llamagate/internal/setup"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Debug, cfg.LogFile)
	log.Info().
		Str("ollama_host", cfg.OllamaHost).
		Str("port", cfg.Port).
		Float64("rate_limit_rps", cfg.RateLimitRPS).
		Bool("debug", cfg.Debug).
		Msg("Starting LlamaGate")

	// Initialize cache
	cacheInstance := cache.New()

	// Initialize proxy with configurable timeout
	proxyInstance := proxy.NewWithTimeout(cfg.OllamaHost, cacheInstance, cfg.Timeout)

	// Initialize MCP clients if enabled
	var mcpComponents *setup.MCPComponents
	if cfg.MCP != nil && cfg.MCP.Enabled {
		var initErr error
		mcpComponents, initErr = setup.InitializeMCP(cfg.MCP)
		if initErr != nil {
			log.Fatal().Err(initErr).Msg("Failed to initialize MCP")
		}

		// Configure proxy with MCP components
		if mcpComponents != nil {
			setup.ConfigureProxy(proxyInstance, mcpComponents, cfg.MCP)
		}
	} else {
		log.Debug().Msg("MCP not enabled")
	}

	// Set Gin mode
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())

	// Logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID := c.GetString("request_id")

		if raw != "" {
			path = fmt.Sprintf("%s?%s", path, raw)
		}

		log.Info().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Msg("HTTP request")
	})

	// Health check endpoint - register BEFORE auth middleware
	// This allows monitoring tools to check health without API key
	router.GET("/health", func(c *gin.Context) {
		// Check Ollama connectivity with a timeout
		healthClient := &http.Client{
			Timeout: cfg.HealthCheckTimeout,
		}

		ollamaHealthURL := fmt.Sprintf("%s/api/tags", cfg.OllamaHost)
		resp, err := healthClient.Get(ollamaHealthURL)
		// Register defer immediately after request to ensure body is always closed
		// This handles both success and error cases where resp might be non-nil
		if resp != nil {
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					log.Warn().
						Str("request_id", c.GetString("request_id")).
						Err(closeErr).
						Msg("Failed to close health check response body")
				}
			}()
		}
		if err != nil {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Err(err).
				Str("ollama_host", cfg.OllamaHost).
				Msg("Health check: Ollama unreachable")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":      "unhealthy",
				"error":       "Ollama unreachable",
				"ollama_host": cfg.OllamaHost,
			})
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Int("status", resp.StatusCode).
				Str("ollama_host", cfg.OllamaHost).
				Msg("Health check: Ollama returned non-OK status")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":      "unhealthy",
				"error":       fmt.Sprintf("Ollama returned status %d", resp.StatusCode),
				"ollama_host": cfg.OllamaHost,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"ollama":      "connected",
			"ollama_host": cfg.OllamaHost,
		})
	})

	// Auth middleware (if API key is configured)
	// Applied to all routes registered AFTER this point
	if cfg.APIKey != "" {
		router.Use(middleware.AuthMiddleware(cfg.APIKey))
		log.Info().Msg("API key authentication enabled")
	} else {
		log.Warn().Msg("API key authentication disabled (API_KEY not set)")
	}

	// Rate limiting middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(cfg.RateLimitRPS)
	router.Use(rateLimitMiddleware.Handler())

	// All routes below will require authentication when API_KEY is set
	// OpenAI-compatible endpoints
	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions", proxyInstance.HandleChatCompletions)
		v1.GET("/models", proxyInstance.HandleModels)

		// MCP management endpoints
		if mcpComponents != nil && mcpComponents.ServerManager != nil {
			toolExecTimeout := 30 * time.Second // Default
			if cfg.MCP != nil && cfg.MCP.ToolExecutionTimeout > 0 {
				toolExecTimeout = cfg.MCP.ToolExecutionTimeout
			}
			mcpHandler := api.NewMCPHandler(mcpComponents.ToolManager, mcpComponents.ServerManager, toolExecTimeout)
			mcp := v1.Group("/mcp")
			{
				// Server management
				mcp.GET("/servers", mcpHandler.ListServers)
				mcp.GET("/servers/health", mcpHandler.GetAllHealth)
				mcp.GET("/servers/:name", mcpHandler.GetServer)
				mcp.GET("/servers/:name/health", mcpHandler.GetServerHealth)
				mcp.GET("/servers/:name/stats", mcpHandler.GetServerStats)

				// Tools, Resources, Prompts
				mcp.GET("/servers/:name/tools", mcpHandler.ListServerTools)
				mcp.GET("/servers/:name/resources", mcpHandler.ListServerResources)
				mcp.GET("/servers/:name/resources/*uri", mcpHandler.ReadServerResource)
				mcp.GET("/servers/:name/prompts", mcpHandler.ListServerPrompts)
				mcp.POST("/servers/:name/prompts/:promptName", mcpHandler.GetServerPrompt)

				// Unified execute endpoint
				mcp.POST("/execute", mcpHandler.ExecuteTool)

				// Refresh and management
				mcp.POST("/servers/:name/refresh", mcpHandler.RefreshServerMetadata)
			}
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("address", srv.Addr).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Stop cache cleanup goroutine
	cacheInstance.StopCleanup()

	// Close MCP clients
	if mcpComponents != nil {
		if mcpComponents.ServerManager != nil {
			if err := mcpComponents.ServerManager.Close(); err != nil {
				log.Warn().Err(err).Msg("Error closing server manager")
			} else {
				log.Info().Msg("Server manager closed")
			}
		}
		if mcpComponents.ToolManager != nil {
			if err := mcpComponents.ToolManager.CloseAll(); err != nil {
				log.Warn().Err(err).Msg("Error closing MCP clients")
			} else {
				log.Info().Msg("MCP clients closed")
			}
		}
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
		cancel() // Ensure cancel is called before exit
		os.Exit(1)
	}

	// Close log file handle to prevent file descriptor leak
	logger.Close()

	log.Info().Msg("Server exited gracefully")
}
