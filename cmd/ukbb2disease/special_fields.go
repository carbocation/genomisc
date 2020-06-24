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
		41210: struct{}{},
		41202: struct{}{},
		41204: struct{}{},
		40001: struct{}{},
		40002: struct{}{},
		41200: struct{}{},
		41203: struct{}{},
		41205: struct{}{},
		40006: struct{}{}, // Cancer Registry ICD10
		40013: struct{}{}, // Cancer Registry ICD9
	}

	MaterializedSpecial = map[int]struct{}{
		42013: struct{}{},
		42011: struct{}{},
		42009: struct{}{},
		42007: struct{}{},
		42001: struct{}{},
		20004: struct{}{},
		20002: struct{}{},
		20001: struct{}{},
		40021: struct{}{},
		40020: struct{}{}, // Death
		42003: struct{}{}, // STEMI
		42005: struct{}{}, // NSTEMI
		42019: struct{}{}, // All-cause dementia
		42021: struct{}{}, // Alzheimer dementia
		42023: struct{}{}, // Vascular dementia
		42025: struct{}{}, // Frontotemporal dementia
		42027: struct{}{}, // ESRD
		6150:  struct{}{}, // Self-reported MI, Angina, CVA, HTN
		6152:  struct{}{}, // Self-reported DVT, PE, Asthma, Hayfever, Emphysema
		2443:  struct{}{}, // Self-reported diabetes
		4041:  struct{}{}, // Self-reported gestational diabetes

		// These are specially constructed but don't have a known date and just
		// get set to the enroll date
		3079: struct{}{}, // Pacemaker ()
	}
)

var (
	// These can be helpful to make sure that the user is including all fields
	// that use the same family of codes

	ICD9 = map[int]struct{}{
		41203: struct{}{},
		41205: struct{}{},
		// 40013 Cancer Registry ICD9 is intentionally excluded, since asking
		// the user to include the cancer registry when they are not including
		// cancer codes would be counterproductive.
		//
		// TODO: recognize specific ICD codes and make the prompt if the user
		// includes cancer-specific codes but not the cancer registry.
	}

	ICD10 = map[int]struct{}{
		41202: struct{}{},
		41204: struct{}{},
		40001: struct{}{},
		40002: struct{}{},
		// 40006 Cancer Registry ICD10 is intentionally excluded, since asking
		// the user to include the cancer registry when they are not including
		// cancer codes would be counterproductive.
		//
		// TODO: recognize specific ICD codes and make the prompt if the user
		// includes cancer-specific codes but not the cancer registry.
		//
		// TODO: Add FieldID 41270 - perhaps optionally? The date for this
		// FieldID is given in 41280
	}

	OPCS = map[int]struct{}{
		41200: struct{}{},
		41210: struct{}{},
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
