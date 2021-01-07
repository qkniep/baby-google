package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"github.com/valyala/fasthttp"
	"runtime"
	"strings"
	"time"
)

const THREADS = 64
var DO_NOT_DOWNLOAD = [...]string{"js", "png", "jpg", "jpeg", "webp", "zip"}
var STARTPAGES = []string{
	"https://wikipedia.org",
	"https://reddit.com",
	"https://yahoo.com",
	"https://github.com",
}

type scrapeResult struct {
	website string
	links   []string
}

func main() {
	numcpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numcpu)

	var toVisit = STARTPAGES
	var visited = make(map[string]bool, 0)
	//var bytesDownloaded int64
	var links = make(map[string][]string, 0)

	var messages = make(chan scrapeResult, THREADS)
	var currentlyCrawling = 0

	for len(links) < 1000 {
		if currentlyCrawling > 0 {
			res := <-messages
			links[res.website] = append(links[res.website], res.links...)
			toVisit = append(toVisit, res.links...)
			currentlyCrawling--
		}

		for len(toVisit) > 0 && currentlyCrawling < THREADS {
			website := toVisit[0]
			toVisit = toVisit[1:]
			if visited[website] {
				continue
			}
			visited[website] = true
			go func(url string) { messages <- scrapeLinksFromPage(url) }(website)
			currentlyCrawling++
		}
	}

	fmt.Println("Calculating PageRank...")
	PageRank(links)
}

func scrapeLinksFromPage(website string) scrapeResult {
	var links []string

	_, body, err := fasthttp.GetTimeout(nil, website, time.Duration(time.Millisecond*500))
	if err != nil {
		log.Println(err)
		return scrapeResult{website, links}
	}
	bodyReader := strings.NewReader(string(body))

	// print stats
	/*bytesDownloaded += int64(len(body))
	fmt.Printf("Current website: %v\n", website)
	fmt.Printf("Number of sites to vist: %v\n", len(toVisit))
	if bytesDownloaded < 1000 {
		fmt.Printf("Downloaded: %v B\n", bytesDownloaded)
	} else if bytesDownloaded < 1000000 {
		fmt.Printf("Downloaded: %v KB\n", bytesDownloaded / 1000)
	} else if bytesDownloaded < 1000000000 {
		fmt.Printf("Downloaded: %v MB\n", bytesDownloaded / 1000000)
	}
	fmt.Printf("Visited pages: %v\n", len(visited))
	fmt.Println()*/

	// search anchor tags
	htmlTokens := html.NewTokenizer(bodyReader)
	loop:
	for {
		tt := htmlTokens.Next()
		//fmt.Printf("%T", tt)
		switch tt {
		case html.ErrorToken:
			break loop
		case html.TextToken:
			//fmt.Println(tt)
			// TODO parse website content here
		case html.StartTagToken:
			t := htmlTokens.Token()
			if t.Data == "a" {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						absURL, success := Resolve(website, attr.Val)
						if success {
							links = append(links, absURL)
						}
					}
				}
			}
		}
	}

	return scrapeResult{website, links}
}
