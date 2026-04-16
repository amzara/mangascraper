-- name: GetAllManga :many
SELECT manga_id, title, slug, created_at
FROM manga ORDER BY created_at DESC;

-- name: GetMangaBySlug :one
SELECT manga_id, title, slug, created_at
FROM manga WHERE slug = $1;

-- name: GetMangaByID :one
SELECT manga_id, title, slug, created_at
FROM manga WHERE manga_id = $1;

-- name: SaveManga :one
INSERT INTO manga (title, slug)
VALUES ($1, $2)
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title
RETURNING manga_id;

-- name: GetChaptersByManga :many
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status, COALESCE(error_message, ''), created_at, updated_at
FROM chapter
WHERE manga_id = $1
ORDER BY chapter_num ASC;

-- name: GetChapterByID :one
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status,
       COALESCE(error_message, ''), created_at, updated_at
FROM chapter
WHERE chapter_id = $1;

-- name: GetChapterBySlug :one
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status,
       COALESCE(error_message, ''), created_at, updated_at
FROM chapter
WHERE manga_id = $1 AND chapter_slug = $2;

-- name: SaveChapter :one
INSERT INTO chapter (manga_id, chapter_slug, chapter_num, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (manga_id, chapter_slug) DO UPDATE SET chapter_num = EXCLUDED.chapter_num
RETURNING chapter_id;

-- name: UpdateChapterStatus :exec
UPDATE chapter
SET status = $1,
    error_message = $2,
    updated_at = NOW()
WHERE chapter_id = $3;

-- name: GetPendingChapters :many
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status, created_at
FROM chapter
WHERE status != $1 AND status != $2
ORDER BY manga_id, chapter_num ASC;

-- name: GetPagesByChapter :many
SELECT page_id, chapter_id, image_url, file_name, file_path, COALESCE(error_message, ''), created_at, downloaded_at
FROM page WHERE chapter_id = $1
ORDER BY page_id;

-- name: SavePage :one
INSERT INTO page (chapter_id, image_url, file_name, file_path, downloaded, error_message)
VALUES ($1, $2, $3, $4, FALSE, NULL)
ON CONFLICT DO NOTHING
RETURNING page_id;

-- name: UpdatePageStatus :exec
UPDATE page
SET downloaded = $1,
    error_message = $2,
    downloaded_at = $3
WHERE page_id = $4;

-- name: GetPendingPages :many
SELECT page_id, chapter_id, image_url, file_name, created_at
FROM page
WHERE downloaded = FALSE AND (error_message IS NULL OR error_message = '')
ORDER BY page_id;
