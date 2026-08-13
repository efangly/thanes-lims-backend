package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	httpenvironment "github.com/efangly/thanes-lims-backend/internal/adapters/http/environment"
	"github.com/efangly/thanes-lims-backend/internal/adapters/minio"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/db"
	postgresdocument "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/document"
	postgresenvironment "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/environment"
	postgresequipment "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/equipment"
	postgresidgen "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/idgen"
	postgresinventory "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/inventory"
	postgresnotification "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/notification"
	postgressample "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/sample"
	postgrestestresult "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/testresult"
	postgresuser "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	applicationdocument "github.com/efangly/thanes-lims-backend/internal/application/document"
	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	applicationequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	applicationnotification "github.com/efangly/thanes-lims-backend/internal/application/notification"
	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	"github.com/efangly/thanes-lims-backend/internal/config"
	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"gorm.io/gorm"
)

// Demo credentials - password is the same across all 4 seed users, purely
// for local/demo convenience. See SEED_CREDENTIALS.md.
const seedPassword = "Passw0rd!Demo"

func main() {
	force := flag.Bool("force", false, "wipe all seeded tables before reseeding")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gdb, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	fileStorage, err := minio.New(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}
	ctx := context.Background()
	if err := fileStorage.EnsureBucket(ctx); err != nil {
		log.Fatalf("minio: ensure bucket: %v", err)
	}

	userRepo := postgresuser.New(gdb)

	existing, err := userRepo.List(ctx)
	if err != nil {
		log.Fatalf("check existing users: %v", err)
	}
	if len(existing) > 0 {
		if !*force {
			log.Println("seed: users already exist, skipping (pass -force to wipe and reseed)")
			return
		}
		if err := wipe(gdb); err != nil {
			log.Fatalf("wipe: %v", err)
		}
	}

	idgen := postgresidgen.New(gdb)

	users := seedUsers(ctx, userRepo)

	sampleRepo := postgressample.New(gdb)
	cocRepo := postgressample.NewCoCRepository(gdb)
	samples := seedSamples(ctx, sampleRepo, cocRepo, idgen, users)

	testResultRepo := postgrestestresult.New(gdb)
	seedTestResults(ctx, testResultRepo, sampleRepo, idgen, samples, users)

	equipmentRepo := postgresequipment.New(gdb)
	seedEquipment(ctx, equipmentRepo, idgen)

	inventoryRepo := postgresinventory.New(gdb)
	seedInventory(ctx, inventoryRepo, idgen)

	documentRepo := postgresdocument.New(gdb)
	docHistoryRepo := postgresdocument.NewHistoryRepository(gdb)
	seedDocuments(ctx, documentRepo, docHistoryRepo, fileStorage, idgen, users)

	notificationRepo := postgresnotification.New(gdb)
	notifier := applicationnotification.NewAsNotifier(applicationnotification.NewCreateNotificationUseCase(notificationRepo, idgen))

	gaugeRepo := postgresenvironment.NewGaugeRepository(gdb)
	readingRepo := postgresenvironment.NewReadingRepository(gdb)
	alertRepo := postgresenvironment.NewAlertRepository(gdb)
	seedEnvironment(ctx, gdb, gaugeRepo, readingRepo, alertRepo, notifier)

	seedNotifications(ctx, notificationRepo, idgen, users)

	log.Println("seed: done")
}

func seedUsers(ctx context.Context, users *postgresuser.Repository) map[string]domainuser.User {
	create := applicationuser.NewCreateUserUseCase(users)

	specs := []struct {
		key   string
		name  string
		email string
		role  domainuser.Role
	}{
		{"admin", "ธเนศ สุขใจ", "admin@thanes-lims.demo", domainuser.RoleAdmin},
		{"qa", "พิมพ์ชนก วารี", "qa@thanes-lims.demo", domainuser.RoleQA},
		{"scientist", "สมชาย เข็มทอง", "scientist@thanes-lims.demo", domainuser.RoleScientist},
		{"general", "วิภา แสงจันทร์", "general@thanes-lims.demo", domainuser.RoleGeneral},
	}

	out := make(map[string]domainuser.User, len(specs))
	for _, s := range specs {
		u, err := create.Execute(ctx, applicationuser.CreateUserInput{
			Name: s.name, Email: s.email, Password: seedPassword, Role: s.role,
		})
		if err != nil {
			log.Fatalf("seed user %s: %v", s.email, err)
		}
		out[s.key] = u
	}
	log.Printf("seed: created %d users (password: %s)", len(out), seedPassword)
	return out
}

