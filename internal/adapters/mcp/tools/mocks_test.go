package tools_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portspurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
	portssample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	portstestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	portsuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/stretchr/testify/mock"
)

// mockSampleRepository implements portssample.SampleRepository. Only
// FindByID and List are exercised by the Sample tools under test; the rest
// are implemented to satisfy the interface, following the existing
// internal/application/chatbot/mocks_test.go house style.
type mockSampleRepository struct{ mock.Mock }

func (m *mockSampleRepository) Create(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) FindByID(ctx context.Context, id string) (sample.Sample, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) FindByBarcodeID(ctx context.Context, barcodeID string) (sample.Sample, error) {
	args := m.Called(ctx, barcodeID)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) List(ctx context.Context, filter portssample.ListFilter) ([]sample.Sample, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) UpdateStatus(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) UpdateLocation(ctx context.Context, sampleID string, locationID, position *string) (sample.Sample, error) {
	args := m.Called(ctx, sampleID, locationID, position)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) ListActiveByLocation(ctx context.Context, locationID string) ([]sample.Sample, error) {
	args := m.Called(ctx, locationID)
	return args.Get(0).([]sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) MoveWithinBox(ctx context.Context, boxID string, moves []portssample.PositionAssignment) ([]sample.Sample, error) {
	args := m.Called(ctx, boxID, moves)
	return args.Get(0).([]sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) UpdateBarcodeID(ctx context.Context, sampleID string, barcodeID *string) (sample.Sample, error) {
	args := m.Called(ctx, sampleID, barcodeID)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepository) ExistsActiveByLocation(ctx context.Context, locationID string) (bool, error) {
	args := m.Called(ctx, locationID)
	return args.Bool(0), args.Error(1)
}
func (m *mockSampleRepository) ExistsActiveByLocationPosition(ctx context.Context, locationID, position string) (bool, error) {
	args := m.Called(ctx, locationID, position)
	return args.Bool(0), args.Error(1)
}
func (m *mockSampleRepository) ExistsByLocation(ctx context.Context, locationID string) (bool, error) {
	args := m.Called(ctx, locationID)
	return args.Bool(0), args.Error(1)
}

// mockUserRepository implements portsuser.UserRepository.
type mockUserRepository struct{ mock.Mock }

func (m *mockUserRepository) Create(ctx context.Context, u user.User) (user.User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepository) FindByID(ctx context.Context, id int64) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepository) List(ctx context.Context) ([]user.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]user.User), args.Error(1)
}
func (m *mockUserRepository) Update(ctx context.Context, u user.User) (user.User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(user.User), args.Error(1)
}
func (m *mockUserRepository) Retire(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockUserRepository) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	args := m.Called(ctx, role)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockUserRepository) CountActiveByRole(ctx context.Context, role user.Role) (int64, error) {
	args := m.Called(ctx, role)
	return args.Get(0).(int64), args.Error(1)
}

var _ portsuser.UserRepository = (*mockUserRepository)(nil)

// mockTestResultRepository implements portstestresult.Repository.
type mockTestResultRepository struct{ mock.Mock }

func (m *mockTestResultRepository) Create(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(testresult.TestResult), args.Error(1)
}
func (m *mockTestResultRepository) FindByID(ctx context.Context, id string) (testresult.TestResult, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(testresult.TestResult), args.Error(1)
}
func (m *mockTestResultRepository) List(ctx context.Context, filter portstestresult.ListFilter) ([]testresult.TestResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]testresult.TestResult), args.Error(1)
}
func (m *mockTestResultRepository) Update(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(testresult.TestResult), args.Error(1)
}

// mockInventoryRepository implements portsinventory.Repository.
type mockInventoryRepository struct{ mock.Mock }

func (m *mockInventoryRepository) Create(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	args := m.Called(ctx, i)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepository) FindByID(ctx context.Context, id string) (inventory.InventoryItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepository) List(ctx context.Context) ([]inventory.InventoryItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepository) Update(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	args := m.Called(ctx, i)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepository) UpdateDefaultVendor(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error) {
	args := m.Called(ctx, id, vendor)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}

// mockLotRepository implements portsinventory.LotRepository.
type mockLotRepository struct{ mock.Mock }

func (m *mockLotRepository) Create(ctx context.Context, l inventory.InventoryLot) (inventory.InventoryLot, error) {
	args := m.Called(ctx, l)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepository) FindByID(ctx context.Context, id string) (inventory.InventoryLot, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepository) FindByItemAndLotNo(ctx context.Context, itemID, lotNo string) (inventory.InventoryLot, error) {
	args := m.Called(ctx, itemID, lotNo)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepository) ListByItem(ctx context.Context, itemID string) ([]inventory.InventoryLot, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepository) UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryLot, error) {
	args := m.Called(ctx, id, quantity)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}

// mockPurchaseOrderRepository implements portspurchaseorder.Repository.
type mockPurchaseOrderRepository struct{ mock.Mock }

func (m *mockPurchaseOrderRepository) Create(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx, po)
	return args.Get(0).(purchaseorder.PurchaseOrder), args.Error(1)
}
func (m *mockPurchaseOrderRepository) FindByID(ctx context.Context, id string) (purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(purchaseorder.PurchaseOrder), args.Error(1)
}
func (m *mockPurchaseOrderRepository) List(ctx context.Context) ([]purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx)
	return args.Get(0).([]purchaseorder.PurchaseOrder), args.Error(1)
}
func (m *mockPurchaseOrderRepository) Update(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx, po)
	return args.Get(0).(purchaseorder.PurchaseOrder), args.Error(1)
}

var (
	_ portssample.SampleRepository  = (*mockSampleRepository)(nil)
	_ portstestresult.Repository    = (*mockTestResultRepository)(nil)
	_ portsinventory.Repository     = (*mockInventoryRepository)(nil)
	_ portsinventory.LotRepository  = (*mockLotRepository)(nil)
	_ portspurchaseorder.Repository = (*mockPurchaseOrderRepository)(nil)
)
