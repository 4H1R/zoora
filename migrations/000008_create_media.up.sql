CREATE TABLE media (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID         REFERENCES organizations (id) ON DELETE CASCADE,
    model_type      VARCHAR(100) NOT NULL,
    model_id        UUID         NOT NULL,
    collection_name VARCHAR(100) NOT NULL DEFAULT '',
    name            VARCHAR(255) NOT NULL DEFAULT '',
    file_name       VARCHAR(255) NOT NULL,
    mime_type       VARCHAR(100) NOT NULL DEFAULT '',
    disk            VARCHAR(50)  NOT NULL DEFAULT 's3',
    size            BIGINT       NOT NULL DEFAULT 0,
    custom_properties JSONB      NOT NULL DEFAULT '{}',
    order_column    INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_org_model_created ON media (organization_id, model_type, created_at DESC);
CREATE INDEX idx_media_model ON media (model_type, model_id);
CREATE INDEX idx_media_collection ON media (model_type, model_id, collection_name);
