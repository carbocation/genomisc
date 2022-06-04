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
	"runtime"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	_ "github.com/carbocation/genomisc/compileinfoprint"
)

var (
	global             *Global
	addSuffix          string
	removeSuffix       string
	onlyRawIsAvailable bool
)

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

	var rawRoot, mergedRoot, outputPath, labelsFile, manifestPath, nestedSuffix string
	var port int
	flag.StringVar(&addSuffix, "add_suffix", "", "(Optional) Suffix to add after the /merged/ filename to obtain the correct filename from the /raw/ data.")
	flag.StringVar(&removeSuffix, "remove_suffix", "", "(Optional) Suffix to remove from the /merged/ filename to obtain the correct filename from the /raw/ data.")
	flag.StringVar(&rawRoot, "raw", "", "(Optional) Path under which all secondary (usually raw/no-overlay) images sit. If --manifest is set, this may be a gs:// URL. Either raw or merged (or both) must be set.")
	flag.StringVar(&manifestPath, "manifest", "", "(Optional) Path with a file whose first column is the file names of the images of interest from the --merged folder.")
	flag.StringVar(&mergedRoot, "merged", "", "(Optional) Path under which all main images of interest sit. If --manifest is set, this may be a gs:// URL. Either raw or merged (or both) must be set.")
	flag.StringVar(&outputPath, "output", "", "Path to a local file where all output will be written. Will be created if it does not yet exist.")
	flag.StringVar(&nestedSuffix, "nested-suffix", "", "If images are nested within .tar.gz files, the wrapper file is assumed to be in a manifest column named 'zip_file' and the image filename assumed to be in a column named 'dicom_file'. --nested-suffix defines how the zip_file name is modified (.zip is removed and nested-suffix is added)")
	flag.IntVar(&port, "port", 9019, "Port for HTTP server")
	flag.StringVar(&labelsFile, "labels", "", "(Optional) json file with labels. E.g.: {Labels: [{'name':'Label 1', 'value':'l1'}]}")
	flag.Parse()

	// We permit passing either
	if (mergedRoot == "" && rawRoot == "") || outputPath == "" {
		flag.PrintDefaults()
		return
	}
	if mergedRoot == "" {
		onlyRawIsAvailable = true
	}

	mergedRoot = strings.TrimSuffix(mergedRoot, "/")
	rawRoot = strings.TrimSuffix(rawRoot, "/")

	var sclient *storage.Client
	var err error

	if strings.HasPrefix(mergedRoot, "gs://") ||
		strings.HasPrefix(manifestPath, "gs://") ||
		strings.HasPrefix(rawRoot, "gs://") {
		sclient, err = storage.NewClient(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
	}

	global = &Global{
		Site:          "Critic",
		Company:       "Broad Institute",
		Email:         "jamesp@broadinstitute.org",
		SnailMail:     "415 Main Street, Cambridge MA",
		log:           log.New(os.Stderr, log.Prefix(), log.Ldate|log.Ltime),
		db:            nil,
		storageClient: sclient,

		Project:    outputPath,
		RawRoot:    rawRoot,
		MergedRoot: mergedRoot,
		OutputPath: outputPath,
		Labels:     []Label{{DisplayName: "Bad Image", Value: "bad-image"}, {DisplayName: "Mistraced Segmentation", Value: "mistrace"}, {DisplayName: "Good", Value: "good"}},
	}

	var sortedAnnotatedManifest *AnnotationTracker

	manReader, _, err := ManifestPathToBufferedReader(manifestPath)
	if manifestPath != "" && err != nil {
		log.Fatalln(err)
	}
	if nestedSuffix != "" {
		sortedAnnotatedManifest, err = ReadNestedManifest(manReader, outputPath, nestedSuffix)
	} else {
		sortedAnnotatedManifest, err = CreateManifestAndOutput(mergedRoot, outputPath, manReader, nestedSuffix)
	}
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
	global.log.Printf("gcloud compute ssh %s@%s -- -NnT -L %d:localhost:%d\n", whoami.Username, hostname, port, port)

	go func() {
		global.log.Println("Starting HTTP server on port", port)

		routing, err := router(global)
		if err != nil {
			errors <- err
			global.log.Println(err)
			sig <- syscall.SIGTERM
			return
		}

		if err := http.ListenAndServe(fmt.Sprintf(`:%d`, port), routing); err != nil {
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
