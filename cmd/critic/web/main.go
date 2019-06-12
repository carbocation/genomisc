package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
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

	manifest := flag.String("manifest", "", "Tab-delimited manifest file which contains a zip_file and a dicom_file column (at least).")
	project := flag.String("project", "", "Project name. Defines a folder into which all overlays will be written.")
	port := flag.Int("port", 9019, "Port for HTTP server")
	//dbName := flag.String("db_name", "pubrank", "Name of the database schema to connect to")
	flag.Parse()

	if *manifest == "" || *project == "" {
		flag.PrintDefaults()
		return
	}

	log.Printf("Creating directory ./%s/ if it does not yet exist\n", *project)
	newpath := filepath.Join(".", *project)
	if err := os.MkdirAll(newpath, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	manifestLines, err := ReadManifest(*manifest, *project)
	if err != nil {
		log.Fatalln(err)
	}

	global = &Global{
		Site:      "TraceOverlay",
		Company:   "Broad Institute",
		Email:     "jamesp@broadinstitute.org",
		SnailMail: "415 Main Street, Cambridge MA",
		log:       log.New(os.Stderr, log.Prefix(), log.Ldate|log.Ltime),
		db:        nil,

		Project:      *project,
		ManifestPath: *manifest,
		manifest:     manifestLines,
	}

	// global.db.SetMaxOpenConns(*maxOpen)
	// global.db.SetMaxIdleConns(*maxIdle)

	global.log.Println("Launching TraceOverlay")

	go func() {
		global.log.Println("Starting HTTP server on port", *port)
		if err := http.ListenAndServe(fmt.Sprintf(`:%d`, *port), router(global)); err != nil {
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
