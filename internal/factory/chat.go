package factory

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewLiveRoomChat(liveRoomID uuid.UUID, opts ...func(*domain.LiveRoomChat)) *domain.LiveRoomChat {
	id := nextID()
	c := &domain.LiveRoomChat{
		Name:        fakeChatName(id),
		Description: fakeSentence(6),
		LiveRoomID:  liveRoomID,
		Status:      domain.LiveRoomChatStatusActive,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func NewLiveRoomMessage(chatID uuid.UUID, senderID *uuid.UUID, opts ...func(*domain.LiveRoomMessage)) *domain.LiveRoomMessage {
	m := &domain.LiveRoomMessage{
		ChatID:      chatID,
		SenderID:    senderID,
		MessageType: domain.LiveRoomMessageTypeText,
		Content:     fakeSentence(fake.IntRange(4, 12)),
		Attachments: json.RawMessage(`[]`),
	}
	for _, o := range opts {
		o(m)
	}
	return m
}
