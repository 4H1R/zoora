package attendance_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/attendance"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/authz"
)

// --- mocks for the live auto-mark dependencies ---

type mLiveRoomRepo struct{ mock.Mock }

func (m *mLiveRoomRepo) Create(ctx context.Context, r *domain.LiveRoom) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mLiveRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveRoom), a.Error(1)
}
func (m *mLiveRoomRepo) Update(ctx context.Context, r *domain.LiveRoom) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mLiveRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mLiveRoomRepo) List(ctx context.Context, scope domain.LiveRoomListScope, p domain.ListParams) ([]domain.LiveRoom, int64, error) {
	a := m.Called(ctx, scope, p)
	res, _ := a.Get(0).([]domain.LiveRoom)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mLiveRoomRepo) FindActiveRoomsWithStaleHost(ctx context.Context, d time.Duration) ([]domain.LiveRoom, error) {
	a := m.Called(ctx, d)
	res, _ := a.Get(0).([]domain.LiveRoom)
	return res, a.Error(1)
}
func (m *mLiveRoomRepo) ListByClassSession(ctx context.Context, sessionID uuid.UUID) ([]domain.LiveRoom, error) {
	a := m.Called(ctx, sessionID)
	res, _ := a.Get(0).([]domain.LiveRoom)
	return res, a.Error(1)
}
func (m *mLiveRoomRepo) FindByLiveKitRoomName(ctx context.Context, name string) (*domain.LiveRoom, error) {
	a := m.Called(ctx, name)
	res, _ := a.Get(0).(*domain.LiveRoom)
	return res, a.Error(1)
}
func (m *mLiveRoomRepo) AdminList(ctx context.Context, q domain.AdminListLiveRoomsQuery) ([]domain.LiveRoom, int64, error) {
	a := m.Called(ctx, q)
	res, _ := a.Get(0).([]domain.LiveRoom)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mLiveRoomRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mLiveRoomRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveRoom), a.Error(1)
}

type mParticipantRepo struct{ mock.Mock }

func (m *mParticipantRepo) Create(ctx context.Context, p *domain.LiveParticipant) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mParticipantRepo) FindActiveByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveParticipant), a.Error(1)
}
func (m *mParticipantRepo) Update(ctx context.Context, p *domain.LiveParticipant) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mParticipantRepo) ListByRoom(ctx context.Context, roomID uuid.UUID, q domain.ListLiveParticipantsQuery) ([]domain.LiveParticipant, int64, error) {
	a := m.Called(ctx, roomID, q)
	res, _ := a.Get(0).([]domain.LiveParticipant)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mParticipantRepo) ListAllByRoom(ctx context.Context, roomID uuid.UUID) ([]domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID)
	res, _ := a.Get(0).([]domain.LiveParticipant)
	return res, a.Error(1)
}
func (m *mParticipantRepo) GetActiveParticipant(ctx context.Context, roomID uuid.UUID, identity string) (*domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID, identity)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveParticipant), a.Error(1)
}
func (m *mParticipantRepo) UpdateParticipantRole(ctx context.Context, roomID uuid.UUID, identity string, role domain.ParticipantRole) error {
	return m.Called(ctx, roomID, identity, role).Error(0)
}
func (m *mParticipantRepo) SetHandRaised(ctx context.Context, roomID uuid.UUID, identity string, raised bool) error {
	return m.Called(ctx, roomID, identity, raised).Error(0)
}
func (m *mParticipantRepo) MarkAllLeft(ctx context.Context, roomID uuid.UUID, leftAt time.Time) error {
	return m.Called(ctx, roomID, leftAt).Error(0)
}

type fakeSettingsProvider struct{ percent int }

func (f fakeSettingsProvider) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationSettings, error) {
	return &domain.OrganizationSettings{OrganizationID: orgID, AttendancePresentThresholdPercent: f.percent}, nil
}

func liveRoom(sessionID uuid.UUID, durationSeconds int) domain.LiveRoom {
	start := time.Unix(1_700_000_000, 0)
	end := start.Add(time.Duration(durationSeconds) * time.Second)
	return domain.LiveRoom{
		ID:              uuid.New(),
		ClassSessionID:  sessionID,
		ActualStartTime: &start,
		ActualEndTime:   &end,
	}
}

func newAutoMarkSvc(repo domain.AttendanceRepository, classes domain.ClassRepository, sessions domain.ClassSessionRepository, members domain.ClassMemberRepository, rooms domain.LiveRoomRepository, parts domain.LiveParticipantRepository, percent int) domain.AttendanceService {
	return attendance.NewService(repo, classes, sessions, members, rooms, parts, nil, nil, fakeSettingsProvider{percent: percent}, authz.NewResolver(nil), slog.Default())
}

