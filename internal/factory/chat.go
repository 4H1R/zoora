package factory

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewChat(modelType string, modelID uuid.UUID, opts ...func(*domain.Chat)) *domain.Chat {
	id := nextID()
	c := &domain.Chat{
		Name:        fakeChatName(id),
		Description: fakeSentence(6),
		ModelType:   modelType,
		ModelID:     modelID,
		Status:      domain.ChatStatusActive,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func NewChatMember(chatID, userID uuid.UUID, role domain.ChatMemberRole, opts ...func(*domain.ChatMember)) *domain.ChatMember {
	m := &domain.ChatMember{
		ChatID:   chatID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

func NewMessage(chatID uuid.UUID, senderID *uuid.UUID, opts ...func(*domain.Message)) *domain.Message {
	m := &domain.Message{
		ChatID:      chatID,
		SenderID:    senderID,
		MessageType: domain.MessageTypeText,
		Content:     fakeSentence(fake.IntRange(4, 12)),
		Attachments: json.RawMessage(`[]`),
		EmojiCounts: json.RawMessage(`{}`),
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

func NewMessageReaction(messageID, userID uuid.UUID, emoji string) *domain.MessageReaction {
	return &domain.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}
}
