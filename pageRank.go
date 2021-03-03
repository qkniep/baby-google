package main

import (
	"fmt"
	"github.com/james-bowman/sparse"
	"gonum.org/v1/gonum/mat"
	"math"
	"sort"
	"time"
)

const convergenceThreshold = 1e-7

type websiteRank struct {
	url  string
	rank float64
}

// PageRank is calculated for all non-dangling sites based on the links map.
// Assumes the adjacency matrix of the link graph is sparse;
// uses this fact to speed up matrix-vector multiplication from O(n^2) to O(n).
func PageRank(links map[string][]string) {
	var filteredLinks = filterLinks(links)
	var indexMap = buildIndexMap(filteredLinks)
	var A = buildAdjMatrix(indexMap, filteredLinks)
	var start = time.Now()

	var initialRanks = make([]float64, len(filteredLinks))
	for i := 0; i < len(initialRanks); i++ {
		initialRanks[i] = 1.0 / float64(len(initialRanks))
	}

	var oneData = make([]float64, len(initialRanks))
	for i := 0; i < len(oneData); i++ {
		oneData[i] = 0.15 / float64(len(oneData))
	}

	var delta float64 = math.Inf(1)
	var R = mat.NewVecDense(len(initialRanks), initialRanks)
	var ONE = mat.NewVecDense(len(oneData), oneData)
	iterations := 0
	for delta > convergenceThreshold {
		oldR := mat.VecDenseCopyOf(R)
		R = sparse.MulMatVec(false, 1.0, A, R, nil)
		R.AddScaledVec(ONE, 0.85, R)
		oldR.SubVec(oldR, R)
		delta = mat.Norm(oldR, 1)
		iterations++
	}

	// print websites by rank
	var pageRanks []websiteRank
	for url, id := range indexMap {
		pageRanks = append(pageRanks, websiteRank{url: url, rank: initialRanks[id]})
	}
	sort.Slice(pageRanks, func(i, j int) bool { return pageRanks[i].rank > pageRanks[j].rank })
	for _, wr := range pageRanks[:3] {
		fmt.Printf("%v - %v\n", wr.url, wr.rank)
	}

	secs := time.Since(start).Seconds()
	fmt.Printf("Calculated the PageRank for %v/%v pages in %.2f seconds.\n", len(pageRanks), len(links), secs)
	fmt.Printf("Number of iterations until convergence: %v\n", iterations)
}

// Filters out dangling links, i.e. links which point to no pages in our crawled set.
func filterLinks(links map[string][]string) map[string][]string {
	var filteredLinks = make(map[string][]string, 0)
	for link, outgoingLinks := range links {
		for _, outgoing := range outgoingLinks {
			_, found := links[outgoing]
			if found {
				filteredLinks[link] = append(filteredLinks[link], outgoing)
			}
		}
	}
	return filteredLinks
}

func buildIndexMap(links map[string][]string) map[string]int {
	var indexMap = make(map[string]int, 0)
	var pageID = 0
	for website := range links {
		indexMap[website] = pageID
		pageID++
	}
	return indexMap
}

func buildAdjMatrix(indexMap map[string]int, links map[string][]string) *sparse.CSR {
	var matrix = sparse.NewDOK(len(links), len(links))
	for website, outgoingLinks := range links {
		websiteID := indexMap[website]
		for _, outgoing := range outgoingLinks {
			outgoingID := indexMap[outgoing]
			newValue := matrix.At(outgoingID, websiteID) + 1.0/float64(len(outgoingLinks))
			matrix.Set(outgoingID, websiteID, newValue)
		}
	}
	return matrix.ToCSR()
}
