package cardiacmri

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Tools for understanding the miscellaneous naming issues with the zip manifests

// The zipfiles look completely unstandardized
func SurveyZipManifests(path string, db *sqlx.DB) error {
	metadata, err := zipPathToMetadata(path)
	if err != nil {
		return err
	}

	rc, err := zip.OpenReader(path)
	if err != nil {
		return err
	}

	for _, v := range rc.File {
		if !strings.HasPrefix(v.Name, "manifest") {
			continue
		}

		zippedFile, err := v.Open()
		if err != nil {
			return err
		}

		fields, err := processManifest(zippedFile)
		if err != nil {
			return err
		}

		dateField := -1
		for i, v := range fields[0] {
			if v == "date" {
				dateField = i
				break
			}
		}

		// Print the first 2 lines (header + 1 example)
		fmt.Println(path)
		fmt.Printf("%s, %+v cols: %d \n", v.Name, metadata, len(fields[0]))
		fmt.Printf("Datefield: %s => %s\n", fields[0][dateField], fields[1][dateField])
		fmt.Printf("%+v\n", fields[0])
		fmt.Printf("%+v\n", fields[1])
		fmt.Println()

		// for _, fld := range fields {
		// 	fmt.Println(fld)
		// }

	}

	return nil
}

//
func surveyManifest(manifest io.ReadCloser) ([][]string, error) {
	output := [][]string{}

	header := []string{}
	dateField := -1
	i := 0

	r := bufio.NewReader(manifest)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if i == 0 {
			record := strings.Split(strings.TrimSpace(strings.Replace(line, "discription", "description", -1)), ",")

			header = append(header, record...)
			for i, col := range record {
				if col == "date" {
					dateField = i
				}
			}

			if dateField == -1 {
				return nil, fmt.Errorf("No field named 'date' was found in the manifest")
			}

			output = append(output, header)
			i++

			continue
		}

		record := strings.Split(strings.TrimSpace(line), ",")

		output = append(output, make([]string, len(header), len(header)))

		for j, col := range record {
			if j == dateField+1 {
				output[i][j-1] += strings.TrimSpace(col)
			} else if j > dateField+1 {
				output[i][j-1] = strings.TrimSpace(col)
			} else {
				output[i][j] = strings.TrimSpace(col)
			}
		}

		i++
	}

	return output, nil
}
