CREATE TABLE chats (
    id          UUID        PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(255) NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    model_type  VARCHAR(100) NOT NULL,
    model_id    UUID        NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT chk_chats_status CHECK (status IN ('active', 'archived'))
);

CREATE INDEX idx_chats_model ON chats (model_type, model_id);
CREATE INDEX idx_chats_deleted_at ON chats (deleted_at);
CREATE INDEX idx_chats_status ON chats (status) WHERE deleted_at IS NULL;

CREATE TABLE chat_members (
    id        UUID        PRIMARY KEY DEFAULT uuidv7(),
    chat_id   UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role      VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_chat_members_chat_user UNIQUE (chat_id, user_id),
    CONSTRAINT chk_chat_members_role CHECK (role IN ('admin', 'member', 'read_only'))
);

CREATE INDEX idx_chat_members_chat_id ON chat_members (chat_id);
CREATE INDEX idx_chat_members_user_id ON chat_members (user_id);

CREATE TABLE messages (
    id                UUID        PRIMARY KEY DEFAULT uuidv7(),
    chat_id           UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id         UUID        REFERENCES users(id) ON DELETE SET NULL,
    parent_message_id UUID        REFERENCES messages(id) ON DELETE SET NULL,
    message_type      VARCHAR(20) NOT NULL DEFAULT 'text',
    content           TEXT        NOT NULL DEFAULT '',
    attachments       JSONB       NOT NULL DEFAULT '[]',
    is_edited         BOOLEAN     NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    CONSTRAINT chk_messages_type CHECK (message_type IN ('text', 'file', 'system'))
);

CREATE INDEX idx_messages_chat_id ON messages (chat_id);
CREATE INDEX idx_messages_sender_id ON messages (sender_id);
CREATE INDEX idx_messages_parent_message_id ON messages (parent_message_id);
CREATE INDEX idx_messages_deleted_at ON messages (deleted_at);
CREATE INDEX idx_messages_chat_created ON messages (chat_id, created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE message_reactions (
    id         UUID        PRIMARY KEY DEFAULT uuidv7(),
    message_id UUID        NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji      VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_message_reactions_msg_user_emoji UNIQUE (message_id, user_id, emoji)
);

CREATE INDEX idx_message_reactions_message_id ON message_reactions (message_id);
