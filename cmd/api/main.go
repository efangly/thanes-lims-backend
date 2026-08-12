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
	postgresaudit "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/audit"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/db"
	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	"github.com/efangly/thanes-lims-backend/internal/config"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

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
	app.Use(cors.New())
	app.Use(middleware.Audit(logAction))

	v1 := app.Group("/api/v1")
	registerRoutes(v1, cfg, gdb, fileStorage)

	go func() {
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