func seedSamples(ctx context.Context, samples *postgressample.Repository, coc *postgressample.CoCRepository, idgen *postgresidgen.Adapter, users map[string]domainuser.User) []sample.Sample {
	create := applicationsample.NewCreateSampleUseCase(samples, coc, idgen)

	specs := []struct {
		name     string
		typ      sample.Type
		location string
	}{
		{"เลือดผู้ป่วย OPD-1042", sample.TypeBlood, "Fridge-A / R2-04"},
		{"ตัวอย่างน้ำดื่มบรรจุขวด", sample.TypeWater, "Shelf-B / R1-01"},
	}

	out := make([]sample.Sample, 0, len(specs))
	for _, s := range specs {
		created, err := create.Execute(ctx, applicationsample.CreateSampleInput{
			Name: s.name, Type: s.typ, Custodian: users["scientist"].Name, Location: s.location,
		})
		if err != nil {
			log.Fatalf("seed sample %s: %v", s.name, err)
		}
		out = append(out, created)
	}
	log.Printf("seed: created %d samples", len(out))
	return out
}

func seedTestResults(ctx context.Context, results *postgrestestresult.Repository, samples *postgressample.Repository, idgen *postgresidgen.Adapter, seededSamples []sample.Sample, users map[string]domainuser.User) {
	create := applicationtestresult.NewCreateTestResultUseCase(results, samples, idgen)

	for _, s := range seededSamples {
		_, err := create.Execute(ctx, applicationtestresult.CreateTestResultInput{
			SampleID: s.ID, TestName: "CBC - Complete Blood Count", Analyst: users["scientist"].Name, RefRange: "4.5-11.0 x10^9/L",
		})
		if err != nil {
			log.Fatalf("seed test result for %s: %v", s.ID, err)
		}
	}
	log.Printf("seed: created %d test results", len(seededSamples))
}

func seedEquipment(ctx context.Context, equipment *postgresequipment.Repository, idgen *postgresidgen.Adapter) {
	create := applicationequipment.NewCreateEquipmentUseCase(equipment, idgen)

	specs := []struct {
		name     string
		typeCode string
		days     int
	}{
		{"เครื่องปั่นเหวี่ยงตกตะกอน", "CENT", 60},
		{"เครื่อง PCR", "PCR", 5},
	}

	for _, s := range specs {
		_, err := create.Execute(ctx, applicationequipment.CreateEquipmentInput{
			Name: s.name, TypeCode: s.typeCode, NextCalibrationDue: time.Now().AddDate(0, 0, s.days),
		})
		if err != nil {
			log.Fatalf("seed equipment %s: %v", s.name, err)
		}
	}
	log.Printf("seed: created %d equipment", len(specs))
}

func seedInventory(ctx context.Context, inventory *postgresinventory.Repository, idgen *postgresidgen.Adapter) {
	create := applicationinventory.NewCreateItemUseCase(inventory, idgen)

	specs := []struct {
		name          string
		category      string
		qty, min, max int
		unit          string
	}{
		{"เอทานอล 95%", "รีเอเจนต์", 5, 20, 100, "ลิตร"}, // deliberately below min to demo reorder flow
		{"ถุงมือไนไตรล์", "วัสดุสิ้นเปลือง", 80, 30, 100, "กล่อง"},
	}

	for _, s := range specs {
		_, err := create.Execute(ctx, applicationinventory.CreateItemInput{
			Name: s.name, Category: s.category, Quantity: s.qty, Unit: s.unit, Min: s.min, Max: s.max,
		})
		if err != nil {
			log.Fatalf("seed inventory %s: %v", s.name, err)
		}
	}
	log.Printf("seed: created %d inventory items", len(specs))
}

