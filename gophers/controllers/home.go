package controllers

import (
	"fmt"
	"gophers/helpers/website"
	"gophers/plate"
	"html/template"
	"math"
	"net/http"
	"strings"
	"time"
)

func Index(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template
	var sites []website.Website
	siteChan := make(chan int)
	tmplChan := make(chan int)

	go func() {
		tmpl, err = server.Template(w)
		if err != nil {
			tmplChan <- 1
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
			"getDomain": func(url string) string {
				urlparts := strings.Split(url, "/")
				domain := url
				if len(urlparts) > 1 {
					domain = urlparts[0] + "//" + urlparts[2] + "/"
				}
				return domain
			},
		}
		tmplChan <- 1
	}()

	go func() {
		sites, err = website.GetAll(r)
		sites = website.PopulateUptime(r, sites)
		siteChan <- 1
	}()

	<-tmplChan
	<-siteChan

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Sites"] = sites
	tmpl.Template = "templates/index.html"

	tmpl.DisplayTemplate()
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	plate.Serve404(w, "")
}
