package chat_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/chat"
	"github.com/4H1R/zoora/internal/domain"
)

type mockChatRepo struct{ mock.Mock }

func (m *mockChatRepo) Create(ctx context.Context, c *domain.LiveRoomChat) error {
	return m.Called(ctx, c).Error(0)
}

func (m *mockChatRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRoomChat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LiveRoomChat), args.Error(1)
}

func (m *mockChatRepo) Update(ctx context.Context, c *domain.LiveRoomChat) error {
	return m.Called(ctx, c).Error(0)
}

func (m *mockChatRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockChatRepo) List(ctx context.Context, q domain.ListChatsQuery) ([]domain.LiveRoomChat, int64, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]domain.LiveRoomChat), args.Get(1).(int64), args.Error(2)
}

func (m *mockChatRepo) FindByRoom(ctx context.Context, liveRoomID uuid.UUID) (*domain.LiveRoomChat, error) {
	args := m.Called(ctx, liveRoomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LiveRoomChat), args.Error(1)
}

type mockMessageRepo struct{ mock.Mock }

func (m *mockMessageRepo) Create(ctx context.Context, msg *domain.LiveRoomMessage) error {
	return m.Called(ctx, msg).Error(0)
}

func (m *mockMessageRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRoomMessage, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LiveRoomMessage), args.Error(1)
}

func (m *mockMessageRepo) Update(ctx context.Context, msg *domain.LiveRoomMessage) error {
	return m.Called(ctx, msg).Error(0)
}

func (m *mockMessageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockMessageRepo) List(ctx context.Context, chatID uuid.UUID, q domain.ListMessagesQuery) ([]domain.LiveRoomMessage, int64, error) {
	args := m.Called(ctx, chatID, q)
	return args.Get(0).([]domain.LiveRoomMessage), args.Get(1).(int64), args.Error(2)
}

func adminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
}

func memberCtx(userID uuid.UUID) context.Context {
	orgID := uuid.New()
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: userID, OrgID: &orgID,
		Ent: domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	})
}

type noopTx struct{}

func (noopTx) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// stubModelAuth is a configurable ModelAuthorizer fake. It returns fixed
// participate/moderate results (and optional errors) regardless of arguments, so
// chat-service authz branches can be exercised without the live-session chain.
type stubModelAuth struct {
	participate    bool
	participateErr error
	moderate       bool
	moderateErr    error
}

func (s stubModelAuth) CanParticipate(context.Context, domain.Caller, string, uuid.UUID) (bool, error) {
	return s.participate, s.participateErr
}

func (s stubModelAuth) CanModerate(context.Context, domain.Caller, string, uuid.UUID) (bool, error) {
	return s.moderate, s.moderateErr
}

