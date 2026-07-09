CREATE TABLE liveroom_chats (
    id           UUID         PRIMARY KEY DEFAULT uuidv7(),
    name         VARCHAR(255) NOT NULL,
    description  TEXT         NOT NULL DEFAULT '',
    live_room_id UUID         NOT NULL REFERENCES live_rooms (id) ON DELETE CASCADE,
    status       VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    CONSTRAINT chk_liveroom_chats_status CHECK (status IN ('active', 'archived'))
);

CREATE INDEX idx_liveroom_chats_room ON liveroom_chats (live_room_id);
CREATE INDEX idx_liveroom_chats_deleted_at ON liveroom_chats (deleted_at);
CREATE INDEX idx_liveroom_chats_status ON liveroom_chats (status) WHERE deleted_at IS NULL;

CREATE TABLE liveroom_messages (
    id                UUID        PRIMARY KEY DEFAULT uuidv7(),
    chat_id           UUID        NOT NULL REFERENCES liveroom_chats(id) ON DELETE CASCADE,
    sender_id         UUID        REFERENCES users(id) ON DELETE SET NULL,
    parent_message_id UUID        REFERENCES liveroom_messages(id) ON DELETE SET NULL,
    message_type      VARCHAR(20) NOT NULL DEFAULT 'text',
    content           TEXT        NOT NULL DEFAULT '',
    attachments       JSONB       NOT NULL DEFAULT '[]',
    is_edited         BOOLEAN     NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    CONSTRAINT chk_liveroom_messages_type CHECK (message_type IN ('text', 'file', 'system'))
);

CREATE INDEX idx_liveroom_messages_chat_id ON liveroom_messages (chat_id);
CREATE INDEX idx_liveroom_messages_sender_id ON liveroom_messages (sender_id);
CREATE INDEX idx_liveroom_messages_parent_id ON liveroom_messages (parent_message_id);
CREATE INDEX idx_liveroom_messages_deleted_at ON liveroom_messages (deleted_at);
CREATE INDEX idx_liveroom_messages_chat_created ON liveroom_messages (chat_id, created_at DESC) WHERE deleted_at IS NULL;
