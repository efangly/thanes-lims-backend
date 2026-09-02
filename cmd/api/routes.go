package main

import (
	"context"
	"database/sql"

	_ "github.com/efangly/thanes-lims-backend/docs"
	anthropicchatbot "github.com/efangly/thanes-lims-backend/internal/adapters/anthropic/chatbot"
	"github.com/efangly/thanes-lims-backend/internal/adapters/cachedenvironment"
	"github.com/efangly/thanes-lims-backend/internal/adapters/cachedlocation"
	"github.com/efangly/thanes-lims-backend/internal/adapters/cachedrbac"
	"github.com/efangly/thanes-lims-backend/internal/adapters/cacheduser"
	httpaudit "github.com/efangly/thanes-lims-backend/internal/adapters/http/audit"
	httpchatbot "github.com/efangly/thanes-lims-backend/internal/adapters/http/chatbot"
	httpdocument "github.com/efangly/thanes-lims-backend/internal/adapters/http/document"
	httpenvironment "github.com/efangly/thanes-lims-backend/internal/adapters/http/environment"
	httpequipment "github.com/efangly/thanes-lims-backend/internal/adapters/http/equipment"
	httpinventory "github.com/efangly/thanes-lims-backend/internal/adapters/http/inventory"
	httplocation "github.com/efangly/thanes-lims-backend/internal/adapters/http/location"
	httpnotification "github.com/efangly/thanes-lims-backend/internal/adapters/http/notification"
	httppurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/http/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	httpsample "github.com/efangly/thanes-lims-backend/internal/adapters/http/sample"
	httptestresult "github.com/efangly/thanes-lims-backend/internal/adapters/http/testresult"
	httpuser "github.com/efangly/thanes-lims-backend/internal/adapters/http/user"
	httpvendor "github.com/efangly/thanes-lims-backend/internal/adapters/http/vendor"
	"github.com/efangly/thanes-lims-backend/internal/adapters/jwt"
	"github.com/efangly/thanes-lims-backend/internal/adapters/minio"
	oraclechatbot "github.com/efangly/thanes-lims-backend/internal/adapters/oracle/chatbot"
	postgresaudit "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/audit"
	postgresdocument "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/document"
	postgresenvironment "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/environment"
	postgresequipment "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/equipment"
	postgresidgen "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/idgen"
	postgresinventory "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/inventory"
	postgreslocation "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/location"
	postgresnotification "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/notification"
	postgrespurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/purchaseorder"
	postgresrbac "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/rbac"
	postgressample "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/sample"
	postgrestestresult "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/testresult"
	postgresuser "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	postgresvendor "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/vendor"
	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	applicationchatbot "github.com/efangly/thanes-lims-backend/internal/application/chatbot"
	applicationdocument "github.com/efangly/thanes-lims-backend/internal/application/document"
	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	applicationequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	applicationlocation "github.com/efangly/thanes-lims-backend/internal/application/location"
	applicationnotification "github.com/efangly/thanes-lims-backend/internal/application/notification"
	applicationpurchaseorder "github.com/efangly/thanes-lims-backend/internal/application/purchaseorder"
	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	applicationvendor "github.com/efangly/thanes-lims-backend/internal/application/vendor"
	"github.com/efangly/thanes-lims-backend/internal/config"
	"github.com/efangly/thanes-lims-backend/internal/ports/cache"
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// registerRoutes mounts each module's routes onto the /api/v1 group. Module
// registration calls are added here as each module is built (user, sample,
// testresult, ... per the implementation plan). It also returns the
// auto-reorder job so main can run it on a schedule alongside the HTTP
// server, since it's composed from the same repositories wired up here.
func registerRoutes(v1 fiber.Router, cfg *config.Config, gdb *gorm.DB, chatbotDB *sql.DB, fileStorage *minio.Adapter, redisCache cache.Cache) *applicationpurchaseorder.AutoReorderJob {
	// /health also probes Redis: per ADR 0005 the refresh path is fail-closed,
	// so a Redis outage logs every user out within one access-token TTL (~15m).
	// External monitoring must be able to alert on that before users notice.
	v1.Get("/health", func(c fiber.Ctx) error {
		if pinger, ok := redisCache.(interface {
			Ping(context.Context) error
		}); ok {
			if err := pinger.Ping(c.Context()); err != nil {
				return c.Status(fiber.StatusServiceUnavailable).
					JSON(fiber.Map{"status": "degraded", "redis": err.Error()})
			}
		}
		return response.OK(c, fiber.Map{"status": "ok", "redis": "ok"})
	})

	v1.Get("/swagger/*", swaggo.New(swaggo.Config{}))

	tokens := jwt.New(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	userRepo := postgresuser.New(gdb)
	refreshRepo := cacheduser.NewCachedRefreshTokenRepository(postgresuser.NewRefreshTokenRepository(gdb), redisCache)
	rbacRepo := cachedrbac.NewCachedRepository(postgresrbac.New(gdb), redisCache)

	userHandler := httpuser.NewHandler(
		applicationuser.NewLoginUseCase(userRepo, refreshRepo, tokens, rbacRepo),
		applicationuser.NewRefreshUseCase(userRepo, refreshRepo, tokens, rbacRepo),
		applicationuser.NewLogoutUseCase(refreshRepo, tokens),
		applicationuser.NewLogoutAllUseCase(refreshRepo),
		applicationuser.NewCreateUserUseCase(userRepo),
		applicationuser.NewListUsersUseCase(userRepo),
		applicationuser.NewGetUserUseCase(userRepo),
		applicationuser.NewUpdateUserUseCase(userRepo, refreshRepo),
		cfg.CookieSecure,
	)
	httpuser.RegisterRoutes(v1, userHandler, tokens)

	idgen := postgresidgen.New(gdb)
	sampleRepo := postgressample.New(gdb)
	cocRepo := postgressample.NewCoCRepository(gdb)
	locationRepo := cachedlocation.NewCachedRepository(postgreslocation.New(gdb), redisCache)

	sampleHandler := httpsample.NewHandler(
		applicationsample.NewCreateSampleUseCase(sampleRepo, cocRepo, userRepo, idgen),
		applicationsample.NewListSamplesUseCase(sampleRepo),
		applicationsample.NewGetSampleUseCase(sampleRepo),
		applicationsample.NewUpdateSampleStatusUseCase(sampleRepo, cocRepo),
		applicationsample.NewListCoCStepsUseCase(cocRepo),
		applicationsample.NewAppendCoCStepUseCase(cocRepo),
		applicationsample.NewAssignLocationUseCase(sampleRepo, locationRepo),
		applicationsample.NewGenerateBarcodeUseCase(sampleRepo, idgen),
		applicationsample.NewStickerDataUseCase(sampleRepo, userRepo, locationRepo),
	)
	httpsample.RegisterRoutes(v1, sampleHandler, tokens)

	locationHandler := httplocation.NewHandler(
		applicationlocation.NewCreateCabinetUseCase(locationRepo, idgen),
		applicationlocation.NewGenerateChildrenUseCase(locationRepo, idgen),
		applicationlocation.NewListChildrenUseCase(locationRepo),
		applicationlocation.NewGetFullPathUseCase(locationRepo),
		applicationlocation.NewDeleteLocationUseCase(locationRepo, sampleRepo),
		applicationlocation.NewLookupByBarcodeUseCase(locationRepo),
		applicationlocation.NewGetLocationUseCase(locationRepo),
		applicationlocation.NewCreateBoxUseCase(locationRepo, idgen),
		applicationlocation.NewEnlargeBoxUseCase(locationRepo),
		applicationsample.NewMoveWithinBoxUseCase(sampleRepo, locationRepo),
	)
	httplocation.RegisterRoutes(v1, locationHandler, tokens)

	vendorRepo := postgresvendor.New(gdb)
	vendorHandler := httpvendor.NewHandler(
		applicationvendor.NewCreateVendorUseCase(vendorRepo, idgen),
		applicationvendor.NewListVendorsUseCase(vendorRepo),
		applicationvendor.NewGetVendorUseCase(vendorRepo),
		applicationvendor.NewUpdateVendorUseCase(vendorRepo),
	)
	httpvendor.RegisterRoutes(v1, vendorHandler, tokens)

	notificationRepo := postgresnotification.New(gdb)
	notifier := applicationnotification.NewAsNotifier(applicationnotification.NewCreateNotificationUseCase(notificationRepo, idgen))

	testResultRepo := postgrestestresult.New(gdb)
	testResultHandler := httptestresult.NewHandler(
		applicationtestresult.NewCreateTestResultUseCase(testResultRepo, sampleRepo, idgen),
		applicationtestresult.NewSubmitResultUseCase(testResultRepo),
		applicationtestresult.NewApproveResultUseCase(testResultRepo, cocRepo, notifier),
		applicationtestresult.NewListTestResultsUseCase(testResultRepo),
		applicationtestresult.NewGetTestResultUseCase(testResultRepo),
		applicationtestresult.NewGenerateReportUseCase(testResultRepo, sampleRepo, cocRepo, locationRepo, userRepo),
	)
	httptestresult.RegisterRoutes(v1, testResultHandler, tokens)

	equipmentRepo := postgresequipment.New(gdb)
	calibrationRepo := postgresequipment.NewCalibrationRepository(gdb)
	calibrationScheduleRepo := postgresequipment.NewScheduleRepository(gdb)
	equipmentHandler := httpequipment.NewHandler(
		applicationequipment.NewCreateEquipmentUseCase(equipmentRepo, idgen, vendorRepo, locationRepo),
		applicationequipment.NewUpdateEquipmentUseCase(equipmentRepo, vendorRepo, locationRepo),
		applicationequipment.NewListEquipmentUseCase(equipmentRepo),
		applicationequipment.NewGetEquipmentUseCase(equipmentRepo),
		applicationequipment.NewRecordCalibrationUseCase(equipmentRepo, calibrationRepo, calibrationScheduleRepo),
		applicationequipment.NewListCalibrationEventsUseCase(calibrationRepo),
		applicationequipment.NewCalibrationScheduleUseCase(equipmentRepo, calibrationScheduleRepo),
		applicationequipment.NewSearchCalibrationResultsUseCase(calibrationRepo),
	)
	httpequipment.RegisterRoutes(v1, equipmentHandler, tokens)

	inventoryRepo := postgresinventory.New(gdb)
	inventoryLotRepo := postgresinventory.NewLotRepository(gdb)
	purchaseOrderRepo := postgrespurchaseorder.New(gdb)

	reorderUseCase := applicationpurchaseorder.NewCreateFromLowStockUseCase(purchaseOrderRepo, inventoryRepo, idgen)

	inventoryHandler := httpinventory.NewHandler(
		applicationinventory.NewCreateItemUseCase(inventoryRepo, idgen, userRepo, vendorRepo, locationRepo),
		applicationinventory.NewListItemsUseCase(inventoryRepo),
		applicationinventory.NewGetItemUseCase(inventoryRepo),
		applicationinventory.NewUpdateItemUseCase(inventoryRepo, userRepo, vendorRepo, locationRepo),
		applicationinventory.NewReceiveStockUseCase(inventoryRepo, inventoryLotRepo, idgen),
		applicationinventory.NewIssueStockUseCase(inventoryRepo, inventoryLotRepo),
		applicationinventory.NewListLotsUseCase(inventoryRepo, inventoryLotRepo),
		applicationinventory.NewUpdateDefaultVendorUseCase(inventoryRepo),
		reorderUseCase,
	)
	httpinventory.RegisterRoutes(v1, inventoryHandler, tokens)

	autoReorderJob := applicationpurchaseorder.NewAutoReorderJob(inventoryRepo, purchaseOrderRepo, reorderUseCase)

	purchaseOrderHandler := httppurchaseorder.NewHandler(
		applicationpurchaseorder.NewListPOsUseCase(purchaseOrderRepo),
		applicationpurchaseorder.NewGetPOUseCase(purchaseOrderRepo),
		applicationpurchaseorder.NewApprovePOUseCase(purchaseOrderRepo),
		applicationpurchaseorder.NewMarkReceivedUseCase(purchaseOrderRepo, inventoryRepo, inventoryLotRepo, idgen),
	)
	httppurchaseorder.RegisterRoutes(v1, purchaseOrderHandler, tokens)

	documentRepo := postgresdocument.New(gdb)
	docHistoryRepo := postgresdocument.NewHistoryRepository(gdb)

	documentHandler := httpdocument.NewHandler(
		applicationdocument.NewUploadDocumentUseCase(documentRepo, docHistoryRepo, fileStorage, idgen, equipmentRepo, calibrationRepo),
		applicationdocument.NewCreateNewVersionUseCase(documentRepo, docHistoryRepo, fileStorage),
		applicationdocument.NewSetLockUseCase(documentRepo),
		applicationdocument.NewListDocumentsUseCase(documentRepo),
		applicationdocument.NewGetDocumentUseCase(documentRepo),
		applicationdocument.NewGetDownloadURLUseCase(documentRepo, fileStorage),
		applicationdocument.NewListHistoryUseCase(docHistoryRepo),
	)
	httpdocument.RegisterRoutes(v1, documentHandler, tokens)

	gaugeRepo := cachedenvironment.NewCachedGaugeRepository(postgresenvironment.NewGaugeRepository(gdb), redisCache)
	readingRepo := postgresenvironment.NewReadingRepository(gdb)
	alertRepo := postgresenvironment.NewAlertRepository(gdb)
	alertHub := httpenvironment.NewHub()

	evaluateThresholds := applicationenvironment.NewEvaluateThresholdsUseCase(gaugeRepo, alertRepo, notifier, alertHub)
	environmentHandler := httpenvironment.NewHandler(
		applicationenvironment.NewRecordReadingUseCase(readingRepo, evaluateThresholds),
		applicationenvironment.NewListGaugesUseCase(gaugeRepo, readingRepo),
		applicationenvironment.NewGetTrendUseCase(readingRepo),
		applicationenvironment.NewListAlertsUseCase(alertRepo),
		alertHub,
	)
	httpenvironment.RegisterRoutes(v1, environmentHandler, tokens)

	notificationHandler := httpnotification.NewHandler(
		applicationnotification.NewListNotificationsUseCase(notificationRepo),
		applicationnotification.NewMarkReadUseCase(notificationRepo),
		applicationnotification.NewMarkAllReadUseCase(notificationRepo),
	)
	httpnotification.RegisterRoutes(v1, notificationHandler, tokens)

	// Chatbot POC (see docs/chatbot-poc-plan.md): only mounted when the
	// optional Oracle ADB connection succeeded at startup.
	if chatbotDB != nil {
		runner := oraclechatbot.New(chatbotDB)
		assistant := anthropicchatbot.New(cfg.AnthropicAPIKey, cfg.ChatbotModel, runner)
		chatbotHandler := httpchatbot.NewHandler(applicationchatbot.NewAskUseCase(assistant))
		httpchatbot.RegisterRoutes(v1, chatbotHandler, tokens)
	}

	auditRepo := postgresaudit.New(gdb)
	auditHandler := httpaudit.NewHandler(
		applicationaudit.NewListAuditLogsUseCase(auditRepo),
		applicationaudit.NewListAuditLogsPageUseCase(auditRepo),
	)
	httpaudit.RegisterRoutes(v1, auditHandler, tokens)

	return autoReorderJob
}
