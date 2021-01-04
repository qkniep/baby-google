package main

import (
	"fmt"
	"net/url"
	"strings"
)

// Resolve converts a relative URL into an absolute URL, using website as the starting point.
// Returns the absolute URL and a boolean which indicates success.
// TODO: Perform path normailzation (https://www.wikiwand.com/en/URI_normalization).
func Resolve(website string, relative string) (string, bool) {
	u,err := url.Parse(strings.TrimSpace(relative))
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	base,err := url.Parse(website)
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	return base.ResolveReference(u).String(), true
}
