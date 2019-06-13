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
	DicomRoot    string

	m        sync.RWMutex
	manifest []Manifest
}

func (g Global) Manifest() []Manifest {
	g.m.RLock()
	defer g.m.RUnlock()

	return g.manifest
}

type logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
