package website

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"encoding/json"
	"errors"
	"gophers/helpers/history"
	"gophers/helpers/notify"
	"gophers/helpers/rest"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Website struct {
	ID            int64
	Name          string
	URL           string
	Interval      int
	Monitoring    bool
	Status        history.History
	Public        bool
	EmailInterval int
	LogDays       int
	Uptime        float32
}

type WebsiteSave struct {
	ID            int64
	Name          string
	URL           string
	Interval      int
	Monitoring    bool
	Public        bool
	EmailInterval int
	LogDays       int
}

func (website Website) IntervalMins() int {
	return website.Interval * website.EmailInterval
}

func GetAll(r *http.Request) (sites []Website, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("website").Order("Name")

	//var sites []QueryResult
	sites = make([]Website, 0)
	_, err = q.GetAll(c, &sites)
	for i := 0; i < len(sites); i++ {
		sites[i].Status, err = history.GetStatus(r, sites[i].ID)
		sites[i].GetUptime(r)
	}

	return sites, err
}

func GetPublic(r *http.Request) (sites []Website, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("website").Filter("Public =", true).Order("Name")

	//var sites []QueryResult
	sites = make([]Website, 0)
	_, err = q.GetAll(c, &sites)
	for i := 0; i < len(sites); i++ {
		sites[i].Status, err = history.GetStatus(r, sites[i].ID)
		sites[i].GetUptime(r)
	}

	return sites, err
}

func GetCachedSites(r *http.Request) (sites []Website, err error) {
	c := appengine.NewContext(r)
	//sites, err := GetAll(r)
	var item *memcache.Item
	// Get the item from the memcache
	if item, err = memcache.Get(c, "sites"); err == memcache.ErrCacheMiss {
		q := datastore.NewQuery("website").Order("Name")
		_, err = q.GetAll(c, &sites)
		if err == nil {
			sitesdata, _ := json.Marshal(sites)
			item := &memcache.Item{
				Key:   "sites",
				Value: sitesdata,
			}
			err = memcache.Add(c, item)
		}
	} else {
		err = json.Unmarshal(item.Value, &sites)
	}
	return sites, err
}

func UpdateCachedSites(r *http.Request) {
	c := appengine.NewContext(r)
	var item *memcache.Item

	// retreive websites
	q := datastore.NewQuery("website").Order("Name")
	sites := make([]Website, 0)
	_, err := q.GetAll(c, &sites)

	if err != nil {
		log.Fatal("Error retreiving sites from datastore")
	}
	sitesdata, err := json.Marshal(sites)

	// Get the item from the memcache
	if item, err = memcache.Get(c, "sites"); err == memcache.ErrCacheMiss {
		// item doesn't exist at all...add it
		item := &memcache.Item{
			Key:   "sites",
			Value: sitesdata,
		}
		err = memcache.Add(c, item)
		log.Println(err)
	} else {
		// item does exist, update
		item.Value = sitesdata
		_ = memcache.Set(c, item)
	}
}

func CleanLogs(r *http.Request) {
	// this task runs only once a day
	c := appengine.NewContext(r)
	q := datastore.NewQuery("website").Filter("Monitoring =", true)

	sites := make([]Website, 0)
	_, err := q.GetAll(c, &sites)
	if err == nil {
		for i := 0; i < len(sites); i++ {
			history.ClearOld(r, sites[i].ID, sites[i].LogDays)
		}
	}

}

func CheckSites(r *http.Request) (err error) {
	sites, err := GetCachedSites(r)
	now := time.Now()
	logs := make(map[int64]history.History)
	if err != nil {
		log.Fatal(err)
	}
	if err == nil {
		for i := 0; i < len(sites); i++ {
			dur := time.Duration(sites[i].Interval) * time.Minute

			status, _ := history.GetStatus(r, sites[i].ID)

			if now.Sub(status.Checked) >= dur {
				logs[sites[i].ID] = sites[i].Check(r)
			}
		}
	}
	history.SaveLogs(r, logs)
	return err
}

