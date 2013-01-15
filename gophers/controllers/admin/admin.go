package admin

import (
	"fmt"
	"gophers/helpers/website"
	"gophers/plate"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func Index(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()
	session := plate.Session.Get(r)
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
	}

	sites := website.GetAll(r)
	tmpl.Bag["Sites"] = sites
	tmpl.Bag["Name"] = session["name"]
	tmpl.Template = "templates/admin/index.html"

	tmpl.DisplayTemplate()
}

func Add(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()
	var err error
	var tmpl plate.Template

	params := r.URL.Query()
	error := params.Get(":error")
	error, _ = url.QueryUnescape(error)

	tmpl, err = server.Template(w)

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Error"] = error
	tmpl.Template = "templates/admin/form.html"

	tmpl.DisplayTemplate()
}

func Edit(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()
	var err error
	var tmpl plate.Template

	params := r.URL.Query()
	error := params.Get(":error")
	error, _ = url.QueryUnescape(error)

	tmpl, err = server.Template(w)

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}
	var keynum int64
	keynum, _ = strconv.ParseInt(params.Get(":key"), 10, 64)
	site, err := website.Get(r, keynum)

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Website"] = site
	tmpl.Bag["Error"] = error
	tmpl.Template = "templates/admin/form.html"

	tmpl.DisplayTemplate()
}

func Delete(w http.ResponseWriter, r *http.Request) {
	err := website.Delete(r)
	if err == nil {
		fmt.Fprint(w, "{\"success\":true}")
	} else {
		fmt.Fprint(w, "{\"success\":false}")
	}
}

func Save(w http.ResponseWriter, r *http.Request) {
	// Save website url
	err := website.Save(r)
	if err != nil {
		http.Redirect(w, r, "/add/"+url.QueryEscape("There was a problem saving to the datastore: "+err.Error()), http.StatusFound)
	} else {
		http.Redirect(w, r, "/admin", http.StatusFound)
	}

}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	plate.Serve404(w, "")
}
