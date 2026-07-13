CREATE TABLE tickets (
    id                  UUID         PRIMARY KEY DEFAULT uuidv7(),
    organization_id     UUID         NOT NULL,
    class_id            UUID         NOT NULL,
    user_id             UUID         NOT NULL, -- creator (student)
    title               VARCHAR(255) NOT NULL,
    type                VARCHAR(20)  NOT NULL,
    -- grade_objection targets: at most one set; both NULL = general objection.
    quiz_room_id        UUID,
    gradebook_column_id UUID,
    status              VARCHAR(20)  NOT NULL DEFAULT 'open',
    closed_at           TIMESTAMPTZ,
    closed_by           UUID,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tickets_org           FOREIGN KEY (organization_id)     REFERENCES organizations (id)     ON DELETE CASCADE,
    CONSTRAINT fk_tickets_class         FOREIGN KEY (class_id)            REFERENCES classes (id)           ON DELETE CASCADE,
    CONSTRAINT fk_tickets_user          FOREIGN KEY (user_id)             REFERENCES users (id)             ON DELETE CASCADE,
    CONSTRAINT fk_tickets_quiz_room     FOREIGN KEY (quiz_room_id)        REFERENCES quiz_rooms (id)        ON DELETE SET NULL,
    CONSTRAINT fk_tickets_gradebook_col FOREIGN KEY (gradebook_column_id) REFERENCES gradebook_columns (id) ON DELETE SET NULL,
    CONSTRAINT fk_tickets_closed_by     FOREIGN KEY (closed_by)           REFERENCES users (id)             ON DELETE SET NULL,
    CONSTRAINT chk_tickets_type   CHECK (type IN ('question', 'grade_objection', 'other')),
    CONSTRAINT chk_tickets_status CHECK (status IN ('open', 'answered', 'closed')),
    -- targets only on grade_objection tickets
    CONSTRAINT chk_tickets_objection_target CHECK (
        type = 'grade_objection' OR (quiz_room_id IS NULL AND gradebook_column_id IS NULL)
    ),
    -- at most one target
    CONSTRAINT chk_tickets_one_target CHECK (quiz_room_id IS NULL OR gradebook_column_id IS NULL)
);

CREATE INDEX idx_tickets_org_status_updated ON tickets (organization_id, status, updated_at DESC);
CREATE INDEX idx_tickets_class        ON tickets (class_id);
CREATE INDEX idx_tickets_user         ON tickets (user_id);
CREATE INDEX idx_tickets_class_status ON tickets (class_id, status);

CREATE TABLE ticket_messages (
    id         UUID        PRIMARY KEY DEFAULT uuidv7(),
    ticket_id  UUID        NOT NULL,
    user_id    UUID        NOT NULL,
    body       TEXT        NOT NULL,
    media_ids  JSONB       NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ticket_messages_ticket FOREIGN KEY (ticket_id) REFERENCES tickets (id) ON DELETE CASCADE,
    CONSTRAINT fk_ticket_messages_user   FOREIGN KEY (user_id)   REFERENCES users (id)   ON DELETE CASCADE
);

-- thread read: oldest-first within a ticket (uuidv7 ids are time-ordered)
CREATE INDEX idx_ticket_messages_ticket ON ticket_messages (ticket_id, id);
CREATE INDEX idx_ticket_messages_user ON ticket_messages (user_id);
