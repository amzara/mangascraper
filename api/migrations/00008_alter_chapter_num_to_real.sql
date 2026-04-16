-- +goose Up
-- +goose StatementBegin
ALTER TABLE chapter ALTER COLUMN chapter_num TYPE REAL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chapter ALTER COLUMN chapter_num TYPE INTEGER;
-- +goose StatementEnd
