CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    class_id UUID NOT NULL REFERENCES classes(id),
    class_session_id UUID NOT NULL REFERENCES class_sessions(id),
    user_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('present', 'absent', 'late', 'excused')),
    is_auto_marked BOOLEAN NOT NULL DEFAULT FALSE,
    remarks TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT attendances_session_user_unique UNIQUE (class_session_id, user_id)
);

CREATE INDEX idx_attendances_organization_id ON attendances(organization_id);
CREATE INDEX idx_attendances_class_id ON attendances(class_id);
CREATE INDEX idx_attendances_class_session_id ON attendances(class_session_id);
CREATE INDEX idx_attendances_user_id ON attendances(user_id);
