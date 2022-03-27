package convertpheno

import (
	"encoding/csv"
	"os"

	"github.com/carbocation/genomisc"
	"github.com/carbocation/pfx"
)

var withdrawnSampleIDs = make(map[string]struct{})

// IsWithdrawnSample returns true if the given sample ID is in the withdrawn
// sample lookup table.
func IsWithdrawnSample(sampleID string) bool {
	_, exists := withdrawnSampleIDs[sampleID]
	return exists
}

// SetWithdrawnSamplesFromFile is not safe for concurrent use. It allows you to
// set the lookup table.
func SetWithdrawnSamplesFromFile(withdrawnSampleFile string) error {
	withdrawnSampleFile = genomisc.ExpandHome(withdrawnSampleFile)

	f, err := os.Open(withdrawnSampleFile)
	if err != nil {
		return pfx.Err(err)
	}
	defer f.Close()

	withdrawals, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return pfx.Err(err)
	}

	// Just one entry per line, no header.
	withdrawalSlice := make([]string, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		if len(withdrawal[0]) <= 0 {
			continue
		}

		withdrawalSlice = append(withdrawalSlice, withdrawal[0])
	}

	SetWithdrawnSamples(withdrawalSlice)

	return nil
}

// SetWithdrawnSamples is not safe for concurrent use.
func SetWithdrawnSamples(withdrawnSampleIDList []string) {
	for _, withdrawnSampleID := range withdrawnSampleIDList {
		withdrawnSampleIDs[withdrawnSampleID] = struct{}{}
	}
}
