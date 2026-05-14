CREATE TABLE classes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    total_users INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_classes_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_classes_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE INDEX idx_classes_organization_id ON classes (organization_id);
CREATE INDEX idx_classes_user_id ON classes (user_id);
CREATE INDEX idx_classes_deleted_at ON classes (deleted_at);

CREATE TABLE class_sessions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    class_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    start_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_class_sessions_class FOREIGN KEY (class_id) REFERENCES classes (id) ON DELETE CASCADE
);

CREATE INDEX idx_class_sessions_class_id ON class_sessions (class_id);
CREATE INDEX idx_class_sessions_start_time ON class_sessions (start_time);
CREATE INDEX idx_class_sessions_deleted_at ON class_sessions (deleted_at);

CREATE TABLE class_members (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    class_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_class_members_class FOREIGN KEY (class_id) REFERENCES classes (id) ON DELETE CASCADE,
    CONSTRAINT fk_class_members_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_class_members_class_user UNIQUE (class_id, user_id)
);

CREATE INDEX idx_class_members_class_id ON class_members (class_id);
CREATE INDEX idx_class_members_user_id ON class_members (user_id);
