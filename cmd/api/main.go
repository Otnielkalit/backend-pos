package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Otnielkalit/backend-pos/internal/infrastructure/cache"
	"github.com/Otnielkalit/backend-pos/internal/infrastructure/config"
	"github.com/Otnielkalit/backend-pos/internal/infrastructure/database"
	"github.com/Otnielkalit/backend-pos/internal/infrastructure/logger"
	"github.com/Otnielkalit/backend-pos/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// @title           POS Inventory Backend API
// @version         1.0
// @description     Backend API untuk sistem pencatatan stok & transaksi retail/grosir.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@pos-backend.local

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"
func main() {
	// ── 1. Load configuration ──────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	// ── 2. Init logger ─────────────────────────────────────────────────────────
	logger.Init(cfg.App.Name, cfg.App.Env)

	log.Info().Str("env", cfg.App.Env).Msg("starting application")

	// ── 3. Init database ───────────────────────────────────────────────────────
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("database connected")

	// ── 4. Init cache ──────────────────────────────────────────────────────────
	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Msg("redis connected")

	// ── 5. Setup Gin ───────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New() // Use gin.New() instead of gin.Default() to control middleware manually

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS([]string{"*"})) // TODO: replace "*" with actual client origins in production

	// Health check — no auth required
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Feature routes will be registered here as features are implemented.
		// Example (uncomment when auth feature is ready):
		// authHandler := auth.NewHandler(...)
		// authHandler.RegisterRoutes(v1)
		_ = v1 // prevent "declared and not used" error
	}

	// ── 6. Start server with graceful shutdown ─────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in background goroutine
	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Block until SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited cleanly")
}
