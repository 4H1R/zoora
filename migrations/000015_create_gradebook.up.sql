CREATE TABLE gradebook_columns (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('auto_attendance', 'auto_practice', 'auto_quiz', 'manual_grade', 'manual_attendance', 'manual_text')),
    source_id UUID,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gradebook_columns_class_id ON gradebook_columns(class_id);

CREATE TABLE gradebook_cells (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    column_id UUID NOT NULL REFERENCES gradebook_columns(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(column_id, student_id)
);

CREATE INDEX idx_gradebook_cells_column_id ON gradebook_cells(column_id);
CREATE INDEX idx_gradebook_cells_student_id ON gradebook_cells(student_id);
