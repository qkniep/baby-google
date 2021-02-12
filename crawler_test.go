package main

import (
	"github.com/valyala/fasthttp"
	"testing"
)

func TestRobots(t *testing.T) {
	url := "https://www.google.com/search?ei=hc4mYOm-DsahgQanwaLIBw&q=hello+world&oq=hello+world&gs_lcp=Cgdnd3Mtd2l6EAMyAggAMgIILjICCAAyAggAMgIIADICCAAyAggAMgIIADICCAAyAggAOgcIABCwAxANOgkIABCwAxANEAo6BQgAEJECOggILhDHARCjAjoICC4QkQIQkwI6CAguEMcBEK8BOgUILhCRAlDMGljCJmCjJ2gBcAB4AIAB1AGIAbQJkgEFOC4zLjGYAQCgAQGqAQdnd3Mtd2l6yAEKwAEB&sclient=gws-wiz&ved=0ahUKEwiplYKmguXuAhXGUMAKHaegCHkQ4dUDCA0&uact=5"
	client := fasthttp.Client{}
	res := scrapeLinksFromPage(&client, url)
	if res.Size != 0 || len(res.Links) != 0 {
		panic("Downloaded page though robots.txt forbids it")
	} else if res.Err != false {
		panic("Error encountered")
	}
}
