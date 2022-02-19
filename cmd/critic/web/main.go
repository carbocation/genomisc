package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	_ "github.com/carbocation/genomisc/compileinfoprint"
)

var global *Global

func init() {
	// Prevent seed re-use
	rand.Seed(int64(time.Now().Nanosecond()))
}

func main() {
	errors := make(chan error, 1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		//syscall.SIGINFO,
	)

	var preParsed bool
	var labelsFile string
	manifest := flag.String("manifest", "", "Tab-delimited manifest file which contains a zip_file and a dicom_file column (at least).")
	dicomRoot := flag.String("dicom-path", "", "Root path under which all DICOM zip files sit, or the root folder for ./dicom_pngs and ./merged_pngs if --preparsed=true. If empty, folder where manifest file resides. May be a Google Storage URL (gs://).")
	outputPath := flag.String("output", "", "Path to a local file where all output will be written. Will be created if it does not yet exist.")
	port := flag.Int("port", 9019, "Port for HTTP server")
	flag.BoolVar(&preParsed, "preparsed", false, "(Optional) If true, looks for pre-parsed images under {dicom-path}/dicom_pngs and {dicom-path}/merged_pngs (depending on whether the UI has toggled the overlay). This is useful if images are pre-parsed as PNGs.")
	flag.StringVar(&labelsFile, "labels", "", "(Optional) json file with labels. E.g.: {Labels: [{'name':'Label 1', 'value':'l1'}]}")
	//dbName := flag.String("db_name", "pubrank", "Name of the database schema to connect to")
	flag.Parse()

	if *manifest == "" || *outputPath == "" {
		flag.PrintDefaults()
		return
	}

	if *dicomRoot == "" {
		*dicomRoot = filepath.Dir(*manifest)
	}

	sclient, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	global = &Global{
		Site:          "Critic",
		Company:       "Broad Institute",
		Email:         "jamesp@broadinstitute.org",
		SnailMail:     "415 Main Street, Cambridge MA",
		log:           log.New(os.Stderr, log.Prefix(), log.Ldate|log.Ltime),
		db:            nil,
		storageClient: sclient,

		Project:      *outputPath,
		ManifestPath: *manifest,
		DicomRoot:    *dicomRoot,
		PreParsed:    preParsed,
		Labels:       []Label{{DisplayName: "Bad Image", Value: "bad-image"}, {DisplayName: "Mistraced Segmentation", Value: "mistrace"}, {DisplayName: "Good", Value: "good"}},
	}

	sortedAnnotatedManifest, err := ReadManifestAndCreateOutput(*manifest, *outputPath)
	if err != nil {
		log.Fatalln(err)
	}

	global.manifest = sortedAnnotatedManifest

	if labelsFile != "" {
		lf, err := os.Open(labelsFile)
		if err != nil {
			log.Fatalln(err)
		}

		if newLabels, err := ParseLabelFile(lf); err != nil {
			log.Fatalln(err)
		} else {
			global.Labels = newLabels
		}
	}

	global.log.Println("Launching", global.Site)

	whoami, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}

	global.log.Println("Locally, you should now run:")
	global.log.Printf("gcloud compute ssh %s@%s -- -NnT -L %d:localhost:%d\n", whoami.Username, hostname, *port, *port)

	go func() {
		global.log.Println("Starting HTTP server on port", *port)

		routing, err := router(global)
		if err != nil {
			errors <- err
			global.log.Println(err)
			sig <- syscall.SIGTERM
			return
		}

		if err := http.ListenAndServe(fmt.Sprintf(`:%d`, *port), routing); err != nil {
			errors <- err
			global.log.Println(err)
			sig <- syscall.SIGTERM
			return
		}
	}()

Outer:
	for {
		select {
		case sigl := <-sig:

			//if sigl == syscall.SIGINFO || sigl == syscall.SIGUSR1 {
			if sigl == syscall.SIGUSR1 {
				SigStatus()
				continue
			}

			// By default, exit
			global.log.Printf("\nExit: %s\n", sigl.String())

			break Outer

		case err := <-errors:
			if err == nil {
				global.log.Println("Finished")
				break Outer
			}

			// Return a status code indicating failure
			global.log.Println("Exiting due to error", err)
			os.Exit(1)
		}
	}
}

func SigStatus() {
	global.log.Println("There are", runtime.NumGoroutine(), "goroutines running")
}
