package admin

import (
	"../../helpers/plate"
	"../../models"
	"fmt"
	"html/template"
	"math"
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
	var sites []models.Website
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
			"hasSites": func(sites []models.Website) bool {
				return len(sites) > 0
			},
		}
		tmplChan <- 1
	}()

	go func() {
		site := models.Website{}
		sites, err = site.GetAll()
		siteChan <- 1
	}()

	<-tmplChan
	<-siteChan

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Sites"] = sites
	tmpl.Bag["Name"] = session["name"]
	tmpl.Template = "templates/admin/index.html"

	tmpl.DisplayTemplate()
}

func Add(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()
	var err error
	var tmpl plate.Template
	tmplChan := make(chan int)

	params := r.URL.Query()
	error := params.Get(":error")
	error, _ = url.QueryUnescape(error)

	go func() {
		tmpl, err = server.Template(w)
		if err != nil {
			tmplChan <- 1
		}
		tmpl.FuncMap = template.FuncMap{
			"daysComparison": func(daysa int, daysb int) bool {
				x := daysa == daysb
				return x
			},
		}
		tmplChan <- 1
	}()

	<-tmplChan

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Website"] = new(models.Website)
	tmpl.Bag["Error"] = error
	tmpl.Template = "templates/admin/form.html"

	tmpl.DisplayTemplate()
}

func Edit(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()
	var err error
	var tmpl plate.Template
	var site models.Website
	tmplChan := make(chan int)
	siteChan := make(chan int)

	params := r.URL.Query()
	error := params.Get(":error")
	error, _ = url.QueryUnescape(error)

	go func() {
		tmpl, err = server.Template(w)
		if err != nil {
			tmplChan <- 1
		}
		tmpl.FuncMap = template.FuncMap{
			"daysComparison": func(daysa int, daysb int) bool {
				x := daysa == daysb
				return x
			},
		}
		tmplChan <- 1
	}()

	go func() {
		id, _ := strconv.Atoi(params.Get(":key"))
		site, err = site.Get(id)
		siteChan <- 1
	}()

	<-tmplChan
	<-siteChan

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
	var err error
	siteChan := make(chan int)

	go func() {
		site := models.Website{}
		err = site.Delete(r)
		siteChan <- 1
	}()
	<-siteChan
	if err == nil {
		fmt.Fprint(w, "{\"success\":true}")
	} else {
		fmt.Fprint(w, "{\"success\":false}")
	}
}

func Save(w http.ResponseWriter, r *http.Request) {
	var err error
	siteChan := make(chan int)

	go func() {
		site := models.Website{}
		err = site.Save(r)
		siteChan <- 1
	}()
	<-siteChan
	if err != nil {
		if r.FormValue("siteID") == "" {
			http.Redirect(w, r, "/add/"+url.QueryEscape("There was a problem saving to the database: "+err.Error()), http.StatusFound)
		} else {
			http.Redirect(w, r, "/edit/"+r.FormValue("siteID")+"/"+url.QueryEscape("There was a problem saving to the database: "+err.Error()), http.StatusFound)
		}
	} else {
		http.Redirect(w, r, "/admin", http.StatusFound)
	}

}

func GetNotifiers(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template
	var site models.Website
	var notifiers []models.Notify
	tmplChan := make(chan int)
	siteChan := make(chan int)

	go func() {
		tmpl, err = server.Template(w)
		if err != nil {
			tmplChan <- 1
		}
		tmplChan <- 1
	}()

	go func() {
		params := r.URL.Query()

		id, _ := strconv.Atoi(params.Get(":key"))
		site, err = site.Get(id)

		if err != nil {
			siteChan <- 1
		}

		notifiers, err = site.GetNotifiers(r)
		siteChan <- 1
	}()

	<-tmplChan
	<-siteChan

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Site"] = site
	tmpl.Bag["Notifiers"] = notifiers
	tmpl.Template = "templates/admin/notifiers.html"

	tmpl.DisplayTemplate()
}

func TestSend(w http.ResponseWriter, r *http.Request) {
	n := models.Notify{}
	notifier, err := n.Get(r)
	if err == nil {
		notifier.Notify("Test", "http://www.test.com", time.Now(), "up", 200, 12)
	}
	fmt.Fprint(w, "Sending Email")
}

func AddNotifier(w http.ResponseWriter, r *http.Request) {
	saveChan := make(chan int)
	var err error
	go func() {
		n := models.Notify{}
		err = n.Save(r)
		saveChan <- 1
	}()

	<-saveChan
	parentID := r.FormValue("parentID")
	if err != nil {
		http.Redirect(w, r, "/emails/"+parentID+"/"+url.QueryEscape("There was a problem saving to the database: "+err.Error()), http.StatusFound)
	} else {
		http.Redirect(w, r, "/emails/"+parentID, http.StatusFound)
	}
}

func DeleteNotifier(w http.ResponseWriter, r *http.Request) {
	delChan := make(chan int)
	var err error
	go func() {
		n := models.Notify{}
		err = n.Delete(r)
		delChan <- 1
	}()
	<-delChan
	if err == nil {
		fmt.Fprint(w, "{\"success\":true}")
	} else {
		fmt.Fprint(w, "{\"success\":false}")
	}
}

func Settings(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template
	var settings models.Setting
	tmplChan := make(chan int)
	settingChan := make(chan int)

	params := r.URL.Query()
	error := params.Get(":error")
	error, _ = url.QueryUnescape(error)

	go func() {
		tmpl, err = server.Template(w)
		if err != nil {
			tmplChan <- 1
		}
		tmplChan <- 1
	}()

	go func() {
		settings, err = settings.Get()
		settingChan <- 1
	}()

	<-tmplChan

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	<-settingChan

	tmpl.Bag["Settings"] = settings
	tmpl.Bag["Error"] = error
	tmpl.Template = "templates/admin/settings.html"

	tmpl.DisplayTemplate()
}

func SaveSettings(w http.ResponseWriter, r *http.Request) {
	settingChan := make(chan int)
	var err error
	go func() {
		setting := models.Setting{}
		err = setting.Save(r)
		settingChan <- 1
	}()
	<-settingChan
	if err != nil {
		http.Redirect(w, r, "/settings/"+url.QueryEscape("There was a problem saving to the database: "+err.Error()), http.StatusFound)
	} else {
		http.Redirect(w, r, "/settings", http.StatusFound)
	}
}

func GetHistory(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template
	var site models.Website
	var logs []models.HistoryGroup
	tmplChan := make(chan int)
	logChan := make(chan int)

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
			"formatDecimal": func(dc float64) string {
				if !math.IsNaN(float64(dc)) {
					return fmt.Sprintf("%.2f", dc) + " ms"
				}
				return "-"
			},
		}
		tmplChan <- 1
	}()

	params := r.URL.Query()
	go func() {
		id, _ := strconv.Atoi(params.Get(":key"))
		s := models.Website{}
		site, err = s.Get(id)

		logs, err = site.GetHistory(r)
		logChan <- 1
	}()

	<-tmplChan
	<-logChan

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Site"] = site
	tmpl.Bag["Logs"] = logs
	tmpl.Template = "templates/admin/history.html"

	tmpl.DisplayTemplate()
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	plate.Serve404(w, "")
}

func Check(w http.ResponseWriter, r *http.Request) {
	models.CheckSites(r)
	fmt.Fprint(w, "Checking sites")
}

func CleanLogs(w http.ResponseWriter, r *http.Request) {
	models.CleanLogs()
	fmt.Fprint(w, "Cleaning Logs")
}
