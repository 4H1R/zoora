-- (users.avatar intentionally NOT added — deferred. conversations.avatar_url
--  already exists from migration 000024.)

-- Full-text search over message content (plaintext, 'simple' config = language-agnostic).
CREATE INDEX idx_conv_messages_content_fts
  ON conversation_messages USING GIN (to_tsvector('simple', content));

-- Mentions ("mentions of me" + notification fan-out).
CREATE TABLE conversation_mentions (
    id         UUID        PRIMARY KEY DEFAULT uuidv7(),
    message_id UUID        NOT NULL,
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_conv_mentions_msg FOREIGN KEY (message_id) REFERENCES conversation_messages (id) ON DELETE CASCADE,
    CONSTRAINT fk_conv_mentions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_conv_mentions_msg_user UNIQUE (message_id, user_id)
);
CREATE INDEX idx_conv_mentions_user ON conversation_mentions (user_id);
