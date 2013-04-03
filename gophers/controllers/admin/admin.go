package admin

import (
	"fmt"
	"gophers/helpers/history"
	"gophers/helpers/notify"
	"gophers/helpers/serversettings"
	"gophers/helpers/website"
	"gophers/plate"
	"html/template"
	"log"
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
		}
		tmplChan <- 1
	}()

	go func() {
		sites, err = website.GetAll(r)
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

	tmpl.Bag["Website"] = new(website.Website)
	tmpl.Bag["Error"] = error
	tmpl.Template = "templates/admin/form.html"

	tmpl.DisplayTemplate()
}

func Edit(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()
	var err error
	var tmpl plate.Template
	var site website.Website
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
		var keynum int64
		keynum, _ = strconv.ParseInt(params.Get(":key"), 10, 64)
		site, _, err = website.Get(r, keynum)
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
		err = website.Delete(r)
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
		err = website.Save(r)
		siteChan <- 1
	}()
	<-siteChan
	if err != nil {
		if r.FormValue("siteID") == "" {
			http.Redirect(w, r, "/add/"+url.QueryEscape("There was a problem saving to the datastore: "+err.Error()), http.StatusFound)
		} else {
			http.Redirect(w, r, "/edit/"+r.FormValue("siteID")+"/"+url.QueryEscape("There was a problem saving to the datastore: "+err.Error()), http.StatusFound)
		}
	} else {
		http.Redirect(w, r, "/admin", http.StatusFound)
	}

}

func GetNotifiers(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template
	var site website.Website
	var notifiers []notify.Notify
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

		var keynum int64
		keynum, _ = strconv.ParseInt(params.Get(":key"), 10, 64)
		site, _, err = website.Get(r, keynum)

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
	notifier, err := notify.Get(r)
	if err == nil {
		log.Println(notifier)
		notifier.Notify(r, "Test", "http://www.test.com", time.Now(), "up", 200, 12)
	}
	fmt.Fprint(w, "Sending Email")
}

func AddNotifier(w http.ResponseWriter, r *http.Request) {
	saveChan := make(chan int)
	var err error
	go func() {
		err = notify.Save(r)
		saveChan <- 1
	}()

	<-saveChan
	parentID := r.FormValue("parentID")
	if err != nil {
		http.Redirect(w, r, "/emails/"+parentID+"/"+url.QueryEscape("There was a problem saving to the datastore: "+err.Error()), http.StatusFound)
	} else {
		http.Redirect(w, r, "/emails/"+parentID, http.StatusFound)
	}
}

func DeleteNotifier(w http.ResponseWriter, r *http.Request) {
	delChan := make(chan int)
	var err error
	go func() {
		err = notify.Delete(r)
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
	var settings *serversettings.Setting
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
		settings, err = serversettings.Get(r)
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
		err = serversettings.Save(r)
		settingChan <- 1
	}()
	<-settingChan
	if err != nil {
		http.Redirect(w, r, "/settings/"+url.QueryEscape("There was a problem saving to the datastore: "+err.Error()), http.StatusFound)
	} else {
		http.Redirect(w, r, "/settings", http.StatusFound)
	}
}

func GetHistory(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	var err error
	var tmpl plate.Template
	var site website.Website
	var logs []history.HistoryGroup
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
		var keynum int64
		keynum, _ = strconv.ParseInt(params.Get(":key"), 10, 64)
		site, _, err = website.Get(r, keynum)

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
	website.CheckSites(r)
	fmt.Fprint(w, "Checking sites")
}

func CleanLogs(w http.ResponseWriter, r *http.Request) {
	website.CleanLogs(r)
	fmt.Fprint(w, "Cleaning Logs")
}
