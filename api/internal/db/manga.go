package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Manga represents a manga record
type Manga struct {
	MangaID   int
	Title     string
	Slug      string
	CreatedAt time.Time
}

// SaveManga inserts a new manga or returns existing one
func SaveManga(title, slug string) (int, error) {
	var mangaID int
	err := DB.QueryRow(`
		INSERT INTO manga (title, slug)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title
		RETURNING manga_id
	`, title, slug).Scan(&mangaID)

	if err != nil {
		return 0, fmt.Errorf("failed to save manga: %w", err)
	}

	fmt.Printf("  Manga saved: %s (ID: %d)\n", title, mangaID)
	return mangaID, nil
}

// GetMangaBySlug retrieves a manga by its slug
func GetMangaBySlug(slug string) (*Manga, error) {
	var m Manga
	err := DB.QueryRow(`
		SELECT manga_id, title, slug, created_at
		FROM manga WHERE slug = $1
	`, slug).Scan(&m.MangaID, &m.Title, &m.Slug, &m.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}
	return &m, nil
}

// GetMangaByID retrieves a manga by its ID
func GetMangaByID(mangaID int) (*Manga, error) {
	var m Manga
	err := DB.QueryRow(`
		SELECT manga_id, title, slug, created_at
		FROM manga WHERE manga_id = $1
	`, mangaID).Scan(&m.MangaID, &m.Title, &m.Slug, &m.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}
	return &m, nil
}

// GetAllManga retrieves all manga records
func GetAllManga() ([]Manga, error) {
	rows, err := DB.Query(`
		SELECT manga_id, title, slug, created_at
		FROM manga ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get manga list: %w", err)
	}
	defer rows.Close()

	var mangaList []Manga
	for rows.Next() {
		var m Manga
		if err := rows.Scan(&m.MangaID, &m.Title, &m.Slug, &m.CreatedAt); err != nil {
			continue
		}
		mangaList = append(mangaList, m)
	}
	return mangaList, nil
}
