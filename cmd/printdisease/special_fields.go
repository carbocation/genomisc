package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	// These identify which FieldIDs should be special-cased: either to the
	// HESIN table or the table with fields that have a specially-known date.
	// All other fields can be queried against the main phenotype table, with an
	// assumed date of whatever makes the most sense for your purposes
	// (generally, I use the enrollment date, but you could use birthdate, etc).

	MaterializedHesin = map[int]struct{}{
		41210: {},
		41202: {},
		41204: {},
		40001: {},
		40002: {},
		41200: {},
		41203: {},
		41205: {},
		40006: {}, // Cancer Registry ICD10
		40013: {}, // Cancer Registry ICD9
	}

	MaterializedSpecial = map[int]struct{}{
		42013: {},
		42011: {},
		42009: {},
		42007: {},
		42001: {},
		20004: {},
		20002: {},
		20001: {},
		40021: {},
		40020: {}, // Death
		42003: {}, // STEMI
		42005: {}, // NSTEMI
		42019: {}, // All-cause dementia
		42021: {}, // Alzheimer dementia
		42023: {}, // Vascular dementia
		42025: {}, // Frontotemporal dementia
		42027: {}, // ESRD
		6150:  {}, // Self-reported MI, Angina, CVA, HTN
		6152:  {}, // Self-reported DVT, PE, Asthma, Hayfever, Emphysema
		2443:  {}, // Self-reported diabetes
		4041:  {}, // Self-reported gestational diabetes

		// These are specially constructed but don't have a known date and just
		// get set to the enroll date
		3079: {}, // Pacemaker ()
	}
)

var (
	// These can be helpful to make sure that the user is including all fields
	// that use the same family of codes

	ICD9 = map[int]struct{}{
		41203: {},
		41205: {},
		// 40013 Cancer Registry ICD9 is intentionally excluded, since asking
		// the user to include the cancer registry when they are not including
		// cancer codes would be counterproductive. TODO: recognize specific ICD
		// codes and make the prompt if the user includes cancer-specific codes
		// but not the cancer registry.
	}

	ICD10 = map[int]struct{}{
		41202: {},
		41204: {},
		40001: {},
		40002: {},
		// 40006 Cancer Registry ICD10 is intentionally excluded, since asking
		// the user to include the cancer registry when they are not including
		// cancer codes would be counterproductive. TODO: recognize specific ICD
		// codes and make the prompt if the user includes cancer-specific codes
		// but not the cancer registry.
	}

	OPCS = map[int]struct{}{
		41200: {},
		41210: {},
	}
)

func IsHesin(fieldID int) bool {
	_, exists := MaterializedHesin[fieldID]

	return exists
}

func IsSpecial(fieldID int) bool {
	_, exists := MaterializedSpecial[fieldID]

	return exists
}

func describeDateFields(verbose bool) {
	fmt.Fprintf(os.Stderr, "\nNote that the tool tries to set accurate dates only for the following FieldIDs:\n")
	fmt.Fprintf(os.Stderr, "\tICD-like FieldIDs:\n")

	out := make([]string, 0, len(MaterializedHesin))
	for icd := range MaterializedHesin {
		out = append(out, strconv.Itoa(icd))
	}
	sort.StringSlice(out).Sort()
	fmt.Fprintf(os.Stderr, "\t\t%s\n", strings.Join(out, ","))

	fmt.Fprintf(os.Stderr, "\tOther FieldIDs:\n")

	if verbose {
		out = make([]string, 0, len(MaterializedSpecial))
		for icd := range MaterializedSpecial {
			out = append(out, strconv.Itoa(icd))
		}
		sort.StringSlice(out).Sort()
		fmt.Fprintf(os.Stderr, "\t\t%s\n", strings.Join(out, ","))
	} else {
		fmt.Fprintf(os.Stderr, "\t\tAnd %d additional fields (use -verbose to see all).\n", len(MaterializedSpecial))
	}
}
