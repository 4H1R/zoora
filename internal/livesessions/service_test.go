package livesessions_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/livesessions"
)

// --- Mocks ---

type mockRoomRepo struct{ mock.Mock }

func (m *mockRoomRepo) Create(ctx context.Context, room *domain.LiveRoom) error {
	return m.Called(ctx, room).Error(0)
}
func (m *mockRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveRoom), a.Error(1)
}
func (m *mockRoomRepo) Update(ctx context.Context, room *domain.LiveRoom) error {
	return m.Called(ctx, room).Error(0)
}
func (m *mockRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) List(ctx context.Context, scope domain.LiveRoomListScope, p domain.ListParams) ([]domain.LiveRoom, int64, error) {
	a := m.Called(ctx, scope, p)
	rooms, _ := a.Get(0).([]domain.LiveRoom)
	return rooms, a.Get(1).(int64), a.Error(2)
}
func (m *mockRoomRepo) FindActiveRoomsWithStaleHost(ctx context.Context, d time.Duration) ([]domain.LiveRoom, error) {
	a := m.Called(ctx, d)
	rooms, _ := a.Get(0).([]domain.LiveRoom)
	return rooms, a.Error(1)
}
func (m *mockRoomRepo) AdminList(ctx context.Context, q domain.AdminListLiveRoomsQuery) ([]domain.LiveRoom, int64, error) {
	a := m.Called(ctx, q)
	rooms, _ := a.Get(0).([]domain.LiveRoom)
	return rooms, a.Get(1).(int64), a.Error(2)
}
func (m *mockRoomRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoomRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.LiveRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveRoom), a.Error(1)
}

type mockParticipantRepo struct{ mock.Mock }

func (m *mockParticipantRepo) Create(ctx context.Context, p *domain.LiveParticipant) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockParticipantRepo) FindActiveByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveParticipant), a.Error(1)
}
func (m *mockParticipantRepo) Update(ctx context.Context, p *domain.LiveParticipant) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockParticipantRepo) ListByRoom(ctx context.Context, roomID uuid.UUID, q domain.ListLiveParticipantsQuery) ([]domain.LiveParticipant, int64, error) {
	a := m.Called(ctx, roomID, q)
	ps, _ := a.Get(0).([]domain.LiveParticipant)
	return ps, a.Get(1).(int64), a.Error(2)
}
func (m *mockParticipantRepo) ListAllByRoom(ctx context.Context, roomID uuid.UUID) ([]domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID)
	ps, _ := a.Get(0).([]domain.LiveParticipant)
	return ps, a.Error(1)
}
func (m *mockParticipantRepo) MarkAllLeft(ctx context.Context, roomID uuid.UUID, leftAt time.Time) error {
	return m.Called(ctx, roomID, leftAt).Error(0)
}

type mockRecordingRepo struct{ mock.Mock }

func (m *mockRecordingRepo) Create(ctx context.Context, r *domain.LiveRecording) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mockRecordingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRecording, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveRecording), a.Error(1)
}
func (m *mockRecordingRepo) FindActiveByRoom(ctx context.Context, roomID uuid.UUID) (*domain.LiveRecording, error) {
	a := m.Called(ctx, roomID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveRecording), a.Error(1)
}
func (m *mockRecordingRepo) Update(ctx context.Context, r *domain.LiveRecording) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mockRecordingRepo) ListByRoom(ctx context.Context, roomID uuid.UUID, q domain.ListLiveRecordingsQuery) ([]domain.LiveRecording, int64, error) {
	a := m.Called(ctx, roomID, q)
	recs, _ := a.Get(0).([]domain.LiveRecording)
	return recs, a.Get(1).(int64), a.Error(2)
}

type mockClassSessionRepo struct{ mock.Mock }

