CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    sender_id       UUID REFERENCES users (id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations (id) ON DELETE CASCADE,
    category        VARCHAR(20)  NOT NULL,
    title           VARCHAR(255) NOT NULL,
    body            TEXT         NOT NULL,
    action_url      VARCHAR(500),
    audience        JSONB        NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Sender history ("my sent notifications"), newest first.
CREATE INDEX idx_notifications_sender ON notifications (sender_id, created_at DESC);

CREATE TABLE notification_recipients (
    notification_id UUID        NOT NULL REFERENCES notifications (id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (notification_id, user_id)
);

-- Inbox feed: one user's rows, newest first.
CREATE INDEX idx_notification_recipients_user ON notification_recipients (user_id, created_at DESC);
-- Unread badge count: partial index keeps it cheap.
CREATE INDEX idx_notification_recipients_unread ON notification_recipients (user_id) WHERE read_at IS NULL;
