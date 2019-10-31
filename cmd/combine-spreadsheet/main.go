package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/extrame/xls"
)

func main() {
	var output [][]string = make([][]string, 0)

	var filename string

	flag.StringVar(&filename, "filename", "", "Name of XLS file")
	flag.Parse()

	if filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	spreadsheet, err := xls.Open(filename, "utf-8")
	if err != nil {
		log.Fatalln(err)
	}

	sheetCount := spreadsheet.NumSheets()
	for sheetID := 0; sheetID < sheetCount; sheetID++ {

		sheet := spreadsheet.GetSheet(sheetID)

		log.Printf("Parsing sheet %d\n", sheetID)

		if sheet == nil {
			log.Fatalf("Sheet %d was nil\n", sheetID)
		}

		for rowID := 0; rowID <= int(sheet.MaxRow); rowID++ {
			row := sheet.Row(rowID)
			if row == nil {
				log.Fatalln("Nil row")
			}

			for colID := 0; colID <= row.LastCol(); colID++ {
				value := row.Col(colID)

				if sheetID == 0 && rowID == 0 {
					output = append(output, make([]string, 0))

					// Add an additional column for the sheet name
					if sheetID == 0 && rowID == 0 && colID == row.LastCol() {
						output = append(output, make([]string, 0))
					}

				} else if rowID == 0 {
					// Skip the header row, except the first time
					continue
				}
				output[colID] = append(output[colID], value)

				// fmt.Println(sheetID, rowID, colID, value)
			}
			if sheetID == 0 && rowID == 0 {
				output[row.LastCol()+1] = append(output[row.LastCol()+1], "Sheet")
			}
			output[row.LastCol()+1] = append(output[row.LastCol()+1], sheet.Name)

		}

	}

	log.Println(len(output), "Columns")
	for i := 0; i < len(output[0]); i++ {

		row := make([]string, 0)
		for _, v := range output {
			row = append(row, v[i])
		}
		fmt.Printf("%s\n", strings.Join(row, "\t"))
	}
}
