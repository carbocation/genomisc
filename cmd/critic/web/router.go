package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/interpose/middleware"
	"github.com/justinas/alice"
)

func router(config *Global) http.Handler {
	router := mux.NewRouter()
	POST := router.Methods("POST").Subrouter()
	GET := router.Methods("GET", "HEAD").Subrouter()

	h := handler{Global: config, router: router}

	GET.HandleFunc("/", h.Index).Name("index")
	GET.HandleFunc("/goroutines", h.Goroutines)
	GET.HandleFunc("/{template:(?:about|privacy|TOS|DMCA)}", h.TemplateOnly)
	GET.HandleFunc("/traceoverlay/{manifest_index}", h.TraceOverlay).Name("traceoverlay")
	GET.HandleFunc("/listproject", h.ListProject).Name("listproject")

	//
	// POST
	//
	POST.Handle("/", http.NotFoundHandler())
	POST.HandleFunc("/traceoverlay/{manifest_index}", h.TraceOverlayPost)

	// Static assets
	GET.PathPrefix(h.Assets()).Handler(
		middleware.MaxAgeHandler(60*60*24*364,
			http.StripPrefix(h.Assets(),
				http.FileServer(http.Dir(fmt.Sprintf(`%s/templates/static/`, h.Folder()))))))

	standard := alice.New(
		// Log all requests to STDOUT
		middleware.GorillaLog(),
	)

	return standard.Then(router)
}
