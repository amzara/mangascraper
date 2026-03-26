package scraper

// This package contains the manga scraping functionality

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	_ "time"

	"mangascraper/internal/db"

	"github.com/gocolly/colly/v2"
)

type body struct {
	Url string
}

var cfClearanceToken string = ""
var currentUserAgent string = ""

// FlareSolverr request/response types
type flareSolverrRequest struct {
	Cmd               string `json:"cmd"`
	URL               string `json:"url"`
	MaxTimeout        int    `json:"maxTimeout"`
	ReturnOnlyCookies bool   `json:"returnOnlyCookies"`
}

type flareSolverrCookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

type flareSolverrSolution struct {
	Cookies   []flareSolverrCookie `json:"cookies"`
	UserAgent string               `json:"userAgent"`
}

type flareSolverrResponse struct {
	Status   string               `json:"status"`
	Message  string               `json:"message"`
	Solution flareSolverrSolution `json:"solution"`
}

type chapterApiResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
		Chapters []struct {
			ChapterSlug string `json:"chapter_slug"`
		} `json:"chapters"`
	} `json:"data"`
}

type DownloadJob struct {
	Title string
	Slug  string
}

func downloadWorkers(id int, mangaID int, jobs <-chan DownloadJob, wg *sync.WaitGroup) {
	defer wg.Done() //decrement counter when this job finish

	for job := range jobs {
		fmt.Printf("Worker %d: processing %s\n", id, job.Slug)

		// Get chapter ID from database
		chapter, err := db.GetChapterBySlug(mangaID, job.Slug)
		if err != nil {
			fmt.Printf("Worker %d: failed to get chapter %s: %v\n", id, job.Slug, err)
			continue
		}
		if chapter == nil {
			fmt.Printf("Worker %d: chapter %s not found in database\n", id, job.Slug)
			continue
		}

		// Mark chapter as downloading
		if err := db.MarkChapterDownloading(chapter.ChapterID); err != nil {
			fmt.Printf("Worker %d: failed to mark chapter as downloading: %v\n", id, err)
		}

		// Download images and save pages to database
		err = DownloadImagesWithDB(job.Title, job.Slug, chapter.ChapterID)
		if err != nil {
			// Mark chapter as error
			db.MarkChapterError(chapter.ChapterID, err.Error())
			fmt.Printf("Worker %d: error downloading %s: %v\n", id, job.Slug, err)
		} else {
			// Mark chapter as downloaded
			db.MarkChapterDownloaded(chapter.ChapterID)
			fmt.Printf("Worker %d: completed %s\n", id, job.Slug)
		}
	}
}

func GetCookies() error {
	// FlareSolverr API request
	reqBody := flareSolverrRequest{
		Cmd:               "request.get",
		URL:               "https://www.mangakakalot.gg/manga/blue-lock",
		MaxTimeout:        60000,
		ReturnOnlyCookies: true,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error on marshaling json")
		return err
	}

	resp, err := http.Post("http://flaresolverr:8191/v1", "application/json", bytes.NewReader(jsonData))

	if err != nil {
		fmt.Printf("err sending POST to fs %s\n", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result flareSolverrResponse
		err := json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			fmt.Printf("Error on decoding json %s\n", err)
		}

		// Convert cookies to map
		cookies := make(map[string]string)
		for _, c := range result.Solution.Cookies {
			cookies[c.Name] = c.Value
		}

		// Check for cf_clearance specifically
		if cfClearance, ok := cookies["cf_clearance"]; ok {
			fmt.Println("✓ cf_clearance retrieved:", cfClearance[:50]+"...")
			cfClearanceToken = cfClearance
			fmt.Printf("This is declared cfClearance = %s\n", cfClearanceToken)

		} else {
			fmt.Println("⚠ cf_clearance not found in cookies")
		}

		fmt.Println("Total cookies retrieved:", len(cookies))
		fmt.Println("User-Agent:", result.Solution.UserAgent)
		currentUserAgent = result.Solution.UserAgent
		fmt.Println("DONE getting cf_clearance token")
	}
	return nil
}

