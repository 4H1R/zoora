CREATE TABLE live_rooms (
    id                UUID PRIMARY KEY DEFAULT uuidv7(),
    class_session_id  UUID NOT NULL,
    name                 VARCHAR(255) NOT NULL DEFAULT '',
    livekit_room_name VARCHAR(255) NOT NULL,
    scheduled_start_time TIMESTAMPTZ,
    status            VARCHAR(20) NOT NULL DEFAULT 'created',
    config            JSONB NOT NULL DEFAULT '{"allow_mic_default":true,"allow_camera_default":true,"allow_screen_share_default":false,"auto_record":false,"max_participants":100}',
    actual_start_time TIMESTAMPTZ,
    actual_end_time   TIMESTAMPTZ,
    host_last_seen_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    CONSTRAINT fk_live_rooms_class_session FOREIGN KEY (class_session_id) REFERENCES class_sessions (id) ON DELETE CASCADE,
    CONSTRAINT uq_live_rooms_livekit_room_name UNIQUE (livekit_room_name),
    CONSTRAINT chk_live_rooms_status CHECK (status IN ('created', 'active', 'finished'))
);

CREATE INDEX idx_live_rooms_class_session_id ON live_rooms (class_session_id);
CREATE INDEX idx_live_rooms_status ON live_rooms (status);
CREATE INDEX idx_live_rooms_deleted_at ON live_rooms (deleted_at);

CREATE TABLE live_participants (
    id                     UUID PRIMARY KEY DEFAULT uuidv7(),
    live_room_id           UUID NOT NULL,
    user_id                UUID NOT NULL,
    identity               VARCHAR(255) NOT NULL,
    joined_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at                TIMESTAMPTZ,
    total_duration_seconds INT NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_live_participants_room FOREIGN KEY (live_room_id) REFERENCES live_rooms (id) ON DELETE CASCADE,
    CONSTRAINT fk_live_participants_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_live_participants_room_id ON live_participants (live_room_id);
CREATE INDEX idx_live_participants_user_id ON live_participants (user_id);
CREATE INDEX idx_live_participants_active ON live_participants (live_room_id, user_id) WHERE left_at IS NULL;

CREATE TABLE live_recordings (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    live_room_id UUID NOT NULL,
    egress_id    VARCHAR(255) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'started',
    file_url     TEXT NOT NULL DEFAULT '',
    duration     INT NOT NULL DEFAULT 0,
    size         BIGINT NOT NULL DEFAULT 0,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at     TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_live_recordings_room FOREIGN KEY (live_room_id) REFERENCES live_rooms (id) ON DELETE CASCADE,
    CONSTRAINT chk_live_recordings_status CHECK (status IN ('started', 'completed', 'failed'))
);

CREATE INDEX idx_live_recordings_room_id ON live_recordings (live_room_id);
CREATE INDEX idx_live_recordings_status ON live_recordings (status);
