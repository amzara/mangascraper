-- +goose Up
ALTER TABLE page ADD COLUMN file_path VARCHAR(500);

-- Update existing records to set file_path based on related tables
UPDATE page p
SET file_path = '/' || m.slug || '/' || c.chapter_slug || '/' || p.file_name
FROM chapter c
JOIN manga m ON c.manga_id = m.manga_id
WHERE p.chapter_id = c.chapter_id;

-- +goose Down
ALTER TABLE page DROP COLUMN file_path;
