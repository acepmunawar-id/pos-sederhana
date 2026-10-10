package service

import (
	"context"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/pkg/pagination"
	"github.com/stretchr/testify/mock"
)

// fakeTx runs fn directly; transactional behaviour is covered by the repository layer.
type fakeTx struct{ calls int }

func (f *fakeTx) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	f.calls++
	return fn(ctx)
}

// fakeHasher treats "hash:<pw>" as the hash of pw.
type fakeHasher struct{}

func (fakeHasher) Hash(pw string) (string, error) { return "hash:" + pw, nil }
func (fakeHasher) Compare(hash, pw string) bool   { return hash == "hash:"+pw }

type fakeIssuer struct{}

func (fakeIssuer) Generate(userID int64, role, name string, now time.Time) (string, time.Time, error) {
	return "access-token", now.Add(15 * time.Minute), nil
}

func fixedClock(t time.Time) Clock { return func() time.Time { return t } }

func ptr[T any](v T) *T { return &v }

// ---------- UserRepository ----------

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id int64, h string) error {
	return m.Called(ctx, id, h).Error(0)
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id int64, at time.Time) error {
	return m.Called(ctx, id, at).Error(0)
}
func (m *mockUserRepo) SoftDelete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*domain.User)
	return u, args.Error(1)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*domain.User)
	return u, args.Error(1)
}
func (m *mockUserRepo) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	args := m.Called(ctx, email, excludeID)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) CountActiveByRole(ctx context.Context, role string) (int64, error) {
	args := m.Called(ctx, role)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockUserRepo) List(ctx context.Context, f domain.UserFilter) ([]domain.User, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

// ---------- RoleRepository ----------

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) List(ctx context.Context) ([]domain.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Role), args.Error(1)
}
func (m *mockRoleRepo) FindByID(ctx context.Context, id int64) (*domain.Role, error) {
	args := m.Called(ctx, id)
	r, _ := args.Get(0).(*domain.Role)
	return r, args.Error(1)
}

// ---------- RefreshTokenRepository ----------

type mockTokenRepo struct{ mock.Mock }

func (m *mockTokenRepo) Create(ctx context.Context, t *domain.RefreshToken) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTokenRepo) FindByHash(ctx context.Context, h string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, h)
	t, _ := args.Get(0).(*domain.RefreshToken)
	return t, args.Error(1)
}
func (m *mockTokenRepo) Revoke(ctx context.Context, id int64, at time.Time) error {
	return m.Called(ctx, id, at).Error(0)
}
func (m *mockTokenRepo) RevokeAllByUser(ctx context.Context, userID int64, at time.Time) error {
	return m.Called(ctx, userID, at).Error(0)
}

// ---------- CategoryRepository ----------

type mockCategoryRepo struct{ mock.Mock }

func (m *mockCategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCategoryRepo) SoftDelete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockCategoryRepo) FindByID(ctx context.Context, id int64) (*domain.Category, error) {
	args := m.Called(ctx, id)
	c, _ := args.Get(0).(*domain.Category)
	return c, args.Error(1)
}
func (m *mockCategoryRepo) NameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}
func (m *mockCategoryRepo) CountProducts(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCategoryRepo) List(ctx context.Context, p pagination.Params) ([]domain.Category, int64, error) {
	args := m.Called(ctx, p)
	return args.Get(0).([]domain.Category), args.Get(1).(int64), args.Error(2)
}

// ---------- ProductRepository ----------

type mockProductRepo struct{ mock.Mock }

func (m *mockProductRepo) Create(ctx context.Context, p *domain.Product) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProductRepo) Update(ctx context.Context, p *domain.Product) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProductRepo) UpdateImage(ctx context.Context, id int64, url *string) error {
	return m.Called(ctx, id, url).Error(0)
}
func (m *mockProductRepo) UpdateStock(ctx context.Context, id int64, stock int) error {
	return m.Called(ctx, id, stock).Error(0)
}
func (m *mockProductRepo) SoftDelete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockProductRepo) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	args := m.Called(ctx, id)
	p, _ := args.Get(0).(*domain.Product)
	return p, args.Error(1)
}
func (m *mockProductRepo) FindByBarcode(ctx context.Context, code string) (*domain.Product, error) {
	args := m.Called(ctx, code)
	p, _ := args.Get(0).(*domain.Product)
	return p, args.Error(1)
}
func (m *mockProductRepo) FindByIDsForUpdate(ctx context.Context, ids []int64) ([]domain.Product, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]domain.Product), args.Error(1)
}
func (m *mockProductRepo) SKUExists(ctx context.Context, sku string, excludeID int64) (bool, error) {
	args := m.Called(ctx, sku, excludeID)
	return args.Bool(0), args.Error(1)
}
func (m *mockProductRepo) BarcodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	args := m.Called(ctx, code, excludeID)
	return args.Bool(0), args.Error(1)
}
func (m *mockProductRepo) CountLowStock(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockProductRepo) List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.Product), args.Get(1).(int64), args.Error(2)
}

