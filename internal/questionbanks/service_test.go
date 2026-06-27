package questionbanks_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/questionbanks"
)

type mockBankRepo struct{ mock.Mock }

func (m *mockBankRepo) Create(ctx context.Context, bank *domain.QuestionBank) error {
	return m.Called(ctx, bank).Error(0)
}
func (m *mockBankRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuestionBank, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuestionBank), a.Error(1)
}
func (m *mockBankRepo) Update(ctx context.Context, bank *domain.QuestionBank) error {
	return m.Called(ctx, bank).Error(0)
}
func (m *mockBankRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockBankRepo) List(ctx context.Context, orgID uuid.UUID, p domain.ListParams) ([]domain.QuestionBank, int64, error) {
	a := m.Called(ctx, orgID, p)
	bs, _ := a.Get(0).([]domain.QuestionBank)
	return bs, a.Get(1).(int64), a.Error(2)
}
func (m *mockBankRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockBankRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.QuestionBank, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuestionBank), a.Error(1)
}
func (m *mockBankRepo) AdminList(ctx context.Context, q domain.AdminListQuestionBanksQuery) ([]domain.QuestionBank, int64, error) {
	a := m.Called(ctx, q)
	bs, _ := a.Get(0).([]domain.QuestionBank)
	return bs, a.Get(1).(int64), a.Error(2)
}

type mockQuestionRepo struct{ mock.Mock }

func (m *mockQuestionRepo) Create(ctx context.Context, q *domain.Question) error {
	return m.Called(ctx, q).Error(0)
}
func (m *mockQuestionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Question), a.Error(1)
}
func (m *mockQuestionRepo) Update(ctx context.Context, q *domain.Question) error {
	return m.Called(ctx, q).Error(0)
}
func (m *mockQuestionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockQuestionRepo) ListByBank(ctx context.Context, bankID uuid.UUID, q domain.ListQuestionsQuery) ([]domain.Question, int64, error) {
	a := m.Called(ctx, bankID, q)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Get(1).(int64), a.Error(2)
}
func (m *mockQuestionRepo) ListAllByBank(ctx context.Context, bankID uuid.UUID) ([]domain.Question, error) {
	a := m.Called(ctx, bankID)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Error(1)
}

type mockMediaRepo struct{ mock.Mock }

func (m *mockMediaRepo) Create(ctx context.Context, media *domain.Media) error {
	return m.Called(ctx, media).Error(0)
}
func (m *mockMediaRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Media), a.Error(1)
}
func (m *mockMediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockMediaRepo) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	a := m.Called(ctx, modelType, modelID, collection)
	ms, _ := a.Get(0).([]domain.Media)
	return ms, a.Error(1)
}
func (m *mockQuestionRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Question, error) {
	a := m.Called(ctx, ids)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Error(1)
}
func (m *mockQuestionRepo) CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error) {
	a := m.Called(ctx, bankID)
	return a.Get(0).(int64), a.Error(1)
}
func (m *mockQuestionRepo) RandomByBank(ctx context.Context, bankID uuid.UUID, count int) ([]domain.Question, error) {
	a := m.Called(ctx, bankID, count)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Error(1)
}
func (m *mockQuestionRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockQuestionRepo) AdminList(ctx context.Context, q domain.AdminListQuestionsQuery) ([]domain.Question, int64, error) {
	a := m.Called(ctx, q)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Get(1).(int64), a.Error(2)
}

func staffCtx(orgIDs ...uuid.UUID) context.Context {
	caller := domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"question_banks:update_any", "question_banks:create", "question_banks:view", "question_banks:delete"},
	}
	if len(orgIDs) > 0 {
		caller.OrgID = &orgIDs[0]
	}
	return domain.WithCaller(context.Background(), caller)
}

func adminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:  uuid.New(),
		IsAdmin: true,
	})
}

func memberCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(),
	})
}

func newTestBankService(bankRepo *mockBankRepo, questionRepo *mockQuestionRepo) domain.QuestionBankService {
	return questionbanks.NewService(bankRepo, questionRepo, &mockMediaRepo{}, slog.Default())
}

func TestBankService_Create_AsStaff(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankRepo := &mockBankRepo{}
	qRepo := &mockQuestionRepo{}

	bankRepo.On("Create", ctx, mock.AnythingOfType("*domain.QuestionBank")).Return(nil)

	svc := newTestBankService(bankRepo, qRepo)
	bank, err := svc.Create(ctx, domain.CreateQuestionBankDTO{Name: "Physics", Description: "desc"})

	assert.NoError(t, err)
	assert.Equal(t, "Physics", bank.Name)
	assert.Equal(t, orgID, bank.OrganizationID)
	bankRepo.AssertExpectations(t)
}

