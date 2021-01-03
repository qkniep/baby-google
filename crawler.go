package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
)

var DO_NOT_DOWNLOAD = [...]string{"js", "png", "jpg", "jpeg", "webp", "zip"}
const STARTPAGE = "https://wikipedia.org"

func main() {
	var current string
	var toVisit = []string{STARTPAGE}

	for len(toVisit) > 0 {
		current, toVisit = toVisit[0], toVisit[1:]

		// download current website
		fmt.Println(current)
		res, err := http.Get(current)
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d\n", len(body))

		// search anchor tags
		htmlTokens := html.NewTokenizer(res.Body)
	loop:
		for {
			tt := htmlTokens.Next()
			fmt.Printf("%T", tt)
			switch tt {
			case html.ErrorToken:
				fmt.Println("End")
				break loop
			case html.TextToken:
				fmt.Println(tt)
			case html.StartTagToken:
				t := htmlTokens.Token()
				isAnchor := t.Data == "a"
				if isAnchor {
					fmt.Println("We found an anchor!")
					// resolve anchor link
					// add URL to toVisit
				}
			}
		}
	}
}
