CREATE TABLE conversations (
    id              UUID         PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID         NOT NULL,
    type            VARCHAR(20)  NOT NULL,
    name            VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url      TEXT         NOT NULL DEFAULT '',
    color_index     SMALLINT     NOT NULL DEFAULT 0,
    created_by      UUID,
    -- direct_key: sorted "minUser:maxUser" for direct convos, NULL otherwise.
    -- Enforces one DM per user pair; NULLs are exempt from the unique index.
    direct_key      TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_conversations_org FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_conversations_creator FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT chk_conversations_type CHECK (type IN ('direct', 'group', 'channel'))
);

CREATE INDEX idx_conversations_org ON conversations (organization_id);
CREATE INDEX idx_conversations_created_by ON conversations (created_by);
CREATE UNIQUE INDEX uq_conversations_direct_key ON conversations (organization_id, direct_key) WHERE direct_key IS NOT NULL;

CREATE TABLE conversation_members (
    id                   UUID        PRIMARY KEY DEFAULT uuidv7(),
    conversation_id      UUID        NOT NULL,
    user_id              UUID        NOT NULL,
    role                 VARCHAR(20) NOT NULL DEFAULT 'member',
    -- last_read_message_id has NO FK on purpose: messages hard-delete, and a
    -- stale pointer is harmless (unread is a keyset compare on uuidv7 ids).
    last_read_message_id UUID,
    last_read_at         TIMESTAMPTZ,
    muted_until          TIMESTAMPTZ,
    joined_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_conv_members_conv FOREIGN KEY (conversation_id) REFERENCES conversations (id) ON DELETE CASCADE,
    CONSTRAINT fk_conv_members_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_conv_members_conv_user UNIQUE (conversation_id, user_id),
    CONSTRAINT chk_conv_members_role CHECK (role IN ('admin', 'member'))
);

CREATE INDEX idx_conv_members_conv ON conversation_members (conversation_id);
CREATE INDEX idx_conv_members_user ON conversation_members (user_id);

CREATE TABLE conversation_messages (
    id                   UUID        PRIMARY KEY DEFAULT uuidv7(),
    conversation_id      UUID        NOT NULL,
    sender_id            UUID,
    reply_to_message_id  UUID,
    content              TEXT        NOT NULL DEFAULT '',
    is_edited            BOOLEAN     NOT NULL DEFAULT FALSE,
    is_pinned            BOOLEAN     NOT NULL DEFAULT FALSE,
    pinned_by            UUID,
    pinned_at            TIMESTAMPTZ,
    media_ids            JSONB       NOT NULL DEFAULT '[]',
    as_document          BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_conv_messages_conv FOREIGN KEY (conversation_id) REFERENCES conversations (id) ON DELETE CASCADE,
    CONSTRAINT fk_conv_messages_sender FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT fk_conv_messages_reply FOREIGN KEY (reply_to_message_id) REFERENCES conversation_messages (id) ON DELETE SET NULL,
    CONSTRAINT fk_conv_messages_pinned_by FOREIGN KEY (pinned_by) REFERENCES users (id) ON DELETE SET NULL
);

-- Keyset scrollback: newest-first within a conversation.
CREATE INDEX idx_conv_messages_conv_id ON conversation_messages (conversation_id, id DESC);
CREATE INDEX idx_conv_messages_sender ON conversation_messages (sender_id);
CREATE INDEX idx_conv_messages_reply ON conversation_messages (reply_to_message_id);
CREATE INDEX idx_conv_messages_pinned_by ON conversation_messages (pinned_by);
CREATE INDEX idx_conv_messages_pinned ON conversation_messages (conversation_id) WHERE is_pinned;

CREATE TABLE conversation_message_reactions (
    id         UUID        PRIMARY KEY DEFAULT uuidv7(),
    message_id UUID        NOT NULL,
    user_id    UUID        NOT NULL,
    emoji      VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_conv_reactions_msg FOREIGN KEY (message_id) REFERENCES conversation_messages (id) ON DELETE CASCADE,
    CONSTRAINT fk_conv_reactions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_conv_reactions_msg_user_emoji UNIQUE (message_id, user_id, emoji)
);

CREATE INDEX idx_conv_reactions_message ON conversation_message_reactions (message_id);
