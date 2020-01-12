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

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
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
	// For now, don't bother checking disk for updates - two people should
	// not run pointing at the same output.

	// if err := UpdateManifest(); err != nil {
	// 	HTTPError(h, w, r, err)
	// 	return
	// }

	output := struct {
		Project  string
		Manifest []ManifestEntry
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

	// Read the zip file into memory still compressed - either from a local
	// file, or from Google storage, depending on the prefix you provide.
	f, nbytes, err := MaybeOpenFromGoogleStorage(fmt.Sprintf("%s/%s", h.Global.DicomRoot, manifestEntry.Zip), h.Global.storageClient)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	showOverlay := true
	r.ParseForm()
	if overlay := r.Form.Get("overlay"); overlay == "off" {
		showOverlay = false
	}

	// Now we have our compressed zip data in an io.ReaderAt, regardless of its
	// origin. The zip library can now consume it.
	im, err := bulkprocess.ExtractDicomFromReaderAt(f, nbytes, manifestEntry.Dicom, showOverlay)
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
		ManifestEntry ManifestEntry
		ManifestIndex int
		EncodedImage  string
		Width         int
		Height        int
		ShowOverlay   bool
	}{
		h.Global.Project,
		manifestEntry,
		manifestIndex,
		strings.NewReplacer("\n", "", "\r", "").Replace(encodedString),
		im.Bounds().Dx(),
		im.Bounds().Dy(),
		showOverlay,
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

	manifestEntry := h.Global.Manifest()[manifestIndex]

	// Apply the annotation
	r.ParseForm()

	userAnno := r.PostForm.Get("value")
	log.Println("Annotation submitted:", manifestEntry.SampleID, userAnno)

	if err := h.Global.manifest.SetAnnotation(manifestIndex, userAnno); err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Write to disk. Can consider launching in a goroutine to reduce delay.
	if err := h.Global.manifest.WriteAnnotationsToDisk(); err != nil {
		HTTPError(h, w, r, err)
		return
	}

	nextURL, err := h.router.Get("critic").URL("manifest_index", strconv.Itoa(manifestIndex+1))
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	http.Redirect(w, r, nextURL.String(), http.StatusSeeOther)
}

func (h *handler) Goroutines(w http.ResponseWriter, r *http.Request) {
	goroutines := fmt.Sprintf("%d goroutines are currently active\n", runtime.NumGoroutine())

	w.Write([]byte(goroutines))
}
