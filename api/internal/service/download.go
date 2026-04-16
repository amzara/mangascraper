package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mangascraper/internal/repository"
	"mangascraper/internal/scraper"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// DownloadService orchestrates manga downloads and persists state via sqlc queries
type DownloadService struct {
	Queries *repository.Queries
}

// NewDownloadService creates a new download service
func NewDownloadService(queries *repository.Queries) *DownloadService {
	return &DownloadService{Queries: queries}
}

// DownloadJob represents a single chapter to download
type DownloadJob struct {
	Title string
	Slug  string
}

// DownloadManga saves the manga, fetches chapters, and spawns workers to download them
func (s *DownloadService) DownloadManga(title string) error {
	ctx := context.Background()

	mangaID, err := s.Queries.SaveManga(ctx, repository.SaveMangaParams{
		Title: title,
		Slug:  title,
	})
	if err != nil {
		return fmt.Errorf("failed to save manga: %w", err)
	}
	fmt.Printf("Manga '%s' saved with ID: %d\n", title, mangaID)

	slugs, err := s.GetChapterListWithDB(mangaID, title)
	if err != nil {
		return fmt.Errorf("failed to get chapter list: %w", err)
	}

	if len(slugs) == 0 {
		return fmt.Errorf("no chapter found for %s", title)
	}

	fmt.Printf("Found %d chapters for %s, starting download . . . \n", len(slugs), title)

	s.BeginJobPool(title, mangaID, slugs, 100)

	fmt.Printf("Finish download %s\n", title)
	return nil
}

// BeginJobPool starts a pool of workers to download chapters concurrently
func (s *DownloadService) BeginJobPool(title string, mangaID int32, slugs []string, numWorkers int) {
	jobs := make(chan DownloadJob, len(slugs))
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go s.downloadWorkers(w, mangaID, jobs, &wg)
	}

	for _, slug := range slugs {
		jobs <- DownloadJob{Title: title, Slug: slug}
	}

	close(jobs)
	wg.Wait()
}

func (s *DownloadService) downloadWorkers(id int, mangaID int32, jobs <-chan DownloadJob, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := context.Background()

	for job := range jobs {
		fmt.Printf("Worker %d: processing %s\n", id, job.Slug)

		chapter, err := s.Queries.GetChapterBySlug(ctx, repository.GetChapterBySlugParams{
			MangaID:     mangaID,
			ChapterSlug: job.Slug,
		})
		if err != nil {
			fmt.Printf("Worker %d: failed to get chapter %s: %v\n", id, job.Slug, err)
			continue
		}

		if err := s.Queries.UpdateChapterStatus(ctx, repository.UpdateChapterStatusParams{
			Status:       pgtype.Text{String: "downloading", Valid: true},
			ErrorMessage: pgtype.Text{Valid: false},
			ChapterID:    chapter.ChapterID,
		}); err != nil {
			fmt.Printf("Worker %d: failed to mark chapter as downloading: %v\n", id, err)
		}

		err = s.DownloadImagesWithDB(job.Title, job.Slug, chapter.ChapterID)
		if err != nil {
			s.Queries.UpdateChapterStatus(ctx, repository.UpdateChapterStatusParams{
				Status:       pgtype.Text{String: "error", Valid: true},
				ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
				ChapterID:    chapter.ChapterID,
			})
			fmt.Printf("Worker %d: error downloading %s: %v\n", id, job.Slug, err)
		} else {
			s.Queries.UpdateChapterStatus(ctx, repository.UpdateChapterStatusParams{
				Status:       pgtype.Text{String: "downloaded", Valid: true},
				ErrorMessage: pgtype.Text{Valid: false},
				ChapterID:    chapter.ChapterID,
			})
			fmt.Printf("Worker %d: completed %s\n", id, job.Slug)
		}
	}
}

