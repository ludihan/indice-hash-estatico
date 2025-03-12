package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"math"
	"os"
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

type Bucket struct {
	values   map[string]uint
	overflow *Bucket
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
	bucketSize := 2

	// (NB)
	bucketCount := (len(db.data) / bucketSize)

	// (NR)
	wordCount := len(db.data)
	fmt.Println(bucketCount, wordCount)

	hashIndex := make([]Bucket, bucketCount)
	for i := range hashIndex {
		hashIndex[i].values = make(map[string]uint)
	}

	for i, v := range db.data {
		hashed := hash(v)
		bucket := &hashIndex[hashed%uint64(len(hashIndex))]

		for bucket.overflow != nil {
			bucket = bucket.overflow
		}

		if len(bucket.values) >= int(bucketSize) {
			newBucket := &Bucket{values: make(map[string]uint)}
			bucket.overflow = newBucket
			bucket = newBucket
		}

		bucket.values[v] = uint(i) / db.pageSize
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')

        text = strings.TrimSpace(text)
		found := false
		page := uint(0)
		bucket := &hashIndex[hash(text)%uint64(len(hashIndex))]
        for bucket != nil {
            fmt.Println(bucket)
			for k, v := range bucket.values {
                if k == text {
                    page = v
                    found = true
                }
			}
			bucket = bucket.overflow
		}
        if found {
            fmt.Println("Found in page", page)
            foundPage, _ := db.getPage(page)
            fmt.Println(foundPage, "\n")
        } else {
            fmt.Println("not found\n")
        }
	}
}