func (s stubModelAuth) OrgForModel(context.Context, string, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

// allowAll authorizes every participate/moderate check; the default for tests
// whose focus is not the authz decision itself.
var allowAll = stubModelAuth{participate: true, moderate: true}

func newService(
	chatRepo *mockChatRepo,
	msgRepo *mockMessageRepo,
) domain.LiveRoomChatService {
	return chat.NewService(chatRepo, msgRepo, noopTx{}, slog.Default(), nil, nil, allowAll)
}

func newServiceWithAuth(
	chatRepo *mockChatRepo,
	msgRepo *mockMessageRepo,
	auth domain.ModelAuthorizer,
) domain.LiveRoomChatService {
	return chat.NewService(chatRepo, msgRepo, noopTx{}, slog.Default(), nil, nil, auth)
}

func TestCreateChat_AdminSuccess(t *testing.T) {
	ctx := adminCtx()
	chatRepo := &mockChatRepo{}
	chatRepo.On("Create", ctx, mock.AnythingOfType("*domain.LiveRoomChat")).Return(nil)

	svc := newService(chatRepo, nil)
	liveRoomID := uuid.New()

	c, err := svc.CreateChat(ctx, domain.CreateChatDTO{
		Name:       "Test Chat",
		LiveRoomID: liveRoomID.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, "Test Chat", c.Name)
	assert.Equal(t, liveRoomID, c.LiveRoomID)
	assert.Equal(t, domain.LiveRoomChatStatusActive, c.Status)
	chatRepo.AssertExpectations(t)
}

func TestCreateChat_NoCaller_Forbidden(t *testing.T) {
	svc := newService(&mockChatRepo{}, nil)
	_, err := svc.CreateChat(context.Background(), domain.CreateChatDTO{
		Name:       "Test",
		LiveRoomID: uuid.New().String(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateChat_NonAdmin_Forbidden(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)

	svc := newService(&mockChatRepo{}, nil)
	_, err := svc.CreateChat(ctx, domain.CreateChatDTO{
		Name:       "Test",
		LiveRoomID: uuid.New().String(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetChat_AnyCaller_Success(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:     chatID,
		Status: domain.LiveRoomChatStatusActive,
	}, nil)

	svc := newService(chatRepo, nil)

	c, err := svc.GetChat(ctx, chatID)
	assert.NoError(t, err)
	assert.Equal(t, chatID, c.ID)
}

func TestSendMessage_Success(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	msgRepo := &mockMessageRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:     chatID,
		Status: domain.LiveRoomChatStatusActive,
	}, nil)
	msgRepo.On("Create", ctx, mock.AnythingOfType("*domain.LiveRoomMessage")).Return(nil)

	svc := newService(chatRepo, msgRepo)

	msg, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.LiveRoomMessageTypeText,
		Content:     "Hello!",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Hello!", msg.Content)
	assert.Equal(t, &userID, msg.SenderID)
}

func TestSendMessage_ArchivedChat_Rejected(t *testing.T) {
	ctx := adminCtx()
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:     chatID,
		Status: domain.LiveRoomChatStatusArchived,
	}, nil)

	svc := newService(chatRepo, nil)

	_, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.LiveRoomMessageTypeText,
		Content:     "Hello!",
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestUpdateMessage_OwnerAllowed(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{
		ID:       msgID,
		SenderID: &userID,
		Content:  "old",
	}, nil)
	msgRepo.On("Update", ctx, mock.MatchedBy(func(m *domain.LiveRoomMessage) bool {
		return m.Content == "new" && m.IsEdited
	})).Return(nil)

	svc := newService(nil, msgRepo)

	msg, err := svc.UpdateMessage(ctx, msgID, domain.UpdateMessageDTO{Content: "new"})
	assert.NoError(t, err)
	assert.Equal(t, "new", msg.Content)
	assert.True(t, msg.IsEdited)
}

func TestUpdateMessage_OtherUser_NonModerator_Forbidden(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	chatID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{
		ID:       msgID,
		ChatID:   chatID,
		SenderID: &otherUserID,
	}, nil)

	svc := newServiceWithAuth(chatRepo, msgRepo, stubModelAuth{moderate: false})

	_, err := svc.UpdateMessage(ctx, msgID, domain.UpdateMessageDTO{Content: "hack"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdateMessage_OtherUser_Moderator_Allowed(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	chatID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{
		ID:       msgID,
		ChatID:   chatID,
		SenderID: &otherUserID,
	}, nil)
	msgRepo.On("Update", ctx, mock.AnythingOfType("*domain.LiveRoomMessage")).Return(nil)

	svc := newServiceWithAuth(chatRepo, msgRepo, stubModelAuth{moderate: true})

	msg, err := svc.UpdateMessage(ctx, msgID, domain.UpdateMessageDTO{Content: "moderated"})
	assert.NoError(t, err)
	assert.Equal(t, "moderated", msg.Content)
}

func TestDeleteMessage_SenderAllowed(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{
		ID:       msgID,
		SenderID: &userID,
	}, nil)
	msgRepo.On("Delete", ctx, msgID).Return(nil)

	svc := newService(nil, msgRepo)

	err := svc.DeleteMessage(ctx, msgID)
	assert.NoError(t, err)
}

func TestDeleteMessage_OtherUser_Moderator_Allowed(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	chatID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{
		ID:       msgID,
		ChatID:   chatID,
		SenderID: &otherUserID,
	}, nil)
	msgRepo.On("Delete", ctx, msgID).Return(nil)

	svc := newServiceWithAuth(chatRepo, msgRepo, stubModelAuth{moderate: true})

	err := svc.DeleteMessage(ctx, msgID)
	assert.NoError(t, err)
}

func TestDeleteMessage_OtherUser_NonModerator_Forbidden(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	chatID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{
		ID:       msgID,
		ChatID:   chatID,
		SenderID: &otherUserID,
	}, nil)

	svc := newServiceWithAuth(chatRepo, msgRepo, stubModelAuth{moderate: false})

	err := svc.DeleteMessage(ctx, msgID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSendMessage_WithThread(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()
	parentMsgID := uuid.New()

	chatRepo := &mockChatRepo{}
	msgRepo := &mockMessageRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:     chatID,
		Status: domain.LiveRoomChatStatusActive,
	}, nil)
	msgRepo.On("FindByID", ctx, parentMsgID).Return(&domain.LiveRoomMessage{
		ID:     parentMsgID,
		ChatID: chatID,
	}, nil)
	msgRepo.On("Create", ctx, mock.MatchedBy(func(m *domain.LiveRoomMessage) bool {
		return m.ParentMessageID != nil && *m.ParentMessageID == parentMsgID
	})).Return(nil)

	svc := newService(chatRepo, msgRepo)

	parentIDStr := parentMsgID.String()
	msg, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType:     domain.LiveRoomMessageTypeText,
		Content:         "Reply",
		ParentMessageID: &parentIDStr,
	})
	assert.NoError(t, err)
	assert.Equal(t, &parentMsgID, msg.ParentMessageID)
}