// ---------- UnitRepository ----------

type mockUnitRepo struct{ mock.Mock }

func (m *mockUnitRepo) Create(ctx context.Context, u *domain.Unit) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUnitRepo) Update(ctx context.Context, u *domain.Unit) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUnitRepo) SoftDelete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUnitRepo) FindByID(ctx context.Context, id int64) (*domain.Unit, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*domain.Unit)
	return u, args.Error(1)
}
func (m *mockUnitRepo) NameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}
func (m *mockUnitRepo) CountProducts(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockUnitRepo) List(ctx context.Context, p pagination.Params) ([]domain.Unit, int64, error) {
	args := m.Called(ctx, p)
	return args.Get(0).([]domain.Unit), args.Get(1).(int64), args.Error(2)
}

// ---------- StockMovementRepository ----------

type mockMovementRepo struct{ mock.Mock }

func (m *mockMovementRepo) Create(ctx context.Context, mv *domain.StockMovement) error {
	return m.Called(ctx, mv).Error(0)
}
func (m *mockMovementRepo) List(ctx context.Context, f domain.StockMovementFilter) ([]domain.StockMovement, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.StockMovement), args.Get(1).(int64), args.Error(2)
}

// ---------- ShiftRepository ----------

type mockShiftRepo struct{ mock.Mock }

func (m *mockShiftRepo) Create(ctx context.Context, s *domain.Shift) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockShiftRepo) Close(ctx context.Context, s *domain.Shift) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockShiftRepo) FindByID(ctx context.Context, id int64) (*domain.Shift, error) {
	args := m.Called(ctx, id)
	s, _ := args.Get(0).(*domain.Shift)
	return s, args.Error(1)
}
func (m *mockShiftRepo) FindOpenByUser(ctx context.Context, userID int64) (*domain.Shift, error) {
	args := m.Called(ctx, userID)
	s, _ := args.Get(0).(*domain.Shift)
	return s, args.Error(1)
}
func (m *mockShiftRepo) Summary(ctx context.Context, id int64) (*domain.ShiftSummary, error) {
	args := m.Called(ctx, id)
	s, _ := args.Get(0).(*domain.ShiftSummary)
	return s, args.Error(1)
}
func (m *mockShiftRepo) List(ctx context.Context, f domain.ShiftFilter) ([]domain.Shift, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.Shift), args.Get(1).(int64), args.Error(2)
}

// ---------- TransactionRepository ----------

type mockTxRepo struct{ mock.Mock }

func (m *mockTxRepo) Create(ctx context.Context, t *domain.Transaction) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTxRepo) FindByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	args := m.Called(ctx, id)
	t, _ := args.Get(0).(*domain.Transaction)
	return t, args.Error(1)
}
func (m *mockTxRepo) FindByIDForUpdate(ctx context.Context, id int64) (*domain.Transaction, error) {
	args := m.Called(ctx, id)
	t, _ := args.Get(0).(*domain.Transaction)
	return t, args.Error(1)
}
func (m *mockTxRepo) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	args := m.Called(ctx, key)
	t, _ := args.Get(0).(*domain.Transaction)
	return t, args.Error(1)
}
func (m *mockTxRepo) MarkVoided(ctx context.Context, id, by int64, reason string, at time.Time) error {
	return m.Called(ctx, id, by, reason, at).Error(0)
}
func (m *mockTxRepo) NextInvoiceSequence(ctx context.Context, day time.Time) (int, error) {
	args := m.Called(ctx, day)
	return args.Int(0), args.Error(1)
}
func (m *mockTxRepo) List(ctx context.Context, f domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]domain.Transaction), args.Get(1).(int64), args.Error(2)
}

// ---------- StoreSettingRepository ----------

type mockSettingRepo struct{ mock.Mock }

func (m *mockSettingRepo) Get(ctx context.Context) (*domain.StoreSetting, error) {
	args := m.Called(ctx)
	s, _ := args.Get(0).(*domain.StoreSetting)
	return s, args.Error(1)
}
func (m *mockSettingRepo) Update(ctx context.Context, s *domain.StoreSetting) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockSettingRepo) UpdateLogo(ctx context.Context, url *string) error {
	return m.Called(ctx, url).Error(0)
}
