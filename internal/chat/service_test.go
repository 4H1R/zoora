package chat_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/chat"
	"github.com/4H1R/zoora/internal/domain"
)

type mockChatRepo struct{ mock.Mock }

func (m *mockChatRepo) Create(ctx context.Context, c *domain.Chat) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockChatRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}
func (m *mockChatRepo) Update(ctx context.Context, c *domain.Chat) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockChatRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockChatRepo) List(ctx context.Context, q domain.ListChatsQuery) ([]domain.Chat, int64, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]domain.Chat), args.Get(1).(int64), args.Error(2)
}
func (m *mockChatRepo) FindByModel(ctx context.Context, modelType string, modelID uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, modelType, modelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

type mockMemberRepo struct{ mock.Mock }

func (m *mockMemberRepo) Create(ctx context.Context, member *domain.ChatMember) error {
	return m.Called(ctx, member).Error(0)
}
func (m *mockMemberRepo) FindByChatAndUser(ctx context.Context, chatID, userID uuid.UUID) (*domain.ChatMember, error) {
	args := m.Called(ctx, chatID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChatMember), args.Error(1)
}
func (m *mockMemberRepo) Delete(ctx context.Context, chatID, userID uuid.UUID) error {
	return m.Called(ctx, chatID, userID).Error(0)
}
func (m *mockMemberRepo) ListByChat(ctx context.Context, chatID uuid.UUID) ([]domain.ChatMember, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]domain.ChatMember), args.Error(1)
}

type mockMessageRepo struct{ mock.Mock }

func (m *mockMessageRepo) Create(ctx context.Context, msg *domain.Message) error {
	return m.Called(ctx, msg).Error(0)
}
func (m *mockMessageRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}
func (m *mockMessageRepo) Update(ctx context.Context, msg *domain.Message) error {
	return m.Called(ctx, msg).Error(0)
}
func (m *mockMessageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockMessageRepo) List(ctx context.Context, chatID uuid.UUID, q domain.ListMessagesQuery) ([]domain.Message, int64, error) {
	args := m.Called(ctx, chatID, q)
	return args.Get(0).([]domain.Message), args.Get(1).(int64), args.Error(2)
}

type mockReactionRepo struct{ mock.Mock }

func (m *mockReactionRepo) Create(ctx context.Context, r *domain.MessageReaction) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mockReactionRepo) Delete(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	return m.Called(ctx, messageID, userID, emoji).Error(0)
}
func (m *mockReactionRepo) FindByMessageAndUser(ctx context.Context, messageID, userID uuid.UUID, emoji string) (*domain.MessageReaction, error) {
	args := m.Called(ctx, messageID, userID, emoji)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageReaction), args.Error(1)
}
func (m *mockReactionRepo) CountByMessage(ctx context.Context, messageID uuid.UUID) (map[string]int, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(map[string]int), args.Error(1)
}

func adminCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), IsAdmin: true})
}

func memberCtx(userID uuid.UUID) context.Context {
	orgID := uuid.New()
	return domain.WithCaller(context.Background(), domain.Caller{UserID: userID, OrgID: &orgID})
}

type noopTx struct{}

func (noopTx) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func newService(
	chatRepo *mockChatRepo,
	memberRepo *mockMemberRepo,
	msgRepo *mockMessageRepo,
	reactionRepo *mockReactionRepo,
) domain.ChatService {
	return chat.NewService(chatRepo, memberRepo, msgRepo, reactionRepo, noopTx{}, slog.Default())
}

