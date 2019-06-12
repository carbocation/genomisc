package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func (h *handler) TemplateOnly(w http.ResponseWriter, r *http.Request) {
	tpl := mux.Vars(r)["template"]
	if tpl == "" {
		tpl = "index"
	}

	Render(h, w, r, strings.Title(tpl), fmt.Sprintf("%s.html", tpl), nil, nil)
}

func (h *handler) Index(w http.ResponseWriter, r *http.Request) {
	output := struct{ Project string }{h.Global.Project}

	Render(h, w, r, h.Global.Site, "index.html", output, nil)
}

func (h *handler) ListProject(w http.ResponseWriter, r *http.Request) {
	if err := UpdateManifest(); err != nil {
		HTTPError(h, w, r, err)
		return
	}

	output := struct {
		Project  string
		Manifest []Manifest
	}{
		h.Global.Project,
		h.Global.Manifest(),
	}

	Render(h, w, r, "List Project", "listproject.html", output, nil)
}

func (h *handler) CriticHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch the desired image from the zip file as described in the manifest
	manifestIdx := mux.Vars(r)["manifest_index"]
	manifestIndex, err := strconv.Atoi(manifestIdx)
	if err != nil {
		HTTPError(h, w, r, fmt.Errorf("No manifest_index passed"))
		return
	}

	if manifestIndex >= len(h.Global.Manifest()) {
		HTTPError(h, w, r, fmt.Errorf("Manifest_index was %d, out of range of the Manifest slice", manifestIndex))
		return
	}

	manifestEntry := h.Global.Manifest()[manifestIndex]

	pathPart := h.Global.DicomRoot
	im, err := ExtractDicom(fmt.Sprintf("%s/%s", pathPart, manifestEntry.Zip), manifestEntry.Dicom, true)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Convert that image to a PNG and base64 encode it so we can show it raw
	var imBuff bytes.Buffer
	png.Encode(&imBuff, im)
	encodedString := base64.StdEncoding.EncodeToString(imBuff.Bytes())

	output := struct {
		Project       string
		ManifestEntry Manifest
		ManifestIndex int
		EncodedImage  string
		Width         int
		Height        int
	}{
		h.Global.Project,
		manifestEntry,
		manifestIndex,
		strings.NewReplacer("\n", "", "\r", "").Replace(encodedString),
		im.Bounds().Dx(),
		im.Bounds().Dy(),
	}

	Render(h, w, r, "Critic Handler", "traceoverlay.html", output, nil)
}

func (h *handler) TraceOverlayPost(w http.ResponseWriter, r *http.Request) {
	manifestIdx := mux.Vars(r)["manifest_index"]
	manifestIndex, err := strconv.Atoi(manifestIdx)
	if err != nil {
		HTTPError(h, w, r, fmt.Errorf("No manifest_index passed"))
		return
	}
	if manifestIndex >= len(h.Global.Manifest()) {
		HTTPError(h, w, r, fmt.Errorf("Manifest_index was %d, out of range of the Manifest slice", manifestIndex))
		return
	}

	// TODO:
	// Here, you need to actually ingest the response, etc
	r.ParseForm()
	log.Println("Annotation submitted:", r.PostForm.Get("is_bad"))

	nextURL, err := h.router.Get("critic").URL("manifest_index", strconv.Itoa(manifestIndex+1))
	if err != nil {
		HTTPError(h, w, r, err)
	}

	http.Redirect(w, r, nextURL.String(), http.StatusSeeOther)
}

func (h *handler) Goroutines(w http.ResponseWriter, r *http.Request) {
	goroutines := fmt.Sprintf("%d goroutines are currently active\n", runtime.NumGoroutine())

	w.Write([]byte(goroutines))
}
