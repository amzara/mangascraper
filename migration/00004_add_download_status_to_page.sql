-- +goose Up
-- Add download tracking columns to page table
ALTER TABLE page
    ADD COLUMN IF NOT EXISTS downloaded BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS error_message VARCHAR(500),
    ADD COLUMN IF NOT EXISTS downloaded_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_page_downloaded ON page(downloaded);

-- +goose Down
-- Remove download tracking columns
DROP INDEX IF EXISTS idx_page_downloaded;

ALTER TABLE page
    DROP COLUMN IF EXISTS downloaded,
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS downloaded_at;
