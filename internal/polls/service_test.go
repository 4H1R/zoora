package polls_test

import (
	"context"
	"log/slog"
	"testing"

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

func newPollService(repo *pollRepoMock, answers *pollAnswerRepoMock) domain.PollService {
	return polls.NewService(repo, answers, slog.Default())
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

func TestPollUpdateAllowsOwnerAdminOrUpdateAnyOnly(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()
	pollID := uuid.New()
	original := &domain.Poll{ID: pollID, UserID: ownerID, Name: "Old", AllowedAnswersCount: 1}

	repo := &pollRepoMock{}
	svc := newPollService(repo, &pollAnswerRepoMock{})

	repo.On("FindByID", mock.Anything, pollID).Return(original, nil).Times(3)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Poll) bool {
		return p.Name == "New" && p.AllowedAnswersCount == 2
	})).Return(nil).Twice()

	newName := "New"
	count := 2
	_, err := svc.Update(pollCaller(context.Background(), otherID), pollID, domain.UpdatePollDTO{Name: &newName})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	updated, err := svc.Update(pollCaller(context.Background(), ownerID), pollID, domain.UpdatePollDTO{
		Name:                &newName,
		AllowedAnswersCount: &count,
	})
	assert.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, 2, updated.AllowedAnswersCount)

	_, err = svc.Update(pollCaller(context.Background(), otherID, string(domain.PermPollsUpdateAny)), pollID, domain.UpdatePollDTO{Name: &newName})
	assert.NoError(t, err)
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
