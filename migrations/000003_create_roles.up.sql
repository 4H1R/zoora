CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID REFERENCES organizations (id) ON DELETE CASCADE,
    is_preset BOOL NOT NULL DEFAULT FALSE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_roles_org_name ON roles (organization_id, name) WHERE organization_id IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX idx_roles_preset_name ON roles (name) WHERE is_preset = true AND deleted_at IS NULL;
CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);

CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    role_id UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT role_permissions_role_id_permission_id_unique UNIQUE (role_id, permission_id)
);
