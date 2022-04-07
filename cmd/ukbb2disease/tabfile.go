package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type TabEntry struct {
	FieldID int
	Values  []string
	Exclude bool
}

// The UK Biobank keys up the ICD and OPCS codes without any decimals, so e.g.,
// K41.2 becomes K412. You can look up the true value using the coding table,
// but for simplicity we just strip the dots.
func (t TabEntry) FormattedValues(biobankSource string) []string {
	out := make([]string, 0, len(t.Values))

	// In the UK Biobank HESIN data, fields are special-cased to exclude "." (So
	// K41.2 becomes K412)
	if biobankSource == BiobankSourceUKBiobank && IsHesin(t.FieldID) {
		for _, v := range t.Values {
			out = append(out, strings.Replace(v, ".", "", -1))
		}
	} else {
		out = append(out, t.Values...)
	}

	// Every field will get leading and trailing spaces trimmed
	for i, v := range out {
		out[i] = strings.TrimSpace(v)
	}

	return out
}

// A tabfile can be represented as a simple list of inclusion and exclusion
// criteria
type TabFile struct {
	Include struct {
		Hesin    []TabEntry
		Special  []TabEntry
		Standard []TabEntry
	}

	Exclude struct {
		Hesin    []TabEntry
		Special  []TabEntry
		Standard []TabEntry
	}
}

// Safely initialize all of the nullable properties of the tabfile
// representation so you don't get nil pointer errors.
func NewTabFile() *TabFile {
	t := &TabFile{}

	t.Include.Hesin = make([]TabEntry, 0)
	t.Include.Special = make([]TabEntry, 0)
	t.Include.Standard = make([]TabEntry, 0)

	t.Exclude.Hesin = make([]TabEntry, 0)
	t.Exclude.Special = make([]TabEntry, 0)
	t.Exclude.Standard = make([]TabEntry, 0)

	return t
}

// Read over the data in this tab file and retrieve the FieldIDs for any
// 'standard' data type, which is neither HESIN data nor the special fields that
// are paired with dates.
func (t *TabFile) AllStandardFields() []int {
	all := make([]int, 0)

	seen := make(map[int]struct{})

	for _, v := range append(t.Include.Standard, t.Exclude.Standard...) {
		if _, exists := seen[v.FieldID]; exists {
			continue
		}

		all = append(all, v.FieldID)
	}

	return all
}

// Get every inclusion criterion from this tabfile
func (t *TabFile) AllIncluded() []TabEntry {
	all := append(t.Include.Hesin, t.Include.Special...)
	return append(all, t.Include.Standard...)
}

// Get every exclusion criterion from this tabfile
func (t *TabFile) AllExcluded() []TabEntry {
	all := append(t.Exclude.Hesin, t.Exclude.Special...)
	return append(all, t.Exclude.Standard...)
}

// ICD9, ICD10, and OPCS codes in the UK Biobank are distributed across many
// FieldIDs. For example, there is a FieldID for "primary ICD10 code" as well as
// a separate FieldID for "secondary ICD10 code". For our purposes, we almost
// always want to consider them equally. Therefore, if a tabfile includes one of
// those fields but not the other, that's usually in error. (It's not *always*
// in error or else we could just override it: there may be special
// circumstances where an analyist only wants to review primary diagnoses.) So,
// this method looks over all of the fields in a tabfile and lets you know
// whether it thinks there are additional FieldIDs you should be adding.
func (t *TabFile) CheckSensibility() ([]string, error) {
	batchnames := []string{"Inclusion", "Exclusion"}
	batches := [][]TabEntry{t.AllIncluded(), t.AllExcluded()}

	missingFields := make(map[int]struct{})
	errs := make([]string, 0)

	// Iterate over all the fields you're including and excluding separately.
	// Rationale: if you successfully listed all the primary and secondary ICD10
	// fields for inclusion, but only listed one of them for exclusion and
	// forgot to list the other, that's probably unintentional and should be
	// flagged.
	for batchID, batch := range batches {
		icd9 := make(map[int]struct{})
		icd10 := make(map[int]struct{})
		opcs := make(map[int]struct{})

		for _, v := range batch {
			// Capitalized ICD9 (etc) here refer to the manual global maps that
			// we create in special_fields.go which list all possible FieldIDs
			// for ICD9 codes. Lower case icd9 represents the local map that
			// will track which of those FieldIDs actually appeared in the
			// tabfile. Basically, this is a lookup to see whether the field we
			// are currently looking at is an ICD9 field, etc.
			if _, exists := ICD9[v.FieldID]; exists {
				icd9[v.FieldID] = struct{}{}
			}
			if _, exists := ICD10[v.FieldID]; exists {
				icd10[v.FieldID] = struct{}{}
			}
			if _, exists := OPCS[v.FieldID]; exists {
				opcs[v.FieldID] = struct{}{}
			}
		}

		// So, you included at least one ICD9 field, but you failed to include
		// all of them. Probably an error.
		if len(icd9) > 0 && len(icd9) < len(ICD9) {
			for known := range ICD9 {
				if _, exists := icd9[known]; !exists {
					missingFields[known] = struct{}{}
				}
			}
			errs = append(errs, fmt.Sprintf("In the %s subset, you included ICD9 fields %v. Consider including all IC9 fields: %v", batchnames[batchID], icd9, ICD9))
		}

		if len(icd10) > 0 && len(icd10) < len(ICD10) {
			for known := range ICD10 {
				if _, exists := icd10[known]; !exists {
					missingFields[known] = struct{}{}
				}
			}
			errs = append(errs, fmt.Sprintf("In the %s subset, you included ICD10 fields %v. Consider including all IC10 fields: %v", batchnames[batchID], icd10, ICD10))
		}

		if len(opcs) > 0 && len(opcs) < len(OPCS) {
			for known := range OPCS {
				if _, exists := opcs[known]; !exists {
					missingFields[known] = struct{}{}
				}
			}
			errs = append(errs, fmt.Sprintf("In the %s subset, you included OPCS fields %v. Consider including all OPCS fields: %v", batchnames[batchID], opcs, OPCS))
		}
	}

	// If we think you forgot one or more FieldIDs, we'll let you know about all
	// of them.
	if len(errs) > 0 {
		missingSlice := make([]string, 0, len(missingFields))
		for missing := range missingFields {
			missingSlice = append(missingSlice, strconv.Itoa(missing))
		}

		return missingSlice, errors.New(strings.Join(errs, " | "))
	}

	return nil, nil
}

