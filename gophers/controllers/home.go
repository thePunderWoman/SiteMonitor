package controllers

import (
	"fmt"
	"gophers/helpers/website"
	"gophers/plate"
	"html/template"
	"math"
	"net/http"
	"time"
)

func Index(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template

	tmpl, err = server.Template(w)

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.FuncMap = template.FuncMap{
		"formatDate": func(dt time.Time) string {
			layout := "Mon, 01/02/06, 3:04PM MST"
			Local, _ := time.LoadLocation("US/Central")
			return dt.In(Local).Format(layout)
		},
		"formatDecimal": func(dc float32) string {
			if !math.IsNaN(float64(dc)) {
				return fmt.Sprintf("%.2f", dc) + "%"
			}
			return "-"
		},
	}

	sites, err := website.GetPublic(r)

	tmpl.Bag["Sites"] = sites
	tmpl.Template = "templates/index.html"

	tmpl.DisplayTemplate()
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	plate.Serve404(w, "")
}
