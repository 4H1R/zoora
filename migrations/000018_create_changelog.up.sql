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

-- NOTE: users.changelog_last_seen_at lives in 000004_create_users.
