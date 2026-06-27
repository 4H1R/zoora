CREATE TABLE live_whiteboards (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    live_room_id UUID NOT NULL,
    snapshot     JSONB NOT NULL DEFAULT '{}',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_live_whiteboards_room FOREIGN KEY (live_room_id) REFERENCES live_rooms (id) ON DELETE CASCADE,
    CONSTRAINT uq_live_whiteboards_room UNIQUE (live_room_id)
);
