CREATE TABLE question_banks (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'ready',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_question_banks_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT chk_question_banks_status CHECK (status IN ('ready', 'copying', 'failed'))
);

CREATE INDEX idx_question_banks_organization_id ON question_banks (organization_id);
CREATE INDEX idx_question_banks_deleted_at ON question_banks (deleted_at);

CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    bank_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    text TEXT NOT NULL,
    type VARCHAR(20) NOT NULL,
    options JSONB NOT NULL DEFAULT '[]',
    model_answer TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '[]',
    negative_mark_mode VARCHAR(20) NOT NULL DEFAULT 'none',
    negative_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    wrongs_per_point INTEGER NOT NULL DEFAULT 0,
    min_seconds INTEGER NOT NULL DEFAULT 0,
    image_render_status VARCHAR(20) NOT NULL DEFAULT 'none',
    system_image_media_id UUID,
    system_image_content_hash VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_questions_bank FOREIGN KEY (bank_id) REFERENCES question_banks (id) ON DELETE CASCADE,
    CONSTRAINT fk_questions_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT chk_questions_type CHECK (type IN ('descriptive', 'short_answer', 'choice')),
    CONSTRAINT chk_questions_image_render_status CHECK (image_render_status IN ('none', 'pending', 'ready', 'failed'))
);

CREATE INDEX idx_questions_bank_id ON questions (bank_id);
CREATE INDEX idx_questions_organization_id ON questions (organization_id);
CREATE INDEX idx_questions_deleted_at ON questions (deleted_at);

-- Share codes let a teacher hand a bank to another class/org: redeeming clones
-- the bank (questions + media) into the redeemer's org as an independent copy.
CREATE TABLE question_bank_share_codes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    bank_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    code VARCHAR(32) NOT NULL,
    created_by UUID NOT NULL,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_question_bank_share_codes_bank FOREIGN KEY (bank_id) REFERENCES question_banks (id) ON DELETE CASCADE,
    CONSTRAINT fk_question_bank_share_codes_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_question_bank_share_codes_created_by FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_question_bank_share_codes_code ON question_bank_share_codes (code);
-- One active (non-revoked) code per bank; generating a new one revokes the old.
CREATE UNIQUE INDEX idx_question_bank_share_codes_active_bank ON question_bank_share_codes (bank_id) WHERE revoked_at IS NULL;