//todo: make thi

func DownloadManga(title string) error {
	// 1. Save manga to database
	mangaID, err := db.SaveManga(title, title)
	if err != nil {
		return fmt.Errorf("failed to save manga: %w", err)
	}
	fmt.Printf("Manga '%s' saved with ID: %d\n", title, mangaID)

	err = GetCookies()
	if err != nil {
		fmt.Println("Failed to get cf clearance token")
		return err
	}

	// 2. Get chapter list and save chapters to database
	slugs, err := GetChapterListWithDB(mangaID, title)
	if err != nil {
		return fmt.Errorf("failed to get chapter list: %w", err)
	}

	if len(slugs) == 0 {
		return fmt.Errorf("no chapter found for %s", title)
	}

	fmt.Printf("Found %d chapters for %s, starting download . . . \n", len(slugs), title)

	BeginJobPool(title, mangaID, slugs, 8)

	fmt.Printf("Finish download %s\n", title)

	return nil
}

var validImageRegex = regexp.MustCompile(`^\d+\.webp$`)

func DownloadImages(title string, slug string) { //this func should take in title, and list of chapters. then you spawn goroutines to parallel download
	downloadDir := filepath.Join("downloads", title, slug)

	err := os.MkdirAll(downloadDir, 0755)

	if err != nil {
		fmt.Printf("Problem creating dir, %v", err)
	}

	fmt.Printf("Current clearance token is %s\n", cfClearanceToken)

	c := colly.NewCollector()

	imgCollector := colly.NewCollector()

	imgCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://www.mangakakalot.gg/")
		fmt.Println("Downloading:", r.URL.String())

	})

	imgCollector.OnResponse(func(r *colly.Response) {
		fmt.Println("Response received")
		filename := filepath.Base(r.Request.URL.Path)

		if !validImageRegex.MatchString(filename) {
			fmt.Printf("Skipping non-manga image: %s\n", filename)
			return
		}

		filepath := filepath.Join(downloadDir, filename)
		err := r.Save(filepath)
		if err != nil {
			fmt.Printf("Error saving image %v\n", err)
		} else {
			fmt.Printf("Saved: %s\n", filename)
		}
	})

	c.OnRequest(func(r *colly.Request) {

		r.Headers.Set("Cookie", "cf_clearance="+cfClearanceToken)
		r.Headers.Set("Referer", "https://www.mangakakalot.gg")
		r.Headers.Set("User-Agent", ""+currentUserAgent)

	})

	c.OnResponse(func(r *colly.Response) {

		contentType := r.Headers.Get("Content-Type")

		if strings.Contains(contentType, "image/webp") {
			filename := filepath.Base(r.Request.URL.Path)
			r.Save("downloads/" + filename)
		}

	})

	c.OnHTML("img", func(e *colly.HTMLElement) {
		imgURL := e.Attr("src")
		fmt.Println("Found image:", imgURL)
		imgCollector.Visit(imgURL)
	})
	fmt.Println(slug)
	fmt.Printf(title)
	linkToDownloadFrom := fmt.Sprintf("https://www.mangakakalot.gg/manga/%s/%s", title, slug)
	err = c.Visit(linkToDownloadFrom)
	if err != nil {
		fmt.Printf("Error visiting page: %v\n", err)
	}

	//on succesful inserts

}

func BeginJobPool(title string, mangaID int, slugs []string, numWorkers int) {
	jobs := make(chan DownloadJob, len(slugs))
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go downloadWorkers(w, mangaID, jobs, &wg)
	}

	for _, slug := range slugs {
		jobs <- DownloadJob{Title: title, Slug: slug}
	}

	close(jobs)

	wg.Wait()

}

