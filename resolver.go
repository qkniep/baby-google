package main

import (
	"fmt"
	"net/url"
	"strings"
)

// TODO
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
