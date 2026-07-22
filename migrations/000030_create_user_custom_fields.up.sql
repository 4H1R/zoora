CREATE TABLE user_custom_field_definitions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    label VARCHAR(255) NOT NULL,
    field_type VARCHAR(32) NOT NULL,
    options JSONB NOT NULL DEFAULT '[]',
    is_required BOOL NOT NULL DEFAULT FALSE,
    is_unique BOOL NOT NULL DEFAULT FALSE,
    visible_to_user BOOL NOT NULL DEFAULT FALSE,
    position INT NOT NULL DEFAULT 0,
    description TEXT,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_custom_field_definitions_organization FOREIGN KEY (organization_id)
        REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE INDEX idx_ucfd_org_position ON user_custom_field_definitions (organization_id, position);