func GetChapterList(title string) ([]string, error) {
	var chapterApi string = fmt.Sprintf("https://www.mangakakalot.gg/api/manga/%s/chapters?limit=999", title)
	fmt.Printf("Chapter API is %s", chapterApi)
	resp, err := http.Get(chapterApi)
	if err != nil {
		fmt.Println("err getting chapter amount")
	}

	//add interceptor middleware for 401 unauthorized, call GetCookies and try again?
	//is there any golang native way to intercept status codes? like attaching interceptor to httpClient

	defer resp.Body.Close() //dont forget do this ree amza

	var clResponse chapterApiResponse

	err = json.NewDecoder(resp.Body).Decode(&clResponse)

	fmt.Printf("Amount of chapter is %d", clResponse.Data.Pagination.Total)

	chapterApiSlug := fmt.Sprintf("https://www.mangakakalot.gg/api/manga/%s/chapters?limit=%d", title, clResponse.Data.Pagination.Total)
	//redundant, just do limit = -1 LOL
	resp, err = http.Get(chapterApiSlug)

	defer resp.Body.Close()

	var csResponse chapterApiResponse

	err = json.NewDecoder(resp.Body).Decode(&csResponse)

	if err != nil {
		fmt.Println("Fail to decode")
	}

	var slugs []string
	for _, chapter := range csResponse.Data.Chapters {
		slugs = append(slugs, chapter.ChapterSlug)
	}

	fmt.Println(slugs)
	return slugs, nil

}

// GetChapterListWithDB fetches chapters and saves them to the database
func GetChapterListWithDB(mangaID int, title string) ([]string, error) {
	var chapterApi string = fmt.Sprintf("https://www.mangakakalot.gg/api/manga/%s/chapters?limit=999", title)
	fmt.Printf("Chapter API is %s\n", chapterApi)
	resp, err := http.Get(chapterApi)
	if err != nil {
		return nil, fmt.Errorf("error getting chapter amount: %w", err)
	}
	defer resp.Body.Close()

	var clResponse chapterApiResponse
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

	var csResponse chapterApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&csResponse); err != nil {
		return nil, fmt.Errorf("error decoding chapters: %w", err)
	}

	var slugs []string
	fmt.Println("Saving chapters to database...")
	for _, chapter := range csResponse.Data.Chapters {
		slugs = append(slugs, chapter.ChapterSlug)
		// Save chapter to database
		_, err := db.SaveChapter(mangaID, chapter.ChapterSlug)
		if err != nil {
			fmt.Printf("Warning: failed to save chapter %s: %v\n", chapter.ChapterSlug, err)
		}
	}

	fmt.Printf("Total chapters: %d\n", len(slugs))
	return slugs, nil
}

// DownloadImagesWithDB downloads images and saves page records to the database
func DownloadImagesWithDB(title string, slug string, chapterID int) error {
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

		if !validImageRegex.MatchString(filename) {
			fmt.Printf("Skipping non-manga image: %s\n", filename)
			return
		}

		filepath := filepath.Join(downloadDir, filename)
		
		// Save page to database with pending status
		pageID, err := db.SavePage(chapterID, r.Request.URL.String(), filename)
		if err != nil {
			fmt.Printf("Error saving page to DB: %v\n", err)
			return
		}

		// Mark page as downloading
		if err := db.MarkPageDownloading(pageID); err != nil {
			fmt.Printf("Error marking page as downloading: %v\n", err)
		}

		// Save the file
		if err := r.Save(filepath); err != nil {
			fmt.Printf("Error saving image %v\n", err)
			// Mark page as error
			db.MarkPageError(pageID, err.Error())
		} else {
			fmt.Printf("Saved: %s\n", filename)
			// Mark page as downloaded
			db.MarkPageDownloaded(pageID)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "cf_clearance="+cfClearanceToken)
		r.Headers.Set("Referer", "https://www.mangakakalot.gg")
		r.Headers.Set("User-Agent", ""+currentUserAgent)
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

// helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
