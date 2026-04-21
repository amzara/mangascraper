-- +goose Up
-- +goose StatementBegin
ALTER TABLE chapter
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS error_message VARCHAR(500),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;

UPDATE chapter SET status = 'downloaded' WHERE downloaded = TRUE AND status = 'pending';

CREATE INDEX IF NOT EXISTS idx_chapter_status ON chapter(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_chapter_status;

ALTER TABLE chapter
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
