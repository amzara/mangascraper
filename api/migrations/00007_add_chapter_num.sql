-- +goose Up
-- +goose StatementBegin
ALTER TABLE chapter ADD COLUMN chapter_num INTEGER;

CREATE INDEX IF NOT EXISTS idx_chapter_num ON chapter(chapter_num);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chapter DROP COLUMN chapter_num;
DROP INDEX IF EXISTS idx_chapter_num;
-- +goose StatementEnd
