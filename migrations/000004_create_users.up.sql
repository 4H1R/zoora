CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    organization_id UUID,
    role_id UUID,
    username VARCHAR(255) NOT NULL,
    is_admin BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    disabled_by UUID,
    disabled_reason TEXT,
    -- Per-user changelog "seen" marker. Defaults to row creation time so a fresh
    -- signup is considered caught-up on everything before they joined.
    changelog_last_seen_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_users_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE SET NULL,
    CONSTRAINT fk_users_role_id FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE SET NULL
);

CREATE INDEX idx_users_deleted_at ON users (deleted_at);

-- Single role per user via FK.
CREATE INDEX idx_users_role_id ON users (role_id);

-- Backs per-org member counts: COUNT(*) WHERE organization_id = ? AND deleted_at IS NULL.
CREATE INDEX idx_users_organization_id ON users (organization_id) WHERE deleted_at IS NULL;

-- Backs the disabled filter on user lists.
CREATE INDEX idx_users_disabled_at ON users (disabled_at) WHERE deleted_at IS NULL;

-- Username is unique per organization. COALESCE maps NULL org_id to a sentinel
-- so two users without an org cannot share the same username.
CREATE UNIQUE INDEX idx_users_org_username
    ON users (COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'), username)
    WHERE deleted_at IS NULL;
