package scraper

// This package contains the manga scraping functionality

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

type body struct {
	Url string
}

var cfClearanceToken string = ""
var currentUserAgent string = ""

// ValidImageRegex matches valid manga page filenames like "001.webp"
var ValidImageRegex = regexp.MustCompile(`^\d+\.webp$`)

// GetCfClearanceToken returns the current cf_clearance token
func GetCfClearanceToken() string {
	return cfClearanceToken
}

// GetCurrentUserAgent returns the current user agent
func GetCurrentUserAgent() string {
	return currentUserAgent
}

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

// ChapterAPIResponse represents the mangakakalot chapter list API response
type ChapterAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
		Chapters []struct {
			ChapterSlug string  `json:"chapter_slug"`
			ChapterNum  float64 `json:"chapter_num"`
		} `json:"chapters"`
	} `json:"data"`
}

// GetCookies retrieves Cloudflare clearance cookies from FlareSolverr
func GetCookies() error {
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

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("flareSolverr returned status %d: %s", resp.StatusCode, string(body))
	}

	var result flareSolverrResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding flareSolverr response: %w", err)
	}

	cookies := make(map[string]string)
	for _, c := range result.Solution.Cookies {
		cookies[c.Name] = c.Value
	}

	if cfClearance, ok := cookies["cf_clearance"]; ok {
		fmt.Println("Retrieved cf clearance token", cfClearance[:50]+"...")
		cfClearanceToken = cfClearance
		fmt.Printf("Token stored successfully\n")
	} else {
		return fmt.Errorf("cf_clearance not found in cookies")
	}

	fmt.Println("Total cookies retrieved:", len(cookies))
	fmt.Println("User-Agent:", result.Solution.UserAgent)
	currentUserAgent = result.Solution.UserAgent
	fmt.Println("DONE getting cf_clearance token")
	return nil
}

// DownloadImages downloads all images for a single chapter without touching the database
func DownloadImages(title string, slug string) {
	downloadDir := filepath.Join("downloads", title, slug)

	if err := os.MkdirAll(downloadDir, 0755); err != nil {
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

		if !ValidImageRegex.MatchString(filename) {
			fmt.Printf("Skipping non-manga image: %s\n", filename)
			return
		}

		fpath := filepath.Join(downloadDir, filename)
		if err := r.Save(fpath); err != nil {
			fmt.Printf("Error saving image %v\n", err)
		} else {
			fmt.Printf("Saved: %s\n", filename)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "cf_clearance="+cfClearanceToken)
		r.Headers.Set("Referer", "https://www.mangakakalot.gg")
		r.Headers.Set("User-Agent", currentUserAgent)
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
	if err := c.Visit(linkToDownloadFrom); err != nil {
		fmt.Printf("Error visiting page: %v\n", err)
	}
}

// GetChapterList fetches the list of chapter slugs from the mangakakalot API
func GetChapterList(title string) ([]string, error) {
	chapterApi := fmt.Sprintf("https://www.mangakakalot.gg/api/manga/%s/chapters?limit=999", title)
	fmt.Printf("Chapter API is %s", chapterApi)
	resp, err := http.Get(chapterApi)
	if err != nil {
		fmt.Println("err getting chapter amount")
	}
	defer resp.Body.Close()

	var clResponse ChapterAPIResponse
	_ = json.NewDecoder(resp.Body).Decode(&clResponse)

	fmt.Printf("Amount of chapter is %d", clResponse.Data.Pagination.Total)

	chapterApiSlug := fmt.Sprintf("https://www.mangakakalot.gg/api/manga/%s/chapters?limit=%d", title, clResponse.Data.Pagination.Total)
	resp, err = http.Get(chapterApiSlug)
	defer resp.Body.Close()

	var csResponse ChapterAPIResponse
	_ = json.NewDecoder(resp.Body).Decode(&csResponse)
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

func deleteManga(title string) error {

	return nil
}

// min returns the smaller of a or b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
