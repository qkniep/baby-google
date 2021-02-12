package main

import (
	"golang.org/x/net/html"
	"io"
)

type ScrapeResult struct {
	Website string
	Links   []string
	Size    int
	Err     bool
}

func ScrapeLinksFromHTML(r io.Reader, baseURL string) ScrapeResult {
	var links []string
	var htmlTokens = html.NewTokenizer(r)

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
						absURL, success := Resolve(baseURL, attr.Val)
						if success {
							links = append(links, absURL)
						}
					}
				}
			}
		}
	}

	return ScrapeResult{baseURL, links, 0, false}
}
