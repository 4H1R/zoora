CREATE TABLE plan_prices (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    plan       VARCHAR(20) NOT NULL,
    interval   VARCHAR(20) NOT NULL,
    currency   CHAR(3) NOT NULL DEFAULT 'IRR',
    amount     BIGINT NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Only one active price per (plan, interval, currency).
CREATE UNIQUE INDEX idx_plan_prices_active
    ON plan_prices (plan, interval, currency)
    WHERE active;

CREATE TABLE invoices (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    number          VARCHAR(30),
    organization_id UUID NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    currency        CHAR(3) NOT NULL DEFAULT 'IRR',
    subtotal        BIGINT NOT NULL DEFAULT 0,
    tax_percent     INT NOT NULL DEFAULT 0,
    tax_amount      BIGINT NOT NULL DEFAULT 0,
    total           BIGINT NOT NULL DEFAULT 0,
    description     TEXT NOT NULL DEFAULT '',
    expires_at      TIMESTAMPTZ,
    issued_at       TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    pdf_object_key  TEXT,
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT fk_invoices_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_invoices_number ON invoices (number) WHERE number IS NOT NULL;
CREATE INDEX idx_invoices_organization_id ON invoices (organization_id);
CREATE INDEX idx_invoices_status ON invoices (status);
CREATE INDEX idx_invoices_deleted_at ON invoices (deleted_at);
CREATE INDEX idx_invoices_pending_issued ON invoices (issued_at) WHERE status = 'pending';
CREATE INDEX idx_invoices_pending_expires ON invoices (expires_at) WHERE status = 'pending';

CREATE TABLE invoice_items (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    invoice_id   UUID NOT NULL,
    kind         VARCHAR(30) NOT NULL,
    description  TEXT NOT NULL,
    plan         VARCHAR(20),
    interval     VARCHAR(20),
    period_start TIMESTAMPTZ,
    period_end   TIMESTAMPTZ,
    quantity     INT NOT NULL DEFAULT 1,
    unit_amount  BIGINT NOT NULL,
    amount       BIGINT NOT NULL,
    currency     CHAR(3) NOT NULL DEFAULT 'IRR',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_invoice_items_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE CASCADE
);

CREATE INDEX idx_invoice_items_invoice_id ON invoice_items (invoice_id);

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    invoice_id      UUID NOT NULL,
    organization_id UUID NOT NULL,
    gateway         VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount          BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL DEFAULT 'IRR',
    authority       VARCHAR(255),
    ref_id          VARCHAR(255),
    raw_response    JSONB,
    note            TEXT NOT NULL DEFAULT '',
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at     TIMESTAMPTZ,
    CONSTRAINT fk_payments_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE CASCADE,
    CONSTRAINT fk_payments_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE INDEX idx_payments_invoice_id ON payments (invoice_id);
CREATE INDEX idx_payments_organization_id ON payments (organization_id);
CREATE UNIQUE INDEX idx_payments_gateway_authority ON payments (gateway, authority) WHERE authority IS NOT NULL;

CREATE TABLE billing_reminders_sent (
    id         UUID PRIMARY KEY DEFAULT uuidv7(),
    kind       VARCHAR(30) NOT NULL,
    subject_id UUID NOT NULL,
    period_key VARCHAR(40) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_billing_reminders_unique
    ON billing_reminders_sent (kind, subject_id, period_key);
