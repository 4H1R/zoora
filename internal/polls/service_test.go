package polls_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/polls"
)

type pollRepoMock struct{ mock.Mock }

func (m *pollRepoMock) Create(ctx context.Context, poll *domain.Poll) error {
	return m.Called(ctx, poll).Error(0)
}

func (m *pollRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*domain.Poll, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Poll), args.Error(1)
}

func (m *pollRepoMock) Update(ctx context.Context, poll *domain.Poll) error {
	return m.Called(ctx, poll).Error(0)
}

func (m *pollRepoMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *pollRepoMock) List(ctx context.Context, scope domain.PollListScope, p domain.ListParams) ([]domain.Poll, int64, error) {
	args := m.Called(ctx, scope, p)
	items, _ := args.Get(0).([]domain.Poll)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *pollRepoMock) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *pollRepoMock) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Poll, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Poll), args.Error(1)
}

func (m *pollRepoMock) AdminList(ctx context.Context, q domain.AdminListPollsQuery) ([]domain.Poll, int64, error) {
	args := m.Called(ctx, q)
	items, _ := args.Get(0).([]domain.Poll)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *pollRepoMock) CloseByModel(ctx context.Context, modelType string, modelID uuid.UUID) error {
	return m.Called(ctx, modelType, modelID).Error(0)
}

type pollAnswerRepoMock struct{ mock.Mock }

func (m *pollAnswerRepoMock) Create(ctx context.Context, answer *domain.PollAnswer) error {
	return m.Called(ctx, answer).Error(0)
}

func (m *pollAnswerRepoMock) FindByPollAndUser(ctx context.Context, pollID, userID uuid.UUID) ([]domain.PollAnswer, error) {
	args := m.Called(ctx, pollID, userID)
	items, _ := args.Get(0).([]domain.PollAnswer)
	return items, args.Error(1)
}

func (m *pollAnswerRepoMock) ListByPoll(ctx context.Context, pollID uuid.UUID, q domain.ListPollAnswersQuery) ([]domain.PollAnswer, int64, error) {
	args := m.Called(ctx, pollID, q)
	items, _ := args.Get(0).([]domain.PollAnswer)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *pollAnswerRepoMock) DeleteByPollAndUser(ctx context.Context, pollID, userID uuid.UUID) error {
	return m.Called(ctx, pollID, userID).Error(0)
}

func (m *pollAnswerRepoMock) CountByOption(ctx context.Context, pollID uuid.UUID) (map[string]int, int, error) {
	args := m.Called(ctx, pollID)
	counts, _ := args.Get(0).(map[string]int)
	return counts, args.Int(1), args.Error(2)
}

// authorizerMock is a controllable domain.ModelAuthorizer. Zero value allows
// everything (CanParticipate/CanModerate => true, nil error) so tests that don't
// care about authz keep their original behavior; set the fields to override.
type authorizerMock struct {
	participate    bool
	participateErr error
	moderate       bool
	moderateErr    error
}

func (a authorizerMock) CanParticipate(_ context.Context, _ domain.Caller, _ string, _ uuid.UUID) (bool, error) {
	return a.participate, a.participateErr
}

func (a authorizerMock) CanModerate(_ context.Context, _ domain.Caller, _ string, _ uuid.UUID) (bool, error) {
	return a.moderate, a.moderateErr
}

func allowAll() authorizerMock { return authorizerMock{participate: true, moderate: true} }

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

func newPollService(repo *pollRepoMock, answers *pollAnswerRepoMock) domain.PollService {
	return polls.NewService(repo, answers, allowAll(), fakeTransactor{}, &auditSpy{}, slog.Default())
}

func newPollServiceWithAuth(repo *pollRepoMock, answers *pollAnswerRepoMock, auth authorizerMock) domain.PollService {
	return polls.NewService(repo, answers, auth, fakeTransactor{}, &auditSpy{}, slog.Default())
}

func newPollServiceWithAudit(repo *pollRepoMock, answers *pollAnswerRepoMock, audit *auditSpy) domain.PollService {
	return polls.NewService(repo, answers, allowAll(), fakeTransactor{}, audit, slog.Default())
}

func pollCaller(ctx context.Context, userID uuid.UUID, perms ...string) context.Context {
	return domain.WithCaller(ctx, domain.Caller{UserID: userID, Permissions: perms})
}

func pollAdminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
}

func TestPollCreateRequiresCallerAndSetsOwner(t *testing.T) {
	repo := &pollRepoMock{}
	svc := newPollService(repo, &pollAnswerRepoMock{})

	_, err := svc.Create(context.Background(), domain.CreatePollDTO{})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	userID := uuid.New()
	modelID := uuid.New()
	ctx := pollCaller(context.Background(), userID)
	repo.On("Create", ctx, mock.MatchedBy(func(p *domain.Poll) bool {
		return p.UserID == userID &&
			p.ModelType == "class" &&
			p.ModelID == modelID &&
			p.Name == "Favorite?" &&
			p.AllowedAnswersCount == 1
	})).Return(nil)

	poll, err := svc.Create(ctx, domain.CreatePollDTO{
		ModelType:           "class",
		ModelID:             modelID,
		Name:                "Favorite?",
		AllowedAnswersCount: 1,
		Options:             []domain.PollOption{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
	})

	assert.NoError(t, err)
	assert.Equal(t, userID, poll.UserID)
	repo.AssertExpectations(t)
}

func TestPollCreateRecordsAudit(t *testing.T) {
	repo := &pollRepoMock{}
	audit := &auditSpy{}
	svc := newPollServiceWithAudit(repo, &pollAnswerRepoMock{}, audit)

	userID := uuid.New()
	modelID := uuid.New()
	ctx := pollCaller(context.Background(), userID)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Poll")).Return(nil)

	poll, err := svc.Create(ctx, domain.CreatePollDTO{
		ModelType:           "live_session",
		ModelID:             modelID,
		Name:                "Favorite topic?",
		AllowedAnswersCount: 1,
		Options:             []domain.PollOption{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
	})

	assert.NoError(t, err)
	assert.Len(t, audit.records, 1)
	assert.Equal(t, domain.AuditCreated, audit.records[0].Action)
	assert.Equal(t, domain.AuditTargetPoll, audit.records[0].TargetType)
	assert.Equal(t, "Favorite topic?", audit.records[0].TargetLabel)
	assert.NotNil(t, audit.records[0].TargetID)
	assert.Equal(t, poll.ID, *audit.records[0].TargetID)
	repo.AssertExpectations(t)
}

func TestPollUpdateRequiresModeration(t *testing.T) {
	pollID := uuid.New()
	original := &domain.Poll{ID: pollID, UserID: uuid.New(), ModelType: "class", ModelID: uuid.New(), Name: "Old", AllowedAnswersCount: 1}
	newName := "New"
	count := 2

	// Non-moderator of the owning model is rejected.
	repo := &pollRepoMock{}
	repo.On("FindByID", mock.Anything, pollID).Return(original, nil).Once()
	svc := newPollServiceWithAuth(repo, &pollAnswerRepoMock{}, authorizerMock{moderate: false})
	_, err := svc.Update(pollCaller(context.Background(), uuid.New()), pollID, domain.UpdatePollDTO{Name: &newName})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)

	// Moderator (owning teacher / update_any / admin) succeeds.
	repo2 := &pollRepoMock{}
	repo2.On("FindByID", mock.Anything, pollID).Return(original, nil).Once()
	repo2.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Poll) bool {
		return p.Name == "New" && p.AllowedAnswersCount == 2
	})).Return(nil).Once()
	svc2 := newPollServiceWithAuth(repo2, &pollAnswerRepoMock{}, authorizerMock{moderate: true})
	updated, err := svc2.Update(pollCaller(context.Background(), uuid.New()), pollID, domain.UpdatePollDTO{
		Name:                &newName,
		AllowedAnswersCount: &count,
	})
	assert.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, 2, updated.AllowedAnswersCount)
	repo2.AssertExpectations(t)
}

func TestPollDeleteRequiresModeration(t *testing.T) {
	pollID := uuid.New()
	poll := &domain.Poll{ID: pollID, UserID: uuid.New(), ModelType: "class", ModelID: uuid.New()}

	// Non-moderator is rejected and no delete happens.
	repo := &pollRepoMock{}
	repo.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	svc := newPollServiceWithAuth(repo, &pollAnswerRepoMock{}, authorizerMock{moderate: false})
	err := svc.Delete(pollCaller(context.Background(), uuid.New()), pollID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)

	// Moderator succeeds.
	repo2 := &pollRepoMock{}
	repo2.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	repo2.On("Delete", mock.Anything, pollID).Return(nil).Once()
	svc2 := newPollServiceWithAuth(repo2, &pollAnswerRepoMock{}, authorizerMock{moderate: true})
	assert.NoError(t, svc2.Delete(pollCaller(context.Background(), uuid.New()), pollID))
	repo2.AssertExpectations(t)
}

