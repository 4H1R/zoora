CREATE TABLE ai_grading_jobs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    quiz_id UUID NOT NULL,
    created_by UUID NOT NULL,
    mode VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total INTEGER NOT NULL DEFAULT 0,
    done INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ai_grading_jobs_quiz FOREIGN KEY (quiz_id) REFERENCES quizzes (id) ON DELETE CASCADE,
    CONSTRAINT fk_ai_grading_jobs_org FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE,
    CONSTRAINT chk_ai_grading_jobs_mode CHECK (mode IN ('apply', 'suggest')),
    CONSTRAINT chk_ai_grading_jobs_status CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);

CREATE INDEX idx_ai_grading_jobs_quiz_created ON ai_grading_jobs (quiz_id, created_at DESC);
