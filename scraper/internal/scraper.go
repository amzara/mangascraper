package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type body struct {
	Url string
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
	Status    string               `json:"status"`
	Message   string               `json:"message"`
	Solution  flareSolverrSolution `json:"solution"`
}

func GetCookies() (map[string]string, string, error) {
	// FlareSolverr API request
	reqBody := flareSolverrRequest{
		Cmd:               "request.get",
		URL:               "https://www.mangakakalot.gg/manga/blue-lock",
		MaxTimeout:        60000,
		ReturnOnlyCookies: true,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	}

	resp, err := http.Post("http://flaresolverr:8191/v1", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result flareSolverrResponse
		err := json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, "", err
		}

		if result.Status != "ok" {
			return nil, "", fmt.Errorf("flaresolverr error: %s", result.Message)
		}

		// Convert cookies to map
		cookies := make(map[string]string)
		for _, c := range result.Solution.Cookies {
			cookies[c.Name] = c.Value
		}

		// Check for cf_clearance specifically
		if cfClearance, ok := cookies["cf_clearance"]; ok {
			fmt.Println("✓ cf_clearance retrieved:", cfClearance[:50]+"...")
		} else {
			fmt.Println("⚠ cf_clearance not found in cookies")
		}

		fmt.Println("Total cookies retrieved:", len(cookies))
		fmt.Println("User-Agent:", result.Solution.UserAgent)
		return cookies, result.Solution.UserAgent, nil
	}
	return nil, "", fmt.Errorf("failed to get cookies, status: %d", resp.StatusCode)
}

// helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DownloadImages(cookies map[string]string, userAgent string) {
	fmt.Println("Starting download with", len(cookies), "cookies...")
	
	// Print all cookies received from nodriver
	fmt.Println("=== COOKIES FROM NODRIVER ===")
	for name, value := range cookies {
		// Truncate long values for readability
		displayValue := value
		if len(displayValue) > 50 {
			displayValue = displayValue[:50] + "..."
		}
		fmt.Printf("  %s: %s\n", name, displayValue)
	}
	fmt.Println("=============================")

	url := "https://www.mangakakalot.gg/manga/hajime-no-ippo/chapter-1"

	// Use the User-Agent from FlareSolverr (MUST match cookies!)
	// If not provided, fall back to a fixed Chrome UA
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	// Create collector
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
	)

	// Build cookie header
	var cookieParts []string
	for name, value := range cookies {
		cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", name, value))
	}
	cookieHeader := strings.Join(cookieParts, "; ")
	fmt.Println("Cookie header length:", len(cookieHeader), "bytes")
	fmt.Println("=== FULL COOKIE HEADER ===")
	fmt.Println(cookieHeader)
	fmt.Println("==========================")

	// Use FlareSolverr's User-Agent (must match cookies!)
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", userAgent)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Referer", "https://www.mangakakalot.gg/")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		
		// Set the cookies
		if cookieHeader != "" {
			r.Headers.Set("Cookie", cookieHeader)
		}
		
		fmt.Println("[REQUEST] User-Agent:", r.Headers.Get("User-Agent"))
		fmt.Println("[REQUEST] Visiting:", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("[RESPONSE] Status: %d\n", r.StatusCode)
		
		if r.StatusCode == 403 {
			fmt.Println("[RESPONSE] BLOCKED by Cloudflare!")
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("[ERROR] %v\n", err)
		if r != nil {
			fmt.Printf("[ERROR] Status: %d\n", r.StatusCode)
		}
	})

	// Find images
	c.OnHTML(".container-chapter-reader img", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		fmt.Println("[FOUND] Image:", src)
	})

	// Fallback - find any images
	c.OnHTML("img", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		if src != "" && (strings.Contains(src, "http") || strings.Contains(src, "//")) {
			fmt.Printf("[DEBUG] img src=%s\n", src)
		}
	})

	// Rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*mangakakalot.gg",
		Parallelism: 1,
		Delay:       2 * time.Second,
	})

	// Visit
	err := c.Visit(url)
	if err != nil {
		fmt.Println("[VISIT ERROR]", err)
	}

	fmt.Println("Waiting...")
	c.Wait()
	fmt.Println("Done!")
}
