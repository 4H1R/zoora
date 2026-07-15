-- Anti-cheat image rendering for questions. When render_as_image is on, the
-- worker renders the body text (and each option value, stored in the options
-- JSONB) to distorted PNGs and the take endpoint withholds the raw text.
ALTER TABLE questions
    ADD COLUMN render_as_image       BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN image_render_status   VARCHAR(20) NOT NULL DEFAULT 'none'
        CHECK (image_render_status IN ('none', 'pending', 'ready', 'failed')),
    ADD COLUMN system_image_media_id UUID;
