package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var DO_NOT_DOWNLOAD = [...]string{"js", "png", "jpg", "jpeg", "webp", "zip"}
const STARTPAGE = "https://wikipedia.org"

func main() {
	var current string
	var toVisit = []string{STARTPAGE}
	var visited = make(map[string]bool, 0)
	var visitLast = false
	var bytesDownloaded int64

	// index structures
	var links = make(map[string][]string, 0)
	//var backLinks = make(map[string][]string, 0)

	//for len(toVisit) > 0 {
	for len(visited) < 1000 {
		if visitLast {
			current, toVisit = toVisit[len(toVisit)-1], toVisit[:len(toVisit)-1]
		} else {
			current, toVisit = toVisit[0], toVisit[1:]
		}
		visitLast = !visitLast

		if visited[current] {
			continue
		}
		visited[current] = true

		// download current website
		fmt.Printf("Current website: %v\n", current)
		fmt.Printf("Number of sites to vist: %v\n", len(toVisit))
		res, err := http.Get(current)
		if err != nil {
			log.Println(err)
			continue
		}
		defer res.Body.Close()

		// count bytes
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			continue
		}
		bytesDownloaded += int64(len(body))
		fmt.Printf("Bytes downloaded: %v\n", bytesDownloaded)
		fmt.Printf("Visited pages: %v\n", len(visited))
		bodyReader := strings.NewReader(string(body))

		// search anchor tags
		htmlTokens := html.NewTokenizer(bodyReader)
	loop:
		for {
			tt := htmlTokens.Next()
			//fmt.Printf("%T", tt)
			switch tt {
			case html.ErrorToken:
				//fmt.Println("End")
				break loop
			case html.TextToken:
				//fmt.Println(tt)
			case html.StartTagToken:
				t := htmlTokens.Token()
				isAnchor := t.Data == "a"
				if isAnchor {
					for _, attr := range t.Attr {
						if attr.Key == "href" {
							absURL, success := Resolve(current, attr.Val)
							if success {
								links[current] = append(links[current], absURL)
								//backLinks[absURL] = append(backLinks[absURL], current)
								if!visited[absURL] {
									toVisit = append(toVisit, absURL)
								}
							}
							//fmt.Printf("We found an anchor: %s\n", absURL)
						}
					}
					// add URL to toVisit
				}
			}
		}
	}

	PageRank(links)
}
