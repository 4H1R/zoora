-- Collapse polymorphic chat into dedicated live-room chat.
-- model_id already holds a live_rooms.id for every existing row (only
-- 'live_session' chats were ever created), so the FK backfill is a rename.

-- 1. Drop the machinery we no longer use.
DROP TABLE IF EXISTS message_reactions;
DROP TABLE IF EXISTS chat_members;

-- 2. Rename core tables.
ALTER TABLE chats    RENAME TO liveroom_chats;
ALTER TABLE messages RENAME TO liveroom_messages;

-- 3. Replace polymorphism with a real FK to live_rooms.
-- NOTE: dropping model_type implicitly drops the composite
-- idx_chats_model (model_type, model_id) index, so we create the
-- single-column replacement fresh instead of renaming it.
ALTER TABLE liveroom_chats RENAME COLUMN model_id TO live_room_id;
ALTER TABLE liveroom_chats DROP COLUMN model_type;
ALTER TABLE liveroom_chats
  ADD CONSTRAINT fk_liveroom_chats_room
  FOREIGN KEY (live_room_id) REFERENCES live_rooms (id) ON DELETE CASCADE;
CREATE INDEX idx_liveroom_chats_room ON liveroom_chats (live_room_id);

-- 4. Rename surviving indexes for clarity (Postgres keeps old names after table rename).
ALTER INDEX IF EXISTS idx_chats_deleted_at    RENAME TO idx_liveroom_chats_deleted_at;
ALTER INDEX IF EXISTS idx_chats_status        RENAME TO idx_liveroom_chats_status;
ALTER INDEX IF EXISTS idx_messages_chat_id    RENAME TO idx_liveroom_messages_chat_id;
ALTER INDEX IF EXISTS idx_messages_sender_id  RENAME TO idx_liveroom_messages_sender_id;
ALTER INDEX IF EXISTS idx_messages_parent_message_id RENAME TO idx_liveroom_messages_parent_id;
ALTER INDEX IF EXISTS idx_messages_deleted_at RENAME TO idx_liveroom_messages_deleted_at;
ALTER INDEX IF EXISTS idx_messages_chat_created RENAME TO idx_liveroom_messages_chat_created;
