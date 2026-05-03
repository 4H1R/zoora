CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    organization_id UUID,
    username VARCHAR(255) NOT NULL,
    is_admin BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_deleted_at ON users (deleted_at);

-- Username is unique per organization. COALESCE maps NULL org_id to a sentinel
-- so two users without an org cannot share the same username.
CREATE UNIQUE INDEX idx_users_org_username
    ON users (COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'), username)
    WHERE deleted_at IS NULL;
