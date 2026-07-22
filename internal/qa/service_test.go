package qa_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/qa"
)

// --- mocks ---

type repoMock struct{ mock.Mock }

func (m *repoMock) Create(ctx context.Context, q *domain.QAQuestion) error {
	return m.Called(ctx, q).Error(0)
}

func (m *repoMock) FindByID(ctx context.Context, id uuid.UUID) (*domain.QAQuestion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.QAQuestion), args.Error(1)
}

func (m *repoMock) Update(ctx context.Context, q *domain.QAQuestion) error {
	return m.Called(ctx, q).Error(0)
}

func (m *repoMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *repoMock) List(ctx context.Context, scope domain.QAListScope, p domain.ListParams) ([]domain.QAQuestionView, int64, error) {
	args := m.Called(ctx, scope, p)
	items, _ := args.Get(0).([]domain.QAQuestionView)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *repoMock) CountOpenByUser(ctx context.Context, modelType string, modelID, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, modelType, modelID, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *repoMock) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *repoMock) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.QAQuestion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.QAQuestion), args.Error(1)
}

func (m *repoMock) AdminList(ctx context.Context, q domain.AdminListQAQuestionsQuery) ([]domain.QAQuestion, int64, error) {
	args := m.Called(ctx, q)
	items, _ := args.Get(0).([]domain.QAQuestion)
	return items, args.Get(1).(int64), args.Error(2)
}

type voteMock struct{ mock.Mock }

func (m *voteMock) Create(ctx context.Context, v *domain.QAVote) error {
	return m.Called(ctx, v).Error(0)
}

func (m *voteMock) Delete(ctx context.Context, questionID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, questionID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *voteMock) Exists(ctx context.Context, questionID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, questionID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *voteMock) CountByQuestion(ctx context.Context, questionID uuid.UUID) (int64, error) {
	args := m.Called(ctx, questionID)
	return args.Get(0).(int64), args.Error(1)
}

type authzMock struct{ mock.Mock }

func (m *authzMock) CanParticipate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
	args := m.Called(ctx, caller, modelType, modelID)
	return args.Bool(0), args.Error(1)
}

func (m *authzMock) CanModerate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
	args := m.Called(ctx, caller, modelType, modelID)
	return args.Bool(0), args.Error(1)
}

// --- helpers ---

// fakeTransactor runs fn inline with no real DB — unit tests exercise the audit
// same-tx wiring without a database.
type fakeTransactor struct{}

func (fakeTransactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// auditSpy captures the records a service emits so tests can assert on them.
type auditSpy struct{ records []domain.AuditRecord }

func (a *auditSpy) Record(_ context.Context, r domain.AuditRecord) error {
	a.records = append(a.records, r)
	return nil
}

func (a *auditSpy) RecordDenied(_ context.Context, _ domain.AuditRecord) error { return nil }

func newSvc(repo *repoMock, votes *voteMock, authz *authzMock) domain.QAService {
	return qa.NewService(repo, votes, authz, fakeTransactor{}, &auditSpy{}, slog.Default(), nil)
}

func newSvcWithAudit(repo *repoMock, votes *voteMock, authz *authzMock, audit *auditSpy) domain.QAService {
	return qa.NewService(repo, votes, authz, fakeTransactor{}, audit, slog.Default(), nil)
}

func ctxWith(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: userID})
}

var modelID = uuid.New()

// --- tests ---

