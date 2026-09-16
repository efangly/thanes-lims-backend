package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/objectstorage"
	"github.com/efangly/thanes-lims-backend/internal/adapters/partnergrpc"
	postgresaudit "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/audit"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/db"
	redisadapter "github.com/efangly/thanes-lims-backend/internal/adapters/redis"
	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
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

	// Composition root: wire adapters -> ports -> use cases -> handlers.
	auditRepo := postgresaudit.New(gdb)
	logAction := applicationaudit.NewLogActionUseCase(auditRepo)

	fileStorage, err := objectstorage.New(cfg.StorageEndpoint, cfg.StorageRegion, cfg.StorageAccessKey, cfg.StorageSecretKey, cfg.StorageBucket, cfg.StorageUseSSL)
	if err != nil {
		log.Fatalf("objectstorage: %v", err)
	}
	if err := fileStorage.EnsureBucket(context.Background()); err != nil {
		log.Fatalf("objectstorage: ensure bucket: %v", err)
	}

	redisCache, err := redisadapter.New(context.Background(), cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisCache.Close()

	// Partner Device (SMtrack third-party device data via gRPC, see
	// docs/partner-api-guide.md and ADR 0012) - optional, off by default.
	// PARTNER_API_ENABLED=true is an explicit operator opt-in already gated
	// by config.validate(), so a dial failure here is fatal, matching
	// objectstorage/redis above.
	var partnerClient *partnergrpc.Client
	if cfg.PartnerAPIEnabled {
		pc, err := partnergrpc.New(cfg.PartnerGRPCAddr, cfg.PartnerAPIKey, 10*time.Second)
		if err != nil {
			log.Fatalf("partnergrpc: %v", err)
		}
		defer pc.Close()
		partnerClient = pc
	}

	fiberCfg := fiber.Config{
		ErrorHandler: middleware.ErrorMapper,
		// fasthttp's default ReadBufferSize (4096) is too small once a real
		// browser's standard headers (sec-ch-ua, sec-fetch-*, accept-*,
		// cookies, etc.) are combined with our JWT access token - it embeds
		// the full RBAC permission list and easily runs 1.5-2KB for an admin
		// (see internal/domain/rbac). A sufficiently long request path (e.g.
		// a Partner Device serial in the URL) pushed real requests over 4096
		// bytes, and fasthttp hard-fails with 431 Request Header Fields Too
		// Large before the request ever reaches a handler - this isn't
		// attacker-sized input, it's normal traffic. 16KB matches common
		// reverse-proxy defaults (e.g. nginx's doubled 8k) with headroom.
		ReadBufferSize: 16384,
	}
	// Only trust proxy headers (X-Forwarded-For etc.) from the configured
	// load balancers, so the client IP stored in a Token Family can't be
	// spoofed by an arbitrary caller sending its own X-Forwarded-For.
	if len(cfg.TrustedProxies) > 0 {
		fiberCfg.TrustProxy = true
		fiberCfg.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: cfg.TrustedProxies}
		if cfg.ProxyHeader != "" {
			fiberCfg.ProxyHeader = cfg.ProxyHeader
		}
	}
	app := fiber.New(fiberCfg)

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(fiberlogger.New())
	app.Use(corsMiddleware(cfg))
	app.Use(middleware.Audit(logAction))

	v1 := app.Group("/api/v1")
	jobs := registerRoutes(v1, cfg, gdb, fileStorage, redisCache, partnerClient)

	go func() {
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	jobCtx, cancelJob := context.WithCancel(context.Background())
	jobDone := make(chan struct{})
	if cfg.AutoReorderEnabled {
		go runAutoReorderJob(jobCtx, jobDone, jobs.AutoReorder, cfg.AutoReorderInterval)
	} else {
		close(jobDone)
	}

	partnerDeviceJobDone := make(chan struct{})
	if cfg.PartnerAPIEnabled {
		go runPollPartnerDevicesJob(jobCtx, partnerDeviceJobDone, jobs.PollPartnerDevices, cfg.PartnerAPIPollInterval)
	} else {
		close(partnerDeviceJobDone)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancelJob()
	<-jobDone
	<-partnerDeviceJobDone

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
		// A wildcard CORS policy is only ever acceptable for local dev. In any
		// real deployment an unset CORS_ALLOW_ORIGINS is a misconfiguration -
		// fail loudly at boot rather than silently serving Access-Control-
		// Allow-Origin: * to the whole internet.
		if cfg.AppEnv != "local" {
			log.Fatalf("CORS_ALLOW_ORIGINS must be set when APP_ENV=%q (refusing to start with a wildcard CORS policy)", cfg.AppEnv)
		}
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

// runPollPartnerDevicesJob periodically polls every Active Partner Device's
// metadata/reading from the Partner API, refreshing the cached snapshot and
// pushing it to SSE subscribers (see ADR 0011). Same run-once-then-tick,
// cancel-to-exit shape as runAutoReorderJob.
func runPollPartnerDevicesJob(ctx context.Context, done chan<- struct{}, job *applicationenvironment.PollPartnerDevicesJob, interval time.Duration) {
	defer close(done)

	job.Run(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			job.Run(ctx)
		}
	}
}
