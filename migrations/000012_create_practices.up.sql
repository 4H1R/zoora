CREATE TABLE practice_rooms (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    class_id UUID NOT NULL,
    class_session_id UUID NOT NULL,
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    max_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_practice_rooms_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_practice_rooms_class FOREIGN KEY (class_id) REFERENCES classes (id) ON DELETE CASCADE,
    CONSTRAINT fk_practice_rooms_class_session FOREIGN KEY (class_session_id) REFERENCES class_sessions (id) ON DELETE CASCADE,
    CONSTRAINT fk_practice_rooms_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT chk_practice_rooms_time CHECK (end_time > start_time)
);

CREATE INDEX idx_practice_rooms_organization_id ON practice_rooms (organization_id);
CREATE INDEX idx_practice_rooms_class_id ON practice_rooms (class_id);
CREATE INDEX idx_practice_rooms_class_session_id ON practice_rooms (class_session_id);
CREATE INDEX idx_practice_rooms_user_id ON practice_rooms (user_id);
CREATE INDEX idx_practice_rooms_deleted_at ON practice_rooms (deleted_at);

CREATE TABLE practice_submissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    practice_room_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    score DOUBLE PRECISION,
    teacher_comment TEXT NOT NULL DEFAULT '',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_practice_submissions_room FOREIGN KEY (practice_room_id) REFERENCES practice_rooms (id) ON DELETE CASCADE,
    CONSTRAINT fk_practice_submissions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_practice_submissions_room_user UNIQUE (practice_room_id, user_id)
);

CREATE INDEX idx_practice_submissions_practice_room_id ON practice_submissions (practice_room_id);
CREATE INDEX idx_practice_submissions_user_id ON practice_submissions (user_id);
CREATE INDEX idx_practice_submissions_deleted_at ON practice_submissions (deleted_at);
