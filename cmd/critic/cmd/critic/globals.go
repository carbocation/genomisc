package main

import (
	"sync"

	"cloud.google.com/go/storage"
	"github.com/jmoiron/sqlx"
)

type Global struct {
	log           logger
	db            *sqlx.DB
	storageClient *storage.Client

	Site      string
	Company   string
	Email     string
	SnailMail string

	Project      string
	ManifestPath string
	RawRoot      string
	MergedRoot   string
	OutputPath   string
	Labels       []Label

	m        sync.RWMutex
	manifest *AnnotationTracker
}

func (g *Global) Manifest() []ManifestEntry {
	g.m.RLock()
	defer g.m.RUnlock()

	if g.manifest == nil {
		return nil
	}

	return g.manifest.GetEntries()
}

type logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
