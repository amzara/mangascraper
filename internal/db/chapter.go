package db

import (
	"database/sql"
	"fmt"
	"time"
)

// ChapterStatus represents the download state of a chapter
type ChapterStatus string

const (
	ChapterStatusPending     ChapterStatus = "pending"
	ChapterStatusDownloading ChapterStatus = "downloading"
	ChapterStatusDownloaded  ChapterStatus = "downloaded"
	ChapterStatusError       ChapterStatus = "error"
)

// Chapter represents a chapter record
type Chapter struct {
	ChapterID    int
	MangaID      int
	ChapterSlug  string
	Status       ChapterStatus
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

// SaveChapter inserts a new chapter or returns existing one
func SaveChapter(mangaID int, chapterSlug string) (int, error) {
	var chapterID int
	err := DB.QueryRow(`
		INSERT INTO chapter (manga_id, chapter_slug, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (manga_id, chapter_slug) DO NOTHING
		RETURNING chapter_id
	`, mangaID, chapterSlug, ChapterStatusPending).Scan(&chapterID)

	if err == sql.ErrNoRows {
		// Chapter already exists, get its ID
		err = DB.QueryRow(`
			SELECT chapter_id FROM chapter 
			WHERE manga_id = $1 AND chapter_slug = $2
		`, mangaID, chapterSlug).Scan(&chapterID)
		if err != nil {
			return 0, fmt.Errorf("failed to get existing chapter: %w", err)
		}
		fmt.Printf("    Chapter exists: %s (ID: %d)\n", chapterSlug, chapterID)
	} else if err != nil {
		return 0, fmt.Errorf("failed to save chapter: %w", err)
	} else {
		fmt.Printf("    Chapter saved: %s (ID: %d)\n", chapterSlug, chapterID)
	}

	return chapterID, nil
}

// UpdateChapterStatus updates the download status of a chapter
func UpdateChapterStatus(chapterID int, status ChapterStatus, errorMsg string) error {
	_, err := DB.Exec(`
		UPDATE chapter 
		SET status = $1, 
		    error_message = $2,
		    updated_at = NOW()
		WHERE chapter_id = $3
	`, status, errorMsg, chapterID)

	if err != nil {
		return fmt.Errorf("failed to update chapter status: %w", err)
	}

	fmt.Printf("    Chapter %d: %s\n", chapterID, status)
	return nil
}

// MarkChapterPending sets chapter status to pending
func MarkChapterPending(chapterID int) error {
	return UpdateChapterStatus(chapterID, ChapterStatusPending, "")
}

// MarkChapterDownloading sets chapter status to downloading
func MarkChapterDownloading(chapterID int) error {
	return UpdateChapterStatus(chapterID, ChapterStatusDownloading, "")
}

// MarkChapterDownloaded sets chapter status to downloaded
func MarkChapterDownloaded(chapterID int) error {
	return UpdateChapterStatus(chapterID, ChapterStatusDownloaded, "")
}

// MarkChapterError sets chapter status to error with error message
func MarkChapterError(chapterID int, errorMsg string) error {
	return UpdateChapterStatus(chapterID, ChapterStatusError, errorMsg)
}

// GetChapterBySlug retrieves a chapter by manga ID and chapter slug
func GetChapterBySlug(mangaID int, chapterSlug string) (*Chapter, error) {
	var c Chapter
	var statusStr string
	var errorMsg string
	var updatedAt sql.NullTime

	err := DB.QueryRow(`
		SELECT chapter_id, manga_id, chapter_slug, status, 
		       COALESCE(error_message, ''), created_at, updated_at
		FROM chapter 
		WHERE manga_id = $1 AND chapter_slug = $2
	`, mangaID, chapterSlug).Scan(
		&c.ChapterID, &c.MangaID, &c.ChapterSlug, &statusStr,
		&errorMsg, &c.CreatedAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}

	c.Status = ChapterStatus(statusStr)
	c.ErrorMessage = errorMsg
	if updatedAt.Valid {
		c.UpdatedAt = &updatedAt.Time
	}

	return &c, nil
}

// GetChaptersByManga retrieves all chapters for a manga
func GetChaptersByManga(mangaID int) ([]Chapter, error) {
	rows, err := DB.Query(`
		SELECT chapter_id, manga_id, chapter_slug, status,
		       COALESCE(error_message, ''), created_at, updated_at
		FROM chapter 
		WHERE manga_id = $1
		ORDER BY chapter_slug
	`, mangaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapters: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var c Chapter
		var statusStr string
		var errorMsg string
		var updatedAt sql.NullTime

		err := rows.Scan(&c.ChapterID, &c.MangaID, &c.ChapterSlug, &statusStr,
			&errorMsg, &c.CreatedAt, &updatedAt)
		if err != nil {
			continue
		}

		c.Status = ChapterStatus(statusStr)
		c.ErrorMessage = errorMsg
		if updatedAt.Valid {
			c.UpdatedAt = &updatedAt.Time
		}

		chapters = append(chapters, c)
	}
	return chapters, nil
}

// GetPendingChapters retrieves all chapters that haven't been downloaded yet
func GetPendingChapters() ([]Chapter, error) {
	rows, err := DB.Query(`
		SELECT chapter_id, manga_id, chapter_slug, status, created_at
		FROM chapter 
		WHERE status != $1 AND status != $2
		ORDER BY created_at
	`, ChapterStatusDownloaded, ChapterStatusError)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending chapters: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var c Chapter
		err := rows.Scan(&c.ChapterID, &c.MangaID, &c.ChapterSlug, &c.Status, &c.CreatedAt)
		if err != nil {
			continue
		}
		chapters = append(chapters, c)
	}
	return chapters, nil
}
