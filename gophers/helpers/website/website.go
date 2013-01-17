package website

import (
	"appengine"
	"appengine/datastore"
	"errors"
	"gophers/helpers/notify"
	"gophers/helpers/rest"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Website struct {
	ID          int64
	Name        string
	URL         string
	Interval    int
	LastChecked time.Time
	Monitoring  bool
	Status      string
	Public      bool
}

func GetAll(r *http.Request) (sites []Website, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("website").Order("Name")

	//var sites []QueryResult
	sites = make([]Website, 0)
	_, err = q.GetAll(c, &sites)

	return sites, err
}

func GetPublic(r *http.Request) (sites []Website, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("website").Filter("Public =", true).Order("Name")

	//var sites []QueryResult
	sites = make([]Website, 0)
	_, err = q.GetAll(c, &sites)

	return sites, err
}

func CheckSites(r *http.Request) (err error) {
	c := appengine.NewContext(r)
	now := time.Now()
	q := datastore.NewQuery("website").Filter("Monitoring =", true)
	//var sites []QueryResult
	sites := make([]Website, 0)
	_, err = q.GetAll(c, &sites)
	if err == nil {
		for i := 0; i < len(sites); i++ {
			dur := time.Duration(sites[i].Interval) * time.Minute
			if now.Sub(sites[i].LastChecked) >= dur {
				sites[i].Check(r)
			}
		}
	}
	return err
}

func Get(r *http.Request, key int64) (site *Website, err error) {

	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "website", "", key, nil)
	w := new(Website)
	err = datastore.Get(c, k, w)

	return w, err
}

func Delete(r *http.Request) (err error) {
	c := appengine.NewContext(r)
	var keynum int64
	keynum, _ = strconv.ParseInt(r.FormValue("key"), 10, 64)
	k := datastore.NewKey(c, "website", "", keynum, nil)
	err = datastore.Delete(c, k)
	return
}

func Save(r *http.Request) (err error) {
	c := appengine.NewContext(r)
	name := r.FormValue("name")
	urlstr := r.FormValue("url")
	interval, err := strconv.Atoi(r.FormValue("interval"))
	var monitoring bool
	var public bool
	if r.FormValue("monitoring") == "" {
		monitoring = false
	} else {
		monitoring = true
	}
	if r.FormValue("public") == "" {
		public = false
	} else {
		public = true
	}

	if strings.TrimSpace(name) == "" || strings.TrimSpace(urlstr) == "" || err != nil || interval < 5 {
		err = errors.New("Name and URL are required. Interval must be an integer greater than 5.")
		return
	}

	var keynum int64
	keynum, err = strconv.ParseInt(r.FormValue("siteID"), 10, 64)

	if err != nil {
		// new Website
		site := Website{
			Name:        name,
			URL:         urlstr,
			Interval:    interval,
			Monitoring:  monitoring,
			Status:      "unknown",
			LastChecked: time.Now(),
			Public:      public,
		}

		key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "website", nil), &site)

		if err == nil {
			site.ID = key.IntID()
			key, err = datastore.Put(c, key, &site)
		}

		return err
	} else {
		// update website
		k := datastore.NewKey(c, "website", "", keynum, nil)
		site, err := Get(r, keynum)
		if err != nil {
			return err
		}
		site.Name = name
		site.URL = urlstr
		site.Interval = interval
		site.Monitoring = monitoring
		site.Public = public

		_, err = datastore.Put(c, k, site)
		return err
	}
	return

}

func (website Website) GetNotifiers(r *http.Request) (notifiers []notify.Notify, err error) {
	notifiers, err = notify.GetAllBySite(r, website.ID)
	return
}

func (website Website) Check(r *http.Request) {
	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "website", "", website.ID, nil)
	status := rest.Get(website.URL, r)
	website.LastChecked = time.Now()
	if status {
		if website.Status == "down" {
			err := website.EmailNotifiers(r, "up")
			if err != nil {
				log.Println(err)
			}
		}
		website.Status = "up"
		_, _ = datastore.Put(c, k, &website)
	} else {
		website.Status = "down"
		_, err := datastore.Put(c, k, &website)
		err = website.EmailNotifiers(r, "down")
		if err != nil {
			log.Println(err)
		}
	}
}

func (website Website) EmailNotifiers(r *http.Request, template string) (err error) {
	notifiers, err := website.GetNotifiers(r)
	if err == nil {
		for i := 0; i < len(notifiers); i++ {
			notifiers[i].Notify(r, website.Name, website.URL, website.LastChecked, template)
		}
	}
	return
}
