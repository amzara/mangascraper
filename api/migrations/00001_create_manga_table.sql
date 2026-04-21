-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS manga (
    manga_id    SERIAL PRIMARY KEY,
    title       VARCHAR(500) NOT NULL,
    slug        VARCHAR(255) NOT NULL UNIQUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chapter;
-- +goose StatementEnd