// CheckUndatedFields notifies the user when we're unaware of a complementary
// date field for a disease field. In this circumstance, any matching
// participant will be set to prevalent (on the day of enrollment) and therefore
// can never have incident disease. This is desirable when there is truly no
// informative field for timing of an event. But, most often, this indicates
// that a disease<->onset field pair is just not yet recognized. (E.g., the
// cancer registry as of mid-2019). Therefore, this is something that may merit
// a warning.
func (t *TabFile) CheckUndatedFields() ([]int, error) {
	stdFields := t.AllStandardFields()

	if len(stdFields) < 1 {
		return nil, nil
	}

	return stdFields, fmt.Errorf("The following fields are not recognized as having a date: %+v. Currently, if matched, they will cause the person to have *prevalent* disease only and will prevent them from having incident disease. If this is in error, please contact the maintainer of this software", stdFields)
}

func consolidateDuplicates(list []TabEntry) ([]TabEntry, error) {
	outputMap := make(map[int]TabEntry)

	// Consolidate so that each ID is seen in only one TabEntry, which may still
	// have duplicate values
	for _, row := range list {
		field, exists := outputMap[row.FieldID]
		if !exists {
			field = row
			outputMap[row.FieldID] = field
			continue
		}
		mixedList := field.Values
		mixedList = append(mixedList, row.Values...)
		field.Values = mixedList
		outputMap[row.FieldID] = field
	}

	// Now consolidate values within each ID, so that each ID will only see each
	// value at most once
	for i, field := range outputMap {
		seenValue := make(map[string]struct{})
		replacementSlice := make([]string, 0)
		for _, value := range field.Values {
			if _, exists := seenValue[value]; exists {
				continue
			}

			seenValue[value] = struct{}{}
			replacementSlice = append(replacementSlice, value)
		}

		sort.Strings(replacementSlice)
		field.Values = replacementSlice
		outputMap[i] = field
	}

	output := make([]TabEntry, 0, len(outputMap))
	for _, v := range outputMap {
		output = append(output, v)
	}

	return output, nil
}

func (t *TabFile) consolidateDuplicates() error {
	hesin, err := consolidateDuplicates(t.Include.Hesin)
	if err != nil {
		return err
	}
	t.Include.Hesin = hesin

	special, err := consolidateDuplicates(t.Include.Special)
	if err != nil {
		return err
	}
	t.Include.Special = special

	standard, err := consolidateDuplicates(t.Include.Standard)
	if err != nil {
		return err
	}
	t.Include.Standard = standard

	// Exclusions

	hesin, err = consolidateDuplicates(t.Exclude.Hesin)
	if err != nil {
		return err
	}
	t.Exclude.Hesin = hesin

	special, err = consolidateDuplicates(t.Exclude.Special)
	if err != nil {
		return err
	}
	t.Exclude.Special = special

	standard, err = consolidateDuplicates(t.Exclude.Standard)
	if err != nil {
		return err
	}
	t.Exclude.Standard = standard

	return nil
}

// ParseTabFile consumes a tabfile and returns our machine representation of
// that file
func ParseTabFile(tabPath string) (*TabFile, error) {
	f, err := os.Open(tabPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileCSV := csv.NewReader(f)
	fileCSV.Comma = '\t'
	fileCSV.Comment = '#'

	recs, err := fileCSV.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	output := NewTabFile()

	for i, row := range recs {
		if l := len(row); l != 3 {
			return nil, fmt.Errorf("Tabfile %s row %d had %d columns, expected 3", tabPath, i, l)
		}

		if i == 0 {
			// header
			continue
		}
		entry := TabEntry{
			Values:  strings.Split(row[1], ","),
			Exclude: row[2] == "1",
		}
		entry.FieldID, err = strconv.Atoi(row[0])
		if err != nil {
			return nil, err
		}

		// Assign to the right field type
		switch entry.Exclude {
		case true:
			if IsHesin(entry.FieldID) {
				output.Exclude.Hesin = append(output.Exclude.Hesin, entry)
			} else if IsSpecial(entry.FieldID) {
				output.Exclude.Special = append(output.Exclude.Special, entry)
			} else {
				output.Exclude.Standard = append(output.Exclude.Standard, entry)
			}
		default:
			if IsHesin(entry.FieldID) {
				output.Include.Hesin = append(output.Include.Hesin, entry)
			} else if IsSpecial(entry.FieldID) {
				output.Include.Special = append(output.Include.Special, entry)
			} else {
				output.Include.Standard = append(output.Include.Standard, entry)
			}
		}
	}

	if err := output.consolidateDuplicates(); err != nil {
		return nil, err
	}

	return output, nil
}
