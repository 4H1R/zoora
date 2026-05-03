package practices_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/practices"
)

// --- Mock repositories ---

type mockRoomRepo struct{ mock.Mock }

func (m *mockRoomRepo) Create(ctx context.Context, room *domain.PracticeRoom) error {
	return m.Called(ctx, room).Error(0)
}
func (m *mockRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeRoom), a.Error(1)
}
func (m *mockRoomRepo) Update(ctx context.Context, room *domain.PracticeRoom) error {
	return m.Called(ctx, room).Error(0)
}
func (m *mockRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) List(ctx context.Context, scope domain.PracticeRoomListScope, q domain.ListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	a := m.Called(ctx, scope, q)
	rs, _ := a.Get(0).([]domain.PracticeRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockRoomRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeRoom), a.Error(1)
}
func (m *mockRoomRepo) AdminList(ctx context.Context, q domain.AdminListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	a := m.Called(ctx, q)
	rs, _ := a.Get(0).([]domain.PracticeRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}

type mockSubRepo struct{ mock.Mock }

func (m *mockSubRepo) Create(ctx context.Context, sub *domain.PracticeSubmission) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *mockSubRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSubmission, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeSubmission), a.Error(1)
}
func (m *mockSubRepo) Update(ctx context.Context, sub *domain.PracticeSubmission) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *mockSubRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSubRepo) FindByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*domain.PracticeSubmission, error) {
	a := m.Called(ctx, roomID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeSubmission), a.Error(1)
}
func (m *mockSubRepo) ListByRoom(ctx context.Context, roomID uuid.UUID, p domain.ListParams) ([]domain.PracticeSubmission, int64, error) {
	a := m.Called(ctx, roomID, p)
	ss, _ := a.Get(0).([]domain.PracticeSubmission)
	return ss, a.Get(1).(int64), a.Error(2)
}

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Create(ctx context.Context, s *domain.ClassSession) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockSessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ClassSession), a.Error(1)
}
func (m *mockSessionRepo) Update(ctx context.Context, s *domain.ClassSession) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSessionRepo) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, classID, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
}
func (m *mockSessionRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSessionRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ClassSession), a.Error(1)
}

type mockClassRepo struct{ mock.Mock }

func (m *mockClassRepo) Create(ctx context.Context, c *domain.Class) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockClassRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Class), a.Error(1)
}
func (m *mockClassRepo) Update(ctx context.Context, c *domain.Class) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockClassRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassRepo) List(ctx context.Context, scope domain.ClassListScope, p domain.ListParams) ([]domain.Class, int64, error) {
	a := m.Called(ctx, scope, p)
	cs, _ := a.Get(0).([]domain.Class)
	return cs, a.Get(1).(int64), a.Error(2)
}
func (m *mockClassRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Class), a.Error(1)
}
func (m *mockClassRepo) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
	a := m.Called(ctx, q)
	cs, _ := a.Get(0).([]domain.Class)
	return cs, a.Get(1).(int64), a.Error(2)
}

type mockMemberRepo struct{ mock.Mock }

func (m *mockMemberRepo) Create(ctx context.Context, mem *domain.ClassMember) error {
	return m.Called(ctx, mem).Error(0)
}
func (m *mockMemberRepo) Delete(ctx context.Context, classID, userID uuid.UUID) error {
	return m.Called(ctx, classID, userID).Error(0)
}
func (m *mockMemberRepo) Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error) {
	a := m.Called(ctx, classID, userID)
	return a.Bool(0), a.Error(1)
}
func (m *mockMemberRepo) CountByClass(ctx context.Context, classID uuid.UUID) (int64, error) {
	a := m.Called(ctx, classID)
	return a.Get(0).(int64), a.Error(1)
}
func (m *mockMemberRepo) ListByClass(ctx context.Context, classID uuid.UUID, p domain.ListParams) ([]domain.ClassMember, int64, error) {
	a := m.Called(ctx, classID, p)
	ms, _ := a.Get(0).([]domain.ClassMember)
	return ms, a.Get(1).(int64), a.Error(2)
}
func (m *mockMemberRepo) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.ClassMember, error) {
	a := m.Called(ctx, classID)
	ms, _ := a.Get(0).([]domain.ClassMember)
	return ms, a.Error(1)
}

// --- Helpers ---

