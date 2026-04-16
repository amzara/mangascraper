# API Endpoints & SQL Queries

This document lists all available HTTP API endpoints and the underlying SQL queries they execute.

---

## `GET /api/health`
Health check endpoint.

**SQL Queries:** None

---

## `POST /api/manga/download`
Starts a manga download in the background.

**Request Body:**
```json
{
  "title": "manga-slug"
}
```

**SQL Queries:**

### 1. `GetMangaBySlug`
Retrieves the manga record to return its `manga_id` in the response.

```sql
SELECT manga_id, title, slug, created_at
FROM manga WHERE slug = $1
```

### 2. `SaveManga` (called by background scraper)
Inserts a new manga or updates the title if the slug already exists.

```sql
INSERT INTO manga (title, slug)
VALUES ($1, $2)
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title
RETURNING manga_id
```

---

## `GET /api/manga`
Lists all downloaded manga.

**SQL Queries:**

### `GetAllManga`
```sql
SELECT manga_id, title, slug, created_at
FROM manga ORDER BY created_at DESC
```

---

## `GET /api/manga/{slug}`
Returns manga details together with its chapters.

**SQL Queries:**

### 1. `GetMangaBySlug`
```sql
SELECT manga_id, title, slug, created_at
FROM manga WHERE slug = $1
```

### 2. `GetChaptersByManga`
```sql
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status,
       COALESCE(error_message, ''), created_at, updated_at
FROM chapter
WHERE manga_id = $1
ORDER BY chapter_num ASC
```

---

## `GET /api/chapter/{id}`
Returns chapter metadata and its pages.

**SQL Queries:**

### 1. `GetChapterByID`
```sql
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status,
       COALESCE(error_message, ''), created_at, updated_at
FROM chapter
WHERE chapter_id = $1
```

### 2. `GetPagesByChapter`
```sql
SELECT page_id, chapter_id, image_url, file_name, file_path,
       COALESCE(error_message, ''), created_at, downloaded_at
FROM page WHERE chapter_id = $1
ORDER BY page_id
```

---

## `POST /api/token`
Refreshes the Cloudflare clearance token via FlareSolverr.

**SQL Queries:** None

---

## `GET /api/search?q=query`
Searches manga on `mangakakalot.gg`.

**SQL Queries:** None (proxied to external search API)

---

# Internal DB Functions & SQL Reference

The following functions are used by the endpoints above or by the background scraper.

## Manga (`internal/db/manga.go`)

### `GetMangaByID`
```sql
SELECT manga_id, title, slug, created_at
FROM manga WHERE manga_id = $1
```

## Chapter (`internal/db/chapter.go`)

### `SaveChapter`
```sql
INSERT INTO chapter (manga_id, chapter_slug, chapter_num, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (manga_id, chapter_slug) DO UPDATE SET chapter_num = EXCLUDED.chapter_num
RETURNING chapter_id
```

*Fallback when `ON CONFLICT` returns no rows:*
```sql
SELECT chapter_id FROM chapter
WHERE manga_id = $1 AND chapter_slug = $2
```

### `UpdateChapterStatus`
```sql
UPDATE chapter
SET status = $1,
    error_message = $2,
    updated_at = NOW()
WHERE chapter_id = $3
```

### `GetChapterBySlug`
```sql
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status,
       COALESCE(error_message, ''), created_at, updated_at
FROM chapter
WHERE manga_id = $1 AND chapter_slug = $2
```

### `GetPendingChapters`
```sql
SELECT chapter_id, manga_id, chapter_slug, chapter_num, status, created_at
FROM chapter
WHERE status != $1 AND status != $2
ORDER BY manga_id, chapter_num ASC
```

## Page (`internal/db/page.go`)

### `SavePage`
```sql
INSERT INTO page (chapter_id, image_url, file_name, file_path, downloaded, error_message)
VALUES ($1, $2, $3, $4, FALSE, NULL)
ON CONFLICT DO NOTHING
RETURNING page_id
```

*Fallback when page already exists:*
```sql
SELECT page_id FROM page
WHERE chapter_id = $1 AND image_url = $2
```

### `UpdatePageStatus`
```sql
UPDATE page
SET downloaded = $1,
    error_message = $2,
    downloaded_at = $3
WHERE page_id = $4
```

### `GetPendingPages`
```sql
SELECT page_id, chapter_id, image_url, file_name, created_at
FROM page
WHERE downloaded = FALSE AND (error_message IS NULL OR error_message = '')
ORDER BY page_id
```
