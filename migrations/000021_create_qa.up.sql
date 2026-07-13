CREATE TABLE qa_questions (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id    UUID NOT NULL,
    model_type VARCHAR(100) NOT NULL,
    model_id   UUID NOT NULL,
    text       VARCHAR(500) NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'open',
    closed_at  TIMESTAMPTZ,
    closed_by  UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_qa_questions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_qa_questions_user_id ON qa_questions (user_id);
CREATE INDEX idx_qa_questions_model ON qa_questions (model_type, model_id);
CREATE INDEX idx_qa_questions_deleted_at ON qa_questions (deleted_at);

CREATE TABLE qa_votes (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    question_id UUID NOT NULL,
    user_id     UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_qa_votes_question FOREIGN KEY (question_id) REFERENCES qa_questions (id) ON DELETE CASCADE,
    CONSTRAINT fk_qa_votes_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_qa_votes_question_user ON qa_votes (question_id, user_id);
CREATE INDEX idx_qa_votes_user ON qa_votes (user_id);
