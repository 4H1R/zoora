-- Per-org, append-only, immutable audit log of structural mutations.
-- Denormalized actor_name/target_label snapshots so an entry survives the
-- deletion of the user or resource it describes. org_id is the TARGET's org.
CREATE TABLE audit_entries (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    -- Actor: NULL user id + actor_name 'System' for non-human actions.
    actor_id     UUID,
    actor_name   VARCHAR(255) NOT NULL DEFAULT '',
    actor_username VARCHAR(255) NOT NULL DEFAULT '',
    action       VARCHAR(40)  NOT NULL,
    target_type  VARCHAR(64)  NOT NULL,
    -- NULL for denied attempts where no valid resource id was in the route.
    target_id    UUID,
    target_label VARCHAR(512) NOT NULL DEFAULT '',
    outcome      VARCHAR(16)  NOT NULL DEFAULT 'success'
                 CHECK (outcome IN ('success', 'denied')),
    metadata     JSONB        NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_audit_entries_organization_id
        FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_entries_actor_id
        FOREIGN KEY (actor_id) REFERENCES users (id) ON DELETE SET NULL
);

-- Primary read: org-wide list, newest first.
CREATE INDEX idx_audit_entries_org_created_at
    ON audit_entries (organization_id, created_at DESC);

-- Per-resource history: "everything that happened to this class".
CREATE INDEX idx_audit_entries_org_target
    ON audit_entries (organization_id, target_type, target_id);

-- Actor filter: "everything user X did".
CREATE INDEX idx_audit_entries_org_actor_created_at
    ON audit_entries (organization_id, actor_id, created_at DESC);
