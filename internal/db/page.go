package db

import (
	"database/sql"
	"fmt"
	"time"
)

// PageStatus represents the download state of a page
type PageStatus string

const (
	StatusPending     PageStatus = "pending"
	StatusDownloading PageStatus = "downloading"
	StatusDownloaded  PageStatus = "downloaded"
	StatusError       PageStatus = "error"
)

// Page represents a page record
type Page struct {
	PageID       int
	ChapterID    int
	ImageURL     string
	FileName     string
	Status       PageStatus
	ErrorMessage string
	CreatedAt    time.Time
	DownloadedAt *time.Time
}

// SavePage inserts a new page with pending status
func SavePage(chapterID int, imageURL, fileName string) (int, error) {
	var pageID int
	err := DB.QueryRow(`
		INSERT INTO page (chapter_id, image_url, file_name, downloaded, error_message)
		VALUES ($1, $2, $3, FALSE, NULL)
		ON CONFLICT DO NOTHING
		RETURNING page_id
	`, chapterID, imageURL, fileName).Scan(&pageID)

	if err == sql.ErrNoRows {
		// Page already exists, get its ID
		err = DB.QueryRow(`
			SELECT page_id FROM page 
			WHERE chapter_id = $1 AND image_url = $2
		`, chapterID, imageURL).Scan(&pageID)
		if err != nil {
			return 0, fmt.Errorf("failed to get existing page: %w", err)
		}
	} else if err != nil {
		return 0, fmt.Errorf("failed to save page: %w", err)
	}

	return pageID, nil
}

// UpdatePageStatus updates the download status of a page
func UpdatePageStatus(pageID int, status PageStatus, errorMsg string) error {
	var downloadedAt interface{}
	
	if status == StatusDownloaded {
		downloadedAt = time.Now()
	} else {
		downloadedAt = nil
	}

	_, err := DB.Exec(`
		UPDATE page 
		SET downloaded = $1, 
		    error_message = $2,
		    downloaded_at = $3
		WHERE page_id = $4
	`, status == StatusDownloaded, errorMsg, downloadedAt, pageID)

	if err != nil {
		return fmt.Errorf("failed to update page status: %w", err)
	}

	fmt.Printf("    Page %d: %s\n", pageID, status)
	return nil
}

// MarkPagePending sets page status to pending
func MarkPagePending(pageID int) error {
	return UpdatePageStatus(pageID, StatusPending, "")
}

// MarkPageDownloading sets page status to downloading
func MarkPageDownloading(pageID int) error {
	return UpdatePageStatus(pageID, StatusDownloading, "")
}

// MarkPageDownloaded sets page status to downloaded
func MarkPageDownloaded(pageID int) error {
	return UpdatePageStatus(pageID, StatusDownloaded, "")
}

// MarkPageError sets page status to error with error message
func MarkPageError(pageID int, errorMsg string) error {
	return UpdatePageStatus(pageID, StatusError, errorMsg)
}

// GetPagesByChapter retrieves all pages for a chapter
func GetPagesByChapter(chapterID int) ([]Page, error) {
	rows, err := DB.Query(`
		SELECT page_id, chapter_id, image_url, file_name, 
		       COALESCE(error_message, ''), created_at, downloaded_at
		FROM page WHERE chapter_id = $1
		ORDER BY page_id
	`, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var p Page
		var errorMsg string
		err := rows.Scan(&p.PageID, &p.ChapterID, &p.ImageURL, &p.FileName,
			&errorMsg, &p.CreatedAt, &p.DownloadedAt)
		if err != nil {
			continue
		}
		p.ErrorMessage = errorMsg
		
		// Determine status from fields
		if p.DownloadedAt != nil {
			p.Status = StatusDownloaded
		} else if errorMsg != "" {
			p.Status = StatusError
		} else {
			p.Status = StatusPending
		}
		
		pages = append(pages, p)
	}
	return pages, nil
}

// GetPendingPages retrieves all pages that haven't been downloaded yet
func GetPendingPages() ([]Page, error) {
	rows, err := DB.Query(`
		SELECT page_id, chapter_id, image_url, file_name, created_at
		FROM page 
		WHERE downloaded = FALSE AND (error_message IS NULL OR error_message = '')
		ORDER BY page_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var p Page
		err := rows.Scan(&p.PageID, &p.ChapterID, &p.ImageURL, &p.FileName, &p.CreatedAt)
		if err != nil {
			continue
		}
		p.Status = StatusPending
		pages = append(pages, p)
	}
	return pages, nil
}
