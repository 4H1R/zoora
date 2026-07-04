ALTER TABLE quiz_rules
    ADD COLUMN negative_default_mode varchar(20),
    ADD COLUMN negative_default_value double precision,
    ADD COLUMN negative_default_wrongs_per_point integer;