func Get(r *http.Request, key int64) (site Website, sitesave WebsiteSave, err error) {

	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "website", "", key, nil)
	err = datastore.Get(c, k, &sitesave)
	if err == nil {
		site.ID = sitesave.ID
		site.Name = sitesave.Name
		site.URL = sitesave.URL
		site.Interval = sitesave.Interval
		site.Monitoring = sitesave.Monitoring
		site.Public = sitesave.Public
		site.EmailInterval = sitesave.EmailInterval
		site.LogDays = sitesave.LogDays
		site.Status, err = history.GetStatus(r, sitesave.ID)
		site.GetUptime(r)
	}
	return site, sitesave, err
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
	emailInterval, err := strconv.Atoi(r.FormValue("emailinterval"))
	logdays, err := strconv.Atoi(r.FormValue("logdays"))
	if err != nil || logdays < 1 {
		logdays = 1
	}
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

	if strings.TrimSpace(name) == "" || strings.TrimSpace(urlstr) == "" || err != nil || interval < 5 || logdays < 1 {
		err = errors.New("Name and URL are required. Interval must be an integer greater than 5. Log Days kept must be greater than 1.")
		return
	}

	var keynum int64
	keynum, err = strconv.ParseInt(r.FormValue("siteID"), 10, 64)

	if err != nil {
		// new Website
		site := WebsiteSave{
			Name:          name,
			URL:           urlstr,
			Interval:      interval,
			Monitoring:    monitoring,
			Public:        public,
			EmailInterval: emailInterval,
			LogDays:       logdays,
		}

		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			// Note: this function's argument c shadows the variable c
			//       from the surrounding function.

			key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "website", nil), &site)

			if err == nil {
				site.ID = key.IntID()
				key, err = datastore.Put(c, key, &site)
			}

			return err
		}, nil)
		UpdateCachedSites(r)
		return err
	} else {
		// update website
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			k := datastore.NewKey(c, "website", "", keynum, nil)
			_, site, err := Get(r, keynum)
			if err != nil {
				return err
			}
			site.Name = name
			site.URL = urlstr
			site.Interval = interval
			site.Monitoring = monitoring
			site.Public = public
			site.LogDays = logdays
			site.EmailInterval = emailInterval

			_, err = datastore.Put(c, k, &site)
			return err
		}, nil)
		UpdateCachedSites(r)
		return err
	}
	return

}

func (website *Website) GetUptime(r *http.Request) {
	website.Uptime = history.Uptime(r, website.ID)
}

func (website Website) GetNotifiers(r *http.Request) (notifiers []notify.Notify, err error) {
	notifiers, err = notify.GetAllBySite(r, website.ID)
	return
}

func (website *Website) Check(r *http.Request) history.History {
	status := rest.Get(website.URL, r)
	prevStatus := website.Status.Status
	var err error
	if status {
		send := prevStatus == "down"
		website.Status = history.Log(r, website.ID, time.Now(), "up", send)
		if send {
			err := website.EmailNotifiers(r)
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		send := (prevStatus == "up") || (website.OkToSend(r))
		website.Status = history.Log(r, website.ID, time.Now(), "down", send)
		if send {
			err = website.EmailNotifiers(r)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return website.Status
}

func (website Website) EmailNotifiers(r *http.Request) (err error) {
	notifiers, err := website.GetNotifiers(r)
	if err == nil {
		for i := 0; i < len(notifiers); i++ {
			notifiers[i].Notify(r, website.Name, website.URL, website.Status.Checked, website.Status.Status)
		}
	}
	return
}

func (website Website) GetHistory(r *http.Request, page int, perpage int) (logs []history.History, pages int, err error) {
	logs, pages, err = history.GetHistory(r, website.ID, page, perpage)
	return
}

func (website Website) OkToSend(r *http.Request) bool {
	lastChange, err := history.GetLastEmail(r, website.ID)
	if err != nil {
		return true
	}
	sinceLast := time.Now().Sub(lastChange.Checked).Minutes()
	dur := (time.Duration(website.Interval*website.EmailInterval) * time.Minute).Minutes()
	return sinceLast > dur
}
