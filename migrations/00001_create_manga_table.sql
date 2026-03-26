-- +goose Up
CREATE TABLE IF NOT EXISTS manga (
    manga_id    SERIAL PRIMARY KEY,
    title       VARCHAR(500) NOT NULL,
    slug        VARCHAR(255) NOT NULL UNIQUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_manga_slug ON manga(slug);

-- +goose Down
DROP TABLE IF EXISTS manga;