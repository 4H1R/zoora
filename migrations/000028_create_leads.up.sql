CREATE TABLE leads (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    name            VARCHAR(255) NOT NULL,
    phone           VARCHAR(32)  NOT NULL,
    org_name        VARCHAR(255) NOT NULL,
    plan            VARCHAR(32)  NOT NULL DEFAULT '',
    note            TEXT         NOT NULL DEFAULT '',
    status          VARCHAR(20)  NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'contacted', 'converted', 'rejected')),
    -- Set on convert: the org this lead became. NULL until then; NULLed if that
    -- org is later hard-deleted.
    organization_id UUID,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_leads_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE SET NULL
);

-- Admin list defaults to newest-first, filterable by status.
CREATE INDEX idx_leads_status_created_at ON leads (status, created_at DESC);

-- Backs submit-time dedupe: look up an open (new/contacted) lead by phone.
CREATE INDEX idx_leads_phone ON leads (phone);
