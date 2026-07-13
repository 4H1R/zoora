CREATE TABLE user_connectors (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type        VARCHAR(20)  NOT NULL,
    target      VARCHAR(500) NOT NULL,
    verified_at TIMESTAMPTZ,
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, type, target)
);

CREATE INDEX idx_user_connectors_user ON user_connectors (user_id);
CREATE INDEX idx_user_connectors_type_target ON user_connectors (type, target);

CREATE TABLE notification_deliveries (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    notification_id UUID         NOT NULL REFERENCES notifications (id) ON DELETE CASCADE,
    user_id         UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    channel         VARCHAR(20)  NOT NULL,
    target          VARCHAR(500) NOT NULL,
    status          VARCHAR(10)  NOT NULL DEFAULT 'pending',
    error           TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    -- fan-out retries must not duplicate deliveries
    UNIQUE (notification_id, user_id, channel, target)
);

-- Delivery report aggregation per notification.
CREATE INDEX idx_notification_deliveries_notification ON notification_deliveries (notification_id, channel, status);
CREATE INDEX idx_notification_deliveries_user ON notification_deliveries (user_id);

-- NOTE: organization_settings.sms_enabled lives in 000016_create_organization_settings.
