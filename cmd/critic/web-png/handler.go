package main

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/kardianos/osext"
)

const (
	BaseFilename = "_base.html"
)

// handler provides global values that must be
// safe for concurrent use from multiple goroutines
// to each handler method.
type handler struct {
	*Global

	router *mux.Router

	// Cached values / do not use directly. If they
	// need to be dynamic in the future, put them
	// under mutex protection.
	assets *string
	folder *string

	// Mutex protected values
	mu       sync.RWMutex
	template map[string]*template.Template
}

func (h *handler) Assets() string {
	if h.assets == nil {
		h.Global.log.Println("Initializing Assets")

		glyphs := fmt.Sprintf("/%s", RandHeteroglyphs(10))
		h.assets = &glyphs
	}

	return *h.assets
}

func (h *handler) Folder() string {
	if h.folder == nil {
		folder, err := osext.ExecutableFolder()
		if err != nil {
			panic(fmt.Errorf(`handlers.go:Folder: %s`, err))
		}
		h.folder = &folder

		h.Global.log.Printf("The binary is running in folder %s\n", *h.folder)
	}

	return *h.folder
}

func (h *handler) Template(templateFilename string) *template.Template {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.template == nil {
		func() {
			h.mu.RUnlock()
			h.mu.Lock()
			defer func() {
				h.mu.Unlock()
				h.mu.RLock()
			}()

			h.Global.log.Println("Initializing HTML templates")
			h.template = make(map[string]*template.Template, 0)

			//h.template = template.New("unused") // Build off a blank canvas

			/*
				tpl, err := h.template.ParseGlob(fmt.Sprintf(`%s/templates/*.html`, h.Folder()))
				if err != nil {
					panic(fmt.Errorf(`handlers.go:Template: %s`, err))
				}
			*/
			tpl, err := template.New(BaseFilename).Funcs(template.FuncMap{
				"add":       func(a, b int) int { return a + b },
				"cleanDate": func(d time.Time) string { return d.Format("January 02, 2006") },
				"year":      func(d time.Time) string { return d.Format("2006") },
				"noescape": func(s string) template.HTML {
					return template.HTML(s)
				},
			}).ParseGlob(fmt.Sprintf(`%s/templates/%s`, h.Folder(), "_*.html"))
			//}).ParseFiles(fmt.Sprintf(`%s/templates/%s`, h.Folder(), BaseFilename))

			if err != nil {
				h.Global.log.Printf("handlers.go:Template: %s\n", err)
				panic(fmt.Errorf(`handlers.go:Template: %s`, err))
			}

			h.template[BaseFilename] = tpl
		}()
	}

	// Prevent execution of the BaseFilename template, which would prevent future copies
	templateName := templateFilename
	if templateFilename == BaseFilename {
		templateName = fmt.Sprintf("CLONE%s", BaseFilename)
	}

	// Specific sub-template has already been generated
	if tpl, ok := h.template[templateName]; ok {
		return tpl
	}

	// Generate a clone of the base template so you don't contaminate it with the
	// derivative template's `define` statements.
	h.Global.log.Println("Initializing HTML template for", templateFilename)
	tpl, err := template.Must(h.template[BaseFilename].Clone()).ParseFiles(fmt.Sprintf(`%s/templates/%s`, h.Folder(), templateFilename))
	if err != nil {
		panic(fmt.Errorf(`handlers.go:Template: %s`, err))
	}
	h.mu.RUnlock()
	h.mu.Lock()
	h.template[templateName] = tpl
	h.mu.Unlock()
	h.mu.RLock()

	return tpl
}
