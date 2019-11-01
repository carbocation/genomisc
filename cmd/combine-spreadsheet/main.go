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
	// [Cols][Rows]value
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

		log.Printf("Parsing sheet #%d (\"%s\")\n", sheetID, sheet.Name)

		// if sheet.Name != "EU_HRC_LVEDVI_BSA" && sheet.Name != "EAS_1000GP3_LVEDVI_BSA" {
		// 	continue
		// }

		if sheet == nil {
			log.Fatalf("Sheet %d was nil\n", sheetID)
		}

		for rowID := 0; rowID <= int(sheet.MaxRow); rowID++ {
			row := sheet.Row(rowID)
			if row == nil {
				log.Fatalln("Nil row")
			}

			outputExists := false
			if len(output) > 0 {
				outputExists = true
			}

			for colID := 0; colID < row.LastCol(); colID++ {
				value := row.Col(colID)

				// Make a container for each column
				if !outputExists && rowID == 0 {
					output = append(output, make([]string, 0))

					// Add an additional column for the sheet name
					if !outputExists && rowID == 0 && colID == row.LastCol()-1 {
						output = append(output, make([]string, 0))
					}

				} else if rowID == 0 {
					// Skip the header row, except the first time
					continue
				}
				output[colID] = append(output[colID], value)

				// fmt.Println(sheetID, rowID, colID, value)
			}
			if !outputExists && rowID == 0 {
				output[row.LastCol()] = append(output[row.LastCol()], "Sheet")
				continue
			} else if rowID == 0 {
				continue
			}
			output[row.LastCol()] = append(output[row.LastCol()], sheet.Name)

		}

	}

	if x, y := len(output[0]), len(output[len(output)-1]); x != y {
		log.Fatalf("%d != %d\n", x, y)
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
