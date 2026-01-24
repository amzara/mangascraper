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
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br, zstd")
		r.Headers.Set("Accept-Language", "en-CA,en-GB;q=0.9,en-US;q=0.8,en;q=0.7")
		r.Headers.Set("Cookie", "__cflb=02DiuD2F3RAeJGVVKWvfHx9L6dXK13KrhbngRPZNy7fAs; cf_clearance=Zq0AR_5jvoYeFUe2zXcLd0OVt_PCuzRD4ofTRrJE6RI-1769264208-1.2.1.1-_DHlmlmh1DND8Nv.O2CUlThblzA2f5gFm..Dbqb6KP_2I7bjUgjO71fsnq.lGT0oetC5sGwKAN_GsD9Or4B2lzODiT2cKdznkhuE638oGOHo7lrwnwy9dYrM8bOJNPPYPRLcTfLA3zeGGy8lpxcpKvwvky.Fm7MmO.RRAl7gnxQFnkr2nV3iG33l58qtCPvlqVUzC7ZquXb7QlLtoYwk6X.BCozmcSZBF6q4fjm37nw; _pu_clickadu=1769264273381; _pu_last=monetag; _pu_monetag=1769264768380")
		r.Headers.Set("Dnt", "1")
		r.Headers.Set("If-Modified-Since", "Sat, 24 Jan 2026 14:22:37 GMT")
		r.Headers.Set("Priority", "u=0, i")
		r.Headers.Set("Sec-Ch-Ua", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
		r.Headers.Set("Sec-Ch-Ua-Arch", `"x86"`)
		r.Headers.Set("Sec-Ch-Ua-Bitness", `"64"`)
		r.Headers.Set("Sec-Ch-Ua-Full-Version", `"143.0.7499.169"`)
		r.Headers.Set("Sec-Ch-Ua-Full-Version-List", `"Google Chrome";v="143.0.7499.169", "Chromium";v="143.0.7499.169", "Not A(Brand";v="24.0.0.0"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", `?0`)
		r.Headers.Set("Sec-Ch-Ua-Model", `""`)
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Linux"`)
		r.Headers.Set("Sec-Ch-Ua-Platform-Version", `""`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	})

	//anonymous function, callback, closure
	//todo: compare with named function so u understand better

	c.OnHTML("img", func(e *colly.HTMLElement) {
		fmt.Println(e.Attr("src"))

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("https://www.mangakakalot.gg")
}
