CREATE TABLE polls (
    id                   UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id              UUID NOT NULL,
    model_type           VARCHAR(100) NOT NULL,
    model_id             UUID NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    allowed_answers_count INTEGER NOT NULL DEFAULT 1,
    options              JSONB NOT NULL DEFAULT '[]',
    closed_at            TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ,
    CONSTRAINT fk_polls_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_polls_user_id ON polls (user_id);
CREATE INDEX idx_polls_model ON polls (model_type, model_id);
CREATE INDEX idx_polls_deleted_at ON polls (deleted_at);

CREATE TABLE poll_answers (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id    UUID NOT NULL,
    poll_id    UUID NOT NULL,
    option     VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_poll_answers_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_poll_answers_poll FOREIGN KEY (poll_id) REFERENCES polls (id) ON DELETE CASCADE
);

CREATE INDEX idx_poll_answers_user_id ON poll_answers (user_id);
CREATE INDEX idx_poll_answers_poll_option ON poll_answers (poll_id, option);
CREATE UNIQUE INDEX idx_poll_answers_poll_user_option ON poll_answers (poll_id, user_id, option);
