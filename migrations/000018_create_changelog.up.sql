CREATE TABLE changelog_entries (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    version      VARCHAR(50),
    title_en     VARCHAR(255) NOT NULL,
    title_fa     VARCHAR(255) NOT NULL DEFAULT '',
    body_en      TEXT         NOT NULL,
    body_fa      TEXT         NOT NULL DEFAULT '',
    is_major     BOOLEAN      NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Feed ordering + "latest" lookups: newest published first.
CREATE INDEX idx_changelog_published ON changelog_entries (published_at DESC, id);

-- Per-user "seen" marker. New rows default to their creation time (a fresh
-- signup is considered caught-up on everything before they joined).
ALTER TABLE users ADD COLUMN changelog_last_seen_at TIMESTAMPTZ DEFAULT NOW();

-- Backfill existing users so nobody gets a wall of historical entries on launch.
UPDATE users SET changelog_last_seen_at = NOW() WHERE changelog_last_seen_at IS NULL;
