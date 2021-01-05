package main

// Performs a breadth-first search of the web graph, starting with a fixed set of seed pages.

import (
	//"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const numGoroutines = 16
var doNotDownload = [...]string{"js", "png", "jpg", "jpeg", "webp", "zip"}
var startpages = []string{
	"https://wikipedia.org",
	"https://reddit.com",
	"https://yahoo.com",
	"https://github.com",
	"https://msn.com",
	"https://sueddeutsche.de",
	"https://nytimes.com",
	"https://bbc.com",
	"https://newyorker.com",
	"https://apnews.com",
	"https://nature.com",
	"https://economist.com",
	"https://wired.com",
	"https://mashable.com",
	"https://quora.com",
	"https://news.ycombinator.com",
}

type scrapeResult struct {
	website string
	links   []string
}

func main() {
	var current []string
	var toVisit = startpages
	var visited = make(map[string]bool, 0)
	//var bytesDownloaded int64
	var links = make(map[string][]string, 0)

	for len(toVisit) >= numGoroutines && len(visited) < 200 {
		messages := make(chan scrapeResult, numGoroutines)
		current, toVisit = toVisit[0:numGoroutines], toVisit[numGoroutines:]

		// ensure no duplicate visits
		for i := range current {
			for visited[current[i]] {
				current[i], toVisit = toVisit[0], toVisit[1:]
			}
			visited[current[i]] = true
		}

		// starting scraping in goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(x int) { messages <- scrapeLinksFromPage(current[x]) }(i)
		}

		// receive and merge results
		for i := 0; i < numGoroutines; i++ {
			res := <-messages

			for _, link := range res.links {
				links[res.website] = append(links[res.website], link)
				if !visited[link] {
					toVisit = append(toVisit, link)
				}
			}
		}
	}

	PageRank(links)
}

// Downloads the website and parses the HTML to find anchor tags.
func scrapeLinksFromPage(website string) scrapeResult {
	var links []string

	// download website
	res, err := http.Get(website)
	if err != nil {
		log.Println(err)
		return scrapeResult{website, links}
	}
	defer res.Body.Close()

	// store body in string to count bytes
	body, err := ioutil.ReadAll(res.Body)
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
		switch tt {
		case html.ErrorToken:
			break loop
		case html.TextToken:
			//fmt.Println(tt)
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