// GetChapterListWithDB fetches chapters from the API and saves them to the database
func (s *DownloadService) GetChapterListWithDB(mangaID int32, title string) ([]string, error) {
	ctx := context.Background()

	chapterApi := fmt.Sprintf("https://www.mangakakalot.gg/api/manga/%s/chapters?limit=999", title)
	fmt.Printf("Chapter API is %s\n", chapterApi)
	resp, err := http.Get(chapterApi)
	if err != nil {
		return nil, fmt.Errorf("error getting chapter amount: %w", err)
	}
	defer resp.Body.Close()

	var clResponse scraper.ChapterAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&clResponse); err != nil {
		return nil, fmt.Errorf("error decoding chapter list: %w", err)
	}

	fmt.Printf("Amount of chapter is %d\n", clResponse.Data.Pagination.Total)

	chapterApiSlug := fmt.Sprintf("https://www.mangakakalot.gg/api/manga/%s/chapters?limit=%d", title, clResponse.Data.Pagination.Total)
	resp, err = http.Get(chapterApiSlug)
	if err != nil {
		return nil, fmt.Errorf("error getting chapter list: %w", err)
	}
	defer resp.Body.Close()

	var csResponse scraper.ChapterAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&csResponse); err != nil {
		return nil, fmt.Errorf("error decoding chapters: %w", err)
	}

	sort.Slice(csResponse.Data.Chapters, func(i, j int) bool {
		return csResponse.Data.Chapters[i].ChapterNum < csResponse.Data.Chapters[j].ChapterNum
	})

	var slugs []string
	fmt.Println("Saving chapters to database...")
	for _, chapter := range csResponse.Data.Chapters {
		slugs = append(slugs, chapter.ChapterSlug)
		_, err := s.Queries.SaveChapter(ctx, repository.SaveChapterParams{
			MangaID:     mangaID,
			ChapterSlug: chapter.ChapterSlug,
			ChapterNum:  pgtype.Float4{Float32: float32(chapter.ChapterNum), Valid: true},
			Status:      pgtype.Text{String: "pending", Valid: true},
		})
		if err != nil {
			fmt.Printf("Warning: failed to save chapter %s: %v\n", chapter.ChapterSlug, err)
		}
	}

	fmt.Printf("Total chapters: %d\n", len(slugs))
	return slugs, nil
}

// DownloadImagesWithDB downloads images for a chapter and persists page records
func (s *DownloadService) DownloadImagesWithDB(title string, slug string, chapterID int32) error {
	ctx := context.Background()

	downloadDir := filepath.Join("downloads", title, slug)
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	c := colly.NewCollector()
	imgCollector := colly.NewCollector()

	imgCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://www.mangakakalot.gg/")
	})

	imgCollector.OnResponse(func(r *colly.Response) {
		filename := filepath.Base(r.Request.URL.Path)

		if !scraper.ValidImageRegex.MatchString(filename) {
			fmt.Printf("Skipping non-manga image: %s\n", filename)
			return
		}

		filePathLocal := filepath.Join(downloadDir, filename)
		filePathURL := "/" + title + "/" + slug + "/" + filename

		pageID, err := s.Queries.SavePage(ctx, repository.SavePageParams{
			ChapterID: chapterID,
			ImageUrl:  r.Request.URL.String(),
			FileName:  pgtype.Text{String: filename, Valid: true},
			FilePath:  pgtype.Text{String: filePathURL, Valid: true},
		})
		if err != nil {
			fmt.Printf("Error saving page to DB: %v\n", err)
			return
		}

		if err := s.Queries.UpdatePageStatus(ctx, repository.UpdatePageStatusParams{
			Downloaded:   pgtype.Bool{Bool: false, Valid: true},
			ErrorMessage: pgtype.Text{Valid: false},
			DownloadedAt: pgtype.Timestamptz{Valid: false},
			PageID:       pageID,
		}); err != nil {
			fmt.Printf("Error marking page as downloading: %v\n", err)
		}

		if err := r.Save(filePathLocal); err != nil {
			fmt.Printf("Error saving image %v\n", err)
			s.Queries.UpdatePageStatus(ctx, repository.UpdatePageStatusParams{
				Downloaded:   pgtype.Bool{Bool: false, Valid: true},
				ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
				DownloadedAt: pgtype.Timestamptz{Valid: false},
				PageID:       pageID,
			})
		} else {
			fmt.Printf("Saved: %s\n", filename)
			s.Queries.UpdatePageStatus(ctx, repository.UpdatePageStatusParams{
				Downloaded:   pgtype.Bool{Bool: true, Valid: true},
				ErrorMessage: pgtype.Text{Valid: false},
				DownloadedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				PageID:       pageID,
			})
		}
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "cf_clearance="+scraper.GetCfClearanceToken())
		r.Headers.Set("Referer", "https://www.mangakakalot.gg")
		r.Headers.Set("User-Agent", scraper.GetCurrentUserAgent())
	})

	c.OnHTML("img", func(e *colly.HTMLElement) {
		imgURL := e.Attr("src")
		imgCollector.Visit(imgURL)
	})

	linkToDownloadFrom := fmt.Sprintf("https://www.mangakakalot.gg/manga/%s/%s", title, slug)
	if err := c.Visit(linkToDownloadFrom); err != nil {
		return fmt.Errorf("error visiting page: %w", err)
	}

	return nil
}