func TestCreateChat_AdminSuccess(t *testing.T) {
	ctx := adminCtx()
	chatRepo := &mockChatRepo{}
	chatRepo.On("Create", ctx, mock.AnythingOfType("*domain.Chat")).Return(nil)

	svc := newService(chatRepo, nil, nil, nil)
	modelID := uuid.New()

	c, err := svc.CreateChat(ctx, domain.CreateChatDTO{
		Name:      "Test Chat",
		ModelType: "live_session",
		ModelID:   modelID.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, "Test Chat", c.Name)
	assert.Equal(t, domain.ChatStatusActive, c.Status)
	chatRepo.AssertExpectations(t)
}

func TestCreateChat_NoCaller_Forbidden(t *testing.T) {
	svc := newService(&mockChatRepo{}, nil, nil, nil)
	_, err := svc.CreateChat(context.Background(), domain.CreateChatDTO{
		Name:      "Test",
		ModelType: "live_session",
		ModelID:   uuid.New().String(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateChat_NonAdmin_Forbidden(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)

	svc := newService(&mockChatRepo{}, nil, nil, nil)
	_, err := svc.CreateChat(ctx, domain.CreateChatDTO{
		Name:      "Test",
		ModelType: "live_session",
		ModelID:   uuid.New().String(),
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetChat_LiveSession_NoMembershipRequired(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "live_session",
		Status:    domain.ChatStatusActive,
	}, nil)

	svc := newService(chatRepo, nil, nil, nil)

	c, err := svc.GetChat(ctx, chatID)
	assert.NoError(t, err)
	assert.Equal(t, chatID, c.ID)
}

func TestGetChat_SupportTicket_RequiresMembership(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	memberRepo := &mockMemberRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "support_ticket",
		Status:    domain.ChatStatusActive,
	}, nil)
	memberRepo.On("FindByChatAndUser", ctx, chatID, userID).Return(nil, domain.ErrNotFound)

	svc := newService(chatRepo, memberRepo, nil, nil)

	_, err := svc.GetChat(ctx, chatID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetChat_SupportTicket_MemberAllowed(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	memberRepo := &mockMemberRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "support_ticket",
		Status:    domain.ChatStatusActive,
	}, nil)
	memberRepo.On("FindByChatAndUser", ctx, chatID, userID).Return(&domain.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   domain.ChatMemberRoleMember,
	}, nil)

	svc := newService(chatRepo, memberRepo, nil, nil)

	c, err := svc.GetChat(ctx, chatID)
	assert.NoError(t, err)
	assert.Equal(t, chatID, c.ID)
}

func TestSendMessage_Success(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	memberRepo := &mockMemberRepo{}
	msgRepo := &mockMessageRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "support_ticket",
		Status:    domain.ChatStatusActive,
	}, nil)
	memberRepo.On("FindByChatAndUser", ctx, chatID, userID).Return(&domain.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   domain.ChatMemberRoleMember,
	}, nil)
	msgRepo.On("Create", ctx, mock.AnythingOfType("*domain.Message")).Return(nil)

	svc := newService(chatRepo, memberRepo, msgRepo, nil)

	msg, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.MessageTypeText,
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
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:     chatID,
		Status: domain.ChatStatusArchived,
	}, nil)

	svc := newService(chatRepo, nil, nil, nil)

	_, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.MessageTypeText,
		Content:     "Hello!",
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestSendMessage_ReadOnlyMember_Forbidden(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	memberRepo := &mockMemberRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "support_ticket",
		Status:    domain.ChatStatusActive,
	}, nil)
	memberRepo.On("FindByChatAndUser", ctx, chatID, userID).Return(&domain.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   domain.ChatMemberRoleReadOnly,
	}, nil)

	svc := newService(chatRepo, memberRepo, nil, nil)

	_, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.MessageTypeText,
		Content:     "Hello!",
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSendMessage_LiveSession_NoMembershipRequired(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	msgRepo := &mockMessageRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "live_session",
		Status:    domain.ChatStatusActive,
	}, nil)
	msgRepo.On("Create", ctx, mock.AnythingOfType("*domain.Message")).Return(nil)

	svc := newService(chatRepo, nil, msgRepo, nil)

	msg, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType: domain.MessageTypeText,
		Content:     "Hello from live session!",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Hello from live session!", msg.Content)
}

