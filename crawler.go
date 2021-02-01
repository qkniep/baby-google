package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	userAgent = "Baby-Google"
	pagesToCrawl = 1000
	numGoroutines = 16
	toVisitFile = "toVisit.gob"
	visitedFile = "visited.gob"
)

//var doNotDownload = [...]string{"js", "png", "jpg", "jpeg", "webp", "zip"}
var startpages = []string{
	"https://wikipedia.org",
	"https://reddit.com",
	"https://yahoo.com",
	"https://msn.com",
}

type scrapeResult struct {
	website string
	links   []string
	size    int
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var toVisit []string
	load(&toVisit, toVisitFile)
	//var toVisit = startpages
	var visited map[string]bool
	load(&visited, visitedFile)
	//var bytesDownloaded int64
	var links = make(map[string][]string, 0)

	var messages = make(chan scrapeResult, numGoroutines)
	var currentlyCrawling = 0

	var client = &fasthttp.Client{
		Name: userAgent,
		ReadBufferSize: 8192,
		MaxConnsPerHost: 8,
		ReadTimeout: 2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	var bytesDownloaded int
	start := time.Now()

	for len(links) < pagesToCrawl {
		// wait for on goroutine to finish
		if currentlyCrawling > 0 {
			res := <-messages
			links[res.website] = append(links[res.website], res.links...)
			toVisit = append(toVisit, res.links...)
			bytesDownloaded += res.size
			currentlyCrawling--
		}

		// assign new tasks until max. number of goroutines is reached
		for currentlyCrawling < numGoroutines && len(toVisit) > 0 {
			website := toVisit[0]
			toVisit = toVisit[1:]
			if visited[website] {
				continue
			}
			visited[website] = true
			go func(url string) { messages <- scrapeLinksFromPage(client, url) }(website)
			currentlyCrawling++
		}

		// print stats
		if bytesDownloaded < 1e9 {
			fmt.Printf("Downloaded: %.1f MB\n", float32(bytesDownloaded) / 1e6)
		} else if bytesDownloaded < 1e12 {
			fmt.Printf("Downloaded: %.1f GB\n", float32(bytesDownloaded) / 1e9)
		}
		fmt.Println("Visited pages:", len(visited))
		fmt.Println("Number of sites to vist:", len(toVisit))
		fmt.Println("Currently crawling:", currentlyCrawling)
		fmt.Printf("Average crawl rate: %.2f pages/sec\n", float64(len(visited)) / time.Since(start).Seconds())
		fmt.Println()
	}

	dump(toVisit, toVisitFile)
	dump(visited, visitedFile)

	fmt.Println("Calculating PageRank...")
	PageRank(links)
}

func scrapeLinksFromPage(client *fasthttp.Client, website string) scrapeResult {
	var links []string

	_, body, err := client.Get(nil, website)
	if err != nil {
		log.Println(err)
		return scrapeResult{website, links, 0}
	}
	dumpPageToDisk(website, body)
	bodyReader := strings.NewReader(string(body))

	// search anchor tags
	htmlTokens := html.NewTokenizer(bodyReader)
loop:
	for {
		tt := htmlTokens.Next()
		switch tt {
		case html.ErrorToken:
			break loop
		case html.TextToken:
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

	return scrapeResult{website, links, len(body)}
}

func dumpPageToDisk(website string, body []byte) {
	hash := sha256.Sum256([]byte(website))
	filename := hex.EncodeToString(hash[:])
	err := ioutil.WriteFile("data/" + filename, body, 0644)
	if err != nil {
		log.Panicf("Error: Failed to write website %v to disk.\n", website)
	}
}

func dump(obj interface{}, filename string) {
	f, err := os.Create("data/" + filename)
	if err != nil {
		log.Panicf("Error: Failed to create file %v.\n", filename)
	}
	w := bufio.NewWriter(f)
	enc := gob.NewEncoder(w)
	err = enc.Encode(obj)
	if err != nil {
		log.Panicf("Error: Failed to dump object %v to disk.\n", filename)
	}
}

func load(obj interface{}, filename string) {
	f, err := os.Open("data/" + filename)
	if err != nil {
		log.Panicf("Error: Failed to open file %v.\n", filename)
	}
	r := bufio.NewReader(f)
	dec := gob.NewDecoder(r)
	err = dec.Decode(obj)
	if err != nil {
		log.Panicf("Error: Failed to load object %v from disk.\n", filename)
	}
}
