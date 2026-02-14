package main

import (
	// "fmt"
	// "log"
	"fmt"
	scraper "mangascraper/internal"
	"os"
)

func main() {

	err := scraper.GetCookies()

	if err != nil {
		fmt.Println(err)
	}

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
	os.Exit(0)
}
