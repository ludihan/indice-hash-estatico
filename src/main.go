package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"math"
	"strings"
)

// https://wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1_hash
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

//go:embed words.txt
var rawData string

type Database struct {
	data     []string
	pageSize uint
}

func (db Database) getPage(index uint) ([]string, bool) {
	clamp := func(low, high uint) uint {
		if low < high {
			return low
		} else {
			return high
		}
	}

	if index*db.pageSize < uint(len(db.data)) {
		return db.data[index*db.pageSize : clamp((index*db.pageSize)+db.pageSize, uint(len(db.data)))], true
	}

	return []string{}, false
}

func (db Database) pageCount() int {
	return int(math.Ceil(float64(len(db.data)) / float64(db.pageSize)))
}


var bucketSize = 0
type Bucket struct {
	values     map[string]uint
	overflow   *Bucket
}

func main() {
	// arquivo de db
	file := bufio.NewScanner(strings.NewReader(rawData))
	db := Database{
		data: []string{},
		// user input
		pageSize: 5,
	}

	for file.Scan() {
		text := file.Text()
		if len(strings.TrimSpace(text)) != 0 {
			db.data = append(db.data, file.Text())
		}
	}

	// (FR)
	bucketSize = 2

	// (NB)
	bucketCount := (len(db.data) / bucketSize) + 1
	fmt.Println(len(db.data), bucketSize, bucketCount)

	hashIndex := make([]Bucket, bucketCount)
	for i, v := range db.data {
		hashed := hash(v)
		bucket := &hashIndex[hashed%uint64(len(hashIndex))]
		if len(bucket.values) >= int(bucketSize) {
			overflow := bucket.overflow

			for overflow != nil {
				overflow = bucket.overflow
			}

            overflow = &Bucket{make(map[string]uint), nil}
			bucket.values[v] = uint(i) / db.pageSize
		} else {
			bucket.values[v] = uint(i) / db.pageSize
		}
	}

	fmt.Println(hashIndex)
}
