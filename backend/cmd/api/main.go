// Command api runs the POS REST API.
//
// @title                      POS UMKM API
// @version                    1.0
// @description                REST API Point of Sale untuk UMKM. Semua nilai uang dalam rupiah (integer), persentase dalam basis point (1100 = 11%).
// @BasePath                   /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Format: "Bearer {access_token}"
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

	"github.com/gin-gonic/gin"
	_ "github.com/posumkm/backend/docs"
	"github.com/posumkm/backend/internal/handler"
	"github.com/posumkm/backend/internal/repository/postgres"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/internal/storage"
	"github.com/posumkm/backend/pkg/config"
	"github.com/posumkm/backend/pkg/hash"
	"github.com/posumkm/backend/pkg/jwt"
	"github.com/posumkm/backend/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.AppEnv)
	slog.SetDefault(log)
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.DBAutoMigrate {
		if err := postgres.MigrateUp(cfg.DatabaseURL); err != nil {
			return err
		}
		log.Info("database migrations applied")
	}
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	files, err := storage.NewLocal(cfg.UploadDir, cfg.UploadBaseURL)
	if err != nil {
		return err
	}

	// Repositories
	txm := postgres.NewTxManager(pool)
	roleRepo := postgres.NewRoleRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	tokenRepo := postgres.NewRefreshTokenRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	unitRepo := postgres.NewUnitRepository(pool)
	productRepo := postgres.NewProductRepository(pool)
	movementRepo := postgres.NewStockMovementRepository(pool)
	shiftRepo := postgres.NewShiftRepository(pool)
	txRepo := postgres.NewTransactionRepository(pool)
	settingRepo := postgres.NewStoreSettingRepository(pool)
	reportRepo := postgres.NewReportRepository(pool, cfg.Timezone.String())

	// Services
	now := time.Now
	hasher := hash.NewBcrypt()
	tokens := jwt.NewManager(cfg.JWTAccessSecret, cfg.JWTAccessTTL)
	authSvc := service.NewAuthService(userRepo, tokenRepo, txm, hasher, tokens, cfg.JWTRefreshTTL, now)
	userSvc := service.NewUserService(userRepo, roleRepo, tokenRepo, txm, hasher, now)
	categorySvc := service.NewCategoryService(categoryRepo)
	unitSvc := service.NewUnitService(unitRepo)
	productSvc := service.NewProductService(productRepo, categoryRepo, unitRepo, movementRepo, txm, files)
	stockSvc := service.NewStockService(productRepo, movementRepo, txm)
	shiftSvc := service.NewShiftService(shiftRepo, txm, now)
	txSvc := service.NewTransactionService(txRepo, productRepo, movementRepo, shiftRepo, settingRepo, txm, cfg.Timezone, now)
	reportSvc := service.NewReportService(reportRepo, productRepo, cfg.Timezone, now)
	settingSvc := service.NewSettingService(settingRepo, files)

	router := handler.NewRouter(handler.RouterConfig{
		Logger:         log,
		Tokens:         tokens,
		AllowedOrigins: cfg.CORSAllowedOrigins,
		UploadDir:      cfg.UploadDir,
		Health:         pool.Ping,
	}, handler.Handlers{
		Auth:        handler.NewAuthHandler(authSvc),
		User:        handler.NewUserHandler(userSvc),
		Category:    handler.NewCategoryHandler(categorySvc),
		Unit:        handler.NewUnitHandler(unitSvc),
		Product:     handler.NewProductHandler(productSvc, cfg.MaxUploadSize),
		Stock:       handler.NewStockHandler(stockSvc, cfg.Timezone),
		Shift:       handler.NewShiftHandler(shiftSvc, cfg.Timezone),
		Transaction: handler.NewTransactionHandler(txSvc, cfg.Timezone),
		Report:      handler.NewReportHandler(reportSvc),
		Setting:     handler.NewSettingHandler(settingSvc, cfg.MaxUploadSize),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("server listening", "addr", srv.Addr, "swagger", "http://localhost:"+cfg.HTTPPort+"/swagger/index.html")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
