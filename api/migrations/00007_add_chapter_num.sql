-- +goose Up
-- Add chapter_num column
ALTER TABLE chapter ADD COLUMN chapter_num INTEGER;

-- Create index for ordering
CREATE INDEX IF NOT EXISTS idx_chapter_num ON chapter(chapter_num);

-- +goose Down
ALTER TABLE chapter DROP COLUMN chapter_num;
DROP INDEX IF EXISTS idx_chapter_num;