func TestAsk_ForbiddenWhenNotParticipant(t *testing.T) {
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(false, nil)

	_, err := newSvc(repo, votes, authz).Ask(ctxWith(uuid.New()), domain.CreateQAQuestionDTO{
		ModelType: domain.QAModelLiveSession, ModelID: modelID, Text: "hello?",
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAsk_BlockedAtOpenCap(t *testing.T) {
	userID := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)
	repo.On("CountOpenByUser", mock.Anything, domain.QAModelLiveSession, modelID, userID).
		Return(int64(domain.MaxOpenQuestionsPerUser), nil)

	_, err := newSvc(repo, votes, authz).Ask(ctxWith(userID), domain.CreateQAQuestionDTO{
		ModelType: domain.QAModelLiveSession, ModelID: modelID, Text: "another?",
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAsk_Success(t *testing.T) {
	userID := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)
	repo.On("CountOpenByUser", mock.Anything, domain.QAModelLiveSession, modelID, userID).Return(int64(0), nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.QAQuestion")).Return(nil)

	q, err := newSvc(repo, votes, authz).Ask(ctxWith(userID), domain.CreateQAQuestionDTO{
		ModelType: domain.QAModelLiveSession, ModelID: modelID, Text: "what is the deadline?",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.QAStatusOpen, q.Status)
	assert.Equal(t, userID, q.UserID)
}

func TestAsk_RecordsAudit(t *testing.T) {
	userID := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	audit := &auditSpy{}
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)
	repo.On("CountOpenByUser", mock.Anything, domain.QAModelLiveSession, modelID, userID).Return(int64(0), nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.QAQuestion")).Return(nil)

	q, err := newSvcWithAudit(repo, votes, authz, audit).Ask(ctxWith(userID), domain.CreateQAQuestionDTO{
		ModelType: domain.QAModelLiveSession, ModelID: modelID, Text: "what is the deadline?",
	})
	require.NoError(t, err)
	require.Len(t, audit.records, 1)
	assert.Equal(t, domain.AuditCreated, audit.records[0].Action)
	assert.Equal(t, domain.AuditTargetQA, audit.records[0].TargetType)
	assert.Equal(t, "what is the deadline?", audit.records[0].TargetLabel)
	require.NotNil(t, audit.records[0].TargetID)
	assert.Equal(t, q.ID, *audit.records[0].TargetID)
}

func TestToggleVote_RejectsSelfVote(t *testing.T) {
	userID := uuid.New()
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: userID, ModelType: domain.QAModelLiveSession, ModelID: modelID, Status: domain.QAStatusOpen,
	}, nil)
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)

	_, _, err := newSvc(repo, votes, authz).ToggleVote(ctxWith(userID), qid)
	assert.ErrorIs(t, err, domain.ErrValidation)
	votes.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	votes.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
}

func TestToggleVote_RejectsClosedQuestion(t *testing.T) {
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: uuid.New(), ModelType: domain.QAModelLiveSession, ModelID: modelID, Status: domain.QAStatusResolved,
	}, nil)
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)

	_, _, err := newSvc(repo, votes, authz).ToggleVote(ctxWith(uuid.New()), qid)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestToggleVote_AddsWhenAbsent(t *testing.T) {
	voter := uuid.New()
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: uuid.New(), ModelType: domain.QAModelLiveSession, ModelID: modelID, Status: domain.QAStatusOpen,
	}, nil)
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)
	votes.On("Delete", mock.Anything, qid, voter).Return(false, nil)
	votes.On("Create", mock.Anything, mock.AnythingOfType("*domain.QAVote")).Return(nil)
	votes.On("CountByQuestion", mock.Anything, qid).Return(int64(1), nil)

	voted, count, err := newSvc(repo, votes, authz).ToggleVote(ctxWith(voter), qid)
	require.NoError(t, err)
	assert.True(t, voted)
	assert.Equal(t, int64(1), count)
}

func TestToggleVote_RemovesWhenPresent(t *testing.T) {
	voter := uuid.New()
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: uuid.New(), ModelType: domain.QAModelLiveSession, ModelID: modelID, Status: domain.QAStatusOpen,
	}, nil)
	authz.On("CanParticipate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)
	votes.On("Delete", mock.Anything, qid, voter).Return(true, nil)
	votes.On("CountByQuestion", mock.Anything, qid).Return(int64(0), nil)

	voted, count, err := newSvc(repo, votes, authz).ToggleVote(ctxWith(voter), qid)
	require.NoError(t, err)
	assert.False(t, voted)
	assert.Equal(t, int64(0), count)
	votes.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestResolve_ForbiddenForNonModerator(t *testing.T) {
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: uuid.New(), ModelType: domain.QAModelLiveSession, ModelID: modelID, Status: domain.QAStatusOpen,
	}, nil)
	authz.On("CanModerate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(false, nil)

	_, err := newSvc(repo, votes, authz).Resolve(ctxWith(uuid.New()), qid)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestResolve_SetsStatusAndClosedFields(t *testing.T) {
	teacher := uuid.New()
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: uuid.New(), ModelType: domain.QAModelLiveSession, ModelID: modelID, Status: domain.QAStatusOpen,
	}, nil)
	authz.On("CanModerate", mock.Anything, mock.Anything, domain.QAModelLiveSession, modelID).Return(true, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.QAQuestion")).Return(nil)

	q, err := newSvc(repo, votes, authz).Resolve(ctxWith(teacher), qid)
	require.NoError(t, err)
	assert.Equal(t, domain.QAStatusResolved, q.Status)
	require.NotNil(t, q.ClosedBy)
	assert.Equal(t, teacher, *q.ClosedBy)
	require.NotNil(t, q.ClosedAt)
}

func TestUpdateText_ForbiddenWhenNotAuthor(t *testing.T) {
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: uuid.New(), Status: domain.QAStatusOpen,
	}, nil)

	_, err := newSvc(repo, votes, authz).UpdateText(ctxWith(uuid.New()), qid, domain.UpdateQAQuestionDTO{Text: "edited"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateText_BlockedWhenClosed(t *testing.T) {
	author := uuid.New()
	qid := uuid.New()
	repo, votes, authz := &repoMock{}, &voteMock{}, &authzMock{}
	repo.On("FindByID", mock.Anything, qid).Return(&domain.QAQuestion{
		ID: qid, UserID: author, Status: domain.QAStatusResolved,
	}, nil)

	_, err := newSvc(repo, votes, authz).UpdateText(ctxWith(author), qid, domain.UpdateQAQuestionDTO{Text: "edited"})
	assert.ErrorIs(t, err, domain.ErrValidation)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