func (m *mockClassSessionRepo) Create(ctx context.Context, s *domain.ClassSession) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockClassSessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ClassSession), a.Error(1)
}
func (m *mockClassSessionRepo) Update(ctx context.Context, s *domain.ClassSession) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockClassSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassSessionRepo) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, classID, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
}
func (m *mockClassSessionRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockClassSessionRepo) AdminList(ctx context.Context, q domain.AdminListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
}
func (m *mockClassSessionRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
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

func (m *mockMemberRepo) Create(ctx context.Context, cm *domain.ClassMember) error {
	return m.Called(ctx, cm).Error(0)
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

type mockChatService struct{ mock.Mock }

func (m *mockChatService) CreateChat(ctx context.Context, dto domain.CreateChatDTO) (*domain.Chat, error) {
	a := m.Called(ctx, dto)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Chat), a.Error(1)
}
func (m *mockChatService) GetChat(ctx context.Context, id uuid.UUID) (*domain.Chat, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Chat), a.Error(1)
}
func (m *mockChatService) UpdateChat(ctx context.Context, id uuid.UUID, dto domain.UpdateChatDTO) (*domain.Chat, error) {
	a := m.Called(ctx, id, dto)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Chat), a.Error(1)
}
func (m *mockChatService) DeleteChat(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockChatService) ListChats(ctx context.Context, q domain.ListChatsQuery) ([]domain.Chat, int64, error) {
	a := m.Called(ctx, q)
	return a.Get(0).([]domain.Chat), a.Get(1).(int64), a.Error(2)
}
func (m *mockChatService) AddMember(ctx context.Context, chatID uuid.UUID, dto domain.AddChatMemberDTO) (*domain.ChatMember, error) {
	a := m.Called(ctx, chatID, dto)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ChatMember), a.Error(1)
}
func (m *mockChatService) RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error {
	return m.Called(ctx, chatID, userID).Error(0)
}
func (m *mockChatService) ListMembers(ctx context.Context, chatID uuid.UUID) ([]domain.ChatMember, error) {
	a := m.Called(ctx, chatID)
	return a.Get(0).([]domain.ChatMember), a.Error(1)
}
func (m *mockChatService) SendMessage(ctx context.Context, chatID uuid.UUID, dto domain.SendMessageDTO) (*domain.Message, error) {
	a := m.Called(ctx, chatID, dto)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Message), a.Error(1)
}
func (m *mockChatService) GetMessage(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Message), a.Error(1)
}
func (m *mockChatService) UpdateMessage(ctx context.Context, id uuid.UUID, dto domain.UpdateMessageDTO) (*domain.Message, error) {
	a := m.Called(ctx, id, dto)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Message), a.Error(1)
}
func (m *mockChatService) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockChatService) ListMessages(ctx context.Context, chatID uuid.UUID, q domain.ListMessagesQuery) ([]domain.Message, int64, error) {
	a := m.Called(ctx, chatID, q)
	return a.Get(0).([]domain.Message), a.Get(1).(int64), a.Error(2)
}
func (m *mockChatService) ToggleReaction(ctx context.Context, messageID uuid.UUID, dto domain.ToggleReactionDTO) (*domain.Message, error) {
	a := m.Called(ctx, messageID, dto)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Message), a.Error(1)
}
func (m *mockChatService) FindChatByModel(ctx context.Context, modelType string, modelID uuid.UUID) (*domain.Chat, error) {
	a := m.Called(ctx, modelType, modelID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Chat), a.Error(1)
}
func (m *mockChatService) ArchiveByModel(ctx context.Context, modelType string, modelID uuid.UUID) error {
	return m.Called(ctx, modelType, modelID).Error(0)
}

type noopTx struct{}

func (noopTx) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// --- Test helpers ---

func newTestService(t *testing.T) (
	domain.LiveSessionService,
	*mockRoomRepo, *mockParticipantRepo, *mockRecordingRepo,
	*mockClassSessionRepo, *mockClassRepo, *mockMemberRepo,
	*mockChatService,
) {
	t.Helper()
	roomRepo := &mockRoomRepo{}
	partRepo := &mockParticipantRepo{}
	recRepo := &mockRecordingRepo{}
	sessRepo := &mockClassSessionRepo{}
	classRepo := &mockClassRepo{}
	memberRepo := &mockMemberRepo{}
	chatSvc := &mockChatService{}

	svc := livesessions.NewService(
		roomRepo, partRepo, recRepo,
		sessRepo, classRepo, memberRepo,
		chatSvc, noopTx{},
		nil, // livekit client
		slog.Default(),
	)
	return svc, roomRepo, partRepo, recRepo, sessRepo, classRepo, memberRepo, chatSvc
}

var (
	testTeacherID = uuid.New()
	testStudentID = uuid.New()
	testClassID   = uuid.New()
	testSessionID = uuid.New()
	testRoomID    = uuid.New()
)

func adminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
}

func teacherCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      testTeacherID,
		Permissions: []string{"live_sessions:manage", "live_sessions:create", "live_sessions:view", "live_sessions:update", "live_sessions:join"},
	})
}

func studentCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: testStudentID,
	})
}

func testClass() *domain.Class {
	return &domain.Class{
		ID:     testClassID,
		UserID: testTeacherID,
	}
}

