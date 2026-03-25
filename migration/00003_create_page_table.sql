-- +goose Up
CREATE TABLE IF NOT EXISTS page (
    page_id         SERIAL PRIMARY KEY,
    chapter_id      INTEGER NOT NULL REFERENCES chapter(chapter_id) ON DELETE CASCADE,
    image_url       VARCHAR(1000) NOT NULL,
    file_name       VARCHAR(255),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_page_chapter_id ON page(chapter_id);

-- +goose Down
DROP TABLE IF EXISTS page;