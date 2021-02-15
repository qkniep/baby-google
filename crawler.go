package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/temoto/robotstxt"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	userAgent = "Baby-Google"
	pagesToCrawl = 500
	numGoroutines = 32
	dataDir = "data"
	toVisitFile = "toVisit.gob"
	visitedFile = "visited.gob"
)

//var doNotDownload = [...]string{"js", "png", "jpg", "jpeg", "webp", "zip"}

/*func main() {
	var visited map[string]bool
	load(&visited, visitedFile)
	var links = make(map[string][]string, 0)

	for site := range visited {
		res := scrapeLinksFromPage(site)
		links[res.Website] = append(links[res.Website], res.Links...)
	}

	fmt.Println("Calculating PageRank...")
	PageRank(links)
}*/

func main() {
	var toVisit []string
	load(&toVisit, toVisitFile)
	//var toVisit = startpages
	var visited map[string]bool
	load(&visited, visitedFile)
	//var visited = make(map[string]bool, 0)
	var bytesDownloaded int64
	var links = make(map[string][]string, 0)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(toVisit), func(i, j int) {
		toVisit[i], toVisit[j] = toVisit[j], toVisit[i]
	})

	var messages = make(chan ScrapeResult, numGoroutines)
	var currentlyCrawling = 0

	var client = &fasthttp.Client{
		Name: userAgent,
		ReadBufferSize: 8192,
		MaxConnsPerHost: 8,
		ReadTimeout: 2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	start := time.Now()

	for len(links) < pagesToCrawl {
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
		fmt.Println("Crawled pages:", len(links))
		fmt.Println("Sites to vist:", len(toVisit))
		fmt.Printf("Average crawl rate: %.2f pages/sec\n", float64(len(links)) / time.Since(start).Seconds())
		fmt.Println()

		// wait for one goroutine to finish
		res := <-messages
		if res.Err {
			toVisit = append(toVisit, res.Website)
			visited[res.Website] = false
		} else {
			links[res.Website] = append(links[res.Website], res.Links...)
			toVisit = append(toVisit, res.Links...)
			bytesDownloaded += int64(res.Size)
			fmt.Println("Finished crawling:", res.Website)
		}
		currentlyCrawling--
	}

	// remove duplicates from toVisit
	for i := 0; i < len(toVisit); i++ {
		if visited[toVisit[i]] {
			toVisit[i] = toVisit[len(toVisit)-1]
			toVisit = toVisit[:len(toVisit)-1]
			i--
		}
	}

	dump(toVisit, toVisitFile)
	dump(visited, visitedFile)

	fmt.Println("Calculating PageRank...")
	PageRank(links)
}

/*func scrapeLinksFromPage(website string) ScrapeResult {
	hash := sha256.Sum256([]byte(website))
	filename := hex.EncodeToString(hash[:])
	file, err := os.Open(dataDir + "/" + filename)
	if err != nil {
		return ScrapeResult{website, []string{}, 0, true}
	}

	return ScrapeLinksFromHTML(file, website)
}*/

func grabRobotsTxt(client *fasthttp.Client, website string) (r *robotstxt.RobotsData, err error) {
	// get robots.txt URL
	siteURL, err := url.Parse(website)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	robotsURL := siteURL.Scheme + "://" + siteURL.Host + "/robots.txt"

	// get robots.txt file from disk or website
	tryLoadPageFromDisk(robotsURL)
	status, body, err := client.Get(nil, robotsURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	dumpPageToDisk(robotsURL, body)
	return robotstxt.FromStatusAndBytes(status, body)
}

func scrapeLinksFromPage(client *fasthttp.Client, website string) ScrapeResult {
	// download & check robots.txt
	robots, err := grabRobotsTxt(client, website)
	if err != nil {
		log.Println(err)
		return ScrapeResult{website, []string{}, 0, true}
	}
	siteURL, err := url.Parse(website)
	if err != nil {
		log.Println(err)
		return ScrapeResult{website, []string{}, 0, true}
	}
	group := robots.FindGroup(userAgent)
	if !group.Test(siteURL.Path) {
		log.Println("Access disallowed by robots.txt")
		return ScrapeResult{website, []string{}, 0, false}
	}

	// download website body
	_, body, err := client.Get(nil, website)
	if err != nil {
		log.Println(err)
		return ScrapeResult{website, []string{}, 0, true}
	}
	dumpPageToDisk(website, body)
	bodyReader := strings.NewReader(string(body))

	// scrape website for anchor tags (links)
	res := ScrapeLinksFromHTML(bodyReader, website)
	return ScrapeResult{res.Website, res.Links, len(body), false}
}

func dumpPageToDisk(website string, body []byte) {
	hash := sha256.Sum256([]byte(website))
	filename := hex.EncodeToString(hash[:])
	err := ioutil.WriteFile(dataDir + "/" + filename, body, 0644)
	if err != nil {
		log.Panicf("Error: Failed to write website %v to disk.\n", website)
	}
}

func tryLoadPageFromDisk(website string) ([]byte, error) {
	hash := sha256.Sum256([]byte(website))
	filename := hex.EncodeToString(hash[:])
	file, err := os.Open(dataDir + "/" + filename)
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(file)
}

// Dump any type of Go object into a gob file of the given name.
func dump(obj interface{}, filename string) {
	f, err := os.Create(dataDir + "/" + filename)
	if err != nil {
		log.Panicf("Error: Failed to create file %v.\n", filename)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	enc := gob.NewEncoder(w)
	if err := enc.Encode(obj); err != nil {
		log.Panicf("Error: Failed to dump object %v to disk.\n", filename)
	}
}

// Load any type of Go object from a gob file of the given name.
func load(obj interface{}, filename string) {
	f, err := os.Open(dataDir + "/" + filename)
	if err != nil {
		log.Panicf("Error: Failed to open file %v.\n", filename)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	dec := gob.NewDecoder(r)
	if err := dec.Decode(obj); err != nil {
		log.Panicf("Error: Failed to load object %v from disk.\n", filename)
	}
}
