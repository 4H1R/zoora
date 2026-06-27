ALTER TABLE quizzes
    ADD COLUMN negative_mark_mode varchar(20) NOT NULL DEFAULT 'none',
    ADD COLUMN negative_value double precision NOT NULL DEFAULT 0,
    ADD COLUMN wrongs_per_point integer NOT NULL DEFAULT 0;

ALTER TABLE quiz_rules
    ADD COLUMN negative_overrides jsonb NOT NULL DEFAULT '[]';