func newTestService(t *testing.T) (domain.PracticeService, *mockRoomRepo, *mockSubRepo, *mockSessionRepo, *mockClassRepo, *mockMemberRepo) {
	t.Helper()
	roomRepo := &mockRoomRepo{}
	subRepo := &mockSubRepo{}
	sessionRepo := &mockSessionRepo{}
	classRepo := &mockClassRepo{}
	memberRepo := &mockMemberRepo{}
	svc := practices.NewService(roomRepo, subRepo, sessionRepo, classRepo, memberRepo, slog.Default())
	return svc, roomRepo, subRepo, sessionRepo, classRepo, memberRepo
}

func callerCtx(userID uuid.UUID, isAdmin bool, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		IsAdmin:     isAdmin,
		Permissions: perms,
	})
}

// --- CreateRoom tests ---

func TestCreateRoom_Success(t *testing.T) {
	svc, roomRepo, _, sessionRepo, classRepo, _ := newTestService(t)

	userID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	orgID := uuid.New()
	ctx := callerCtx(userID, false, "practices:create")

	sessionRepo.On("FindByID", ctx, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: userID}, nil)
	roomRepo.On("Create", ctx, mock.AnythingOfType("*domain.PracticeRoom")).Return(nil)

	room, err := svc.CreateRoom(ctx, domain.CreatePracticeRoomDTO{
		ClassSessionID: sessionID,
		Title:          "Homework 1",
		MaxScore:       100,
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(24 * time.Hour),
	})

	assert.NoError(t, err)
	assert.Equal(t, "Homework 1", room.Title)
	assert.Equal(t, classID, room.ClassID)
	assert.Equal(t, orgID, room.OrganizationID)
	roomRepo.AssertExpectations(t)
}