func TestBankService_Create_NoCaller_Forbidden(t *testing.T) {
	svc := newTestBankService(&mockBankRepo{}, &mockQuestionRepo{})
	_, err := svc.Create(context.Background(), domain.CreateQuestionBankDTO{Name: "X"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestBankService_Create_NonStaff_Forbidden(t *testing.T) {
	ctx := memberCtx()
	svc := newTestBankService(&mockBankRepo{}, &mockQuestionRepo{})
	_, err := svc.Create(ctx, domain.CreateQuestionBankDTO{Name: "X"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestBankService_GetByID_Success(t *testing.T) {
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(),
		OrgID:  &orgID,
	})
	bankRepo := &mockBankRepo{}
	bankID := uuid.New()
	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID, Name: "Physics"}, nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	bank, err := svc.GetByID(ctx, bankID)
	assert.NoError(t, err)
	assert.Equal(t, "Physics", bank.Name)
}

func TestBankService_Update_AsStaff_Success(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankRepo := &mockBankRepo{}
	bankID := uuid.New()
	bank := &domain.QuestionBank{ID: bankID, OrganizationID: orgID, Name: "Old"}

	bankRepo.On("FindByID", ctx, bankID).Return(bank, nil)
	bankRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuestionBank")).Return(nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	newName := "New"
	updated, err := svc.Update(ctx, bankID, domain.UpdateQuestionBankDTO{Name: &newName})
	assert.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
}

func TestBankService_Delete_CrossOrg_Forbidden(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	ctx := staffCtx(orgA)
	bankRepo := &mockBankRepo{}
	bankID := uuid.New()
	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgB}, nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	err := svc.Delete(ctx, bankID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestBankService_CreateQuestion_Success(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankRepo := &mockBankRepo{}
	qRepo := &mockQuestionRepo{}
	bankID := uuid.New()

	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID}, nil)
	qRepo.On("Create", ctx, mock.AnythingOfType("*domain.Question")).Return(nil)

	svc := newTestBankService(bankRepo, qRepo)
	q, err := svc.CreateQuestion(ctx, bankID, domain.CreateQuestionDTO{
		Text: "What is 2+2?",
		Type: domain.QuestionTypeChoice,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "4", Score: 1},
			{ID: "b", Value: "3", Score: 0},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, bankID, q.BankID)
	assert.Equal(t, orgID, q.OrganizationID)
}

func TestBankService_CreateQuestion_NegativeMarkValidated(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankRepo := &mockBankRepo{}
	qRepo := &mockQuestionRepo{}
	bankID := uuid.New()
	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID}, nil)

	svc := newTestBankService(bankRepo, qRepo)
	_, err := svc.CreateQuestion(ctx, bankID, domain.CreateQuestionDTO{
		Text: "Q", Type: domain.QuestionTypeChoice,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "x", Score: 1},
			{ID: "b", Value: "y", Score: 0},
		},
		NegativeMarkMode: domain.NegativeMarkPerWrong,
		NegativeValue:    0,
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestBankService_CreateQuestion_NegativeMarkPersisted(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankRepo := &mockBankRepo{}
	qRepo := &mockQuestionRepo{}
	bankID := uuid.New()
	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID}, nil)
	qRepo.On("Create", ctx, mock.AnythingOfType("*domain.Question")).Return(nil)

	svc := newTestBankService(bankRepo, qRepo)
	q, err := svc.CreateQuestion(ctx, bankID, domain.CreateQuestionDTO{
		Text: "Q", Type: domain.QuestionTypeChoice,
		Options: []domain.QuestionOption{
			{ID: "a", Value: "x", Score: 1},
			{ID: "b", Value: "y", Score: 0},
		},
		NegativeMarkMode: domain.NegativeMarkAccumulative,
		WrongsPerPoint:   3,
		NegativeValue:    9, // should be zeroed by normalize
	})
	assert.NoError(t, err)
	assert.Equal(t, domain.NegativeMarkAccumulative, q.NegativeMarkMode)
	assert.Equal(t, 3, q.WrongsPerPoint)
	assert.Equal(t, float64(0), q.NegativeValue)
}

func TestBankService_CreateQuestion_OptionImageClearedForNonChoice(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankRepo := &mockBankRepo{}
	qRepo := &mockQuestionRepo{}
	bankID := uuid.New()
	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID}, nil)
	qRepo.On("Create", ctx, mock.AnythingOfType("*domain.Question")).Return(nil)

	imgID := uuid.New()
	svc := newTestBankService(bankRepo, qRepo)
	q, err := svc.CreateQuestion(ctx, bankID, domain.CreateQuestionDTO{
		Text: "Q", Type: domain.QuestionTypeShortAnswer,
		Options: []domain.QuestionOption{{ID: "a", Value: "ans", Score: 1, ImageMediaID: &imgID}},
	})
	assert.NoError(t, err)
	assert.Nil(t, q.Options[0].ImageMediaID)
	assert.Equal(t, domain.NegativeMarkNone, q.NegativeMarkMode)
}

func TestBankService_AdminHardDelete_NonAdmin_Forbidden(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	svc := newTestBankService(&mockBankRepo{}, &mockQuestionRepo{})
	err := svc.AdminHardDelete(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestBankService_AdminHardDelete_Success(t *testing.T) {
	ctx := adminCtx()
	bankRepo := &mockBankRepo{}
	bankID := uuid.New()
	bankRepo.On("HardDelete", ctx, bankID).Return(nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	err := svc.AdminHardDelete(ctx, bankID)
	assert.NoError(t, err)
	bankRepo.AssertExpectations(t)
}
