package main

import (
	"encoding/json"
	"fmt"
	"mangascraper/internal/db"
	scraper "mangascraper/internal"
	"net/http"
	"time"
)

func healthCheck() bool {

	resp, err := http.Get("http://flaresolverr:8191/health") //change to service name in docker compose
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var r struct {
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&r)

	fmt.Println(resp.Body)

	return r.Status == "ok"

}

func main() {
	// Initialize database
	if err := db.InitDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer db.CloseDB()

	var fsIsAlive = false

	for fsIsAlive == false {
		if healthCheck() {
			fsIsAlive = true
		} else {
			fmt.Println("Flaresolver not ready, waiting")
			time.Sleep(10 * time.Second)
		}

	}

	scraper.GetCookies()
	scraper.DownloadManga("choujin-x")

	// slugs := scraper.GetChapterList("kagurabachi")
	// scraper.BeginJobPool("kagurabachi", slugs, 16)

	// scraper.DownloadImages("blue-lock", "chapter-330")
	fmt.Println("donezo")

	// Get cookies from the local server

	// if err != nil {
	// 	log.Printf("Warning: Could not get cookies: %v", err)
	// 	log.Println("Proceeding without cookies...")
	// } else {
	// 	fmt.Printf("Successfully retrieved %d cookies\n", len(cookies))
	// }

	// Chapter URL to download
	// chapterURL := "https://www.mangakakalot.gg/manga/hajime-no-ippo/chapter-1513"

	// fmt.Printf("Downloading images from: %s\n", chapterURL)

	// Download images from the chapter
	// if err := scraper.DownloadChapterImages(chapterURL, cookies); err != nil {
	// 	log.Fatalf("Error downloading chapter images: %v", err)
	// }

	// fmt.Println("Download complete! Check the 'downloads' directory.")

	time.Sleep(10 * time.Hour)

}