func TestCreateRoom_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	_, err := svc.CreateRoom(context.Background(), domain.CreatePracticeRoomDTO{
		Title: "HW", ClassSessionID: uuid.New(),
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateRoom_NotClassOwner_Forbidden(t *testing.T) {
	svc, _, _, sessionRepo, classRepo, _ := newTestService(t)

	callerID := uuid.New()
	ownerID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	ctx := callerCtx(callerID, false, "practices:create")

	sessionRepo.On("FindByID", ctx, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: ownerID}, nil)

	_, err := svc.CreateRoom(ctx, domain.CreatePracticeRoomDTO{
		ClassSessionID: sessionID, Title: "HW",
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateRoom_Admin_CanCreateForAnyClass(t *testing.T) {
	svc, roomRepo, _, sessionRepo, classRepo, _ := newTestService(t)

	adminID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	ownerID := uuid.New()
	ctx := callerCtx(adminID, true)

	sessionRepo.On("FindByID", ctx, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: ownerID, OrganizationID: uuid.New()}, nil)
	roomRepo.On("Create", ctx, mock.Anything).Return(nil)

	room, err := svc.CreateRoom(ctx, domain.CreatePracticeRoomDTO{
		ClassSessionID: sessionID, Title: "Admin HW",
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, adminID, room.UserID)
}

// --- GetRoom tests ---

func TestGetRoom_Owner_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(userID, false, "practices:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: userID}, nil)

	room, err := svc.GetRoom(ctx, roomID)
	assert.NoError(t, err)
	assert.Equal(t, roomID, room.ID)
}

func TestGetRoom_Member_Success(t *testing.T) {
	svc, roomRepo, _, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	ownerID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ctx := callerCtx(userID, false, "practices:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: ownerID, ClassID: classID}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(true, nil)

	room, err := svc.GetRoom(ctx, roomID)
	assert.NoError(t, err)
	assert.Equal(t, roomID, room.ID)
}

func TestGetRoom_NonMember_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	ownerID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ctx := callerCtx(userID, false, "practices:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: ownerID, ClassID: classID}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(false, nil)

	_, err := svc.GetRoom(ctx, roomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// --- UpdateRoom tests ---

func TestUpdateRoom_Owner_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(userID, false, "practices:update")

	room := &domain.PracticeRoom{ID: roomID, UserID: userID, Title: "Old"}
	roomRepo.On("FindByID", ctx, roomID).Return(room, nil)
	roomRepo.On("Update", ctx, mock.AnythingOfType("*domain.PracticeRoom")).Return(nil)

	newTitle := "New"
	updated, err := svc.UpdateRoom(ctx, roomID, domain.UpdatePracticeRoomDTO{Title: &newTitle})
	assert.NoError(t, err)
	assert.Equal(t, "New", updated.Title)
}

func TestUpdateRoom_NotOwner_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	callerID := uuid.New()
	ownerID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(callerID, false, "practices:update")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: ownerID}, nil)

	title := "New"
	_, err := svc.UpdateRoom(ctx, roomID, domain.UpdatePracticeRoomDTO{Title: &title})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// --- DeleteRoom tests ---

func TestDeleteRoom_Owner_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(userID, false, "practices:delete")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: userID}, nil)
	roomRepo.On("Delete", ctx, roomID).Return(nil)

	err := svc.DeleteRoom(ctx, roomID)
	assert.NoError(t, err)
}

// --- Submit tests ---

func TestSubmit_Member_InWindow_Success(t *testing.T) {
	svc, roomRepo, subRepo, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ctx := callerCtx(userID, false, "practices:submit")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{
			ID:        roomID,
			ClassID:   classID,
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now().Add(1 * time.Hour),
		}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(true, nil)
	subRepo.On("Create", ctx, mock.AnythingOfType("*domain.PracticeSubmission")).Return(nil)

	sub, err := svc.Submit(ctx, roomID, domain.CreatePracticeSubmissionDTO{Content: "my work"})
	assert.NoError(t, err)
	assert.Equal(t, userID, sub.UserID)
	assert.Equal(t, "my work", sub.Content)
}

func TestSubmit_OutsideWindow_ValidationError(t *testing.T) {
	svc, roomRepo, _, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ctx := callerCtx(userID, false, "practices:submit")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{
			ID:        roomID,
			ClassID:   classID,
			StartTime: time.Now().Add(1 * time.Hour),
			EndTime:   time.Now().Add(2 * time.Hour),
		}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(true, nil)

	_, err := svc.Submit(ctx, roomID, domain.CreatePracticeSubmissionDTO{Content: "late"})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestSubmit_AfterDeadline_ValidationError(t *testing.T) {
	svc, roomRepo, _, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ctx := callerCtx(userID, false, "practices:submit")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{
			ID:        roomID,
			ClassID:   classID,
			StartTime: time.Now().Add(-2 * time.Hour),
			EndTime:   time.Now().Add(-1 * time.Hour),
		}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(true, nil)

	_, err := svc.Submit(ctx, roomID, domain.CreatePracticeSubmissionDTO{Content: "past deadline"})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestSubmit_NonMember_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ownerID := uuid.New()
	ctx := callerCtx(userID, false, "practices:submit")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{
			ID:        roomID,
			ClassID:   classID,
			UserID:    ownerID,
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now().Add(1 * time.Hour),
		}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(false, nil)

	_, err := svc.Submit(ctx, roomID, domain.CreatePracticeSubmissionDTO{Content: "x"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// --- Grade tests ---

func TestGrade_RoomOwner_Success(t *testing.T) {
	svc, roomRepo, subRepo, _, _, _ := newTestService(t)

	teacherID := uuid.New()
	roomID := uuid.New()
	subID := uuid.New()
	ctx := callerCtx(teacherID, false, "practices:grade")

	subRepo.On("FindByID", ctx, subID).
		Return(&domain.PracticeSubmission{ID: subID, PracticeRoomID: roomID}, nil)
	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: teacherID, MaxScore: 100}, nil)
	subRepo.On("Update", ctx, mock.AnythingOfType("*domain.PracticeSubmission")).Return(nil)

	score := 85.0
	comment := "Well done"
	sub, err := svc.Grade(ctx, subID, domain.GradePracticeSubmissionDTO{
		Score:          &score,
		TeacherComment: &comment,
	})
	assert.NoError(t, err)
	assert.Equal(t, &score, sub.Score)
	assert.Equal(t, "Well done", sub.TeacherComment)
}

func TestGrade_ScoreExceedsMax_ValidationError(t *testing.T) {
	svc, roomRepo, subRepo, _, _, _ := newTestService(t)

	teacherID := uuid.New()
	roomID := uuid.New()
	subID := uuid.New()
	ctx := callerCtx(teacherID, false, "practices:grade")

	subRepo.On("FindByID", ctx, subID).
		Return(&domain.PracticeSubmission{ID: subID, PracticeRoomID: roomID}, nil)
	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: teacherID, MaxScore: 100}, nil)

	score := 150.0
	_, err := svc.Grade(ctx, subID, domain.GradePracticeSubmissionDTO{Score: &score})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestGrade_NotOwner_Forbidden(t *testing.T) {
	svc, roomRepo, subRepo, _, _, _ := newTestService(t)

	callerID := uuid.New()
	ownerID := uuid.New()
	roomID := uuid.New()
	subID := uuid.New()
	ctx := callerCtx(callerID, false, "practices:grade")

	subRepo.On("FindByID", ctx, subID).
		Return(&domain.PracticeSubmission{ID: subID, PracticeRoomID: roomID}, nil)
	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: ownerID}, nil)

	score := 50.0
	_, err := svc.Grade(ctx, subID, domain.GradePracticeSubmissionDTO{Score: &score})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// --- GetSubmission tests ---

func TestGetSubmission_OwnSubmission_Success(t *testing.T) {
	svc, _, subRepo, _, _, _ := newTestService(t)

	userID := uuid.New()
	subID := uuid.New()
	ctx := callerCtx(userID, false, "practices:view")

	subRepo.On("FindByID", ctx, subID).
		Return(&domain.PracticeSubmission{ID: subID, UserID: userID}, nil)

	sub, err := svc.GetSubmission(ctx, subID)
	assert.NoError(t, err)
	assert.Equal(t, subID, sub.ID)
}

func TestGetSubmission_OtherUser_RoomOwner_Success(t *testing.T) {
	svc, roomRepo, subRepo, _, _, _ := newTestService(t)

	teacherID := uuid.New()
	studentID := uuid.New()
	roomID := uuid.New()
	subID := uuid.New()
	ctx := callerCtx(teacherID, false, "practices:view")

	subRepo.On("FindByID", ctx, subID).
		Return(&domain.PracticeSubmission{ID: subID, UserID: studentID, PracticeRoomID: roomID}, nil)
	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: teacherID}, nil)

	sub, err := svc.GetSubmission(ctx, subID)
	assert.NoError(t, err)
	assert.Equal(t, subID, sub.ID)
}

