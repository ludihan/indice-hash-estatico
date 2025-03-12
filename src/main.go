package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
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
	isCli := flag.Bool("cli", false, "Starts a simple cli instead of a gui.")
	flag.Parse()
	if *isCli {
		cli()
	} else {
		go func() {
			window := new(app.Window)
			window.Option(app.Title("Hash Index"))
			if err := gui(window); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}()
		app.Main()
	}
}

func cli() {
	// db file
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
	bucketSize := 5

	// (NB)
	bucketCount := (len(db.data) / bucketSize) + int(float64(len(db.data))*0.05)

	// (NR)
	wordCount := len(db.data)

	fmt.Println("bucketSize: ", bucketSize)
	fmt.Println("bucketCount:", bucketCount)
	fmt.Println("wordCount:  ", wordCount)

	hashIndex := make([]Bucket, bucketCount)
	for i := range hashIndex {
		hashIndex[i].values = make(map[string]uint)
	}

	collisions := 0
	overflows := 0
	isOverflow := false
	for i, v := range db.data {
		hashed := hash(v)
		bucket := &hashIndex[hashed%uint64(len(hashIndex))]
		for bucket.overflow != nil {
			bucket = bucket.overflow
			isOverflow = true
		}

		if len(bucket.values) >= int(bucketSize) {
			overflows++
			newBucket := &Bucket{values: make(map[string]uint)}
			bucket.overflow = newBucket
			bucket = newBucket
		}

		if isOverflow {
			collisions++
		}
		bucket.values[v] = uint(i) / db.pageSize
	}

	fmt.Println("collisions: ", collisions, float64(collisions)/float64(wordCount))
	fmt.Println("overflows:  ", overflows, float64(overflows)/float64(wordCount))
	fmt.Println("")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">>> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if len(text) == 0 {
			continue
		}
		if text[0] == ':' {
			num, err := strconv.Atoi(text[1:])
			if err == nil {
				if num >= 0 {
					fmt.Println(db.getPage(uint(num)))
				} else {
					fmt.Println("only positive")
				}
				continue
			}
		}

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
			fmt.Println(foundPage)
			fmt.Println("")
		} else {
			fmt.Println("not found")
			fmt.Println("")
		}
	}

}

func gui(window *app.Window) error {
	// db file
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
	bucketSize := 5

	// (NB)
	bucketCount := (len(db.data) / bucketSize) + int(float64(len(db.data))*0.05)

	// (NR)
	wordCount := len(db.data)

	fmt.Println("bucketSize: ", bucketSize)
	fmt.Println("bucketCount:", bucketCount)
	fmt.Println("wordCount:  ", wordCount)

	hashIndex := make([]Bucket, bucketCount)
	for i := range hashIndex {
		hashIndex[i].values = make(map[string]uint)
	}

	collisions := 0
	overflows := 0
	isOverflow := false
	for i, v := range db.data {
		hashed := hash(v)
		bucket := &hashIndex[hashed%uint64(len(hashIndex))]
		for bucket.overflow != nil {
			bucket = bucket.overflow
			isOverflow = true
		}

		if len(bucket.values) >= int(bucketSize) {
			overflows++
			newBucket := &Bucket{values: make(map[string]uint)}
			bucket.overflow = newBucket
			bucket = newBucket
		}

		if isOverflow {
			collisions++
		}
		bucket.values[v] = uint(i) / db.pageSize
	}

	fmt.Println("collisions: ", collisions, float64(collisions)/float64(wordCount))
	fmt.Println("overflows:  ", overflows, float64(overflows)/float64(wordCount))
	fmt.Println("")

	// GUI stuff
	var ops op.Ops

	var startButton widget.Clickable

	theme := material.NewTheme()

	for {
		switch e := window.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						margins := layout.UniformInset(unit.Dp(10))

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(theme, &startButton, "Start")
								return btn.Layout(gtx)
							},
						)

					},
				),
			)
			e.Frame(gtx.Ops)
		case app.DestroyEvent:
			return e.Err
		}
	}

}
