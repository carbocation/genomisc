package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JSONError(h *handler, w http.ResponseWriter, r *http.Request, err error, code ...int) {
	unifiedError(h, w, r, err, code...)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(struct {
		Success bool
		Message string
	}{
		false,
		err.Error(),
	})
}

func HTTPError(h *handler, w http.ResponseWriter, r *http.Request, err error, code ...int) {
	unifiedError(h, w, r, err, code...)

	output := struct {
		StatusCode     int
		StatusCodeText string
		Error          string
	}{
		StatusCode:     http.StatusInternalServerError,
		StatusCodeText: http.StatusText(http.StatusInternalServerError),
		Error:          err.Error(),
	}

	for _, c := range code {
		output.StatusCode = c
		output.StatusCodeText = http.StatusText(c)
		break // Take the first, if any is given
	}

	//r.Context()

	// Add a flash message about our error
	/*
		if err := s.AddFlashMessage(w, r, FlashMessage{Alert: "danger", Message: userError.Error()}); err != nil {
			// If there is an error adding this flash message, print what we can as plaintext and abort
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(code, err, userError)
			return
		}

		err = s.web.t.ExecuteTemplate(w, "error.html", Page(w, r, fmt.Sprintf(`%d %s`, code, http.StatusText(code)), nil))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(code, err, userError)
			return
		}
	*/

	/*
		Built from the Render() function, but not calling Render()
		to avoid possibility of infinite loop
	*/
	page := Page{
		Title:     "Error",
		Site:      h.Global.Site,
		Company:   h.Global.Company,
		Email:     h.Global.Email,
		SnailMail: h.Global.SnailMail,
		Assets:    h.Assets(),
		Data:      output,
	}

	if err := h.Template("error.html").Execute(w, page); err != nil {
		fmt.Fprintf(w, "Error (%d) (%v) with %+v", output.StatusCode, err, page)
	}
}

func unifiedError(h *handler, w http.ResponseWriter, r *http.Request, err error, code ...int) {
	usedCode := http.StatusInternalServerError
	if len(code) > 0 {
		usedCode = code[0]
	}
	w.WriteHeader(usedCode)
	h.log.Println(r.Host, r.URL.Path, ":", usedCode, err)
}