func TestGetSubmission_OtherUser_NotOwner_Forbidden(t *testing.T) {
	svc, roomRepo, subRepo, _, _, _ := newTestService(t)

	callerID := uuid.New()
	studentID := uuid.New()
	ownerID := uuid.New()
	roomID := uuid.New()
	subID := uuid.New()
	ctx := callerCtx(callerID, false, "practices:view")

	subRepo.On("FindByID", ctx, subID).
		Return(&domain.PracticeSubmission{ID: subID, UserID: studentID, PracticeRoomID: roomID}, nil)
	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: ownerID}, nil)

	_, err := svc.GetSubmission(ctx, subID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// --- ListSubmissions tests ---

func TestListSubmissions_RoomOwner_Success(t *testing.T) {
	svc, roomRepo, subRepo, _, _, _ := newTestService(t)

	teacherID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(teacherID, false, "practices:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: teacherID}, nil)
	subRepo.On("ListByRoom", ctx, roomID, mock.AnythingOfType("domain.ListParams")).
		Return([]domain.PracticeSubmission{{ID: uuid.New()}}, int64(1), nil)

	subs, total, err := svc.ListSubmissions(ctx, roomID, domain.ListPracticeSubmissionsQuery{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, subs, 1)
}

func TestListSubmissions_NotOwner_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	callerID := uuid.New()
	ownerID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(callerID, false, "practices:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.PracticeRoom{ID: roomID, UserID: ownerID}, nil)

	_, _, err := svc.ListSubmissions(ctx, roomID, domain.ListPracticeSubmissionsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// --- Admin tests ---

func TestAdminList_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	ctx := callerCtx(uuid.New(), true)
	roomRepo.On("AdminList", ctx, mock.AnythingOfType("domain.AdminListPracticeRoomsQuery")).
		Return([]domain.PracticeRoom{{ID: uuid.New()}}, int64(1), nil)

	rooms, total, err := svc.AdminList(ctx, domain.AdminListPracticeRoomsQuery{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rooms, 1)
}

func TestAdminList_NonAdmin_Forbidden(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	ctx := callerCtx(uuid.New(), false)
	_, _, err := svc.AdminList(ctx, domain.AdminListPracticeRoomsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAdminHardDelete_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	roomID := uuid.New()
	ctx := callerCtx(uuid.New(), true)
	roomRepo.On("HardDelete", ctx, roomID).Return(nil)

	err := svc.AdminHardDelete(ctx, roomID)
	assert.NoError(t, err)
}

func TestAdminHardDelete_NonAdmin_Forbidden(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	ctx := callerCtx(uuid.New(), false)
	err := svc.AdminHardDelete(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}
