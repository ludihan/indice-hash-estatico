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
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// https://wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1_hash
func hash(word string) uint64 {
	FNVOffsetBasis := uint64(14695981039346656037)
	FNVPrime := uint64(1099511628211)
	hash := FNVOffsetBasis
	for i := 0; i < len(word); i++ {
		hash = hash * FNVPrime
		hash = hash ^ uint64(word[i])
	}

	return hash
}

func nextPrime(n int) int {
	for !isPrime(n) {
		n++
	}
	return n
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

//go:embed words.txt
var rawData string

type Database struct {
	data     []string
	pageSize uint
}

func (db *Database) getPage(index uint) ([]string, bool) {
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

func (db *Database) pageCount() int {
	return int(math.Ceil(float64(len(db.data)) / float64(db.pageSize)))
}

// pagina, se foi achado, acesso de disco e tempo levado
func (db *Database) search(word string) (uint, bool, uint, time.Duration) {
	start := time.Now()
	pages := uint(0)
	for i, v := range db.data {
		if word == v {
			pages = uint(i) / db.pageSize
			return pages, true, pages + 1, time.Now().Sub(start)
		}
	}
	return pages, false, pages, time.Now().Sub(start)
}

type Bucket struct {
	values   map[string]uint
	overflow *Bucket
}

type HashIndex []Bucket

// pagina, se foi achado, acesso de disco e tempo levado
func (hi HashIndex) search(word string) (uint, bool, uint, time.Duration) {
	start := time.Now()
	totalAccess := uint(0)
	hashed := hash(word)
	bucket := &hi[hashed%uint64(len(hi))]
	totalAccess++
	for bucket != nil {
		for k, v := range bucket.values {
			if k == word {
				return v, true, totalAccess, time.Now().Sub(start)
			}
		}
		bucket = bucket.overflow
	}
	return 0, false, totalAccess, time.Now().Sub(start)
}

type App struct {
	db          Database
	bucketSize  uint
	bucketCount uint
	wordCount   uint
	collisions  uint
	overflows   uint
	hashIndex   HashIndex
}

func rehash(pageSize uint) App {
	// db file
	file := bufio.NewScanner(strings.NewReader(rawData))
	db := Database{
		data: []string{},
		// user input
		pageSize: pageSize,
	}

	for file.Scan() {
		text := file.Text()
		if len(strings.TrimSpace(text)) != 0 {
			db.data = append(db.data, text)
		}
	}

	// (FR)
	bucketSize := 5

	// (NB)
	bucketCount := nextPrime(int(math.Ceil(float64(len(db.data))/float64(bucketSize)) + float64(len(db.data))*0.2))

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

	return App{
		db:          db,
		bucketSize:  uint(bucketSize),
		bucketCount: uint(bucketCount),
		wordCount:   uint(wordCount),
		collisions:  uint(collisions),
		overflows:   uint(overflows),
		hashIndex:   hashIndex,
	}
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
			if err := initialWindow(window); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}()
		app.Main()
	}
}

func cli() {
	reader := bufio.NewReader(os.Stdin)
	var pageSize uint
	for {
		fmt.Print("page size:   ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		size, err := strconv.ParseUint(text, 10, 0)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if size == 0 {
			fmt.Println("cant use 0")
			continue
		}
		pageSize = uint(size)
		break
	}

	app := rehash(pageSize)
	db := app.db
	hashIndex := app.hashIndex
	for {
		fmt.Print(">>> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if len(text) == 0 {
			continue
		}
		if text[0] == ':' {
			num, err := strconv.Atoi(text[1:])
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				if num >= 0 {
					fmt.Println(db.getPage(uint(num)))
				} else {
					fmt.Println("only positive")
				}
			}
			continue
		}

		page, found, access, time := uint(0), false, uint(0), time.Duration(0)

		if text[0] == '.' {
			text := text[1:]
			page, found, access, time = db.search(text)
			fmt.Println("Table scan:")
		} else {
			page, found, access, time = hashIndex.search(text)
			fmt.Println("Hash index:")
		}

		if found {
			fmt.Println("Found in page", page, "Access:", access, "Time:", time)
			foundPage, _ := db.getPage(page)
			fmt.Println(foundPage)
			fmt.Println("")
		} else {
			fmt.Println("not found")
			fmt.Println("")
		}
	}

}

