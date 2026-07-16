CREATE TABLE ai_usage_events (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    feature VARCHAR(40) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    model VARCHAR(60) NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    cost_micros BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ai_usage_events_org FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE INDEX idx_ai_usage_events_org_created ON ai_usage_events (organization_id, created_at DESC);