// spySender records SendData calls to assert realtime broadcast happened.
type spySender struct {
	rooms    []string
	payloads [][]byte
}

func (s *spySender) SendData(_ context.Context, room string, payload []byte, _ []string) error {
	s.rooms = append(s.rooms, room)
	s.payloads = append(s.payloads, payload)
	return nil
}

// stubLiveRooms resolves a live room by ID. Embeds the interface so only the one
// method under test needs implementing; any other call would panic (and none do).
type stubLiveRooms struct {
	domain.LiveRoomRepository
	room *domain.LiveRoom
}

func (s stubLiveRooms) FindByID(_ context.Context, _ uuid.UUID) (*domain.LiveRoom, error) {
	return s.room, nil
}

func TestSendMessage_BroadcastsOverDataChannel(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()
	roomID := uuid.New()

	chatRepo := &mockChatRepo{}
	msgRepo := &mockMessageRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:         chatID,
		LiveRoomID: roomID,
		Status:     domain.LiveRoomChatStatusActive,
	}, nil)
	msgRepo.On("Create", ctx, mock.AnythingOfType("*domain.LiveRoomMessage")).Return(nil)

	sender := &spySender{}
	rooms := stubLiveRooms{room: &domain.LiveRoom{ID: roomID, LiveKitRoomName: "session-abc"}}
	svc := chat.NewService(chatRepo, msgRepo, noopTx{}, slog.Default(), sender, rooms, allowAll)

	_, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.LiveRoomMessageTypeText,
		Content:     "Hi room!",
	})
	assert.NoError(t, err)

	if assert.Len(t, sender.rooms, 1) {
		assert.Equal(t, "session-abc", sender.rooms[0])
		var env struct {
			Type string `json:"type"`
			Data struct {
				Content string `json:"content"`
			} `json:"data"`
		}
		assert.NoError(t, json.Unmarshal(sender.payloads[0], &env))
		assert.Equal(t, "chat_message", env.Type)
		assert.Equal(t, "Hi room!", env.Data.Content)
	}
}

func TestSendMessage_NoLiveKitWiring_NoBroadcast(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	msgRepo := &mockMessageRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:     chatID,
		Status: domain.LiveRoomChatStatusActive,
	}, nil)
	msgRepo.On("Create", ctx, mock.AnythingOfType("*domain.LiveRoomMessage")).Return(nil)

	svc := newService(chatRepo, msgRepo)

	_, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.LiveRoomMessageTypeText,
		Content:     "no broadcast wiring",
	})
	assert.NoError(t, err)
}

// managerCtx builds a caller holding chats:manage so it clears the capability
// pre-filter and reaches the object-level canModerate check.
func managerCtx(userID uuid.UUID) context.Context {
	orgID := uuid.New()
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		OrgID:       &orgID,
		Ent:         domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
		Permissions: []string{string(domain.PermChatsManage)},
	})
}

