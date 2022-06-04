package main

import (
	"encoding/json"
	"net/http"
)

const (
	JSON = "json"
	HTML = "html"
)

type Page struct {
	Title     string
	Site      string
	Company   string
	Email     string
	SnailMail string
	Assets    string
	Data      interface{}
}

type renderOpts struct {
	OutputFormat string
}

func NewRenderOpts() *renderOpts {
	return &renderOpts{
		OutputFormat: HTML,
	}
}

func Render(h *handler, w http.ResponseWriter, r *http.Request, title string, tpl string, data interface{}, opts *renderOpts) {
	if opts == nil {
		opts = NewRenderOpts()
	}

	if opts.OutputFormat == JSON {
		renderJSON(h, w, r, data, *opts)
		return
	}

	page := Page{
		Title:     title,
		Site:      h.Global.Site,
		Company:   h.Global.Company,
		Email:     h.Global.Email,
		SnailMail: h.Global.SnailMail,
		Assets:    h.Assets(),
		Data:      data,
	}

	renderHTML(h, w, r, tpl, page, *opts)
}

func renderJSON(h *handler, w http.ResponseWriter, r *http.Request, data interface{}, opts renderOpts) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func renderHTML(h *handler, w http.ResponseWriter, r *http.Request, tpl string, page Page, opts renderOpts) {
	if tpl == "" {
		tpl = "_base.html"
	}

	if err := h.Template(tpl).Execute(w, page); err != nil {
		HTTPError(h, w, r, err)
	}
}
