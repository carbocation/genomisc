package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/image/bmp"

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

func (h *handler) TraceOverlay(w http.ResponseWriter, r *http.Request) {
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

	pathPart := path.Dir(h.Global.ManifestPath)
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

	Render(h, w, r, "Trace Overlay", "traceoverlay.html", output, nil)
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

	output := struct {
		Project       string
		ManifestIndex int
		ManifestEntry Manifest
		EncodedImage  string
		Width         int
		Height        int
	}{
		Project:       h.Global.Project,
		ManifestIndex: manifestIndex,
		ManifestEntry: manifestEntry,
	}

	parsedImage := strings.SplitAfterN(r.PostFormValue("imgBase64"), ",", 2)
	if len(parsedImage) < 2 {
		log.Println(r.FormValue("imgBase64"))
		HTTPError(h, w, r, fmt.Errorf("Parsed image was shorter than expected"))
		return
	}

	imgRdr := strings.NewReader(parsedImage[1])
	dec := base64.NewDecoder(base64.StdEncoding, imgRdr)

	bmpImage, err := bmp.Decode(dec)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Save the BMP to disk under your project folder
	f, err := os.Create(filepath.Join(".", global.Project, manifestEntry.Zip+"_"+manifestEntry.Dicom+".png"))
	if err != nil {
		HTTPError(h, w, r, err)
	}
	defer f.Close()

	// BMP encoding yields all black files for some reason?
	// if err := bmp.Encode(f, bmpImage); err != nil {
	// 	HTTPError(h, w, r, err)
	// 	return
	// }
	if err := png.Encode(f, bmpImage); err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Convert that image to a PNG and base64 encode it so we can show it raw
	var imBuff bytes.Buffer
	png.Encode(&imBuff, bmpImage)
	encodedString := base64.StdEncoding.EncodeToString(imBuff.Bytes())

	output.EncodedImage = encodedString
	output.Width = bmpImage.Bounds().Dx()
	output.Height = bmpImage.Bounds().Dy()

	Render(h, w, r, "Trace Overlay", "traceoverlay-POST.html", output, nil)
}

func (h *handler) Goroutines(w http.ResponseWriter, r *http.Request) {
	goroutines := fmt.Sprintf("%d goroutines are currently active\n", runtime.NumGoroutine())

	w.Write([]byte(goroutines))
}
