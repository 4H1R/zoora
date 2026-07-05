package offlines_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/offlines"
)

type mockRoomRepo struct{ mock.Mock }

func (m *mockRoomRepo) Create(ctx context.Context, room *domain.OfflineRoom) error {
	return m.Called(ctx, room).Error(0)
}
func (m *mockRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.OfflineRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.OfflineRoom), a.Error(1)
}
func (m *mockRoomRepo) Update(ctx context.Context, room *domain.OfflineRoom) error {
	return m.Called(ctx, room).Error(0)
}
func (m *mockRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) List(ctx context.Context, scope domain.OfflineRoomListScope, q domain.ListOfflineRoomsQuery) ([]domain.OfflineRoom, int64, error) {
	a := m.Called(ctx, scope, q)
	rs, _ := a.Get(0).([]domain.OfflineRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mockRoomRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.OfflineRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.OfflineRoom), a.Error(1)
}
func (m *mockRoomRepo) AdminList(ctx context.Context, q domain.AdminListOfflineRoomsQuery) ([]domain.OfflineRoom, int64, error) {
	a := m.Called(ctx, q)
	rs, _ := a.Get(0).([]domain.OfflineRoom)
	return rs, a.Get(1).(int64), a.Error(2)
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
func (m *mockSessionRepo) AdminList(ctx context.Context, q domain.AdminListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
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

type mockViewRepo struct{ mock.Mock }

func (m *mockViewRepo) Create(ctx context.Context, v *domain.OfflineRoomView) error {
	return m.Called(ctx, v).Error(0)
}
func (m *mockViewRepo) ListByRoom(ctx context.Context, roomID uuid.UUID) ([]domain.OfflineRoomView, error) {
	a := m.Called(ctx, roomID)
	vs, _ := a.Get(0).([]domain.OfflineRoomView)
	return vs, a.Error(1)
}
func (m *mockViewRepo) ListDistinctUsersByRoom(ctx context.Context, roomID uuid.UUID) ([]uuid.UUID, error) {
	a := m.Called(ctx, roomID)
	ids, _ := a.Get(0).([]uuid.UUID)
	return ids, a.Error(1)
}

func newTestService(t *testing.T) (domain.OfflineService, *mockRoomRepo, *mockViewRepo, *mockSessionRepo, *mockClassRepo, *mockMemberRepo) {
	t.Helper()
	roomRepo := &mockRoomRepo{}
	viewRepo := &mockViewRepo{}
	sessionRepo := &mockSessionRepo{}
	classRepo := &mockClassRepo{}
	memberRepo := &mockMemberRepo{}
	svc := offlines.NewService(roomRepo, viewRepo, sessionRepo, classRepo, memberRepo, slog.Default())
	return svc, roomRepo, viewRepo, sessionRepo, classRepo, memberRepo
}

func callerCtx(userID uuid.UUID, isAdmin bool, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		IsAdmin:     isAdmin,
		Permissions: perms,
		Ent:         domain.PlanCatalog[domain.PlanPro],
	})
}

// freeCallerCtx builds a non-admin caller on the Free plan (no gated features).
func freeCallerCtx(userID uuid.UUID, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		Permissions: perms,
		Ent:         domain.PlanCatalog[domain.PlanFree],
	})
}

