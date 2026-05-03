CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'trial', 'suspended', 'archived')),
    total_users INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_organizations_deleted_at ON organizations (deleted_at);

ALTER TABLE users
    ADD CONSTRAINT fk_users_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE SET NULL;
