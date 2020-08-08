package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
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

	showOverlay := true
	r.ParseForm()
	if overlay := r.Form.Get("overlay"); overlay == "off" {
		showOverlay = false
	}

	var im image.Image

	if !h.Global.PreParsed {
		// Read the zip file either from a local file, or from Google storage,
		// depending on the prefix you provide.
		im, err = bulkprocess.ExtractDicomFromGoogleStorage(
			fmt.Sprintf("%s/%s", h.Global.DicomRoot, manifestEntry.Zip),
			manifestEntry.Dicom,
			showOverlay,
			h.Global.storageClient)
	} else {
		// If there is a suffix, we're directly reading an image, not through a
		// zipped DICOM

		// TODO: Stop hardcoding the subfolders and suffixes - make this configurable
		if showOverlay {
			im, err = bulkprocess.ExtractImageFromGoogleStorage(manifestEntry.Dicom, ".png.overlay.png", h.Global.DicomRoot+"/merged_pngs", h.Global.storageClient)
		} else {
			im, err = bulkprocess.ExtractImageFromGoogleStorage(manifestEntry.Dicom, ".png", h.Global.DicomRoot+"/dicom_pngs", h.Global.storageClient)
		}
	}
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
		Labels        []Label
	}{
		h.Global.Project,
		manifestEntry,
		manifestIndex,
		strings.NewReplacer("\n", "", "\r", "").Replace(encodedString),
		im.Bounds().Dx(),
		im.Bounds().Dy(),
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
	if err := h.Global.manifest.WriteAnnotationsToDisk(h.Global.DicomRoot); err != nil {
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
