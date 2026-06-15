ALTER TABLE live_rooms ADD COLUMN IF NOT EXISTS scheduled_start_time TIMESTAMPTZ;
