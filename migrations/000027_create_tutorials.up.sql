CREATE TABLE tutorials (
    id             UUID PRIMARY KEY DEFAULT uuidv7(),
    title_en       VARCHAR(255) NOT NULL,
    title_fa       VARCHAR(255) NOT NULL DEFAULT '',
    description_en TEXT         NOT NULL DEFAULT '',
    description_fa TEXT         NOT NULL DEFAULT '',
    aparat_hash    VARCHAR(64)  NOT NULL,
    thumbnail_url  TEXT         NOT NULL DEFAULT '',
    position       INTEGER      NOT NULL DEFAULT 0,
    published_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Grid ordering: published library is read in curated (position ASC) order.
CREATE INDEX idx_tutorials_position ON tutorials (position, id);
