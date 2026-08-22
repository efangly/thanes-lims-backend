package main

import (
	"context"
	"flag"
	"fmt"
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
	postgrespurchaseorder "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/purchaseorder"
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
	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
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

	notificationRepo := postgresnotification.New(gdb)
	notifier := applicationnotification.NewAsNotifier(applicationnotification.NewCreateNotificationUseCase(notificationRepo, idgen))

	users := seedUsers(ctx, userRepo)

	sampleRepo := postgressample.New(gdb)
	cocRepo := postgressample.NewCoCRepository(gdb)
	samples := seedSamples(ctx, gdb, sampleRepo, cocRepo, idgen, users)

	testResultRepo := postgrestestresult.New(gdb)
	seedTestResults(ctx, testResultRepo, sampleRepo, cocRepo, idgen, samples, users, notifier)

	equipmentRepo := postgresequipment.New(gdb)
	seedEquipment(ctx, equipmentRepo, idgen)

	inventoryRepo := postgresinventory.New(gdb)
	_ = seedInventory(ctx, inventoryRepo, idgen)

	purchaseOrderRepo := postgrespurchaseorder.New(gdb)
	seedPurchaseOrders(ctx, purchaseOrderRepo, inventoryRepo, idgen)

	documentRepo := postgresdocument.New(gdb)
	docHistoryRepo := postgresdocument.NewHistoryRepository(gdb)
	seedDocuments(ctx, documentRepo, docHistoryRepo, fileStorage, idgen, users)

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

// sampleCustodians mirrors the custodian names used in the Oracle Select AI
// POC seed (scripts/oracle/002_seed.sql) so demo questions read the same
// across both databases.
var sampleCustodians = []string{"สมชาย ใจดี", "วิภา สายใจ", "ประยุทธ์ แสงทอง", "อรทัย พงษ์ศรี"}

var sampleTypeSpecs = []struct {
	typ      sample.Type
	name     string
	location string
}{
	{sample.TypeBlood, "เลือดผู้ป่วย OPD", "Fridge-A / R2-04"},
	{sample.TypeUrine, "ปัสสาวะผู้ป่วย", "Shelf-C / R1-02"},
	{sample.TypeWater, "ตัวอย่างน้ำดื่มบรรจุขวด", "Shelf-B / R1-01"},
	{sample.TypeTissue, "ชิ้นเนื้อตรวจพยาธิวิทยา", "Freezer-1 / R3-01"},
	{sample.TypeFood, "ตัวอย่างอาหารแปรรูป", "Shelf-D / R2-01"},
	{sample.TypeSerum, "ซีรัมผู้ป่วย", "Fridge-A / R2-05"},
}

// seedSamples creates 40 samples cycling through type/custodian/location,
// then drives most of them through the state machine (UpdateSampleStatus)
// so the resulting mix of pending/testing/completed/transferred looks like
// a lab that's been running for a while - not just freshly-inserted rows.
// A handful are left "pending" with a backdated received_at (via direct SQL,
// since CreateSample always stamps time.Now()) to demo "overdue" queries,
// mirroring the Oracle POC's demo scenario #1.
func seedSamples(ctx context.Context, gdb *gorm.DB, samples *postgressample.Repository, coc *postgressample.CoCRepository, idgen *postgresidgen.Adapter, users map[string]domainuser.User) []sample.Sample {
	create := applicationsample.NewCreateSampleUseCase(samples, coc, idgen)
	updateStatus := applicationsample.NewUpdateSampleStatusUseCase(samples, coc)

	const total = 40
	overdueDays := []int{15, 14, 10, 7} // applied in order to the pending samples below
	pendingSeen := 0

	out := make([]sample.Sample, 0, total)
	for i := 0; i < total; i++ {
		spec := sampleTypeSpecs[i%len(sampleTypeSpecs)]
		custodian := sampleCustodians[i%len(sampleCustodians)]

		created, err := create.Execute(ctx, applicationsample.CreateSampleInput{
			Name:      fmt.Sprintf("%s #%02d", spec.name, i+1),
			Type:      spec.typ,
			Custodian: custodian,
			Location:  spec.location,
		})
		if err != nil {
			log.Fatalf("seed sample %d: %v", i, err)
		}

		switch i % 10 {
		case 0: // stays pending, backdated to look overdue
			if pendingSeen < len(overdueDays) {
				backdated := time.Now().AddDate(0, 0, -overdueDays[pendingSeen])
				pendingSeen++
				if err := gdb.WithContext(ctx).Exec("UPDATE samples SET received_at = ? WHERE id = ?", backdated, created.ID).Error; err != nil {
					log.Fatalf("backdate sample %s: %v", created.ID, err)
				}
				created.ReceivedAt = backdated
			}
		case 1: // received today, transferred to another department
			created, err = updateStatus.Execute(ctx, applicationsample.UpdateSampleStatusInput{
				SampleID: created.ID, NewStatus: sample.StatusTransferred, ActorRole: domainuser.RoleQA, ActorName: users["qa"].Name,
			})
		case 2, 3: // in testing
			created, err = updateStatus.Execute(ctx, applicationsample.UpdateSampleStatusInput{
				SampleID: created.ID, NewStatus: sample.StatusTesting, ActorRole: domainuser.RoleScientist, ActorName: users["scientist"].Name,
			})
		default: // completed (testing -> completed)
			if created, err = updateStatus.Execute(ctx, applicationsample.UpdateSampleStatusInput{
				SampleID: created.ID, NewStatus: sample.StatusTesting, ActorRole: domainuser.RoleScientist, ActorName: users["scientist"].Name,
			}); err == nil {
				created, err = updateStatus.Execute(ctx, applicationsample.UpdateSampleStatusInput{
					SampleID: created.ID, NewStatus: sample.StatusCompleted, ActorRole: domainuser.RoleQA, ActorName: users["qa"].Name,
				})
			}
		}
		if err != nil {
			log.Fatalf("seed sample %s status transition: %v", created.ID, err)
		}

		out = append(out, created)
	}
	log.Printf("seed: created %d samples", len(out))
	return out
}

var testNameSpecs = []struct {
	name     string
	refRange string
}{
	{"CBC - Complete Blood Count", "4.5-11.0 x10^9/L"},
	{"IgG", "700-1600 mg/dL"},
	{"Coliform Count", "0 CFU/100mL"},
	{"LDL Cholesterol", "<130 mg/dL"},
	{"HDL Cholesterol", ">40 mg/dL"},
	{"BOD", "<20 mg/L"},
	{"Lead (Pb)", "<0.01 mg/L"},
	{"Glucose (Fasting)", "70-100 mg/dL"},
	{"ALT", "7-56 U/L"},
	{"HBsAg", "Non-reactive"},
	{"Arsenic (As)", "<0.01 mg/L"},
	{"Urinalysis - WBC", "0-5 /HPF"},
	{"Histopathology", "No malignancy"},
	{"Creatinine", "0.6-1.3 mg/dL"},
	{"TSH", "0.4-4.0 mIU/L"},
}

// flagCycle assigns hi/lo flags to ~40% of submitted results, mirroring the
// abnormal-result ratio in the Oracle POC's demo scenario #2.
var flagCycle = []testresult.Flag{testresult.FlagHi, testresult.FlagOk, testresult.FlagOk, testresult.FlagLo, testresult.FlagOk}

// seedTestResults creates one test result per sample (skipping the last
// still-pending sample, index 30, which has no result yet) and progresses
// most of them through submit -> approve so the demo has a realistic spread
// of analyzing/pending_verification/approved statuses and hi/lo/ok flags.
func seedTestResults(ctx context.Context, results *postgrestestresult.Repository, samples *postgressample.Repository, coc *postgressample.CoCRepository, idgen *postgresidgen.Adapter, seededSamples []sample.Sample, users map[string]domainuser.User, notifier *applicationnotification.AsNotifier) {
	create := applicationtestresult.NewCreateTestResultUseCase(results, samples, idgen)
	submit := applicationtestresult.NewSubmitResultUseCase(results)
	approve := applicationtestresult.NewApproveResultUseCase(results, coc, notifier)

	created := 0
	submitted := 0
	approved := 0
	for i, s := range seededSamples {
		if i == 30 {
			continue // leave this pending sample with no test result at all
		}

		spec := testNameSpecs[i%len(testNameSpecs)]
		tr, err := create.Execute(ctx, applicationtestresult.CreateTestResultInput{
			SampleID: s.ID, TestName: spec.name, Analyst: users["scientist"].Name, RefRange: spec.refRange,
		})
		if err != nil {
			log.Fatalf("seed test result for %s: %v", s.ID, err)
		}
		created++

		if s.Status == sample.StatusPending {
			continue // leave analyzing - matches the sample not being worked on yet
		}

		tr, err = submit.Execute(ctx, applicationtestresult.SubmitResultInput{
			ID: tr.ID, Result: "see report", Flag: flagCycle[submitted%len(flagCycle)],
		})
		if err != nil {
			log.Fatalf("submit test result %s: %v", tr.ID, err)
		}
		submitted++

		if s.Status == sample.StatusCompleted {
			if _, err := approve.Execute(ctx, applicationtestresult.ApproveResultInput{
				ID: tr.ID, ActorRole: domainuser.RoleQA, ActorName: users["qa"].Name,
			}); err != nil {
				log.Fatalf("approve test result %s: %v", tr.ID, err)
			}
			approved++
		}
	}
	log.Printf("seed: created %d test results (%d submitted, %d approved)", created, submitted, approved)
}

func seedEquipment(ctx context.Context, equipment *postgresequipment.Repository, idgen *postgresidgen.Adapter) {
	create := applicationequipment.NewCreateEquipmentUseCase(equipment, idgen)

	specs := []struct {
		name     string
		typeCode string
		days     int // NextCalibrationDue offset; negative = already overdue
	}{
		{"เครื่องปั่นเหวี่ยงตกตะกอน", "CENT", 60},
		{"เครื่อง PCR", "PCR", 5},
		{"เครื่องนึ่งฆ่าเชื้อ Autoclave", "AUTOCLAVE", -10},
		{"เครื่อง Spectrophotometer", "SPECTRO", 20},
		{"ตู้บ่มเชื้อ Incubator", "INCUBATOR", 90},
		{"เครื่องชั่งวิเคราะห์ Analytical Balance", "BALANCE", -3},
		{"Micropipette ชุดปรับปริมาตร", "PIPETTE", 45},
		{"ตู้เย็นเก็บตัวอย่าง Fridge-A", "FRIDGE", 15},
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

func seedInventory(ctx context.Context, inventoryRepo *postgresinventory.Repository, idgen *postgresidgen.Adapter) []inventory.InventoryItem {
	create := applicationinventory.NewCreateItemUseCase(inventoryRepo, idgen)

	specs := []struct {
		name          string
		category      string
		qty, min, max int
		unit          string
		vendor        string
	}{
		// below min - demo the low-stock / auto-reorder flow (see seedPurchaseOrders)
		{"เอทานอล 95%", "รีเอเจนต์", 5, 20, 100, "ลิตร", "บริษัท เคมีภัณฑ์ไทย จำกัด"},
		{"น้ำยา Glucose Reagent", "รีเอเจนต์", 5, 20, 50, "ขวด", "บริษัท เคมีภัณฑ์ไทย จำกัด"},
		{"หลอดเก็บซีรัม (Serum Vial)", "วัสดุสิ้นเปลือง", 8, 30, 100, "กล่อง", "บริษัท เซฟตี้ ซัพพลาย จำกัด"},
		{"ชุดทดสอบโลหะหนัก (Heavy Metal Kit)", "ชุดทดสอบ", 6, 12, 30, "ชุด", "บริษัท ไดแอกโนสติก ซัพพลาย จำกัด"},
		{"น้ำยา TSH Immunoassay Kit", "ชุดทดสอบ", 14, 15, 40, "ชุด", "บริษัท ไดแอกโนสติก ซัพพลาย จำกัด"},
		{"น้ำยา HBsAg Test Kit", "ชุดทดสอบ", 9, 15, 40, "ชุด", "บริษัท ไดแอกโนสติก ซัพพลาย จำกัด"},
		// above min - normal stock
		{"ถุงมือไนไตรล์ M", "วัสดุสิ้นเปลือง", 80, 30, 100, "กล่อง", "บริษัท เซฟตี้ ซัพพลาย จำกัด"},
		{"กระดาษกรอง Whatman No.1", "วัสดุสิ้นเปลือง", 200, 50, 300, "แผ่น", "บริษัท แล็บแวร์ไทย จำกัด"},
		{"โซเดียมคลอไรด์", "รีเอเจนต์", 40, 15, 80, "กิโลกรัม", "บริษัท เคมีภัณฑ์ไทย จำกัด"},
		{"หน้ากากอนามัย N95", "อุปกรณ์ป้องกัน", 60, 20, 150, "ชิ้น", "บริษัท เซฟตี้ ซัพพลาย จำกัด"},
		{"Micropipette Tips 1000uL", "วัสดุสิ้นเปลือง", 500, 100, 1000, "กล่อง", "บริษัท แล็บแวร์ไทย จำกัด"},
		{"ชุดตรวจ HIV Rapid Test", "ชุดทดสอบ", 25, 10, 60, "ชุด", "บริษัท ไดแอกโนสติก ซัพพลาย จำกัด"},
	}

	out := make([]inventory.InventoryItem, 0, len(specs))
	for _, s := range specs {
		item, err := create.Execute(ctx, applicationinventory.CreateItemInput{
			Name: s.name, Category: s.category, Quantity: s.qty, Unit: s.unit, Min: s.min, Max: s.max,
			DefaultVendor: s.vendor,
		})
		if err != nil {
			log.Fatalf("seed inventory %s: %v", s.name, err)
		}
		out = append(out, item)
	}
	log.Printf("seed: created %d inventory items", len(out))
	return out
}

// seedPurchaseOrders reorders every below-min inventory item (see
// seedInventory) via the same use case the API's manual-reorder endpoint
// uses, then varies each PO's lifecycle stage so the demo has a mix of
// pending_approval/sent_to_vendor/received - not just freshly-created rows.
func seedPurchaseOrders(ctx context.Context, purchaseOrders *postgrespurchaseorder.Repository, inventoryRepo *postgresinventory.Repository, idgen *postgresidgen.Adapter) {
	reorder := applicationpurchaseorder.NewCreateFromLowStockUseCase(purchaseOrders, inventoryRepo, idgen)
	approve := applicationpurchaseorder.NewApprovePOUseCase(purchaseOrders)
	receive := applicationpurchaseorder.NewMarkReceivedUseCase(purchaseOrders, inventoryRepo)

	items, err := inventoryRepo.List(ctx)
	if err != nil {
		log.Fatalf("list inventory for reorder: %v", err)
	}

	created := 0
	for _, item := range items {
		if !item.BelowMin() {
			continue
		}
		po, err := reorder.Execute(ctx, applicationpurchaseorder.CreateFromLowStockInput{ItemID: item.ID, Vendor: item.DefaultVendor})
		if err != nil {
			log.Fatalf("seed PO for %s: %v", item.ID, err)
		}
		created++

		switch created % 3 {
		case 0: // leave pending_approval
		case 1: // approve to sent_to_vendor
			if _, err := approve.Execute(ctx, applicationpurchaseorder.ApprovePOInput{ID: po.ID, ActorRole: domainuser.RoleQA}); err != nil {
				log.Fatalf("approve PO %s: %v", po.ID, err)
			}
		case 2: // full lifecycle: approve then receive (also restocks the item)
			if _, err := approve.Execute(ctx, applicationpurchaseorder.ApprovePOInput{ID: po.ID, ActorRole: domainuser.RoleQA}); err != nil {
				log.Fatalf("approve PO %s: %v", po.ID, err)
			}
			if _, err := receive.Execute(ctx, po.ID); err != nil {
				log.Fatalf("receive PO %s: %v", po.ID, err)
			}
		}
	}
	log.Printf("seed: created %d purchase orders from low-stock items", created)
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
		{"SOP การตรวจสอบคุณภาพน้ำดื่ม", document.TypeSOP, "sop-water-quality.txt", "ขั้นตอนมาตรฐานการตรวจคุณภาพน้ำดื่ม (demo placeholder)"},
		{"คู่มือการใช้เครื่อง PCR", document.TypeManual, "pcr-manual.txt", "คู่มือการใช้งานเครื่อง PCR (demo placeholder)"},
		{"คู่มือการใช้เครื่องปั่นเหวี่ยง", document.TypeManual, "centrifuge-manual.txt", "คู่มือการใช้งานเครื่องปั่นเหวี่ยงตกตะกอน (demo placeholder)"},
		{"นโยบายความปลอดภัยห้องปฏิบัติการ", document.TypePolicy, "lab-safety-policy.txt", "นโยบายความปลอดภัยในห้องปฏิบัติการ (demo placeholder)"},
		{"แบบฟอร์มขอเบิกวัสดุสิ้นเปลือง", document.TypeForm, "supply-request-form.txt", "แบบฟอร์มขอเบิกวัสดุสิ้นเปลือง (demo placeholder)"},
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
		{Location: "Incubator-1", Unit: "°C", RangeMin: 35, RangeMax: 39},
	}
	if err := gdb.WithContext(ctx).Create(&gaugeSpecs).Error; err != nil {
		log.Fatalf("seed gauges: %v", err)
	}

	evaluate := applicationenvironment.NewEvaluateThresholdsUseCase(gauges, alerts, notifier, httpenvironment.NewHub())
	record := applicationenvironment.NewRecordReadingUseCase(readings, evaluate)

	// Freezer-1's second reading is deliberately out of range (max -15) so
	// the seed leaves one open EnvAlert - matching the Oracle POC's "flagged
	// result" demo pattern for the environment monitoring module.
	readingSpecs := []struct {
		location string
		value    float64
	}{
		{"Fridge-A", 4.5},
		{"Freezer-1", -18.2},
		{"Freezer-1", -12.1},
		{"Incubator-1", 37.2},
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
	qaID := users["qa"].ID
	scientistID := users["scientist"].ID
	specs := []applicationnotification.CreateNotificationInput{
		{RecipientUserID: &adminID, Tone: notification.ToneAmber, Icon: "AlertTriangle", Title: "สต๊อกใกล้หมด", Message: "เอทานอล 95% เหลือต่ำกว่าเกณฑ์ขั้นต่ำ"},
		{RecipientUserID: &qaID, Tone: notification.ToneRed, Icon: "AlertCircle", Title: "ผลทดสอบผิดปกติ", Message: "พบผลทดสอบ flag ผิดปกติหลายรายการ รอตรวจสอบ"},
		{RecipientUserID: &scientistID, Tone: notification.ToneAmber, Icon: "Clock", Title: "ตัวอย่างค้างสถานะ pending", Message: "มีตัวอย่างค้างสถานะ pending เกิน 7 วัน กรุณาตรวจสอบ"},
		{RecipientUserID: &qaID, Tone: notification.ToneTeal, Icon: "FileText", Title: "ใบสั่งซื้อรออนุมัติ", Message: "มีใบสั่งซื้อ (PO) รออนุมัติจากการ reorder อัตโนมัติ"},
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
