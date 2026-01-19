package main

import (
	"fmt"

	"github.com/gocolly/colly/v2"
)

func main() {

	// var aggregatorSite String = "https//mangakakalot.gg"
	// var targetManga String = "Rust"
	// var targetChapter int = 1

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://www.mangakakalot.gg/")

	})

	//anonymous function, callback, closure
	//todo: compare with named function so u understand better

	c.OnHTML("img", func(e *colly.HTMLElement) {
		fmt.Println(e.Attr("src"))

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("https://www.mangakakalot.gg/manga/rust/chapter-37")
}
