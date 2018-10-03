package bulkmanifest

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Many of the manifest.csv files are invalid. They use commas as delimiters and
// do not use quotes, and yet they have fields (like date) that use commas... So
// this requires some processing.
func processInvalidManifest(manifest io.Reader) ([][]string, error) {
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
					break
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
				output[i][j-1] += ", " + strings.TrimSpace(col)
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

func processManifest(manifest io.Reader) ([][]string, error) {
	r := csv.NewReader(manifest)
	return r.ReadAll()
}

// Iterating over the directory will be delegated to the caller.
// Here we will process one zip file.
func ProcessCardiacMRIZip(zipPath string, db *sqlx.DB) error {
	metadata, err := zipPathToMetadata(zipPath)
	if err != nil {
		return err
	}

	zipName := path.Base(zipPath)

	rc, err := zip.OpenReader(zipPath)
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

		// Consume the full manifest into a buffer so we can re-read it if it is
		// invalid
		manifestBuffer := &bytes.Buffer{}
		if _, err := io.Copy(manifestBuffer, zippedFile); err != nil {
			return err
		}

		manifestReader := bytes.NewReader(manifestBuffer.Bytes())

		// Try to process it as a CSV
		fields, err := processManifest(manifestReader)
		if err != nil {

			// If that fails, try again with our manual algorithm
			manifestReader.Seek(0, 0)

			fields, err = processInvalidManifest(manifestReader)
			if err != nil {
				return err
			}
		}

		dicoms := make([]DicomOutput, len(fields), len(fields))
		for loc := range fields {
			dicoms[loc], err = stringSliceToDicomStruct(fields[loc])
			if err != nil {
				return err
			}
			dicoms[loc].SampleID = metadata.SampleID
			dicoms[loc].ZipFile = zipName
			dicoms[loc].FieldID = metadata.FieldID
			dicoms[loc].Instance = metadata.Instance
			dicoms[loc].Index = metadata.Index
		}

		for _, dcm := range dicoms {
			fmt.Printf("%+v\n", dcm)
		}

		// fmt.Printf("%s | %s, %+v\n", zipName, v.Name, metadata)
		// for _, fld := range fields {
		// 	fmt.Println(fld)
		// }

		zippedFile.Close()
	}

	return nil
}

//   Take the manifest and modify its contents to specify the sample,
//     then update the SQLite
//   Annotate with lookup of the axis type *and field ID* if not yet done
//   Take the dicoms and consider renaming them vs keeping the same name.
