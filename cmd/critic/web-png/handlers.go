package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
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

	showOverlay := true
	r.ParseForm()
	if overlay := r.Form.Get("overlay"); overlay == "off" {
		showOverlay = false
	}
	// If raw is default, never show the overlay
	if onlyRawIsAvailable {
		showOverlay = false
	}

	imReader, filename, err := h.FetchImageFromContainer(manifestEntry, showOverlay)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}
	defer imReader.Close()

	imageBytes, err := ioutil.ReadAll(imReader)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Since we're consuming images as raw bytes without decoding them, we don't
	// need to encode them either.
	var encodingType string
	if strings.HasSuffix(filename, ".gif") {
		// A gif. Could be animated.
		encodingType = "gif"
	} else {
		// Not a .gif. Decode the stillframe image and encode it as a PNG.
		encodingType = "png"
	}
	encodedString := base64.StdEncoding.EncodeToString(imageBytes)

	output := struct {
		Project       string
		ManifestEntry ManifestEntry
		ManifestIndex int
		EncodingType  string
		EncodedImage  string
		Width         int
		Height        int
		ShowOverlay   bool
		Labels        []Label
	}{
		h.Global.Project,
		manifestEntry,
		manifestIndex,
		encodingType,
		strings.NewReplacer("\n", "", "\r", "").Replace(encodedString),
		10, // Previously im.Bounds().Dx(),
		10, // Previously im.Bounds().Dy(),
		showOverlay,
		h.Global.Labels,
	}

	Render(h, w, r, "Critic Handler", "critic.html", output, nil)
}

func (h *handler) CriticPost(w http.ResponseWriter, r *http.Request) {
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
	notedPath := h.Global.MergedRoot
	if notedPath == "" {
		notedPath = h.Global.RawRoot
	}
	if err := h.Global.manifest.WriteAnnotationsToDisk(notedPath); err != nil {
		HTTPError(h, w, r, err)
		return
	}

	nextURL, err := h.router.Get("critic").URL("manifest_index", strconv.Itoa(manifestIndex+1))
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Force query string so if your overlays are off, they stay off
	showOverlay := "on"
	log.Println(r.Form)
	if overlay := r.Form.Get("overlay"); overlay == "off" {
		showOverlay = "off"
	}
	qv := nextURL.Query()
	qv.Add("overlay", showOverlay)
	nextURL.RawQuery = qv.Encode()

	http.Redirect(w, r, nextURL.String(), http.StatusSeeOther)
}

func (h *handler) Goroutines(w http.ResponseWriter, r *http.Request) {
	goroutines := fmt.Sprintf("%d goroutines are currently active\n", runtime.NumGoroutine())

	w.Write([]byte(goroutines))
}
