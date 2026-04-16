-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chapter (
    chapter_id      SERIAL PRIMARY KEY,
    manga_id        INTEGER NOT NULL REFERENCES manga(manga_id) ON DELETE CASCADE,
    chapter_slug    VARCHAR(255) NOT NULL,
    downloaded      BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(manga_id, chapter_slug)
);
CREATE INDEX IF NOT EXISTS idx_chapter_manga_id ON chapter(manga_id);
CREATE INDEX IF NOT EXISTS idx_chapter_slug ON chapter(chapter_slug);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chapter;
-- +goose StatementEnd
