package admin

import (
	//	"fmt"
	"appengine"
	"appengine/datastore"
	"gophers/plate"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Website struct {
	Name        string
	URL         string
	Interval    int
	LastChecked time.Time
	Monitoring  bool
	Status      string
}

type QueryResult struct {
	Key     *datastore.Key
	Website Website
}

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

	c := appengine.NewContext(r)
	q := datastore.NewQuery("website").Order("Name")

	//sites := make([]Website, 0)
	/*var keys []*datastore.Key
	if keys, err = q.GetAll(c, &sites); err != nil {
		// handle the error
	}
	log.Println(sites)
	log.Println(keys)
	*/
	var sites []QueryResult
	i := 0
	for t := q.Run(c); ; {
		var x Website
		key, err := t.Next(&x)
		sites[i] = QueryResult{
			Key:     key,
			Website: x,
		}
		i++
		if err == datastore.Done {
			break
		}
		if err != nil {
			break
		}
	}
	log.Println(sites)
	//tmpl.Bag["Keys"] = keys
	tmpl.Bag["SiteCount"] = len(sites)
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
	tmpl.Template = "templates/admin/add.html"

	tmpl.DisplayTemplate()
}

func Save(w http.ResponseWriter, r *http.Request) {
	// Save website url
	c := appengine.NewContext(r)
	name := r.FormValue("name")
	urlstr := r.FormValue("url")
	interval, err := strconv.Atoi(r.FormValue("interval"))

	if strings.TrimSpace(name) == "" || strings.TrimSpace(urlstr) == "" || err != nil || interval < 5 {
		http.Redirect(w, r, "/add/"+url.QueryEscape("Name and URL are required. Interval must be an integer greater than 5."), http.StatusFound)
	}

	site := Website{
		Name:        name,
		URL:         urlstr,
		Interval:    interval,
		Monitoring:  true,
		Status:      "unknown",
		LastChecked: time.Now(),
	}

	_, err1 := datastore.Put(c, datastore.NewIncompleteKey(c, "website", nil), &site)
	if err1 != nil {
		http.Redirect(w, r, "/add/"+url.QueryEscape("There was a problem saving to the datastore: "+err.Error()), http.StatusFound)
	} else {
		http.Redirect(w, r, "/admin", http.StatusFound)
	}

}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	plate.Serve404(w, "")
}
