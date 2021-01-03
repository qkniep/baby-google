package main

import (
    "fmt"
    "net/url"
)

func resolve(website string, relative string) *url.URL {
    u,err := url.Parse(relative)
    base,err := url.Parse(website)

    if err!=nil {
        fmt.Println(err)
    }

    return base.ResolveReference(u)
}
