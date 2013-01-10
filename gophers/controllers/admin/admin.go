package controllers

import (
	//	"fmt"
	"gophers/plate"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template

	tmpl, err = server.Template(w)

	message := "Jessica"

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["message"] = message
	tmpl.Template = "templates/index.html"

	tmpl.DisplayTemplate()
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	plate.Serve404(w, "")
}