func TestUpdateMessage_OwnerAllowed(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.Message{
		ID:       msgID,
		SenderID: &userID,
		Content:  "old",
	}, nil)
	msgRepo.On("Update", ctx, mock.MatchedBy(func(m *domain.Message) bool {
		return m.Content == "new" && m.IsEdited
	})).Return(nil)

	svc := newService(nil, nil, msgRepo, nil)

	msg, err := svc.UpdateMessage(ctx, msgID, domain.UpdateMessageDTO{Content: "new"})
	assert.NoError(t, err)
	assert.Equal(t, "new", msg.Content)
	assert.True(t, msg.IsEdited)
}

func TestUpdateMessage_OtherUser_Forbidden(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.Message{
		ID:       msgID,
		SenderID: &otherUserID,
	}, nil)

	svc := newService(nil, nil, msgRepo, nil)

	_, err := svc.UpdateMessage(ctx, msgID, domain.UpdateMessageDTO{Content: "hack"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestToggleReaction_AddReaction(t *testing.T) {
	userID := uuid.New()
	ctx := adminCtx()
	// Override with known userID
	caller, _ := domain.CallerFromCtx(ctx)
	caller.UserID = userID
	ctx = domain.WithCaller(context.Background(), caller)

	msgID := uuid.New()
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	msgRepo := &mockMessageRepo{}
	reactionRepo := &mockReactionRepo{}

	msgRepo.On("FindByID", mock.Anything, msgID).Return(&domain.Message{
		ID:          msgID,
		ChatID:      chatID,
		EmojiCounts: json.RawMessage(`{}`),
	}, nil)
	chatRepo.On("FindByID", mock.Anything, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "live_session",
	}, nil)
	reactionRepo.On("FindByMessageAndUser", mock.Anything, msgID, userID, "👍").
		Return(nil, domain.ErrNotFound)
	reactionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MessageReaction")).Return(nil)
	reactionRepo.On("CountByMessage", mock.Anything, msgID).Return(map[string]int{"👍": 1}, nil)
	msgRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	svc := chat.NewService(chatRepo, nil, msgRepo, reactionRepo, noopTx{}, slog.Default())

	msg, err := svc.ToggleReaction(ctx, msgID, domain.ToggleReactionDTO{Emoji: "👍"})
	assert.NoError(t, err)

	var counts map[string]int
	_ = json.Unmarshal(msg.EmojiCounts, &counts)
	assert.Equal(t, 1, counts["👍"])
}

func TestToggleReaction_RemoveExistingReaction(t *testing.T) {
	userID := uuid.New()
	caller := domain.Caller{UserID: userID, IsAdmin: true}
	ctx := domain.WithCaller(context.Background(), caller)

	msgID := uuid.New()
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	msgRepo := &mockMessageRepo{}
	reactionRepo := &mockReactionRepo{}

	msgRepo.On("FindByID", mock.Anything, msgID).Return(&domain.Message{
		ID:          msgID,
		ChatID:      chatID,
		EmojiCounts: json.RawMessage(`{"👍":1}`),
	}, nil)
	chatRepo.On("FindByID", mock.Anything, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "live_session",
	}, nil)
	reactionRepo.On("FindByMessageAndUser", mock.Anything, msgID, userID, "👍").
		Return(&domain.MessageReaction{
			MessageID: msgID,
			UserID:    userID,
			Emoji:     "👍",
			CreatedAt: time.Now(),
		}, nil)
	reactionRepo.On("Delete", mock.Anything, msgID, userID, "👍").Return(nil)
	reactionRepo.On("CountByMessage", mock.Anything, msgID).Return(map[string]int{}, nil)
	msgRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Message")).Return(nil)

	svc := chat.NewService(chatRepo, nil, msgRepo, reactionRepo, noopTx{}, slog.Default())

	msg, err := svc.ToggleReaction(ctx, msgID, domain.ToggleReactionDTO{Emoji: "👍"})
	assert.NoError(t, err)

	var counts map[string]int
	_ = json.Unmarshal(msg.EmojiCounts, &counts)
	assert.Empty(t, counts)
	reactionRepo.AssertExpectations(t)
}