func testSession() *domain.ClassSession {
	return &domain.ClassSession{
		ID:      testSessionID,
		ClassID: testClassID,
	}
}

func testRoom() *domain.LiveRoom {
	return &domain.LiveRoom{
		ID:              testRoomID,
		ClassSessionID:  testSessionID,
		LiveKitRoomName: "session-" + testSessionID.String(),
		Status:          domain.LiveRoomStatusCreated,
		Config:          domain.DefaultLiveRoomConfig(),
	}
}

// --- Tests ---

func TestCreateRoom_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.CreateRoom(context.Background(), domain.CreateLiveRoomDTO{ClassSessionID: testSessionID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateRoom_Student_Forbidden(t *testing.T) {
	svc, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	_, err := svc.CreateRoom(studentCtx(), domain.CreateLiveRoomDTO{ClassSessionID: testSessionID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateRoom_Teacher_Success(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, chatSvc := newTestService(t)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)
	chatSvc.On("CreateChat", mock.Anything, mock.AnythingOfType("domain.CreateChatDTO")).
		Return(&domain.Chat{ID: uuid.New()}, nil)

	room, err := svc.CreateRoom(teacherCtx(), domain.CreateLiveRoomDTO{
		ClassSessionID: testSessionID,
		Config:         domain.DefaultLiveRoomConfig(),
	})
	assert.NoError(t, err)
	assert.Equal(t, domain.LiveRoomStatusCreated, room.Status)
	roomRepo.AssertExpectations(t)
	chatSvc.AssertExpectations(t)
}

func TestGetRoom_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.GetRoom(context.Background(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetRoom_Student_NotMember_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, memberRepo, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	memberRepo.On("Exists", mock.Anything, testClassID, testStudentID).Return(false, nil)

	_, err := svc.GetRoom(studentCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetRoom_Student_Member_Success(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, memberRepo, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	memberRepo.On("Exists", mock.Anything, testClassID, testStudentID).Return(true, nil)

	room, err := svc.GetRoom(studentCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, room.ID)
}

func TestGetRoom_Admin_Success(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	room, err := svc.GetRoom(adminCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, room.ID)
}

func TestEndRoom_NotActive_ValidationError(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusCreated
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	_, err := svc.EndRoom(teacherCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestHeartbeat_Student_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.Heartbeat(studentCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestHeartbeat_Teacher_Success(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)

	err := svc.Heartbeat(teacherCtx(), testRoomID)
	assert.NoError(t, err)
	roomRepo.AssertExpectations(t)
}

func TestList_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newTestService(t)
	_, _, err := svc.List(context.Background(), domain.ListLiveRoomsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestList_Admin_All(t *testing.T) {
	svc, roomRepo, _, _, _, _, _, _ := newTestService(t)
	roomRepo.On("List", mock.Anything, domain.LiveRoomListScope{All: true}, mock.Anything).
		Return([]domain.LiveRoom{{ID: testRoomID}}, int64(1), nil)

	rooms, total, err := svc.List(adminCtx(), domain.ListLiveRoomsQuery{})
	assert.NoError(t, err)
	assert.Len(t, rooms, 1)
	assert.Equal(t, int64(1), total)
}

func TestAdminEndRoom_NotAdmin_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.AdminEndRoom(studentCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAdminHardDelete_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _, _, _ := newTestService(t)
	roomRepo.On("HardDelete", mock.Anything, testRoomID).Return(nil)

	err := svc.AdminHardDelete(adminCtx(), testRoomID)
	assert.NoError(t, err)
	roomRepo.AssertExpectations(t)
}

func TestCreateRoom_CreatesChat(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, chatSvc := newTestService(t)
	session := testSession()
	session.Name = "Algebra 101"
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(session, nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)
	chatSvc.On("CreateChat", mock.Anything, mock.MatchedBy(func(dto domain.CreateChatDTO) bool {
		return dto.ModelType == "live_session" && dto.Name == "Chat – Algebra 101"
	})).Return(&domain.Chat{ID: uuid.New()}, nil)

	room, err := svc.CreateRoom(teacherCtx(), domain.CreateLiveRoomDTO{
		ClassSessionID: testSessionID,
		Config:         domain.DefaultLiveRoomConfig(),
	})
	assert.NoError(t, err)
	assert.NotNil(t, room)
	chatSvc.AssertExpectations(t)
}

func TestCreateRoom_ChatFailure_RollsBack(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, chatSvc := newTestService(t)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)
	chatSvc.On("CreateChat", mock.Anything, mock.AnythingOfType("domain.CreateChatDTO")).
		Return(nil, domain.ErrConflict)

	_, err := svc.CreateRoom(teacherCtx(), domain.CreateLiveRoomDTO{
		ClassSessionID: testSessionID,
		Config:         domain.DefaultLiveRoomConfig(),
	})
	assert.Error(t, err)
}

func TestEndRoom_ArchivesChat(t *testing.T) {
	svc, roomRepo, partRepo, recRepo, sessRepo, classRepo, _, chatSvc := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	recRepo.On("FindActiveByRoom", mock.Anything, testRoomID).Return(nil, domain.ErrNotFound)
	roomRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)
	partRepo.On("MarkAllLeft", mock.Anything, testRoomID, mock.AnythingOfType("time.Time")).Return(nil)
	chatSvc.On("ArchiveByModel", mock.Anything, "live_session", testRoomID).Return(nil)

	result, err := svc.EndRoom(teacherCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, domain.LiveRoomStatusFinished, result.Status)
	chatSvc.AssertExpectations(t)
}

func TestLeaveRoom_Success(t *testing.T) {
	svc, _, partRepo, _, _, _, _, _ := newTestService(t)
	joined := time.Now().Add(-30 * time.Minute)
	p := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testStudentID,
		JoinedAt:   joined,
	}
	partRepo.On("FindActiveByRoomAndUser", mock.Anything, testRoomID, testStudentID).Return(p, nil)
	partRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.LiveParticipant")).Return(nil)

	err := svc.LeaveRoom(studentCtx(), testRoomID)
	assert.NoError(t, err)
	partRepo.AssertExpectations(t)
}

func manageAnyCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"live_sessions:manage_any", "live_sessions:view_any", "live_sessions:update_any"},
	})
}


func TestManageAny_NonOwner_CanManageRoom(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)

	err := svc.Heartbeat(manageAnyCtx(), testRoomID)
	assert.NoError(t, err)
}

func TestManage_NonOwner_Forbidden(t *testing.T) {
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"live_sessions:manage"},
	})
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.Heartbeat(ctx, testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestViewAny_ListReturnsAll_ScopedToOrg(t *testing.T) {
	svc, roomRepo, _, _, _, _, _, _ := newTestService(t)
	orgID := uuid.New()
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		OrgID:       &orgID,
		Permissions: []string{"live_sessions:view_any"},
	})
	// view_any is org-wide, NOT cross-tenant: the resolved scope must carry the
	// caller's OrgID so the repo filters live rooms to that organization only.
	roomRepo.On("List", mock.Anything, domain.LiveRoomListScope{All: true, OrganizationID: &orgID}, mock.Anything).
		Return([]domain.LiveRoom{{ID: testRoomID}}, int64(1), nil)

	rooms, total, err := svc.List(ctx, domain.ListLiveRoomsQuery{})
	assert.NoError(t, err)
	assert.Len(t, rooms, 1)
	assert.Equal(t, int64(1), total)
	roomRepo.AssertExpectations(t)
}

