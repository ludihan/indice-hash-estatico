package main

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"strings"
)

const pageSize = 5

//go:embed words.txt
var rawData string

func main() {
	// arquivo de db
	data := []string{}

	todosBuckets := []Bucket{}

	for v := range strings.Lines(rawData) {
		data = append(data, strings.Trim(v, "\n"))
	}

	/*
		for i := 0; i < len(data); i += pageSize {
			fmt.Printf("%#v\n", data[i:pageSize+i])
		}
	*/

	for _, v := range data {
		resultadoHash := sha256.Sum256(v)

	}
}

type Bucket struct {
	keys   []string
	values []int
}
