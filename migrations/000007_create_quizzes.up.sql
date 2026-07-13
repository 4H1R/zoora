CREATE TABLE quizzes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    class_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_minutes INTEGER NOT NULL,
    total_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    no_back_navigation BOOLEAN NOT NULL DEFAULT FALSE,
    shuffle_questions BOOLEAN NOT NULL DEFAULT FALSE,
    shuffle_options BOOLEAN NOT NULL DEFAULT FALSE,
    track_tab_switches BOOLEAN NOT NULL DEFAULT FALSE,
    require_gps BOOLEAN NOT NULL DEFAULT FALSE,
    disable_copy_paste BOOLEAN NOT NULL DEFAULT FALSE,
    disable_right_click_shortcuts BOOLEAN NOT NULL DEFAULT FALSE,
    negative_mark_mode VARCHAR(20) NOT NULL DEFAULT 'none',
    negative_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    wrongs_per_point INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_quizzes_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_quizzes_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_quizzes_class FOREIGN KEY (class_id) REFERENCES classes (id) ON DELETE CASCADE
);

CREATE INDEX idx_quizzes_organization_id ON quizzes (organization_id);
CREATE INDEX idx_quizzes_user_id ON quizzes (user_id);
CREATE INDEX idx_quizzes_class_id ON quizzes (class_id);
CREATE INDEX idx_quizzes_deleted_at ON quizzes (deleted_at);

CREATE TABLE quiz_rules (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    quiz_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    bank_id UUID,
    question_ids JSONB NOT NULL DEFAULT '[]',
    count INTEGER NOT NULL DEFAULT 0,
    is_dynamic BOOLEAN NOT NULL DEFAULT FALSE,
    negative_overrides JSONB NOT NULL DEFAULT '[]',
    negative_default_mode VARCHAR(20),
    negative_default_value DOUBLE PRECISION,
    negative_default_wrongs_per_point INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_quiz_rules_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes (id) ON DELETE CASCADE,
    CONSTRAINT fk_quiz_rules_bank FOREIGN KEY (bank_id) REFERENCES question_banks (id) ON DELETE SET NULL,
    CONSTRAINT chk_quiz_rules_type CHECK (type IN ('manual', 'random'))
);

CREATE INDEX idx_quiz_rules_quiz_id ON quiz_rules (quiz_id);
CREATE INDEX idx_quiz_rules_bank_id ON quiz_rules (bank_id);

CREATE TABLE quiz_rooms (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    quiz_id UUID NOT NULL,
    class_session_id UUID NOT NULL,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_quiz_rooms_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes (id) ON DELETE CASCADE,
    CONSTRAINT fk_quiz_rooms_class_session FOREIGN KEY (class_session_id) REFERENCES class_sessions (id) ON DELETE CASCADE
);

CREATE INDEX idx_quiz_rooms_quiz_id ON quiz_rooms (quiz_id);
CREATE INDEX idx_quiz_rooms_class_session_id ON quiz_rooms (class_session_id);

CREATE TABLE quiz_submissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    quiz_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    answers JSONB NOT NULL DEFAULT '[]',
    question_set JSONB NOT NULL DEFAULT '[]',
    quiz_room_id UUID REFERENCES quiz_rooms (id),
    tab_hidden_count INTEGER NOT NULL DEFAULT 0,
    tab_hidden_seconds INTEGER NOT NULL DEFAULT 0,
    gps_lat DOUBLE PRECISION,
    gps_lng DOUBLE PRECISION,
    gps_accuracy DOUBLE PRECISION,
    gps_denied BOOLEAN NOT NULL DEFAULT FALSE,
    total_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_quiz_submissions_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes (id) ON DELETE CASCADE,
    CONSTRAINT fk_quiz_submissions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_quiz_submissions_quiz_user UNIQUE (quiz_id, user_id),
    CONSTRAINT chk_quiz_submissions_status CHECK (status IN ('in_progress', 'submitted', 'graded'))
);

CREATE INDEX idx_quiz_submissions_quiz_status ON quiz_submissions (quiz_id, status);
CREATE INDEX idx_quiz_submissions_user_id ON quiz_submissions (user_id);