func TestAdmin_ListReturnsAll_NoOrgFilter(t *testing.T) {
	svc, roomRepo, _, _, _, _, _, _ := newTestService(t)
	// Admins are cross-tenant by design: no OrganizationID on the scope.
	roomRepo.On("List", mock.Anything, domain.LiveRoomListScope{All: true}, mock.Anything).
		Return([]domain.LiveRoom{{ID: testRoomID}}, int64(1), nil)

	rooms, _, err := svc.List(adminCtx(), domain.ListLiveRoomsQuery{})
	assert.NoError(t, err)
	assert.Len(t, rooms, 1)
	roomRepo.AssertExpectations(t)
}

func TestViewAny_GetRoom_NonMember_Success(t *testing.T) {
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"live_sessions:view_any"},
	})
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	r, err := svc.GetRoom(ctx, testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, r.ID)
}

func TestUpdateAny_NonOwner_CanUpdateConfig(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)

	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"live_sessions:update_any"},
	})
	cfg := domain.DefaultLiveRoomConfig()
	cfg.MaxParticipants = 50
	_, err := svc.UpdateRoomConfig(ctx, testRoomID, domain.UpdateLiveRoomConfigDTO{Config: &cfg})
	assert.NoError(t, err)
}

func TestUpdate_NonOwner_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"live_sessions:update"},
	})
	cfg := domain.DefaultLiveRoomConfig()
	_, err := svc.UpdateRoomConfig(ctx, testRoomID, domain.UpdateLiveRoomConfigDTO{Config: &cfg})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}
