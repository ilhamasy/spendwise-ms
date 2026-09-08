package service

import (
	"testing"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"
	"spendwise-ms/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------- Mock Implementations ----------

type mockTxRepo struct {
	store map[string]*model.Transaction
}

func newMockTxRepo() *mockTxRepo { return &mockTxRepo{store: map[string]*model.Transaction{}} }
func (m *mockTxRepo) Create(tx *model.Transaction) error {
	if tx.ID == "" { tx.ID = uuid.New().String() }
	m.store[tx.ID] = tx; return nil
}
func (m *mockTxRepo) FindByID(id, _ string) (*model.Transaction, error) {
	if t, ok := m.store[id]; ok { return t, nil }
	return nil, gorm.ErrRecordNotFound
}
func (m *mockTxRepo) FindAll(_ repository.TransactionFilter) ([]model.Transaction, int64, error) {
	r := make([]model.Transaction, 0, len(m.store))
	for _, t := range m.store { r = append(r, *t) }
	return r, int64(len(r)), nil
}
func (m *mockTxRepo) Update(tx *model.Transaction) error { m.store[tx.ID] = tx; return nil }
func (m *mockTxRepo) Delete(id, _ string) error           { delete(m.store, id); return nil }
func (m *mockTxRepo) BatchCreate(txs []model.Transaction) (int, []int, []string) {
	return len(txs), nil, nil
}
func (m *mockTxRepo) FindCategoryByID(id, _ string) (*model.Category, error) {
	return &model.Category{ID: id, Name: "Cat", Type: "expense", Icon: "📁"}, nil
}

type mockCatRepo struct {
	store map[string]*model.Category
}

func newMockCatRepo() *mockCatRepo { return &mockCatRepo{store: map[string]*model.Category{}} }
func (m *mockCatRepo) Create(c *model.Category) error {
	if c.ID == "" { c.ID = uuid.New().String() }
	m.store[c.ID] = c; return nil
}
func (m *mockCatRepo) FindByID(id, _ string) (*model.Category, error) {
	if c, ok := m.store[id]; ok { return c, nil }
	return nil, gorm.ErrRecordNotFound
}
func (m *mockCatRepo) FindAll(_, _ string) ([]model.Category, error) {
	r := make([]model.Category, 0, len(m.store))
	for _, c := range m.store { r = append(r, *c) }
	return r, nil
}
func (m *mockCatRepo) FindAllForSync(_ string) ([]model.Category, error) {
	return m.FindAll("", "")
}
func (m *mockCatRepo) Update(c *model.Category) error { m.store[c.ID] = c; return nil }
func (m *mockCatRepo) Delete(id, _ string) error      { delete(m.store, id); return nil }
func (m *mockCatRepo) FindByNameAndType(_, _, _ string) (*model.Category, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockCatRepo) CountTransactionsByCategoryID(_, _ string) (int64, error) { return 0, nil }
func (m *mockCatRepo) ReassignTransactions(_, _, _ string) error             { return nil }

type mockGoalRepo struct {
	store map[string]*model.SavingGoal
}

func newMockGoalRepo() *mockGoalRepo { return &mockGoalRepo{store: map[string]*model.SavingGoal{}} }
func (m *mockGoalRepo) Create(g *model.SavingGoal) error { m.store[g.ID] = g; return nil }
func (m *mockGoalRepo) FindByID(id, _ string) (*model.SavingGoal, error) {
	if g, ok := m.store[id]; ok { return g, nil }
	return nil, gorm.ErrRecordNotFound
}
func (m *mockGoalRepo) FindAll(_, _ string) ([]model.SavingGoal, error) {
	r := make([]model.SavingGoal, 0, len(m.store))
	for _, g := range m.store { r = append(r, *g) }
	return r, nil
}
func (m *mockGoalRepo) Update(g *model.SavingGoal) error { m.store[g.ID] = g; return nil }
func (m *mockGoalRepo) UpdateStatus(id, _, status string) error {
	if g, ok := m.store[id]; ok { g.Status = status; return nil }
	return nil
}
func (m *mockGoalRepo) Delete(id, _ string) error { delete(m.store, id); return nil }
func (m *mockGoalRepo) CreateContribution(_ *model.GoalContribution) error { return nil }
func (m *mockGoalRepo) FindContributionsByGoalID(_ string, _, _ int) ([]model.GoalContribution, int64, error) {
	return nil, 0, nil
}

type mockBudgetRepo struct {
	store map[string]*model.Budget
}

func newMockBudgetRepo() *mockBudgetRepo { return &mockBudgetRepo{store: map[string]*model.Budget{}} }
func (m *mockBudgetRepo) Create(b *model.Budget) error {
	if b.ID == "" { b.ID = uuid.New().String() }
	m.store[b.ID] = b; return nil
}
func (m *mockBudgetRepo) FindByID(id, _ string) (*model.Budget, error) {
	if b, ok := m.store[id]; ok { return b, nil }
	return nil, gorm.ErrRecordNotFound
}
func (m *mockBudgetRepo) FindAll(_, _ string) ([]model.Budget, error) {
	r := make([]model.Budget, 0, len(m.store))
	for _, b := range m.store { r = append(r, *b) }
	return r, nil
}
func (m *mockBudgetRepo) Update(b *model.Budget) error { m.store[b.ID] = b; return nil }
func (m *mockBudgetRepo) Delete(id, _ string) error    { delete(m.store, id); return nil }
func (m *mockBudgetRepo) GetSpentAmount(_, _, _ string) (int64, error) { return 0, nil }

// ---------- Mock-Based Tests ----------

func TestTxService_CreateGetUpdateDelete_Mock(t *testing.T) {
	svc := NewTransactionServiceWithRepo(newMockTxRepo())
	uid := uuid.New().String()

	tx, err := svc.CreateTransaction(uid, dto.CreateTransactionRequest{
		Type: "expense", Amount: 50000, CategoryID: "c1", OccurredAt: "2026-06-01",
	})
	if err != nil || tx.ID == "" {
		t.Fatalf("Create: err=%v id=%s", err, tx.ID)
	}

	got, _ := svc.GetTransactionByID(uid, tx.ID)
	if got.Amount != 50000 {
		t.Errorf("GetByID: expected 50000, got %d", got.Amount)
	}

	upd, _ := svc.UpdateTransaction(uid, tx.ID, dto.UpdateTransactionRequest{Amount: 99999})
	if upd.Amount != 99999 {
		t.Errorf("Update: expected 99999, got %d", upd.Amount)
	}

	svc.DeleteTransaction(uid, tx.ID)
	if _, err := svc.GetTransactionByID(uid, tx.ID); err == nil {
		t.Error("Should not find deleted tx")
	}
}

func TestCatService_CreateDelete_Mock(t *testing.T) {
	svc := NewCategoryServiceWithRepo(newMockCatRepo())
	uid := uuid.New().String()

	cat, err := svc.CreateCategory(uid, dto.CreateCategoryRequest{Name: "Food", Type: "expense", Icon: "🍔", Color: "#ef4444"})
	if err != nil || cat.Name != "Food" {
		t.Fatalf("CreateCategory: err=%v", err)
	}

	if err := svc.DeleteCategory(uid, cat.ID, ""); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
}

func TestGoalService_CreateArchive_Mock(t *testing.T) {
	svc := NewGoalServiceWithRepo(newMockGoalRepo())
	uid := uuid.New().String()

	goal, err := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Car", TargetAmount: 100000000})
	if err != nil || goal.Name != "Car" {
		t.Fatalf("CreateGoal: err=%v", err)
	}

	archived, _ := svc.ArchiveGoal(uid, goal.ID)
	if archived.Status != "archived" {
		t.Errorf("Expected archived, got %s", archived.Status)
	}

	svc.DeleteGoal(uid, goal.ID)
	if _, err := svc.GetGoal(uid, goal.ID); err == nil {
		t.Error("Should not find deleted goal")
	}
}

func TestBudgetService_Create_Mock(t *testing.T) {
	catRepo := newMockCatRepo()
	svc := NewBudgetServiceWithRepo(newMockBudgetRepo(), catRepo)
	uid := uuid.New().String()

	catRepo.Create(&model.Category{ID: "c1", Name: "Food", Type: "expense"})

	bud, err := svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "Food Budget", Amount: 2000000, Period: "monthly", CategoryID: "c1"})
	if err != nil || bud.Name != "Food Budget" {
		t.Fatalf("CreateBudget: err=%v", err)
	}

	svc.DeleteBudget(uid, bud.ID)
	if _, err := svc.GetBudget(uid, bud.ID); err == nil {
		t.Error("Should not find deleted budget")
	}
}

func TestBudgetService_List_Mock(t *testing.T) {
	catRepo := newMockCatRepo()
	catRepo.Create(&model.Category{ID: "c1", Name: "Food", Type: "expense"})
	svc := NewBudgetServiceWithRepo(newMockBudgetRepo(), catRepo)
	uid := uuid.New().String()

	svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "B1", Amount: 1000, Period: "monthly", CategoryID: "c1"})
	svc.CreateBudget(uid, dto.CreateBudgetRequest{Name: "B2", Amount: 2000, Period: "weekly", CategoryID: "c1"})

	all, _ := svc.ListBudgets(uid, "")
	if len(all) != 2 {
		t.Errorf("Expected 2 budgets, got %d", len(all))
	}
}
