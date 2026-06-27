ALTER TABLE live_participants
    ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    ADD COLUMN hand_raised_at TIMESTAMPTZ;

ALTER TABLE live_participants
    ADD CONSTRAINT chk_live_participants_role CHECK (role IN ('host', 'presenter', 'viewer'));

CREATE INDEX idx_live_participants_role ON live_participants (live_room_id, role);
