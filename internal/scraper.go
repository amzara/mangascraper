package scraper

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

func downloadWorkers(id int, jobs <-chan DownloadJob, wg *sync.WaitGroup) {
	defer wg.Done() //decrement counter when this job finish

	for job := range jobs {
		fmt.Printf("Worker %d: processing %s\n", id, job.Slug)

		DownloadImages(job.Title, job.Slug)
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

	err := GetCookies()
	if err != nil {
		fmt.Println("Failed to get cf clearance token")
		return err
	}

	slugs, err := GetChapterList(title)

	if err != nil {
		return fmt.Errorf("Failed to get chapter %s\n", err)
	}

	if len(slugs) == 0 {
		return fmt.Errorf("No chapter found for %s\n", title)
	}

	fmt.Printf("Found %d chapters for %s, starting download . . . \n", len(slugs), title)

	BeginJobPool(title, slugs, 8)

	fmt.Println("Finish download %s", title)

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

func BeginJobPool(title string, slugs []string, numWorkers int) {
	jobs := make(chan DownloadJob, len(slugs))
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go downloadWorkers(w, jobs, &wg)
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

// helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