func TestAutoMarkSessionLive_MarksPresentAndAbsent(t *testing.T) {
	classID := uuid.New()
	sessionID := uuid.New()
	orgID := uuid.New()
	userA := uuid.New() // 80% -> present
	userB := uuid.New() // 70% -> absent
	userC := uuid.New() // never joined -> absent

	repo := &mAttRepo{}
	classes := &mClassRepo{}
	sessions := &mSessRepo{}
	members := &mMemberRepo{}
	rooms := &mLiveRoomRepo{}
	parts := &mParticipantRepo{}

	room := liveRoom(sessionID, 1000)
	classes.On("FindByID", mock.Anything, classID).Return(&domain.Class{ID: classID, OrganizationID: orgID}, nil)
	sessions.On("FindByID", mock.Anything, sessionID).Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	rooms.On("ListByClassSession", mock.Anything, sessionID).Return([]domain.LiveRoom{room}, nil)
	parts.On("ListAllByRoom", mock.Anything, room.ID).Return([]domain.LiveParticipant{
		{UserID: userA, TotalDurationSeconds: 800},
		{UserID: userB, TotalDurationSeconds: 700},
	}, nil)
	members.On("ListAllByClass", mock.Anything, classID).Return([]domain.ClassMember{
		{ClassID: classID, UserID: userA},
		{ClassID: classID, UserID: userB},
		{ClassID: classID, UserID: userC},
	}, nil)
	repo.On("FindBySessionAndUser", mock.Anything, sessionID, mock.Anything).Return(nil, domain.ErrNotFound)

	created := map[uuid.UUID]domain.AttendanceStatus{}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Attendance")).
		Run(func(args mock.Arguments) {
			a := args.Get(1).(*domain.Attendance)
			created[a.UserID] = a.Status
			require.True(t, a.IsAutoMarked)
		}).Return(nil)

	svc := newAutoMarkSvc(repo, classes, sessions, members, rooms, parts, 75)
	res, err := svc.AutoMarkSessionLive(context.Background(), classID, sessionID)

	require.NoError(t, err)
	require.Equal(t, 3, res.Marked)
	require.Equal(t, domain.AttendanceStatusPresent, created[userA])
	require.Equal(t, domain.AttendanceStatusAbsent, created[userB])
	require.Equal(t, domain.AttendanceStatusAbsent, created[userC])
}

func TestAutoMarkSessionLive_OverwritesAutoPreservesManual(t *testing.T) {
	classID := uuid.New()
	sessionID := uuid.New()
	orgID := uuid.New()
	userA := uuid.New() // present, had an auto absent row -> updated to present
	userB := uuid.New() // present, had a manual excused row -> preserved

	repo := &mAttRepo{}
	classes := &mClassRepo{}
	sessions := &mSessRepo{}
	members := &mMemberRepo{}
	rooms := &mLiveRoomRepo{}
	parts := &mParticipantRepo{}

	room := liveRoom(sessionID, 1000)
	classes.On("FindByID", mock.Anything, classID).Return(&domain.Class{ID: classID, OrganizationID: orgID}, nil)
	sessions.On("FindByID", mock.Anything, sessionID).Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	rooms.On("ListByClassSession", mock.Anything, sessionID).Return([]domain.LiveRoom{room}, nil)
	parts.On("ListAllByRoom", mock.Anything, room.ID).Return([]domain.LiveParticipant{
		{UserID: userA, TotalDurationSeconds: 900},
		{UserID: userB, TotalDurationSeconds: 900},
	}, nil)
	members.On("ListAllByClass", mock.Anything, classID).Return([]domain.ClassMember{
		{ClassID: classID, UserID: userA},
		{ClassID: classID, UserID: userB},
	}, nil)

	autoRow := &domain.Attendance{ID: uuid.New(), ClassSessionID: sessionID, UserID: userA, Status: domain.AttendanceStatusAbsent, IsAutoMarked: true}
	manualRow := &domain.Attendance{ID: uuid.New(), ClassSessionID: sessionID, UserID: userB, Status: domain.AttendanceStatusExcused, IsAutoMarked: false}
	repo.On("FindBySessionAndUser", mock.Anything, sessionID, userA).Return(autoRow, nil)
	repo.On("FindBySessionAndUser", mock.Anything, sessionID, userB).Return(manualRow, nil)

	var updated []*domain.Attendance
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Attendance")).
		Run(func(args mock.Arguments) { updated = append(updated, args.Get(1).(*domain.Attendance)) }).
		Return(nil)

	svc := newAutoMarkSvc(repo, classes, sessions, members, rooms, parts, 75)
	res, err := svc.AutoMarkSessionLive(context.Background(), classID, sessionID)

	require.NoError(t, err)
	require.Equal(t, 1, res.Marked)  // userA updated
	require.Equal(t, 1, res.Skipped) // userB preserved
	require.Len(t, updated, 1)
	require.Equal(t, userA, updated[0].UserID)
	require.Equal(t, domain.AttendanceStatusPresent, updated[0].Status)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAutoMarkSessionLive_ZeroDurationSkips(t *testing.T) {
	classID := uuid.New()
	sessionID := uuid.New()
	orgID := uuid.New()

	repo := &mAttRepo{}
	classes := &mClassRepo{}
	sessions := &mSessRepo{}
	members := &mMemberRepo{}
	rooms := &mLiveRoomRepo{}
	parts := &mParticipantRepo{}

	// Room with no actual times -> excluded -> total duration 0.
	classes.On("FindByID", mock.Anything, classID).Return(&domain.Class{ID: classID, OrganizationID: orgID}, nil)
	sessions.On("FindByID", mock.Anything, sessionID).Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	rooms.On("ListByClassSession", mock.Anything, sessionID).Return([]domain.LiveRoom{
		{ID: uuid.New(), ClassSessionID: sessionID},
	}, nil)

	svc := newAutoMarkSvc(repo, classes, sessions, members, rooms, parts, 75)
	res, err := svc.AutoMarkSessionLive(context.Background(), classID, sessionID)

	require.NoError(t, err)
	require.Equal(t, 0, res.Marked)
	require.Equal(t, 0, res.Skipped)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	members.AssertNotCalled(t, "ListAllByClass", mock.Anything, mock.Anything)
}