func seedDocuments(ctx context.Context, documents *postgresdocument.Repository, history *postgresdocument.HistoryRepository, storage *minio.Adapter, idgen *postgresidgen.Adapter, users map[string]domainuser.User) {
	upload := applicationdocument.NewUploadDocumentUseCase(documents, history, storage, idgen)

	specs := []struct {
		name     string
		docType  document.Type
		filename string
		content  string
	}{
		{"SOP การเก็บตัวอย่างเลือด", document.TypeSOP, "sop-blood-collection.txt", "ขั้นตอนมาตรฐานการเก็บตัวอย่างเลือด (demo placeholder)"},
		{"คู่มือการใช้เครื่อง PCR", document.TypeManual, "pcr-manual.txt", "คู่มือการใช้งานเครื่อง PCR (demo placeholder)"},
	}

	for _, s := range specs {
		_, err := upload.Execute(ctx, applicationdocument.UploadDocumentInput{
			Name: s.name, Type: s.docType, Filename: s.filename, ContentType: "text/plain",
			Size: int64(len(s.content)), Content: strings.NewReader(s.content),
			AccessLevel: "internal", UploadedBy: users["qa"].Name,
		})
		if err != nil {
			log.Fatalf("seed document %s: %v", s.name, err)
		}
	}
	log.Printf("seed: created %d documents", len(specs))
}

func seedEnvironment(ctx context.Context, gdb *gorm.DB, gauges *postgresenvironment.GaugeRepository, readings *postgresenvironment.ReadingRepository, alerts *postgresenvironment.AlertRepository, notifier *applicationnotification.AsNotifier) {
	// Gauges are fixed monitoring points (physical sensors), not something
	// created through the API - seed their config directly.
	gaugeSpecs := []postgresenvironment.GaugeModel{
		{Location: "Fridge-A", Unit: "°C", RangeMin: 2, RangeMax: 8},
		{Location: "Freezer-1", Unit: "°C", RangeMin: -25, RangeMax: -15},
	}
	if err := gdb.WithContext(ctx).Create(&gaugeSpecs).Error; err != nil {
		log.Fatalf("seed gauges: %v", err)
	}

	evaluate := applicationenvironment.NewEvaluateThresholdsUseCase(gauges, alerts, notifier, httpenvironment.NewHub())
	record := applicationenvironment.NewRecordReadingUseCase(readings, evaluate)

	readingSpecs := []struct {
		location string
		value    float64
	}{
		{"Fridge-A", 4.5},
		{"Freezer-1", -18.2},
	}
	for _, r := range readingSpecs {
		if _, err := record.Execute(ctx, r.location, r.value); err != nil {
			log.Fatalf("seed reading %s: %v", r.location, err)
		}
	}
	log.Printf("seed: created %d gauges + readings", len(gaugeSpecs))
}

func seedNotifications(ctx context.Context, notifications *postgresnotification.Repository, idgen *postgresidgen.Adapter, users map[string]domainuser.User) {
	create := applicationnotification.NewCreateNotificationUseCase(notifications, idgen)

	adminID := users["admin"].ID
	specs := []applicationnotification.CreateNotificationInput{
		{RecipientUserID: &adminID, Tone: notification.ToneAmber, Icon: "AlertTriangle", Title: "สต๊อกใกล้หมด", Message: "เอทานอล 95% เหลือต่ำกว่าเกณฑ์ขั้นต่ำ"},
		{RecipientUserID: nil, Tone: notification.ToneTeal, Icon: "Info", Title: "ยินดีต้อนรับสู่ Thanes LIMS", Message: "ระบบพร้อมใช้งานสำหรับการทดสอบ"},
	}

	for _, s := range specs {
		if _, err := create.Execute(ctx, s); err != nil {
			log.Fatalf("seed notification %s: %v", s.Title, err)
		}
	}
	log.Printf("seed: created %d notifications", len(specs))
}

// wipe truncates every seeded table in FK-safe order for -force reseeding.
func wipe(gdb *gorm.DB) error {
	tables := []string{
		"notifications", "env_alerts", "sensor_readings", "gauges",
		"doc_history", "documents",
		"purchase_orders", "inventory_items",
		"equipment",
		"test_results", "coc_steps", "samples",
		"refresh_tokens", "users",
		"audit_logs", "id_sequences",
	}
	for _, t := range tables {
		if err := gdb.Exec("TRUNCATE TABLE " + t + " CASCADE").Error; err != nil {
			return err
		}
	}
	return nil
}
