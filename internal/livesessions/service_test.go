package livesessions_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	lkproto "github.com/livekit/protocol/livekit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
	"github.com/4H1R/zoora/internal/livesessions"
)

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
func (m *mockRoomRepo) ListByClassSession(ctx context.Context, sessionID uuid.UUID) ([]domain.LiveRoom, error) {
	a := m.Called(ctx, sessionID)
	rooms, _ := a.Get(0).([]domain.LiveRoom)
	return rooms, a.Error(1)
}
func (m *mockRoomRepo) Transition(ctx context.Context, room *domain.LiveRoom, from domain.LiveRoomStatus) error {
	return m.Called(ctx, room, from).Error(0)
}
func (m *mockRoomRepo) TouchHostLastSeen(ctx context.Context, roomID uuid.UUID, seenAt time.Time) error {
	return m.Called(ctx, roomID, seenAt).Error(0)
}
func (m *mockRoomRepo) UpdateConfig(ctx context.Context, roomID uuid.UUID, cfg domain.LiveRoomConfig) error {
	return m.Called(ctx, roomID, cfg).Error(0)
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
func (m *mockRoomRepo) FindByLiveKitRoomName(ctx context.Context, name string) (*domain.LiveRoom, error) {
	a := m.Called(ctx, name)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveRoom), a.Error(1)
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
func (m *mockParticipantRepo) GetActiveParticipant(ctx context.Context, roomID uuid.UUID, identity string) (*domain.LiveParticipant, error) {
	a := m.Called(ctx, roomID, identity)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveParticipant), a.Error(1)
}
func (m *mockParticipantRepo) UpdateParticipantRole(ctx context.Context, roomID uuid.UUID, identity string, role domain.ParticipantRole) error {
	return m.Called(ctx, roomID, identity, role).Error(0)
}
func (m *mockParticipantRepo) SetHandRaised(ctx context.Context, roomID uuid.UUID, identity string, raised bool) error {
	return m.Called(ctx, roomID, identity, raised).Error(0)
}
func (m *mockParticipantRepo) MarkAllLeft(ctx context.Context, roomID uuid.UUID, leftAt time.Time) error {
	return m.Called(ctx, roomID, leftAt).Error(0)
}
func (m *mockParticipantRepo) MarkLeftByIdentity(ctx context.Context, roomID uuid.UUID, identity string, leftAt time.Time) error {
	return m.Called(ctx, roomID, identity, leftAt).Error(0)
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
func (m *mockRecordingRepo) FindByEgressID(ctx context.Context, egressID string) (*domain.LiveRecording, error) {
	a := m.Called(ctx, egressID)
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

type mockWhiteboardRepo struct{ mock.Mock }

func (m *mockWhiteboardRepo) Get(ctx context.Context, roomID uuid.UUID) (*domain.LiveWhiteboard, error) {
	a := m.Called(ctx, roomID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveWhiteboard), a.Error(1)
}
func (m *mockWhiteboardRepo) Upsert(ctx context.Context, roomID uuid.UUID, snapshot json.RawMessage) (*domain.LiveWhiteboard, error) {
	a := m.Called(ctx, roomID, snapshot)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.LiveWhiteboard), a.Error(1)
}
func (m *mockWhiteboardRepo) Delete(ctx context.Context, roomID uuid.UUID) error {
	return m.Called(ctx, roomID).Error(0)
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

// mockPollService satisfies domain.PollService; only CloseByModel is exercised
// from livesessions (room-finish teardown), the rest are unused stubs.
type mockPollService struct{ mock.Mock }

func (m *mockPollService) CloseByModel(ctx context.Context, modelType string, modelID uuid.UUID) error {
	return m.Called(ctx, modelType, modelID).Error(0)
}
func (m *mockPollService) Create(context.Context, domain.CreatePollDTO) (*domain.Poll, error) {
	panic("unused")
}
func (m *mockPollService) GetByID(context.Context, uuid.UUID) (*domain.Poll, error) {
	panic("unused")
}
func (m *mockPollService) Update(context.Context, uuid.UUID, domain.UpdatePollDTO) (*domain.Poll, error) {
	panic("unused")
}
func (m *mockPollService) Delete(context.Context, uuid.UUID) error { panic("unused") }
func (m *mockPollService) List(context.Context, domain.ListPollsQuery) ([]domain.Poll, int64, error) {
	panic("unused")
}
func (m *mockPollService) Answer(context.Context, uuid.UUID, domain.AnswerPollDTO) ([]domain.PollAnswer, error) {
	panic("unused")
}
func (m *mockPollService) ListAnswers(context.Context, uuid.UUID, domain.ListPollAnswersQuery) ([]domain.PollAnswer, int64, error) {
	panic("unused")
}
func (m *mockPollService) Results(context.Context, uuid.UUID) (*domain.PollResults, error) {
	panic("unused")
}
func (m *mockPollService) AdminList(context.Context, domain.AdminListPollsQuery) ([]domain.Poll, int64, error) {
	panic("unused")
}
func (m *mockPollService) AdminHardDelete(context.Context, uuid.UUID) error { panic("unused") }

type noopTx struct{}

func (noopTx) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeLiveKit is a permissive LiveKitClient fake: every call succeeds. Tests
// override the function fields to steer or observe specific calls.
type fakeLiveKit struct {
	listParticipantsFn  func(ctx context.Context, roomName string) ([]*lkproto.ParticipantInfo, error)
	removeParticipantFn func(ctx context.Context, roomName, identity string) error

	tokenMetadata  []string
	tokenSources   [][]lkproto.TrackSource
	tokenRoomAdmin []bool
}

func (f *fakeLiveKit) CreateRoom(_ context.Context, roomName string, _ uint32) (*lkproto.Room, error) {
	return &lkproto.Room{Name: roomName}, nil
}
func (f *fakeLiveKit) DeleteRoom(context.Context, string) error { return nil }
func (f *fakeLiveKit) GenerateToken(_, _, _, metadata string, sources []lkproto.TrackSource, roomAdmin bool) (string, error) {
	f.tokenMetadata = append(f.tokenMetadata, metadata)
	f.tokenSources = append(f.tokenSources, sources)
	f.tokenRoomAdmin = append(f.tokenRoomAdmin, roomAdmin)
	return "test-token", nil
}
func (f *fakeLiveKit) StartRecording(context.Context, string, string) (string, error) {
	return "EG_test", nil
}
func (f *fakeLiveKit) StopRecording(context.Context, string) error { return nil }
func (f *fakeLiveKit) ListParticipants(ctx context.Context, roomName string) ([]*lkproto.ParticipantInfo, error) {
	if f.listParticipantsFn != nil {
		return f.listParticipantsFn(ctx, roomName)
	}
	return nil, nil
}
func (f *fakeLiveKit) UpdateParticipant(context.Context, string, string, string, []lkproto.TrackSource) error {
	return nil
}
func (f *fakeLiveKit) MutePublishedTrack(context.Context, string, string, string, bool) error {
	return nil
}
func (f *fakeLiveKit) RemoveParticipant(ctx context.Context, roomName, identity string) error {
	if f.removeParticipantFn != nil {
		return f.removeParticipantFn(ctx, roomName, identity)
	}
	return nil
}
func (f *fakeLiveKit) SendData(context.Context, string, []byte, []string) error { return nil }
func (f *fakeLiveKit) PublicURL() string                                        { return "wss://livekit.test" }

func newTestService(t *testing.T) (
	domain.LiveSessionService,
	*mockRoomRepo, *mockParticipantRepo, *mockRecordingRepo, *mockWhiteboardRepo,
	*mockClassSessionRepo, *mockClassRepo, *mockMemberRepo,
	*mockChatService,
) {
	t.Helper()
	svc, f := newTestServiceLK(t)
	return svc, f.rooms, f.parts, f.recs, f.wb, f.sess, f.classes, f.members, f.chat
}

// lkFixture bundles the mocks plus the LiveKit fake for tests that need to
// steer or inspect LiveKit interactions.
type lkFixture struct {
	rooms   *mockRoomRepo
	parts   *mockParticipantRepo
	recs    *mockRecordingRepo
	wb      *mockWhiteboardRepo
	sess    *mockClassSessionRepo
	classes *mockClassRepo
	members *mockMemberRepo
	chat    *mockChatService
	poll    *mockPollService
	lk      *fakeLiveKit
}

// fakeEntSvc injects canned entitlement-limit results for livesession tests.
type fakeEntSvc struct{ concurrentErr error }

func (f fakeEntSvc) CheckUserLimit(context.Context, uuid.UUID, domain.Entitlements) error { return nil }
func (f fakeEntSvc) CheckStorageLimit(context.Context, uuid.UUID, domain.Entitlements, int64) error {
	return nil
}
func (f fakeEntSvc) CheckConcurrentRoomsLimit(context.Context, uuid.UUID, domain.Entitlements) error {
	return f.concurrentErr
}

func newTestServiceLK(t *testing.T) (domain.LiveSessionService, *lkFixture) {
	return newTestServiceLKEnt(t, fakeEntSvc{})
}

func newTestServiceLKEnt(t *testing.T, ent entitlements.Service) (domain.LiveSessionService, *lkFixture) {
	t.Helper()
	f := &lkFixture{
		rooms:   &mockRoomRepo{},
		parts:   &mockParticipantRepo{},
		recs:    &mockRecordingRepo{},
		wb:      &mockWhiteboardRepo{},
		sess:    &mockClassSessionRepo{},
		classes: &mockClassRepo{},
		members: &mockMemberRepo{},
		chat:    &mockChatService{},
		poll:    &mockPollService{},
		lk:      &fakeLiveKit{},
	}
	// Room-finish teardown closes polls; allow it without forcing every test to
	// register the expectation explicitly.
	f.poll.On("CloseByModel", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	svc := livesessions.NewService(
		f.rooms, f.parts, f.recs, f.wb,
		f.sess, f.classes, f.members,
		f.chat, f.poll, noopTx{},
		f.lk,
		nil, // queue client
		ent,
		15*time.Minute,
		slog.Default(),
	)
	return svc, f
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
		Ent:         domain.PlanCatalog[domain.PlanPro],
	})
}

// freeTeacherCtx is a teacher on the Free plan (no recording feature).
func freeTeacherCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      testTeacherID,
		Permissions: []string{"live_sessions:manage", "live_sessions:create", "live_sessions:view", "live_sessions:update", "live_sessions:join"},
		Ent:         domain.PlanCatalog[domain.PlanFree],
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

func TestCreateRoom_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.CreateRoom(context.Background(), domain.CreateLiveRoomDTO{ClassSessionID: testSessionID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateRoom_Student_Forbidden(t *testing.T) {
	svc, _, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	_, err := svc.CreateRoom(studentCtx(), domain.CreateLiveRoomDTO{ClassSessionID: testSessionID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateRoom_Teacher_Success(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, chatSvc := newTestService(t)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)
	chatSvc.On("CreateChat", mock.Anything, mock.AnythingOfType("domain.CreateChatDTO")).
		Return(&domain.Chat{ID: uuid.New()}, nil)

	room, err := svc.CreateRoom(teacherCtx(), domain.CreateLiveRoomDTO{
		ClassSessionID: testSessionID,
		Name:           "Morning session",
		Config:         domain.DefaultLiveRoomConfig(),
	})
	assert.NoError(t, err)
	assert.Equal(t, domain.LiveRoomStatusCreated, room.Status)
	roomRepo.AssertExpectations(t)
	chatSvc.AssertExpectations(t)
}

func TestGetRoom_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.GetRoom(context.Background(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetRoom_Student_NotMember_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, memberRepo, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	memberRepo.On("Exists", mock.Anything, testClassID, testStudentID).Return(false, nil)

	_, err := svc.GetRoom(studentCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetRoom_Student_Member_Success(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, memberRepo, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	memberRepo.On("Exists", mock.Anything, testClassID, testStudentID).Return(true, nil)

	room, err := svc.GetRoom(studentCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, room.ID)
}

func TestGetRoom_Admin_Success(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	room, err := svc.GetRoom(adminCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, room.ID)
}

func TestEndRoom_NotActive_ValidationError(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusCreated
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	_, err := svc.EndRoom(teacherCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestHeartbeat_Student_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.Heartbeat(studentCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestHeartbeat_Teacher_Success(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("TouchHostLastSeen", mock.Anything, testRoomID, mock.AnythingOfType("time.Time")).Return(nil)

	err := svc.Heartbeat(teacherCtx(), testRoomID)
	assert.NoError(t, err)
	roomRepo.AssertExpectations(t)
}

func TestHeartbeat_NeverFullSaves(t *testing.T) {
	// A heartbeat racing a close must not write the whole row back (that would
	// resurrect a finished room); only the conditional touch is allowed.
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("TouchHostLastSeen", mock.Anything, testRoomID, mock.AnythingOfType("time.Time")).Return(nil)

	err := svc.Heartbeat(teacherCtx(), testRoomID)
	assert.NoError(t, err)
	roomRepo.AssertNotCalled(t, "Transition")
	roomRepo.AssertNotCalled(t, "UpdateConfig")
}

func TestList_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, _, err := svc.List(context.Background(), domain.ListLiveRoomsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestList_Admin_All(t *testing.T) {
	svc, roomRepo, _, _, _, _, _, _, _ := newTestService(t)
	roomRepo.On("List", mock.Anything, domain.LiveRoomListScope{All: true}, mock.Anything).
		Return([]domain.LiveRoom{{ID: testRoomID}}, int64(1), nil)

	rooms, total, err := svc.List(adminCtx(), domain.ListLiveRoomsQuery{})
	assert.NoError(t, err)
	assert.Len(t, rooms, 1)
	assert.Equal(t, int64(1), total)
}

func TestAdminEndRoom_NotAdmin_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.AdminEndRoom(studentCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAdminHardDelete_Success(t *testing.T) {
	svc, roomRepo, _, _, _, _, _, _, _ := newTestService(t)
	roomRepo.On("HardDelete", mock.Anything, testRoomID).Return(nil)

	err := svc.AdminHardDelete(adminCtx(), testRoomID)
	assert.NoError(t, err)
	roomRepo.AssertExpectations(t)
}

func TestCreateRoom_CreatesChat(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, chatSvc := newTestService(t)
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
		Name:           "Morning session",
		Config:         domain.DefaultLiveRoomConfig(),
	})
	assert.NoError(t, err)
	assert.NotNil(t, room)
	chatSvc.AssertExpectations(t)
}

func TestCreateRoom_ChatFailure_RollsBack(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, chatSvc := newTestService(t)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.LiveRoom")).Return(nil)
	chatSvc.On("CreateChat", mock.Anything, mock.AnythingOfType("domain.CreateChatDTO")).
		Return(nil, domain.ErrConflict)

	_, err := svc.CreateRoom(teacherCtx(), domain.CreateLiveRoomDTO{
		ClassSessionID: testSessionID,
		Name:           "Morning session",
		Config:         domain.DefaultLiveRoomConfig(),
	})
	assert.Error(t, err)
}

func TestCreateRoom_BlankName_Validation(t *testing.T) {
	svc, _, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	_, err := svc.CreateRoom(teacherCtx(), domain.CreateLiveRoomDTO{
		ClassSessionID: testSessionID,
		Name:           "   ",
		Config:         domain.DefaultLiveRoomConfig(),
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestEndRoom_ArchivesChat(t *testing.T) {
	svc, roomRepo, partRepo, recRepo, wbRepo, sessRepo, classRepo, _, chatSvc := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	recRepo.On("FindActiveByRoom", mock.Anything, testRoomID).Return(nil, domain.ErrNotFound)
	roomRepo.On("Transition", mock.Anything, mock.AnythingOfType("*domain.LiveRoom"), domain.LiveRoomStatusActive).Return(nil)
	partRepo.On("MarkAllLeft", mock.Anything, testRoomID, mock.AnythingOfType("time.Time")).Return(nil)
	wbRepo.On("Delete", mock.Anything, testRoomID).Return(nil)
	chatSvc.On("ArchiveByModel", mock.Anything, "live_session", testRoomID).Return(nil)

	result, err := svc.EndRoom(teacherCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, domain.LiveRoomStatusFinished, result.Status)
	chatSvc.AssertExpectations(t)
}

func TestLeaveRoom_Success(t *testing.T) {
	svc, _, partRepo, _, _, _, _, _, _ := newTestService(t)
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
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("TouchHostLastSeen", mock.Anything, testRoomID, mock.AnythingOfType("time.Time")).Return(nil)

	err := svc.Heartbeat(manageAnyCtx(), testRoomID)
	assert.NoError(t, err)
}

func TestManage_NonOwner_Forbidden(t *testing.T) {
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID:      uuid.New(),
		Permissions: []string{"live_sessions:manage"},
	})
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.Heartbeat(ctx, testRoomID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestViewAny_ListReturnsAll_ScopedToOrg(t *testing.T) {
	svc, roomRepo, _, _, _, _, _, _, _ := newTestService(t)
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
	svc, roomRepo, _, _, _, _, _, _, _ := newTestService(t)
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
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	r, err := svc.GetRoom(ctx, testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, r.ID)
}

func TestUpdateAny_NonOwner_CanUpdateConfig(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	roomRepo.On("UpdateConfig", mock.Anything, testRoomID, mock.AnythingOfType("domain.LiveRoomConfig")).Return(nil)

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
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
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

// ---------------------------------------------------------------------------
// SetParticipantRole tests
// ---------------------------------------------------------------------------

func TestSetParticipantRole_InvalidRole_Error(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.SetParticipantRole(teacherCtx(), testRoomID, "some-identity", domain.SetParticipantRoleDTO{Role: "invalid"})
	assert.ErrorIs(t, err, domain.ErrInvalidParticipantRole)
}

func TestSetParticipantRole_RoleHost_Error(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.SetParticipantRole(teacherCtx(), testRoomID, "some-identity", domain.SetParticipantRoleDTO{Role: domain.ParticipantRoleHost})
	assert.ErrorIs(t, err, domain.ErrCannotChangeHostRole)
}

func TestSetParticipantRole_Student_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	_, err := svc.SetParticipantRole(studentCtx(), testRoomID, "some-identity", domain.SetParticipantRoleDTO{Role: domain.ParticipantRolePresenter})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSetParticipantRole_TargetIsHost_Error(t *testing.T) {
	svc, roomRepo, partRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	hostParticipant := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testTeacherID,
		Identity:   testTeacherID.String(),
		Role:       domain.ParticipantRoleHost,
	}
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, testTeacherID.String()).Return(hostParticipant, nil)

	_, err := svc.SetParticipantRole(teacherCtx(), testRoomID, testTeacherID.String(), domain.SetParticipantRoleDTO{Role: domain.ParticipantRoleViewer})
	assert.ErrorIs(t, err, domain.ErrCannotChangeHostRole)
}

func TestSetParticipantRole_PromoteToPresenter_Success(t *testing.T) {
	svc, roomRepo, partRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	viewerParticipant := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testStudentID,
		Identity:   testStudentID.String(),
		Role:       domain.ParticipantRoleViewer,
	}
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, testStudentID.String()).Return(viewerParticipant, nil)
	partRepo.On("UpdateParticipantRole", mock.Anything, testRoomID, testStudentID.String(), domain.ParticipantRolePresenter).Return(nil)

	result, err := svc.SetParticipantRole(teacherCtx(), testRoomID, testStudentID.String(), domain.SetParticipantRoleDTO{Role: domain.ParticipantRolePresenter})
	assert.NoError(t, err)
	assert.Equal(t, domain.ParticipantRolePresenter, result.Role)
	partRepo.AssertExpectations(t)
}

func TestSetParticipantRole_DemoteToViewer_Success(t *testing.T) {
	svc, roomRepo, partRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	presenterParticipant := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testStudentID,
		Identity:   testStudentID.String(),
		Role:       domain.ParticipantRolePresenter,
	}
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, testStudentID.String()).Return(presenterParticipant, nil)
	partRepo.On("UpdateParticipantRole", mock.Anything, testRoomID, testStudentID.String(), domain.ParticipantRoleViewer).Return(nil)

	result, err := svc.SetParticipantRole(teacherCtx(), testRoomID, testStudentID.String(), domain.SetParticipantRoleDTO{Role: domain.ParticipantRoleViewer})
	assert.NoError(t, err)
	assert.Equal(t, domain.ParticipantRoleViewer, result.Role)
	partRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// MuteParticipant tests
// ---------------------------------------------------------------------------

func TestMuteParticipant_Student_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.MuteParticipant(studentCtx(), testRoomID, "some-identity", domain.MuteParticipantDTO{TrackSID: "TR_abc", Muted: true})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestMuteParticipant_Teacher_Success(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.MuteParticipant(teacherCtx(), testRoomID, testStudentID.String(), domain.MuteParticipantDTO{TrackSID: "TR_abc", Muted: true})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// RemoveParticipant tests
// ---------------------------------------------------------------------------

func TestRemoveParticipant_Student_Forbidden(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.RemoveParticipant(studentCtx(), testRoomID, testStudentID.String())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestRemoveParticipant_Self_Error(t *testing.T) {
	svc, roomRepo, _, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	err := svc.RemoveParticipant(teacherCtx(), testRoomID, testTeacherID.String())
	assert.ErrorIs(t, err, domain.ErrCannotRemoveSelf)
}

func TestRemoveParticipant_TargetIsHost_Error(t *testing.T) {
	svc, roomRepo, partRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	otherHost := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testStudentID,
		Identity:   testStudentID.String(),
		Role:       domain.ParticipantRoleHost,
	}
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, testStudentID.String()).Return(otherHost, nil)

	err := svc.RemoveParticipant(teacherCtx(), testRoomID, testStudentID.String())
	assert.ErrorIs(t, err, domain.ErrCannotRemoveHost)
}

func TestRemoveParticipant_Teacher_Success(t *testing.T) {
	svc, roomRepo, partRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	viewer := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testStudentID,
		Identity:   testStudentID.String(),
		Role:       domain.ParticipantRoleViewer,
	}
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, testStudentID.String()).Return(viewer, nil)
	partRepo.On("MarkLeftByIdentity", mock.Anything, testRoomID, testStudentID.String(), mock.Anything).Return(nil)

	err := svc.RemoveParticipant(teacherCtx(), testRoomID, testStudentID.String())
	assert.NoError(t, err)
	partRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// SetHand tests
// ---------------------------------------------------------------------------

func TestSetHand_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.SetHand(context.Background(), testRoomID, domain.SetHandDTO{Raised: true})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSetHand_Student_SetsHandAndReturnsParticipant(t *testing.T) {
	svc, roomRepo, partRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)

	identity := testStudentID.String()
	partRepo.On("SetHandRaised", mock.Anything, testRoomID, identity, true).Return(nil)

	now := time.Now()
	expected := &domain.LiveParticipant{
		ID:           uuid.New(),
		LiveRoomID:   testRoomID,
		UserID:       testStudentID,
		Identity:     identity,
		Role:         domain.ParticipantRoleViewer,
		HandRaisedAt: &now,
	}
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, identity).Return(expected, nil)

	result, err := svc.SetHand(studentCtx(), testRoomID, domain.SetHandDTO{Raised: true})
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.NotNil(t, result.HandRaisedAt)
	partRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// SaveWhiteboard / GetWhiteboard tests
// ---------------------------------------------------------------------------

func TestSaveWhiteboard_NoCaller_Forbidden(t *testing.T) {
	svc, _, _, _, _, _, _, _, _ := newTestService(t)
	_, err := svc.SaveWhiteboard(context.Background(), testRoomID, domain.SaveWhiteboardDTO{
		Snapshot: json.RawMessage(`{"shapes":[]}`),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSaveWhiteboard_Viewer_Forbidden(t *testing.T) {
	svc, roomRepo, partRepo, _, _, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, testStudentID.String()).
		Return(&domain.LiveParticipant{Role: domain.ParticipantRoleViewer}, nil)

	_, err := svc.SaveWhiteboard(studentCtx(), testRoomID, domain.SaveWhiteboardDTO{
		Snapshot: json.RawMessage(`{"shapes":[]}`),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSaveWhiteboard_Host_Success(t *testing.T) {
	svc, roomRepo, _, _, wbRepo, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	snapshot := json.RawMessage(`{"shapes":[{"id":"s1"}]}`)
	wbRepo.On("Upsert", mock.Anything, testRoomID, snapshot).
		Return(&domain.LiveWhiteboard{LiveRoomID: testRoomID, Snapshot: snapshot}, nil)

	wb, err := svc.SaveWhiteboard(teacherCtx(), testRoomID, domain.SaveWhiteboardDTO{Snapshot: snapshot})
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, wb.LiveRoomID)
	wbRepo.AssertExpectations(t)
}

func TestSaveWhiteboard_Presenter_Success(t *testing.T) {
	svc, roomRepo, partRepo, _, wbRepo, sessRepo, classRepo, _, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	partRepo.On("GetActiveParticipant", mock.Anything, testRoomID, testStudentID.String()).
		Return(&domain.LiveParticipant{Role: domain.ParticipantRolePresenter}, nil)
	snapshot := json.RawMessage(`{"shapes":[{"id":"s2"}]}`)
	wbRepo.On("Upsert", mock.Anything, testRoomID, snapshot).
		Return(&domain.LiveWhiteboard{LiveRoomID: testRoomID, Snapshot: snapshot}, nil)

	wb, err := svc.SaveWhiteboard(studentCtx(), testRoomID, domain.SaveWhiteboardDTO{Snapshot: snapshot})
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, wb.LiveRoomID)
	wbRepo.AssertExpectations(t)
}

func TestGetWhiteboard_NoRecord_ReturnsEmpty(t *testing.T) {
	svc, roomRepo, _, _, wbRepo, sessRepo, classRepo, memberRepo, _ := newTestService(t)
	roomRepo.On("FindByID", mock.Anything, testRoomID).Return(testRoom(), nil)
	sessRepo.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	classRepo.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	memberRepo.On("Exists", mock.Anything, testClassID, testStudentID).Return(true, nil)
	wbRepo.On("Get", mock.Anything, testRoomID).Return(nil, domain.ErrWhiteboardNotFound)

	wb, err := svc.GetWhiteboard(studentCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, testRoomID, wb.LiveRoomID)
	assert.Equal(t, json.RawMessage("{}"), wb.Snapshot)
}
