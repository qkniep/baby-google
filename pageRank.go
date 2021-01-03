package main

import (
	"fmt"
	"math"
	"sort"
)

const EPS = 0.000001

type webRank struct {
	url  string
	rank float32
}

func PageRank(links map[string][]string) {
	var filteredLinks = filterLinks(links)
	var indexMap = buildIndexMap(filteredLinks)
	var A = buildAdjMatrix(indexMap, filteredLinks)
	var S = make([]float32, len(filteredLinks))
	for i := 0; i < len(S); i++ {
		S[i] = 0.15 / float32(len(S))
	}

	var delta float32 = 999.0
	var R, oldR = make([]float32, len(S)), make([]float32, len(S))
	copy(R, S)
	iterations := 0
	//for delta > EPS {
	for iterations < 10 {
		copy(oldR, R)
		mvMult(A, R, oldR)
		d := vecSum(oldR) - vecSum(R)
		addScaledVec(R, d, S)
		delta = vecDist(oldR, R)
		fmt.Printf("DELTA: %v\n", delta)
		iterations++
	}

	fmt.Printf("after %v iterations: %v\n", iterations, R)

	// print websites by rank
	var pageRanks []webRank
	for url, id := range indexMap {
		pageRanks = append(pageRanks, webRank{url: url, rank: R[id]})
	}
	sort.Slice(pageRanks, func(i, j int) bool { return pageRanks[i].rank > pageRanks[j].rank })
	fmt.Printf("%v\n", pageRanks)
}

func filterLinks(links map[string][]string) map[string][]string {
	// filter dangling links
	filteredLinks := make(map[string][]string, 0)
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
	indexMap := make(map[string]int, 0)
	pageID := 0
	for website := range links {
		indexMap[website] = pageID
		pageID++
	}

	return indexMap
}

func buildAdjMatrix(indexMap map[string]int, links map[string][]string) [][]float32 {
	matrix := make([][]float32, len(links))
	for i := range matrix {
		matrix[i] = make([]float32, len(links))
	}

	// add link weights
	for website, outgoingLinks := range links {
		websiteID := indexMap[website]
		for _, outgoing := range outgoingLinks {
			outgoingID := indexMap[outgoing]
			matrix[websiteID][outgoingID] += 1.0 / float32(len(outgoingLinks))
		}
	}

	return matrix
}

// Matrix-Vector Multiplication
// Assumes newV and oldV to be equal at the start.
// Overwrites the values in newV.
func mvMult(M [][]float32, newV []float32, oldV []float32) {
	for i := range oldV {
		var sum float32 = 0.0
		for j, matVal := range M[i] {
			sum += oldV[j] * matVal
		}
		newV[i] = sum
	}
}

func vecDist(v1 []float32, v2 []float32) (sum float32) {
	for i, val := range v1 {
		sum += float32(math.Abs(float64(val - v2[i])))
	}
	return
}

func vecSum(v []float32) (sum float32) {
	for _, val := range v {
		sum += float32(math.Abs(float64(val)))
	}
	return
}

// Calculates v1 + scalar * v2 and stores it into v1.
func addScaledVec(v1 []float32, scalar float32, v2 []float32) {
	for i := range v1 {
		v1[i] += scalar * v2[i]
	}
}