func TestGetChat_NonParticipant_Forbidden(t *testing.T) {
	ctx := memberCtx(uuid.New())
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:     chatID,
		Status: domain.LiveRoomChatStatusActive,
	}, nil)

	svc := newServiceWithAuth(chatRepo, nil, stubModelAuth{participate: false})

	_, err := svc.GetChat(ctx, chatID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetMessage_NonParticipant_Forbidden(t *testing.T) {
	ctx := memberCtx(uuid.New())
	chatID := uuid.New()
	msgID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{ID: msgID, ChatID: chatID}, nil)

	svc := newServiceWithAuth(chatRepo, msgRepo, stubModelAuth{participate: false})

	_, err := svc.GetMessage(ctx, msgID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetMessage_Participant_Success(t *testing.T) {
	ctx := memberCtx(uuid.New())
	chatID := uuid.New()
	msgID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.LiveRoomMessage{ID: msgID, ChatID: chatID}, nil)

	svc := newServiceWithAuth(chatRepo, msgRepo, stubModelAuth{participate: true})

	msg, err := svc.GetMessage(ctx, msgID)
	assert.NoError(t, err)
	assert.Equal(t, msgID, msg.ID)
}

func TestListMessages_NonParticipant_Forbidden(t *testing.T) {
	ctx := memberCtx(uuid.New())
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)

	svc := newServiceWithAuth(chatRepo, nil, stubModelAuth{participate: false})

	_, _, err := svc.ListMessages(ctx, chatID, domain.ListMessagesQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSendMessage_NonParticipant_Forbidden_NoCreate(t *testing.T) {
	ctx := memberCtx(uuid.New())
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{
		ID:     chatID,
		Status: domain.LiveRoomChatStatusActive,
	}, nil)
	// msgRepo has no Create expectation: a call would panic, proving no message
	// was persisted for a non-participant.
	msgRepo := &mockMessageRepo{}

	svc := newServiceWithAuth(chatRepo, msgRepo, stubModelAuth{participate: false})

	_, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.LiveRoomMessageTypeText,
		Content:     "intrusion",
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	msgRepo.AssertNotCalled(t, "Create")
}

func TestUpdateChat_NonModerator_Forbidden(t *testing.T) {
	ctx := managerCtx(uuid.New())
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)

	svc := newServiceWithAuth(chatRepo, nil, stubModelAuth{moderate: false})

	name := "renamed"
	_, err := svc.UpdateChat(ctx, chatID, domain.UpdateChatDTO{Name: &name})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDeleteChat_NonModerator_Forbidden(t *testing.T) {
	ctx := managerCtx(uuid.New())
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	// no Delete expectation: proving the chat is not removed.

	svc := newServiceWithAuth(chatRepo, nil, stubModelAuth{moderate: false})

	err := svc.DeleteChat(ctx, chatID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	chatRepo.AssertNotCalled(t, "Delete")
}

func TestDeleteChat_Moderator_Allowed(t *testing.T) {
	ctx := managerCtx(uuid.New())
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.LiveRoomChat{ID: chatID}, nil)
	chatRepo.On("Delete", ctx, chatID).Return(nil)

	svc := newServiceWithAuth(chatRepo, nil, stubModelAuth{moderate: true})

	err := svc.DeleteChat(ctx, chatID)
	assert.NoError(t, err)
	chatRepo.AssertExpectations(t)
}

func TestListChats_NonAdmin_NilRoom_Forbidden(t *testing.T) {
	ctx := memberCtx(uuid.New())

	svc := newServiceWithAuth(&mockChatRepo{}, nil, stubModelAuth{participate: true})

	_, _, err := svc.ListChats(ctx, domain.ListChatsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListChats_NonAdmin_UnauthorizedRoom_Forbidden(t *testing.T) {
	ctx := memberCtx(uuid.New())
	roomID := uuid.New()

	svc := newServiceWithAuth(&mockChatRepo{}, nil, stubModelAuth{participate: false})

	_, _, err := svc.ListChats(ctx, domain.ListChatsQuery{LiveRoomID: &roomID})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListChats_NonAdmin_AuthorizedRoom_ReturnsList(t *testing.T) {
	ctx := memberCtx(uuid.New())
	roomID := uuid.New()
	q := domain.ListChatsQuery{LiveRoomID: &roomID}

	chatRepo := &mockChatRepo{}
	chatRepo.On("List", ctx, q).Return([]domain.LiveRoomChat{{ID: uuid.New()}}, int64(1), nil)

	svc := newServiceWithAuth(chatRepo, nil, stubModelAuth{participate: true})

	list, total, err := svc.ListChats(ctx, q)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), total)
}

func TestListChats_Admin_ReturnsList(t *testing.T) {
	ctx := adminCtx()
	q := domain.ListChatsQuery{}

	chatRepo := &mockChatRepo{}
	chatRepo.On("List", ctx, q).Return([]domain.LiveRoomChat{{ID: uuid.New()}}, int64(1), nil)

	svc := newServiceWithAuth(chatRepo, nil, stubModelAuth{})

	list, total, err := svc.ListChats(ctx, q)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), total)
}
