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
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/logger"
	"github.com/llamagate/llamagate/internal/mcpclient"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/proxy"
	"github.com/llamagate/llamagate/internal/tools"
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
	var toolManager *tools.Manager
	var guardrails *tools.Guardrails
	if cfg.MCP != nil && cfg.MCP.Enabled {
		log.Info().Msg("Initializing MCP clients...")

		toolManager = tools.NewManager()

		// Create guardrails
		guardrails, err = tools.NewGuardrails(
			cfg.MCP.AllowTools,
			cfg.MCP.DenyTools,
			cfg.MCP.MaxToolRounds,
			cfg.MCP.MaxToolCallsPerRound,
			cfg.MCP.DefaultToolTimeout,
			cfg.MCP.MaxToolResultSize,
		)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create MCP guardrails")
		}

		// Initialize MCP clients for each configured server
		for _, serverCfg := range cfg.MCP.Servers {
			if !serverCfg.Enabled {
				log.Debug().
					Str("server", serverCfg.Name).
					Msg("MCP server disabled, skipping")
				continue
			}

			var client *mcpclient.Client
			var initErr error

			switch serverCfg.Transport {
			case "stdio":
				// Use server timeout or default
				timeout := serverCfg.Timeout
				if timeout == 0 {
					timeout = 30 * time.Second // Default timeout
				}

				client, initErr = mcpclient.NewClientWithTimeout(serverCfg.Name, serverCfg.Command, serverCfg.Args, serverCfg.Env, timeout)
				if initErr != nil {
					log.Error().
						Str("server", serverCfg.Name).
						Err(initErr).
						Msg("Failed to initialize MCP client")
					continue
				}
			case "sse":
				log.Warn().
					Str("server", serverCfg.Name).
					Msg("SSE transport not yet implemented, skipping")
				continue
			default:
				log.Error().
					Str("server", serverCfg.Name).
					Str("transport", serverCfg.Transport).
					Msg("Unknown transport type")
				continue
			}

			// Add client to tool manager
			if err := toolManager.AddClient(client); err != nil {
				log.Error().
					Str("server", serverCfg.Name).
					Err(err).
					Msg("Failed to add MCP client to tool manager")
				client.Close()
				continue
			}

			log.Info().
				Str("server", serverCfg.Name).
				Str("transport", serverCfg.Transport).
				Msg("MCP client initialized successfully")
		}

		// Set tool manager and guardrails on proxy
		proxyInstance.SetToolManager(toolManager, guardrails)

		toolCount := len(toolManager.GetAllTools())
		log.Info().
			Int("total_tools", toolCount).
			Msg("MCP initialization complete")
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

	// Auth middleware (if API key is configured)
	if cfg.APIKey != "" {
		router.Use(middleware.AuthMiddleware(cfg.APIKey))
		log.Info().Msg("API key authentication enabled")
	} else {
		log.Warn().Msg("API key authentication disabled (API_KEY not set)")
	}

	// Rate limiting middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(cfg.RateLimitRPS)
	router.Use(rateLimitMiddleware.Handler())

	// Health check endpoint - verifies both server and Ollama connectivity
	router.GET("/health", func(c *gin.Context) {
		// Check Ollama connectivity with a timeout
		healthClient := &http.Client{
			Timeout: 5 * time.Second, // Short timeout for health check
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

	// OpenAI-compatible endpoints
	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions", proxyInstance.HandleChatCompletions)
		v1.GET("/models", proxyInstance.HandleModels)
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
	if toolManager != nil {
		if err := toolManager.CloseAll(); err != nil {
			log.Warn().Err(err).Msg("Error closing MCP clients")
		} else {
			log.Info().Msg("MCP clients closed")
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
