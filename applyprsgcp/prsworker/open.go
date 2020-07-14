package prsworker

import (
	"log"
	"time"

	"github.com/carbocation/bgen"
)

// OpenBGIAndBGEN loops until final error or success with opening the BGI and
// BGEN files. Useful because if you use this over an unreliable filesystem,
// you'll run into i/o errors that can be overcome by waiting a bit.
func OpenBGIAndBGEN(bgenPath string) (bgi *bgen.BGIIndex, b *bgen.BGEN, err error) {
	for loadAttempts, maxLoadAttempts := 1, 10; loadAttempts <= maxLoadAttempts; loadAttempts++ {
		bgi, err = bgen.OpenBGI(bgenPath + ".bgi?mode=ro")
		if err != nil && loadAttempts == maxLoadAttempts {
			// Ongoing failure at maxLoadAttempts is a terminal error
			return nil, nil, err
		} else if err != nil {
			// If we had an error, often due to an unreliable underlying
			// filesystem, wait for a substantial amount of time before
			// retrying.
			log.Println("bgen.OpenBGI: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
			time.Sleep(5 * time.Second)
			continue
		}
		bgi.Metadata.FirstThousandBytes = nil
		log.Printf("BGI Metadata: %+v\n", bgi.Metadata)

		// Load the BGEN for the chromosome
		b, err = bgen.Open(bgenPath)
		if err != nil && loadAttempts == maxLoadAttempts {
			// Ongoing failure at maxLoadAttempts is a terminal error
			return nil, nil, err
		} else if err != nil {
			// If we had an error, often due to an unreliable underlying
			// filesystem, wait for a substantial amount of time before retrying.
			log.Println("bgen.Open: Sleeping 5s to recover from", err.Error(), "attempt", loadAttempts)
			time.Sleep(5 * time.Second)

			// In this case, we know that the bgi successfully opened, so we
			// should close it before looping since we will lose that
			// handle.
			bgi.Close()

			continue
		}

		// If loading the bgen was error-free, no additional attempts
		// are required.
		break
	}

	return
}