func TestCreateRoom_Success(t *testing.T) {
	svc, roomRepo, _, sessionRepo, classRepo, _ := newTestService(t)

	userID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	orgID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:create")

	sessionRepo.On("FindByID", ctx, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: userID}, nil)
	roomRepo.On("Create", ctx, mock.AnythingOfType("*domain.OfflineRoom")).Return(nil)

	room, err := svc.CreateRoom(ctx, domain.CreateOfflineRoomDTO{
		ClassSessionID: sessionID,
		Title:          "Lecture 1 Recording",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Lecture 1 Recording", room.Title)
	assert.Equal(t, classID, room.ClassID)
	assert.Equal(t, orgID, room.OrganizationID)
	assert.Equal(t, userID, room.CreatorID)
	roomRepo.AssertExpectations(t)
}

func TestCreateRoom_FreePlanRejected(t *testing.T) {
	svc, roomRepo, _, sessionRepo, classRepo, _ := newTestService(t)

	userID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	orgID := uuid.New()
	ctx := freeCallerCtx(userID, "offlines:create")

	sessionRepo.On("FindByID", ctx, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: userID}, nil)

	_, err := svc.CreateRoom(ctx, domain.CreateOfflineRoomDTO{
		ClassSessionID: sessionID,
		Title:          "Lecture 1 Recording",
	})
	assert.ErrorIs(t, err, domain.ErrFeatureNotInPlan)
	roomRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateRoom_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	_, err := svc.CreateRoom(context.Background(), domain.CreateOfflineRoomDTO{
		Title:          "Lecture",
		ClassSessionID: uuid.New(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateRoom_NotClassOwner_Forbidden(t *testing.T) {
	svc, _, _, sessionRepo, classRepo, _ := newTestService(t)

	callerID := uuid.New()
	ownerID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	ctx := callerCtx(callerID, false, "offlines:create")

	sessionRepo.On("FindByID", ctx, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: ownerID}, nil)

	_, err := svc.CreateRoom(ctx, domain.CreateOfflineRoomDTO{
		ClassSessionID: sessionID, Title: "Lecture",
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

	room, err := svc.CreateRoom(ctx, domain.CreateOfflineRoomDTO{
		ClassSessionID: sessionID, Title: "Admin Lecture",
	})
	assert.NoError(t, err)
	assert.Equal(t, adminID, room.CreatorID)
}

func TestCreateRoom_WithCreateAny_CanCreateForAnyClass(t *testing.T) {
	svc, roomRepo, _, sessionRepo, classRepo, _ := newTestService(t)

	callerID := uuid.New()
	ownerID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	ctx := callerCtx(callerID, false, "offlines:create", "offlines:create_any")

	sessionRepo.On("FindByID", ctx, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: ownerID, OrganizationID: uuid.New()}, nil)
	roomRepo.On("Create", ctx, mock.Anything).Return(nil)

	room, err := svc.CreateRoom(ctx, domain.CreateOfflineRoomDTO{
		ClassSessionID: sessionID, Title: "Staff Lecture",
	})
	assert.NoError(t, err)
	assert.Equal(t, callerID, room.CreatorID)
}

func TestCreateRoom_SessionNotFound(t *testing.T) {
	svc, _, _, sessionRepo, _, _ := newTestService(t)

	sessionID := uuid.New()
	ctx := callerCtx(uuid.New(), false, "offlines:create")

	sessionRepo.On("FindByID", ctx, sessionID).
		Return((*domain.ClassSession)(nil), domain.ErrNotFound)

	_, err := svc.CreateRoom(ctx, domain.CreateOfflineRoomDTO{
		ClassSessionID: sessionID, Title: "Lecture",
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetRoom_Creator_Success(t *testing.T) {
	svc, roomRepo, viewRepo, _, _, _ := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.OfflineRoom{ID: roomID, CreatorID: userID}, nil)
	roomRepo.On("IncrementViewCount", ctx, roomID).Return(nil)
	viewRepo.On("Create", ctx, mock.AnythingOfType("*domain.OfflineRoomView")).Return(nil)

	room, err := svc.GetRoom(ctx, roomID)
	assert.NoError(t, err)
	assert.Equal(t, roomID, room.ID)
	roomRepo.AssertCalled(t, "IncrementViewCount", ctx, roomID)
}

func TestGetRoom_Member_Success(t *testing.T) {
	svc, roomRepo, viewRepo, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	creatorID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.OfflineRoom{ID: roomID, CreatorID: creatorID, ClassID: classID}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(true, nil)
	roomRepo.On("IncrementViewCount", ctx, roomID).Return(nil)
	viewRepo.On("Create", ctx, mock.AnythingOfType("*domain.OfflineRoomView")).Return(nil)

	room, err := svc.GetRoom(ctx, roomID)
	assert.NoError(t, err)
	assert.Equal(t, roomID, room.ID)
}

func TestGetRoom_NonMember_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, memberRepo := newTestService(t)

	userID := uuid.New()
	creatorID := uuid.New()
	roomID := uuid.New()
	classID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.OfflineRoom{ID: roomID, CreatorID: creatorID, ClassID: classID}, nil)
	memberRepo.On("Exists", ctx, classID, userID).Return(false, nil)

	_, err := svc.GetRoom(ctx, roomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetRoom_NotFound(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	roomID := uuid.New()
	ctx := callerCtx(uuid.New(), false, "offlines:view")

	roomRepo.On("FindByID", ctx, roomID).
		Return((*domain.OfflineRoom)(nil), domain.ErrNotFound)

	_, err := svc.GetRoom(ctx, roomID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpdateRoom_Creator_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:update")

	room := &domain.OfflineRoom{ID: roomID, CreatorID: userID, Title: "Old"}
	roomRepo.On("FindByID", ctx, roomID).Return(room, nil)
	roomRepo.On("Update", ctx, mock.AnythingOfType("*domain.OfflineRoom")).Return(nil)

	newTitle := "New"
	updated, err := svc.UpdateRoom(ctx, roomID, domain.UpdateOfflineRoomDTO{Title: &newTitle})
	assert.NoError(t, err)
	assert.Equal(t, "New", updated.Title)
}

func TestUpdateRoom_NotCreator_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	callerID := uuid.New()
	creatorID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(callerID, false, "offlines:update")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.OfflineRoom{ID: roomID, CreatorID: creatorID}, nil)

	title := "New"
	_, err := svc.UpdateRoom(ctx, roomID, domain.UpdateOfflineRoomDTO{Title: &title})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdateRoom_Admin_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	adminID := uuid.New()
	creatorID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(adminID, true)

	room := &domain.OfflineRoom{ID: roomID, CreatorID: creatorID, Title: "Old"}
	roomRepo.On("FindByID", ctx, roomID).Return(room, nil)
	roomRepo.On("Update", ctx, mock.AnythingOfType("*domain.OfflineRoom")).Return(nil)

	newTitle := "Admin Updated"
	updated, err := svc.UpdateRoom(ctx, roomID, domain.UpdateOfflineRoomDTO{Title: &newTitle})
	assert.NoError(t, err)
	assert.Equal(t, "Admin Updated", updated.Title)
}

func TestUpdateRoom_Description(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:update")

	room := &domain.OfflineRoom{ID: roomID, CreatorID: userID, Description: "old desc"}
	roomRepo.On("FindByID", ctx, roomID).Return(room, nil)
	roomRepo.On("Update", ctx, mock.AnythingOfType("*domain.OfflineRoom")).Return(nil)

	newDesc := "new desc"
	updated, err := svc.UpdateRoom(ctx, roomID, domain.UpdateOfflineRoomDTO{Description: &newDesc})
	assert.NoError(t, err)
	assert.Equal(t, "new desc", updated.Description)
}

func TestDeleteRoom_Creator_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:delete")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.OfflineRoom{ID: roomID, CreatorID: userID}, nil)
	roomRepo.On("Delete", ctx, roomID).Return(nil)

	err := svc.DeleteRoom(ctx, roomID)
	assert.NoError(t, err)
}

func TestDeleteRoom_NotCreator_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	callerID := uuid.New()
	creatorID := uuid.New()
	roomID := uuid.New()
	ctx := callerCtx(callerID, false, "offlines:delete")

	roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.OfflineRoom{ID: roomID, CreatorID: creatorID}, nil)

	err := svc.DeleteRoom(ctx, roomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDeleteRoom_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	err := svc.DeleteRoom(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListRooms_Admin_SeesAll(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	ctx := callerCtx(uuid.New(), true)
	roomRepo.On("List", ctx, mock.MatchedBy(func(s domain.OfflineRoomListScope) bool {
		return s.All
	}), mock.AnythingOfType("domain.ListOfflineRoomsQuery")).
		Return([]domain.OfflineRoom{{ID: uuid.New()}}, int64(1), nil)

	rooms, total, err := svc.ListRooms(ctx, domain.ListOfflineRoomsQuery{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rooms, 1)
}

func TestListRooms_RegularUser_ScopedToOwnerAndMember(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:view")
	roomRepo.On("List", ctx, mock.MatchedBy(func(s domain.OfflineRoomListScope) bool {
		return !s.All && s.OwnerID != nil && *s.OwnerID == userID && s.MemberUserID != nil && *s.MemberUserID == userID
	}), mock.AnythingOfType("domain.ListOfflineRoomsQuery")).
		Return([]domain.OfflineRoom{}, int64(0), nil)

	_, _, err := svc.ListRooms(ctx, domain.ListOfflineRoomsQuery{})
	assert.NoError(t, err)
}

func TestListRooms_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	_, _, err := svc.ListRooms(context.Background(), domain.ListOfflineRoomsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListRooms_NonAdmin_CannotIncludeDeleted(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	userID := uuid.New()
	ctx := callerCtx(userID, false, "offlines:view")
	roomRepo.On("List", ctx, mock.Anything, mock.MatchedBy(func(q domain.ListOfflineRoomsQuery) bool {
		return !q.IncludeDeleted
	})).Return([]domain.OfflineRoom{}, int64(0), nil)

	_, _, err := svc.ListRooms(ctx, domain.ListOfflineRoomsQuery{IncludeDeleted: true})
	assert.NoError(t, err)
}

func TestAdminList_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	ctx := callerCtx(uuid.New(), true)
	roomRepo.On("AdminList", ctx, mock.AnythingOfType("domain.AdminListOfflineRoomsQuery")).
		Return([]domain.OfflineRoom{{ID: uuid.New()}}, int64(1), nil)

	rooms, total, err := svc.AdminList(ctx, domain.AdminListOfflineRoomsQuery{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rooms, 1)
}

func TestAdminList_NonAdmin_Forbidden(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	ctx := callerCtx(uuid.New(), false)
	_, _, err := svc.AdminList(ctx, domain.AdminListOfflineRoomsQuery{})
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

func TestAdminHardDelete_NotFound(t *testing.T) {
	svc, roomRepo, _, _, _, _ := newTestService(t)

	roomID := uuid.New()
	ctx := callerCtx(uuid.New(), true)
	roomRepo.On("HardDelete", ctx, roomID).Return(domain.ErrNotFound)

	err := svc.AdminHardDelete(ctx, roomID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
