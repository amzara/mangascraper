-- +goose Up
-- Add status tracking columns to chapter table
ALTER TABLE chapter
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS error_message VARCHAR(500),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;

-- Migrate existing data: convert downloaded boolean to status
UPDATE chapter SET status = 'downloaded' WHERE downloaded = TRUE AND status = 'pending';

-- Create index for status lookups
CREATE INDEX IF NOT EXISTS idx_chapter_status ON chapter(status);

-- +goose Down
-- Remove status tracking columns
DROP INDEX IF EXISTS idx_chapter_status;

ALTER TABLE chapter
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS updated_at;
