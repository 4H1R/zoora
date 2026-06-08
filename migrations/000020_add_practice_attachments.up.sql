ALTER TABLE practice_rooms
    ADD COLUMN attachments JSONB NOT NULL DEFAULT '[]';

ALTER TABLE practice_submissions
    ADD COLUMN attachments JSONB NOT NULL DEFAULT '[]';
