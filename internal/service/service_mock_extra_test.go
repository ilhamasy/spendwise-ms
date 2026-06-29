package service

import (
	"errors"
	"testing"

	"spendwise-ms/internal/dto"
	"spendwise-ms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------- Extended Mock Tests ----------

func TestTxSvc_UpdateNotFound_Mock(t *testing.T) {
	r := &failFindTxRepo{}
	svc := NewTransactionServiceWithRepo(r)
	_, err := svc.UpdateTransaction("u1", "bad", dto.UpdateTransactionRequest{Amount: 100})
	if err == nil { t.Error("expected error") }
}

func TestTxSvc_DeleteNotFound_Mock(t *testing.T) {
	r := &failFindTxRepo{}
	svc := NewTransactionServiceWithRepo(r)
	err := svc.DeleteTransaction("u1", "bad")
	if err == nil { t.Error("expected error") }
}

func TestTxSvc_Sync_Mock(t *testing.T) {
	r := newMockTxRepo()
	svc := NewTransactionServiceWithRepo(r)
	resp, err := svc.SyncTransactions("u1", dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{
			{Type: "expense", Amount: 1000, CategoryID: "c1", OccurredAt: "2026-06-28"},
			{Type: "income", Amount: 5000, CategoryID: "c1", OccurredAt: "2026-06-28"},
		},
	})
	if err != nil { t.Fatalf("sync: %v", err) }
	if resp.Synced != 2 { t.Errorf("synced: %d", resp.Synced) }
}

func TestTxSvc_SyncEmpty_Mock(t *testing.T) {
	svc := NewTransactionServiceWithRepo(newMockTxRepo())
	_, err := svc.SyncTransactions("u1", dto.SyncRequest{Transactions: []dto.SyncTransactionRequest{}})
	if err == nil { t.Error("expected error for empty") }
}

func TestTxSvc_SyncInvalidDate_Mock(t *testing.T) {
	r := newMockTxRepo()
	svc := NewTransactionServiceWithRepo(r)
	resp, err := svc.SyncTransactions("u1", dto.SyncRequest{
		Transactions: []dto.SyncTransactionRequest{{Type: "expense", Amount: 1000, CategoryID: "c1", OccurredAt: "not-a-date"}},
	})
	if err != nil { t.Fatal("should not fail on validation, should return failed items") }
	if len(resp.Failed) != 1 { t.Errorf("expected 1 failed, got %d", len(resp.Failed)) }
}

func TestCatSvc_UpdateNotFound_Mock(t *testing.T) {
	r := newMockCatRepo()
	svc := NewCategoryServiceWithRepo(r)
	_, err := svc.UpdateCategory("u1", "bad", dto.UpdateCategoryRequest{Name: "X"})
	if err == nil { t.Error("expected error") }
}


func TestGoalSvc_UpdateNotFound_Mock(t *testing.T) {
	r := &failFindGoalRepo{}
	svc := NewGoalServiceWithRepo(r)
	_, err := svc.UpdateGoal("u1", "bad", dto.UpdateGoalRequest{Name: "X"})
	if err == nil { t.Error("expected error") }
}


func TestGoalSvc_AddContribution_Mock(t *testing.T) {
	r := newMockGoalRepo()
	uid := uuid.New().String()
	r.Create(&model.SavingGoal{ID: "g1", UserID: uid, Name: "Car", TargetAmount: 1000000, CurrentSaved: 0})
	svc := NewGoalServiceWithRepo(r)

	c, g, err := svc.AddContribution(uid, "g1", dto.CreateContributionRequest{Amount: 200000, Date: "2026-06-28"})
	if err != nil { t.Fatalf("add contribution: %v", err) }
	if c.Amount != 200000 { t.Errorf("contrib amount: %d", c.Amount) }
	if g.CurrentSaved != 200000 { t.Errorf("goal saved: %d", g.CurrentSaved) }
}


func TestBudgetSvc_UpdateNotFound_Mock(t *testing.T) {
	r := &failFindBudgetRepo{}
	cr := newMockCatRepo()
	cr.Create(&model.Category{ID: "c1", Name: "Food", Type: "expense"})
	svc := NewBudgetServiceWithRepo(r, cr)
	_, err := svc.UpdateBudget("u1", "bad", dto.UpdateBudgetRequest{Name: "X"})
	if err == nil { t.Error("expected error") }
}


func TestBudgetSvc_GetNotFound_Mock(t *testing.T) {
	r := &failFindBudgetRepo{}
	cr := newMockCatRepo()
	cr.Create(&model.Category{ID: "c1", Name: "Food", Type: "expense"})
	svc := NewBudgetServiceWithRepo(r, cr)
	_, err := svc.GetBudget("u1", "bad")
	if err == nil { t.Error("expected error") }
}

