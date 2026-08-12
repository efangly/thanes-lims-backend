package main

import (
	httpdocument "github.com/efangly/thanes-lims-backend/internal/adapters/http/document"
	httpenvironment "github.com/efangly/thanes-lims-backend/internal/adapters/http/environment"
	httpequipment "github.com/efangly/thanes-lims-backend/internal/adapters/http/equipment"
	httpinventory "github.com/efangly/thanes-lims-backend/internal/adapters/http/inventory"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	httpnotification "github.com/efangly/thanes-lims-backend/internal/adapters/http/notification"
	httppurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/http/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	httpsample "github.com/efangly/thanes-lims-backend/internal/adapters/http/sample"
	httptestresult "github.com/efangly/thanes-lims-backend/internal/adapters/http/testresult"
	httpuser "github.com/efangly/thanes-lims-backend/internal/adapters/http/user"
	"github.com/efangly/thanes-lims-backend/internal/adapters/jwt"
	"github.com/efangly/thanes-lims-backend/internal/adapters/minio"
	postgresdocument "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/document"
	postgresenvironment "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/environment"
	postgresequipment "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/equipment"
	postgresidgen "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/idgen"
	postgresinventory "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/inventory"
	postgresnotification "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/notification"
	postgrespurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/purchaseorder"
	postgressample "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/sample"
	postgrestestresult "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/testresult"
	postgresuser "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	applicationdocument "github.com/efangly/thanes-lims-backend/internal/application/document"
	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	applicationequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	applicationnotification "github.com/efangly/thanes-lims-backend/internal/application/notification"
	applicationpurchaseorder "github.com/efangly/thanes-lims-backend/internal/application/purchaseorder"
	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	"github.com/efangly/thanes-lims-backend/internal/config"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// registerRoutes mounts each module's routes onto the /api/v1 group. Module
// registration calls are added here as each module is built (user, sample,
// testresult, ... per the implementation plan).
func registerRoutes(v1 fiber.Router, cfg *config.Config, gdb *gorm.DB, fileStorage *minio.Adapter) {
	v1.Get("/health", func(c fiber.Ctx) error {
		return response.OK(c, fiber.Map{"status": "ok"})
	})

	tokens := jwt.New(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	userRepo := postgresuser.New(gdb)
	refreshRepo := postgresuser.NewRefreshTokenRepository(gdb)

	userHandler := httpuser.NewHandler(
		applicationuser.NewLoginUseCase(userRepo, refreshRepo, tokens),
		applicationuser.NewRefreshUseCase(userRepo, refreshRepo, tokens),
		applicationuser.NewLogoutUseCase(refreshRepo, tokens),
		applicationuser.NewCreateUserUseCase(userRepo),
		applicationuser.NewListUsersUseCase(userRepo),
		applicationuser.NewGetUserUseCase(userRepo),
		applicationuser.NewUpdateUserUseCase(userRepo),
	)
	httpuser.RegisterRoutes(v1, userHandler, tokens)

	idgen := postgresidgen.New(gdb)
	sampleRepo := postgressample.New(gdb)
	cocRepo := postgressample.NewCoCRepository(gdb)

	sampleHandler := httpsample.NewHandler(
		applicationsample.NewCreateSampleUseCase(sampleRepo, cocRepo, idgen),
		applicationsample.NewListSamplesUseCase(sampleRepo),
		applicationsample.NewGetSampleUseCase(sampleRepo),
		applicationsample.NewUpdateSampleStatusUseCase(sampleRepo, cocRepo),
		applicationsample.NewListCoCStepsUseCase(cocRepo),
		applicationsample.NewAppendCoCStepUseCase(cocRepo),
	)
	httpsample.RegisterRoutes(v1, sampleHandler, tokens)

	testResultRepo := postgrestestresult.New(gdb)
	testResultHandler := httptestresult.NewHandler(
		applicationtestresult.NewCreateTestResultUseCase(testResultRepo, sampleRepo, idgen),
		applicationtestresult.NewSubmitResultUseCase(testResultRepo),
		applicationtestresult.NewApproveResultUseCase(testResultRepo, cocRepo),
		applicationtestresult.NewListTestResultsUseCase(testResultRepo),
		applicationtestresult.NewGetTestResultUseCase(testResultRepo),
	)
	httptestresult.RegisterRoutes(v1, testResultHandler, tokens)

	equipmentRepo := postgresequipment.New(gdb)
	equipmentHandler := httpequipment.NewHandler(
		applicationequipment.NewCreateEquipmentUseCase(equipmentRepo, idgen),
		applicationequipment.NewListEquipmentUseCase(equipmentRepo),
		applicationequipment.NewGetEquipmentUseCase(equipmentRepo),
		applicationequipment.NewRecordCalibrationUseCase(equipmentRepo),
	)
	httpequipment.RegisterRoutes(v1, equipmentHandler, tokens)

	inventoryRepo := postgresinventory.New(gdb)
	purchaseOrderRepo := postgrespurchaseorder.New(gdb)

	reorderUseCase := applicationpurchaseorder.NewCreateFromLowStockUseCase(purchaseOrderRepo, inventoryRepo, idgen)

	inventoryHandler := httpinventory.NewHandler(
		applicationinventory.NewCreateItemUseCase(inventoryRepo, idgen),
		applicationinventory.NewListItemsUseCase(inventoryRepo),
		applicationinventory.NewGetItemUseCase(inventoryRepo),
		applicationinventory.NewUpdateQuantityUseCase(inventoryRepo),
		reorderUseCase,
	)
	httpinventory.RegisterRoutes(v1, inventoryHandler, tokens)

	purchaseOrderHandler := httppurchaseorder.NewHandler(
		applicationpurchaseorder.NewListPOsUseCase(purchaseOrderRepo),
		applicationpurchaseorder.NewGetPOUseCase(purchaseOrderRepo),
		applicationpurchaseorder.NewApprovePOUseCase(purchaseOrderRepo),
		applicationpurchaseorder.NewMarkReceivedUseCase(purchaseOrderRepo, inventoryRepo),
	)
	httppurchaseorder.RegisterRoutes(v1, purchaseOrderHandler, tokens)

	documentRepo := postgresdocument.New(gdb)
	docHistoryRepo := postgresdocument.NewHistoryRepository(gdb)

	documentHandler := httpdocument.NewHandler(
		applicationdocument.NewUploadDocumentUseCase(documentRepo, docHistoryRepo, fileStorage, idgen),
		applicationdocument.NewCreateNewVersionUseCase(documentRepo, docHistoryRepo, fileStorage),
		applicationdocument.NewSetLockUseCase(documentRepo),
		applicationdocument.NewListDocumentsUseCase(documentRepo),
		applicationdocument.NewGetDocumentUseCase(documentRepo),
		applicationdocument.NewGetDownloadURLUseCase(documentRepo, fileStorage),
		applicationdocument.NewListHistoryUseCase(docHistoryRepo),
	)
	httpdocument.RegisterRoutes(v1, documentHandler, tokens)

	gaugeRepo := postgresenvironment.NewGaugeRepository(gdb)
	readingRepo := postgresenvironment.NewReadingRepository(gdb)
	alertRepo := postgresenvironment.NewAlertRepository(gdb)

	evaluateThresholds := applicationenvironment.NewEvaluateThresholdsUseCase(gaugeRepo, alertRepo)
	environmentHandler := httpenvironment.NewHandler(
		applicationenvironment.NewRecordReadingUseCase(readingRepo, evaluateThresholds),
		applicationenvironment.NewListGaugesUseCase(gaugeRepo, readingRepo),
		applicationenvironment.NewGetTrendUseCase(readingRepo),
		applicationenvironment.NewListAlertsUseCase(alertRepo),
	)
	httpenvironment.RegisterRoutes(v1, environmentHandler, tokens)

	notificationRepo := postgresnotification.New(gdb)
	notificationHandler := httpnotification.NewHandler(
		applicationnotification.NewListNotificationsUseCase(notificationRepo),
		applicationnotification.NewMarkReadUseCase(notificationRepo),
		applicationnotification.NewMarkAllReadUseCase(notificationRepo),
	)
	httpnotification.RegisterRoutes(v1, notificationHandler, tokens)

	registerReportStubs(v1, middleware.Auth(tokens))
}

// registerReportStubs mounts placeholder report/export endpoints - real PDF
// generation (e.g. gofpdf) is deferred to a later phase; for now they return
// 501 in the standard envelope so the frontend's "generate report" /
// "export audit report" actions have something to call against.
func registerReportStubs(v1 fiber.Router, authGuard fiber.Handler) {
	notImplemented := func(c fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotImplemented, "not_implemented", "report generation is not implemented yet")
	}
	v1.Get("/tests/:id/report", authGuard, notImplemented)
	v1.Get("/audit/export", authGuard, notImplemented)
}