func TestDeleteMessage_SenderAllowed(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()

	msgRepo := &mockMessageRepo{}
	msgRepo.On("FindByID", ctx, msgID).Return(&domain.Message{
		ID:       msgID,
		SenderID: &userID,
	}, nil)
	msgRepo.On("Delete", ctx, msgID).Return(nil)

	svc := newService(nil, nil, msgRepo, nil)

	err := svc.DeleteMessage(ctx, msgID)
	assert.NoError(t, err)
}

func TestDeleteMessage_OtherUser_ChatAdmin_Allowed(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	ctx := memberCtx(userID)
	msgID := uuid.New()
	chatID := uuid.New()

	chatRepo := &mockChatRepo{}
	memberRepo := &mockMemberRepo{}
	msgRepo := &mockMessageRepo{}

	msgRepo.On("FindByID", ctx, msgID).Return(&domain.Message{
		ID:       msgID,
		ChatID:   chatID,
		SenderID: &otherUserID,
	}, nil)
	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "support_ticket",
	}, nil)
	memberRepo.On("FindByChatAndUser", ctx, chatID, userID).Return(&domain.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   domain.ChatMemberRoleAdmin,
	}, nil)
	msgRepo.On("Delete", ctx, msgID).Return(nil)

	svc := newService(chatRepo, memberRepo, msgRepo, nil)

	err := svc.DeleteMessage(ctx, msgID)
	assert.NoError(t, err)
}

func TestAddMember_Success(t *testing.T) {
	ctx := adminCtx()
	chatID := uuid.New()
	newUserID := uuid.New()

	chatRepo := &mockChatRepo{}
	memberRepo := &mockMemberRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{ID: chatID}, nil)
	memberRepo.On("Create", ctx, mock.AnythingOfType("*domain.ChatMember")).Return(nil)

	svc := newService(chatRepo, memberRepo, nil, nil)

	member, err := svc.AddMember(ctx, chatID, domain.AddChatMemberDTO{
		UserID: newUserID.String(),
		Role:   domain.ChatMemberRoleMember,
	})
	assert.NoError(t, err)
	assert.Equal(t, chatID, member.ChatID)
	assert.Equal(t, newUserID, member.UserID)
	assert.Equal(t, domain.ChatMemberRoleMember, member.Role)
}

func TestSendMessage_WithThread(t *testing.T) {
	userID := uuid.New()
	ctx := memberCtx(userID)
	chatID := uuid.New()
	parentMsgID := uuid.New()

	chatRepo := &mockChatRepo{}
	memberRepo := &mockMemberRepo{}
	msgRepo := &mockMessageRepo{}

	chatRepo.On("FindByID", ctx, chatID).Return(&domain.Chat{
		ID:        chatID,
		ModelType: "support_ticket",
		Status:    domain.ChatStatusActive,
	}, nil)
	memberRepo.On("FindByChatAndUser", ctx, chatID, userID).Return(&domain.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   domain.ChatMemberRoleMember,
	}, nil)
	msgRepo.On("FindByID", ctx, parentMsgID).Return(&domain.Message{
		ID:     parentMsgID,
		ChatID: chatID,
	}, nil)
	msgRepo.On("Create", ctx, mock.MatchedBy(func(m *domain.Message) bool {
		return m.ParentMessageID != nil && *m.ParentMessageID == parentMsgID
	})).Return(nil)

	svc := newService(chatRepo, memberRepo, msgRepo, nil)

	parentIDStr := parentMsgID.String()
	msg, err := svc.SendMessage(ctx, chatID, domain.SendMessageDTO{
		MessageType:     domain.MessageTypeText,
		Content:         "Reply",
		ParentMessageID: &parentIDStr,
	})
	assert.NoError(t, err)
	assert.Equal(t, &parentMsgID, msg.ParentMessageID)
}
