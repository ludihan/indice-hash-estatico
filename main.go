package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"math"
	"strings"
)

func hash(word string) uint64 {
	const fnvOffsetBasis uint64 = 14695981039346656037
	const fnvPrime uint64 = 1099511628211

	var hash uint64 = fnvOffsetBasis

	for i := range len(word) {
		hash *= fnvPrime
		hash ^= uint64(word[i])
	}

	return hash
}

const pageSize = 5

//go:embed words.txt
var rawData string

type Database []string

type Bucket map[string]uint

type HashIndex []Bucket

func (db Database) getPage(index uint) ([]string, bool) {
	clamp := func(low, high uint) uint {
		if low < high {
			return low
		} else {
			return high
		}
	}
	if index*pageSize < uint(len(db)) {
		return db[index*pageSize : clamp((index*pageSize)+pageSize, uint(len(db)))], true
	}

	return []string{}, false
}

func (db Database) pageCount() int {
	return int(math.Ceil(float64(len(db)) / float64(pageSize)))
}

func main() {
	// arquivo de db
	file := bufio.NewScanner(strings.NewReader(rawData))
	var db Database

	for file.Scan() {
		text := file.Text()
		if len(strings.TrimSpace(text)) != 0 {
			db = append(db, file.Text())
		}
	}

}