// ---------- Failing Mock Repos ----------

type failFindTxRepo struct{ mockTxRepo }

func (f *failFindTxRepo) FindByID(_, _ string) (*model.Transaction, error) { return nil, gorm.ErrRecordNotFound }
func (f *failFindTxRepo) Create(_ *model.Transaction) error                { return nil }
func (f *failFindTxRepo) Delete(_, _ string) error                          { return gorm.ErrRecordNotFound }
func (f *failFindTxRepo) FindCategoryByID(_, _ string) (*model.Category, error) {
	return &model.Category{ID: "c1", Name: "C", Type: "expense"}, nil
}

type failFindGoalRepo struct{ mockGoalRepo }

func (f *failFindGoalRepo) FindByID(_, _ string) (*model.SavingGoal, error) { return nil, gorm.ErrRecordNotFound }

type failFindBudgetRepo struct{ mockBudgetRepo }

func (f *failFindBudgetRepo) FindByID(_, _ string) (*model.Budget, error) { return nil, gorm.ErrRecordNotFound }
func (f *failFindBudgetRepo) Create(_ *model.Budget) error                { return nil }
func (f *failFindBudgetRepo) Delete(_, _ string) error                    { return gorm.ErrRecordNotFound }
func (f *failFindBudgetRepo) FindAll(_, _ string) ([]model.Budget, error) { return nil, nil }




type dupCatRepo struct{ *mockCatRepo }

func (d *dupCatRepo) FindByNameAndType(name, catType, _ string) (*model.Category, error) {
	return &model.Category{ID: "dup", Name: name, Type: catType}, nil
}


func TestGoalSvc_ArchiveUnarchive_Mock(t *testing.T) {
	r := newMockGoalRepo()
	uid := uuid.New().String()
	svc := NewGoalServiceWithRepo(r)
	g, _ := svc.CreateGoal(uid, dto.CreateGoalRequest{Name: "Arc", TargetAmount: 1000})

	a, err := svc.ArchiveGoal(uid, g.ID)
	if err != nil { t.Fatalf("archive: %v", err) }
	if a.Status != "archived" { t.Errorf("status: %s", a.Status) }

	u, err := svc.UnarchiveGoal(uid, g.ID)
	if err != nil { t.Fatalf("unarchive: %v", err) }
	if u.Status != "active" { t.Errorf("status: %s", u.Status) }
}

func TestGoalSvc_GetNotFound_Mock(t *testing.T) {
	svc := NewGoalServiceWithRepo(newMockGoalRepo())
	_, err := svc.GetGoal("u1", "bad")
	if err == nil { t.Error("expected error") }
}



func TestBudgetSvc_DeleteNotFound_Mock(t *testing.T) {
	r := &failFindBudgetRepo{}
	cr := newMockCatRepo()
	cr.Create(&model.Category{ID: "c1", Name: "Food", Type: "expense"})
	svc := NewBudgetServiceWithRepo(r, cr)
	err := svc.DeleteBudget("u1", "bad")
	if err == nil { t.Error("expected error") }
}

type errCategoryRepo struct{}

func (e *errCategoryRepo) Create(_ *model.Category) error                           { return errors.New("db error") }
func (e *errCategoryRepo) FindByID(_, _ string) (*model.Category, error)            { return nil, errors.New("db error") }
func (e *errCategoryRepo) FindAll(_, _ string) ([]model.Category, error)            { return nil, errors.New("db error") }
func (e *errCategoryRepo) FindAllForSync(_ string) ([]model.Category, error)        { return nil, errors.New("db error") }
func (e *errCategoryRepo) Update(_ *model.Category) error                            { return errors.New("db error") }
func (e *errCategoryRepo) Delete(_, _ string) error                                  { return errors.New("db error") }
func (e *errCategoryRepo) FindByNameAndType(_, _, _ string) (*model.Category, error) { return nil, gorm.ErrRecordNotFound }
func (e *errCategoryRepo) CountTransactionsByCategoryID(_ string) (int64, error)     { return 0, nil }
func (e *errCategoryRepo) ReassignTransactions(_, _, _ string) error                  { return nil }

func TestCatSvc_Create_DBError(t *testing.T) {
	svc := NewCategoryServiceWithRepo(&errCategoryRepo{})
	_, err := svc.CreateCategory("u1", dto.CreateCategoryRequest{Name: "Food", Type: "expense"})
	if err == nil { t.Error("expected db error") }
}
