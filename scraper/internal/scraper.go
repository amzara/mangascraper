package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	uaFake "github.com/lib4u/fake-useragent"
)

type body struct {
	Url string
}

type reqCookies struct {
	Url     string            `json:"url"`
	Success bool              `json:"success"`
	Cookies map[string]string `json:"cookies"`
}

func GetCookies() (map[string]string, error) {
	jsonBody := map[string]string{"url": "https://www.mangakakalot.gg/official"}
	jsonData, err := json.Marshal(jsonBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://nodriver:8000/getCfCookies", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result reqCookies
		err := json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, err
		}

		fmt.Println("Cookies retrieved:", len(result.Cookies))
		return result.Cookies, nil
	}
	return nil, fmt.Errorf("failed to get cookies, status: %d", resp.StatusCode)
}

// helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DownloadImages(cookies map[string]string) {
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

	// Init fake user-agent
	ua, err := uaFake.New()
	if err != nil {
		fmt.Println("Error creating user agent faker:", err)
		return
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

	// Set random fake user agent and other headers on every request
	c.OnRequest(func(r *colly.Request) {
		// Set fake user agent - use GetRandom() directly on ua
		r.Headers.Set("User-Agent", ua.GetRandom())
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
	err = c.Visit(url)
	if err != nil {
		fmt.Println("[VISIT ERROR]", err)
	}

	fmt.Println("Waiting...")
	c.Wait()
	fmt.Println("Done!")
}
