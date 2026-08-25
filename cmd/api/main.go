package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/minio"
	oracledb "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/db"
	postgresaudit "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/audit"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/db"
	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	applicationpurchaseorder "github.com/efangly/thanes-lims-backend/internal/application/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/config"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// @title           Thanes LIMS Backend API
// @version         1.0
// @description     REST API สำหรับ Thanes LIMS backend (Auth, Sample/Chain-of-Custody, Test Result, Equipment, Inventory/Purchase Order, Document, Environment, Notification, Audit)
// @BasePath        /api/v1
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                พิมพ์ "Bearer" ตามด้วยเว้นวรรคและ JWT access token เช่น "Bearer eyJhbGciOi..."
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gdb, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	// Oracle ADB (chatbot POC) is optional - failure here must never crash
	// the main API, which stays fully functional on Postgres alone.
	if cfg.OracleDSN != "" {
		oracleDB, err := oracledb.New(cfg.OracleDSN, cfg.OracleTNSAdmin)
		if err != nil {
			log.Printf("oracle: connect failed, ADB features disabled: %v", err)
		} else {
			defer oracleDB.Close()
			log.Println("oracle: connected to ADB")
		}
	}

	// Composition root: wire adapters -> ports -> use cases -> handlers.
	auditRepo := postgresaudit.New(gdb)
	logAction := applicationaudit.NewLogActionUseCase(auditRepo)

	fileStorage, err := minio.New(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}
	if err := fileStorage.EnsureBucket(context.Background()); err != nil {
		log.Fatalf("minio: ensure bucket: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorMapper,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(fiberlogger.New())
	app.Use(corsMiddleware(cfg))
	app.Use(middleware.Audit(logAction))

	v1 := app.Group("/api/v1")
	autoReorderJob := registerRoutes(v1, cfg, gdb, fileStorage)

	go func() {
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	jobCtx, cancelJob := context.WithCancel(context.Background())
	jobDone := make(chan struct{})
	if cfg.AutoReorderEnabled {
		go runAutoReorderJob(jobCtx, jobDone, autoReorderJob, cfg.AutoReorderInterval)
	} else {
		close(jobDone)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancelJob()
	<-jobDone

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// corsMiddleware only needs real configuration for cross-origin deployments
// (e.g. a frontend dev server on a different port) - the Refresh Cookie
// requires AllowCredentials plus an explicit, non-wildcard AllowOrigins list
// to reach the browser at all (see ADR 0004). Same-origin production
// deployments never hit CORS in the first place, so the default (no
// AllowOrigins configured) keeps today's permissive behavior.
func corsMiddleware(cfg *config.Config) fiber.Handler {
	if len(cfg.CORSAllowOrigins) == 0 {
		return cors.New()
	}
	return cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", middleware.CSRFHeaderName},
	})
}

// runAutoReorderJob periodically scans inventory for items below their
// minimum threshold and reorders those with a configured default vendor.
// It runs once immediately on startup, then on every tick, and exits
// (closing done) as soon as ctx is cancelled during shutdown.
func runAutoReorderJob(ctx context.Context, done chan<- struct{}, job *applicationpurchaseorder.AutoReorderJob, interval time.Duration) {
	defer close(done)

	runOnce := func() {
		result, err := job.Run(ctx)
		if err != nil {
			log.Printf("auto-reorder: %v", err)
			return
		}
		if len(result.Created) > 0 {
			log.Printf("auto-reorder: created %d purchase order(s)", len(result.Created))
		}
		if len(result.SkippedNoVendor) > 0 {
			log.Printf("auto-reorder: skipped %d item(s) below min with no default vendor: %v", len(result.SkippedNoVendor), result.SkippedNoVendor)
		}
	}

	runOnce()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runOnce()
		}
	}
}
