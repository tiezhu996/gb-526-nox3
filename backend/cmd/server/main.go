package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/auth"
	"commercial-diving-decompression-control/backend/internal/config"
	"commercial-diving-decompression-control/backend/internal/database"
	"commercial-diving-decompression-control/backend/internal/handler"
	"commercial-diving-decompression-control/backend/internal/middleware"
	"commercial-diving-decompression-control/backend/internal/repository"
	"commercial-diving-decompression-control/backend/internal/router"
	"commercial-diving-decompression-control/backend/internal/service"
	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration rejected", "error", err)
		os.Exit(1)
	}
	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	db, err := database.Open(cfg)
	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("database handle failed", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	auditRepo := audit.NewRepository(db)
	authRepo := auth.NewRepository(db)
	profileRepo := repository.NewDiverProfileRepository(db, auditRepo)
	planRepo := repository.NewDivePlanRepository(db, auditRepo)
	segmentRepo := repository.NewExposureSegmentRepository(db, auditRepo)
	assessmentRepo := repository.NewDecompressionAssessmentRepository(db, auditRepo)

	authService := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTTTL)
	profileService := service.NewDiverProfileService(profileRepo, planRepo)
	planService := service.NewDivePlanService(planRepo, profileRepo)
	segmentService := service.NewExposureSegmentService(segmentRepo, planRepo, cfg.MaxSegments)
	assessmentService := service.NewDecompressionAssessmentService(assessmentRepo, planRepo, profileRepo, segmentRepo, cfg.ModelVersion, cfg.MaxSegments)

	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.AccessLog(logger))
	engine.Use(middleware.CORS(cfg.CORSOrigins))
	engine.Use(middleware.NewLocalRateLimiter(cfg.RateLimitPerMinute).Handler())
	engine.GET("/healthz", func(c *gin.Context) { util.OK(c, gin.H{"status": "ok"}) })
	engine.GET("/readyz", func(c *gin.Context) {
		if err := database.Ping(db); err != nil {
			util.Fail(c, util.Internal(err))
			return
		}
		util.OK(c, gin.H{"status": "ready", "database": cfg.DBDriver, "model_version": cfg.ModelVersion})
	})

	api := engine.Group("/api/v1")
	api.POST("/auth/login", auth.NewHandler(authService).Login)
	protected := api.Group("")
	protected.Use(middleware.Auth(authService))
	write := middleware.RBAC(auth.RolePlanner, auth.RoleAdmin)
	review := middleware.RBAC(auth.RoleSupervisor, auth.RoleAdmin)
	router.RegisterDiverProfileRoutes(protected, handler.NewDiverProfileHandler(profileService), write)
	router.RegisterDivePlanRoutes(protected, handler.NewDivePlanHandler(planService), write, review)
	router.RegisterExposureSegmentRoutes(protected, handler.NewExposureSegmentHandler(segmentService), write)
	router.RegisterDecompressionAssessmentRoutes(protected, handler.NewDecompressionAssessmentHandler(assessmentService), write, review)
	protected.GET("/audit-events", review, audit.NewHandler(auditRepo).List)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: engine, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("server listening", "port", cfg.Port, "database", cfg.DBDriver, "model_version", cfg.ModelVersion)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownPeriod)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		return
	}
	logger.Info("server stopped")
}

func newLogger(level string) *slog.Logger {
	minimum := slog.LevelInfo
	switch level {
	case "debug":
		minimum = slog.LevelDebug
	case "warn":
		minimum = slog.LevelWarn
	case "error":
		minimum = slog.LevelError
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: minimum}))
}