func initialWindow(window *app.Window) error {
	var ops op.Ops

	theme := material.NewTheme()
	theme.TextSize = unit.Sp(18)

	var pageSizeInput = widget.Editor{
		SingleLine: true,
		Submit:     true,
	}

	var db_app App
	var alreadyRehashed = false

	var pageSize uint64
	var inputError string
	var input string
	var mainWindow = false
	var backButton widget.Clickable
	var wordInput = widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	var wordText string
	var dbInfoOutput = widget.Editor{
		ReadOnly: true,
	}

	var tableScanButton widget.Clickable
	var hashIndexScanButton widget.Clickable

	var firstPage = widget.Editor{
		ReadOnly: true,
	}
	var lastPage = widget.Editor{
		ReadOnly: true,
	}
	var pageFound = widget.Editor{
		ReadOnly: true,
	}
	margins := layout.UniformInset(unit.Dp(10))
	for {
		switch e := window.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			if !mainWindow {
				layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return margins.Layout(gtx,
								func(gtx layout.Context) layout.Dimensions {
									title := material.H3(theme, "Tamanho da página:")
									return title.Layout(gtx)
								},
							)
						},
					),
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return margins.Layout(gtx,
								func(gtx layout.Context) layout.Dimensions {
									btnE, ok := pageSizeInput.Update(gtx)
									if ok {
										if btnE, ok := btnE.(widget.SubmitEvent); ok {
											input = btnE.Text
											btnE.Text = ""
										}
									}
									input := material.Editor(theme, &pageSizeInput, "digite aqui")
									return input.Layout(gtx)
								},
							)
						},
					),
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return margins.Layout(gtx,
								func(gtx layout.Context) layout.Dimensions {
									btn := material.H4(theme, inputError)
									return btn.Layout(gtx)
								},
							)
						},
					),
				)

				if input != "" {
					var err error
					pageSize, err = strconv.ParseUint(strings.TrimSpace(input), 10, 0)
					if err != nil {
						inputError = "seu input não é unsigned int,\n"
						inputError += err.Error()
					} else {
						mainWindow = true
						inputError = ""
						input = ""
					}
				}

				e.Frame(gtx.Ops)
			} else {
				if !alreadyRehashed {
					db_app = rehash(uint(pageSize))
					alreadyRehashed = true
					dbInfoOutput.SetText(
						fmt.Sprintf(
							"Tamanho do bucket: %v\n"+
								"Quantidade de buckets: %v\n"+
								"Tamanho da pagina: %v\n"+
								"Quantidade de registros: %v\n"+
								"Colisões: %v %v %%\n"+
								"Overflows: %v %v %%\n",
							db_app.bucketSize, db_app.bucketCount, db_app.db.pageSize, len(db_app.db.data),
							db_app.collisions, float64(db_app.collisions)/float64(len(db_app.db.data))*100,
							db_app.overflows, float64(db_app.overflows)/float64(len(db_app.db.data))*100,
						),
					)
					output := "Primeira página (0):\n\n"
					firstPageSeries, _ := db_app.db.getPage(0)
					for _, v := range firstPageSeries {
						output += v + "\n"
					}
					firstPage.SetText(output)

					pageNum := uint(db_app.db.pageCount()) - 1
					output = fmt.Sprintf("Última página:(%v)\n\n", pageNum)
					lastPageSeries, _ := db_app.db.getPage(uint(db_app.db.pageCount()) - 1)
					for _, v := range lastPageSeries {
						output += v + "\n"
					}
					lastPage.SetText(output)
				}
				layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,

					// botao
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return margins.Layout(gtx,
								func(gtx layout.Context) layout.Dimensions {
									backBtn := material.Button(theme, &backButton, "voltar")
									if backButton.Clicked(gtx) {
										input = ""
										mainWindow = false
										alreadyRehashed = false
									}
									return backBtn.Layout(gtx)
								},
							)
						},
					),

					// tudo abaixo do botao de voltar
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return margins.Layout(gtx,
								func(gtx layout.Context) layout.Dimensions {

									// duas colunas
									return layout.Flex{
										Axis:    layout.Horizontal,
										Spacing: layout.SpaceAround,
									}.Layout(gtx,

										// coluna de informacao
										layout.Flexed(1,
											func(gtx layout.Context) layout.Dimensions {
												return layout.Flex{
													Axis: layout.Vertical,
												}.Layout(gtx,
													layout.Flexed(1,
														// mostra a pagina que ele achou
														func(gtx layout.Context) layout.Dimensions {
															pageOutputEditor := material.Editor(theme, &dbInfoOutput, "")
															return pageOutputEditor.Layout(gtx)
														},
													),
													layout.Flexed(1,
														// primeira e ultima pagina
														func(gtx layout.Context) layout.Dimensions {
															return layout.Flex{
																Axis: layout.Horizontal,
															}.Layout(gtx,
																layout.Flexed(1,
																	func(gtx layout.Context) layout.Dimensions {
																		first := material.Editor(theme, &firstPage, "")
																		return first.Layout(gtx)
																	},
																),
																layout.Flexed(1,
																	func(gtx layout.Context) layout.Dimensions {
																		last := material.Editor(theme, &lastPage, "")
																		return last.Layout(gtx)
																	},
																),
															)
														},
													),
												)
											},
										),

										//coluna de input e pagina
										layout.Flexed(1,
											func(gtx layout.Context) layout.Dimensions {
												return layout.Flex{
													Axis: layout.Vertical,
												}.Layout(gtx,
													// pagina
													layout.Flexed(1,
														func(gtx layout.Context) layout.Dimensions {
															pageFoundEditor := material.Editor(theme, &pageFound, "")
															return pageFoundEditor.Layout(gtx)
														},
													),

													// input
													layout.Rigid(
														func(gtx layout.Context) layout.Dimensions {
															wordInputEditor := material.Editor(theme, &wordInput, "pesquise uma palavra aqui")
															wordText = strings.TrimSpace(wordInput.Text())
															return wordInputEditor.Layout(gtx)
														},
													),

													// botoes
													layout.Rigid(
														func(gtx layout.Context) layout.Dimensions {
															return layout.Flex{
																Axis: layout.Horizontal,
															}.Layout(gtx,
																layout.Flexed(1,
																	func(gtx layout.Context) layout.Dimensions {
																		hashIndexScanButtonbtn := material.Button(theme, &hashIndexScanButton, "Índice Hash")
																		if hashIndexScanButton.Clicked(gtx) {
																			page, found, access, time := db_app.hashIndex.search(wordText)
																			output := ""
																			if found {
																				output += "Índice hash:\n"
																				output += fmt.Sprintf("Encontrado na página %v, com %v acessos e demorando %v\nRegistros da pagina:\n\n", page, access, time)
																				pages, _ := db_app.db.getPage(page)
																				for _, v := range pages {
																					output += v + "\n"
																				}
																				pageFound.SetText(output)
																			} else {
																				pageFound.SetText("Não encontrado: \"" + wordText + "\"")
																			}
																		}
																		return hashIndexScanButtonbtn.Layout(gtx)
																	},
																),
																layout.Flexed(1,
																	func(gtx layout.Context) layout.Dimensions {
																		tableScanButtonbtn := material.Button(theme, &tableScanButton, "Table Scan")
																		if tableScanButton.Clicked(gtx) {
																			page, found, access, time := db_app.db.search(wordText)
																			output := ""
																			if found {
																				output += "Table scan:\n"
																				output += fmt.Sprintf("Encontrado na página %v, com %v acessos e demorando %v\nRegistros da pagina:\n\n", page, access, time)
																				pages, _ := db_app.db.getPage(page)
																				for _, v := range pages {
																					output += v + "\n"
																				}
																				pageFound.SetText(output)
																			} else {
																				pageFound.SetText("Não encontrado: \"" + wordText + "\"")
																			}
																		}
																		return tableScanButtonbtn.Layout(gtx)
																	},
																),
															)
														},
													),
												)
											},
										),
									)
								},
							)
						},
					),
				)

				e.Frame(gtx.Ops)
			}
		case app.DestroyEvent:
			return e.Err
		}
	}
}
