CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(63) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'trial', 'suspended', 'archived')),
    -- No CHECK constraint on plan: the catalog is code-defined (Plan.Valid()
    -- validates at the admin endpoint) and adding a tier later must not require a
    -- migration. An unknown stored value resolves to Free via EffectiveEntitlements.
    plan            VARCHAR(20) NOT NULL DEFAULT 'free',
    plan_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_organizations_deleted_at ON organizations (deleted_at);

-- Slug is the tenant subdomain label; unique among non-deleted orgs.
CREATE UNIQUE INDEX idx_organizations_slug
    ON organizations (slug)
    WHERE deleted_at IS NULL;