func TestPollCreateRequiresModeration(t *testing.T) {
	repo := &pollRepoMock{}
	svc := newPollServiceWithAuth(repo, &pollAnswerRepoMock{}, authorizerMock{moderate: false})

	_, err := svc.Create(pollCaller(context.Background(), uuid.New()), domain.CreatePollDTO{
		ModelType:           "class",
		ModelID:             uuid.New(),
		Name:                "Favorite?",
		AllowedAnswersCount: 1,
		Options:             []domain.PollOption{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestPollGetByIDRequiresParticipation(t *testing.T) {
	pollID := uuid.New()
	poll := &domain.Poll{ID: pollID, ModelType: "class", ModelID: uuid.New()}

	repo := &pollRepoMock{}
	repo.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	svc := newPollServiceWithAuth(repo, &pollAnswerRepoMock{}, authorizerMock{participate: false})
	_, err := svc.GetByID(pollCaller(context.Background(), uuid.New()), pollID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	repo2 := &pollRepoMock{}
	repo2.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	svc2 := newPollServiceWithAuth(repo2, &pollAnswerRepoMock{}, authorizerMock{participate: true})
	got, err := svc2.GetByID(pollCaller(context.Background(), uuid.New()), pollID)
	assert.NoError(t, err)
	assert.Equal(t, pollID, got.ID)
}

func TestPollAnswerRejectedForNonParticipant(t *testing.T) {
	pollID := uuid.New()
	poll := &domain.Poll{ID: pollID, ModelType: "class", ModelID: uuid.New(), AllowedAnswersCount: 1, Options: []domain.PollOption{{Label: "A", Value: "a"}}}

	repo := &pollRepoMock{}
	answers := &pollAnswerRepoMock{}
	repo.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	svc := newPollServiceWithAuth(repo, answers, authorizerMock{participate: false})

	created, err := svc.Answer(pollCaller(context.Background(), uuid.New()), pollID, domain.AnswerPollDTO{Options: []string{"a"}})
	assert.Nil(t, created)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	answers.AssertNotCalled(t, "DeleteByPollAndUser")
	answers.AssertNotCalled(t, "Create")
}

func TestPollResultsRequiresParticipation(t *testing.T) {
	pollID := uuid.New()
	poll := &domain.Poll{ID: pollID, ModelType: "class", ModelID: uuid.New()}

	repo := &pollRepoMock{}
	answers := &pollAnswerRepoMock{}
	repo.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	svc := newPollServiceWithAuth(repo, answers, authorizerMock{participate: false})
	_, err := svc.Results(pollCaller(context.Background(), uuid.New()), pollID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	answers.AssertNotCalled(t, "CountByOption")
}

func TestPollListAnswersRequiresModeration(t *testing.T) {
	pollID := uuid.New()
	poll := &domain.Poll{ID: pollID, ModelType: "class", ModelID: uuid.New()}

	// Participant but not moderator: rejected (voter identity is a privacy boundary).
	repo := &pollRepoMock{}
	answers := &pollAnswerRepoMock{}
	repo.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	svc := newPollServiceWithAuth(repo, answers, authorizerMock{participate: true, moderate: false})
	_, _, err := svc.ListAnswers(pollCaller(context.Background(), uuid.New()), pollID, domain.ListPollAnswersQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	answers.AssertNotCalled(t, "ListByPoll")

	// Moderator: allowed.
	repo2 := &pollRepoMock{}
	answers2 := &pollAnswerRepoMock{}
	repo2.On("FindByID", mock.Anything, pollID).Return(poll, nil).Once()
	answers2.On("ListByPoll", mock.Anything, pollID, mock.Anything).Return([]domain.PollAnswer{{ID: uuid.New()}}, int64(1), nil).Once()
	svc2 := newPollServiceWithAuth(repo2, answers2, authorizerMock{moderate: true})
	items, total, err := svc2.ListAnswers(pollCaller(context.Background(), uuid.New()), pollID, domain.ListPollAnswersQuery{})
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(1), total)
	answers2.AssertExpectations(t)
}

func TestPollListRejectsUnscopedNonAdminAndUnauthorizedModel(t *testing.T) {
	modelID := uuid.New()
	modelType := "class"

	// Non-admin without a model filter is refused (would otherwise hit the
	// cross-org empty scope).
	repo := &pollRepoMock{}
	svc := newPollServiceWithAuth(repo, &pollAnswerRepoMock{}, allowAll())
	_, _, err := svc.List(pollCaller(context.Background(), uuid.New()), domain.ListPollsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertNotCalled(t, "List")

	// Non-admin filtering a model they cannot access is refused.
	repo2 := &pollRepoMock{}
	svc2 := newPollServiceWithAuth(repo2, &pollAnswerRepoMock{}, authorizerMock{participate: false})
	_, _, err = svc2.List(pollCaller(context.Background(), uuid.New()), domain.ListPollsQuery{ModelType: &modelType, ModelID: &modelID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo2.AssertNotCalled(t, "List")

	// Admin lists without a model filter across orgs.
	repo3 := &pollRepoMock{}
	repo3.On("List", mock.Anything, mock.MatchedBy(func(scope domain.PollListScope) bool {
		return scope.AllOrgs
	}), mock.Anything).Return([]domain.Poll{{ID: uuid.New()}}, int64(1), nil).Once()
	svc3 := newPollServiceWithAuth(repo3, &pollAnswerRepoMock{}, authorizerMock{})
	items, total, err := svc3.List(pollAdminCtx(), domain.ListPollsQuery{})
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(1), total)
	repo3.AssertExpectations(t)
}

func TestPollListScopesByCallerAndIncludeDeletedPrivilege(t *testing.T) {
	ownerID := uuid.New()
	modelID := uuid.New()
	modelType := "class"
	params := domain.ListParams{Page: 2, PageSize: 5}

	tests := []struct {
		name      string
		ctx       context.Context
		wantScope func(domain.PollListScope) bool
	}{
		{
			name: "owner sees own non-deleted polls",
			ctx:  pollCaller(context.Background(), ownerID),
			wantScope: func(scope domain.PollListScope) bool {
				return !scope.AllOrgs && scope.OwnerID != nil && *scope.OwnerID == ownerID && !scope.IncludeDeleted
			},
		},
		{
			name: "update_any can include deleted without owner scope",
			ctx:  pollCaller(context.Background(), ownerID, string(domain.PermPollsUpdateAny)),
			wantScope: func(scope domain.PollListScope) bool {
				return !scope.AllOrgs && scope.OwnerID == nil && scope.IncludeDeleted
			},
		},
		{
			name: "admin can include deleted across orgs",
			ctx:  pollAdminCtx(),
			wantScope: func(scope domain.PollListScope) bool {
				return scope.AllOrgs && scope.OwnerID == nil && scope.IncludeDeleted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &pollRepoMock{}
			svc := newPollService(repo, &pollAnswerRepoMock{})
			repo.On("List", tt.ctx, mock.MatchedBy(func(scope domain.PollListScope) bool {
				return tt.wantScope(scope) &&
					scope.ModelType != nil && *scope.ModelType == modelType &&
					scope.ModelID != nil && *scope.ModelID == modelID
			}), params).Return([]domain.Poll{{ID: uuid.New()}}, int64(1), nil)

			items, total, err := svc.List(tt.ctx, domain.ListPollsQuery{
				ModelType:      &modelType,
				ModelID:        &modelID,
				IncludeDeleted: true,
				ListParams:     params,
			})

			assert.NoError(t, err)
			assert.Len(t, items, 1)
			assert.Equal(t, int64(1), total)
			repo.AssertExpectations(t)
		})
	}
}

func TestPollAnswerValidatesSelectionBeforeReplacingExistingAnswers(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()
	ctx := pollCaller(context.Background(), userID)
	poll := &domain.Poll{
		ID:                  pollID,
		AllowedAnswersCount: 1,
		Options:             []domain.PollOption{{Label: "Yes", Value: "yes"}, {Label: "No", Value: "no"}},
	}

	tests := []struct {
		name string
		dto  domain.AnswerPollDTO
	}{
		{name: "too many options", dto: domain.AnswerPollDTO{Options: []string{"yes", "no"}}},
		{name: "invalid option", dto: domain.AnswerPollDTO{Options: []string{"maybe"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &pollRepoMock{}
			answers := &pollAnswerRepoMock{}
			svc := newPollService(repo, answers)
			repo.On("FindByID", ctx, pollID).Return(poll, nil)

			created, err := svc.Answer(ctx, pollID, tt.dto)

			assert.Nil(t, created)
			assert.ErrorIs(t, err, domain.ErrValidation)
			answers.AssertNotCalled(t, "DeleteByPollAndUser")
			answers.AssertNotCalled(t, "Create")
		})
	}
}

func TestPollAnswerReplacesPreviousAnswersAndCreatesOnePerOption(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()
	ctx := pollCaller(context.Background(), userID)
	repo := &pollRepoMock{}
	answers := &pollAnswerRepoMock{}
	svc := newPollService(repo, answers)

	repo.On("FindByID", ctx, pollID).Return(&domain.Poll{
		ID:                  pollID,
		AllowedAnswersCount: 2,
		Options:             []domain.PollOption{{Label: "A", Value: "a"}, {Label: "B", Value: "b"}},
	}, nil)
	answers.On("DeleteByPollAndUser", ctx, pollID, userID).Return(nil).Once()
	answers.On("Create", ctx, mock.MatchedBy(func(a *domain.PollAnswer) bool {
		return a.UserID == userID && a.PollID == pollID && (a.Option == "a" || a.Option == "b")
	})).Return(nil).Twice()

	created, err := svc.Answer(ctx, pollID, domain.AnswerPollDTO{Options: []string{"a", "b"}})

	assert.NoError(t, err)
	assert.Len(t, created, 2)
	answers.AssertExpectations(t)
}

func TestPollAnswerRejectedWhenPollClosed(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()
	ctx := pollCaller(context.Background(), userID)
	repo := &pollRepoMock{}
	answers := &pollAnswerRepoMock{}
	svc := newPollService(repo, answers)

	closedAt := time.Unix(1700000000, 0)
	repo.On("FindByID", ctx, pollID).Return(&domain.Poll{
		ID:                  pollID,
		AllowedAnswersCount: 1,
		Options:             []domain.PollOption{{Label: "A", Value: "a"}},
		ClosedAt:            &closedAt,
	}, nil)

	created, err := svc.Answer(ctx, pollID, domain.AnswerPollDTO{Options: []string{"a"}})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, domain.ErrPollClosed)
	// No answer mutation should occur once the poll is closed.
	answers.AssertNotCalled(t, "DeleteByPollAndUser", mock.Anything, mock.Anything, mock.Anything)
	answers.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestPollCloseByModelDelegatesToRepo(t *testing.T) {
	repo := &pollRepoMock{}
	svc := newPollService(repo, &pollAnswerRepoMock{})
	modelID := uuid.New()

	repo.On("CloseByModel", mock.Anything, "live_session", modelID).Return(nil).Once()

	assert.NoError(t, svc.CloseByModel(context.Background(), "live_session", modelID))
	repo.AssertExpectations(t)
}

func TestPollAdminMethodsRequireAdminAndDefaultPagination(t *testing.T) {
	repo := &pollRepoMock{}
	svc := newPollService(repo, &pollAnswerRepoMock{})

	_, _, err := svc.AdminList(pollCaller(context.Background(), uuid.New()), domain.AdminListPollsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.ErrorIs(t, svc.AdminHardDelete(context.Background(), uuid.New()), domain.ErrForbidden)

	ctx := pollAdminCtx()
	repo.On("AdminList", ctx, mock.MatchedBy(func(q domain.AdminListPollsQuery) bool {
		return q.ListParams.Page == 1 && q.ListParams.PageSize == domain.DefaultPageSize
	})).Return([]domain.Poll{}, int64(0), nil)
	_, total, err := svc.AdminList(ctx, domain.AdminListPollsQuery{ListParams: domain.ListParams{Page: -1}})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)

	pollID := uuid.New()
	repo.On("HardDelete", ctx, pollID).Return(nil)
	assert.NoError(t, svc.AdminHardDelete(ctx, pollID))
	repo.AssertExpectations(t)
}
