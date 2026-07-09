CREATE TABLE organization_settings (
    id                                   UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id                      UUID NOT NULL,
    attendance_present_threshold_percent INT NOT NULL DEFAULT 75,
    sms_enabled                          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at                           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_organization_settings_org FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT chk_attendance_percent CHECK (attendance_present_threshold_percent BETWEEN 1 AND 100)
);

CREATE UNIQUE INDEX idx_organization_settings_org_id ON organization_settings (organization_id);
