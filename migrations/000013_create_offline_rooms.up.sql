CREATE TABLE offline_rooms (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    class_id UUID NOT NULL,
    class_session_id UUID NOT NULL,
    creator_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    view_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_offline_rooms_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_offline_rooms_class FOREIGN KEY (class_id) REFERENCES classes (id) ON DELETE CASCADE,
    CONSTRAINT fk_offline_rooms_class_session FOREIGN KEY (class_session_id) REFERENCES class_sessions (id) ON DELETE CASCADE,
    CONSTRAINT fk_offline_rooms_creator FOREIGN KEY (creator_id) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE INDEX idx_offline_rooms_organization_id ON offline_rooms (organization_id);
CREATE INDEX idx_offline_rooms_class_id ON offline_rooms (class_id);
CREATE INDEX idx_offline_rooms_class_session_id ON offline_rooms (class_session_id);
CREATE INDEX idx_offline_rooms_creator_id ON offline_rooms (creator_id);
CREATE INDEX idx_offline_rooms_deleted_at ON offline_rooms (deleted_at);

CREATE TABLE offline_room_views (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    offline_room_id UUID NOT NULL REFERENCES offline_rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    duration_seconds INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_offline_room_views_room_id ON offline_room_views(offline_room_id);
CREATE INDEX idx_offline_room_views_user_id ON offline_room_views(user_id);
