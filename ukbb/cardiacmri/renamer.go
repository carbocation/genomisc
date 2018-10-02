package cardiacmri

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/jmoiron/sqlx"
)

// The manifest.csv files are invalid. They use commas as delimiters and do not
// use quotes, and yet they have fields (like date) that use commas... So this
// requires some processing.
func processManifest(manifest io.ReadCloser) ([][]string, error) {
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

// Iterating over the directory will be delegated to the caller.
// Here we will process one zip file.

func ProcessCardiacMRIZip(path string, db *sqlx.DB) error {
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

		for _, fld := range fields {
			fmt.Println(fld)
		}
		fmt.Printf("%s, %+v\n", v.Name, metadata)
	}

	return nil
}

//   Take the manifest and modify its contents to specify the sample,
//     then update the SQLite
//   Annotate with lookup of the axis type *and field ID* if not yet done
//   Take the dicoms and consider renaming them vs keeping the same name.
