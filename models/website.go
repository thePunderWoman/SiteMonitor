package models

import (
	"../helpers/database"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
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
	Status        History
	Public        bool
	EmailInterval int
	LogDays       int
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

func (website Website) GetAll(r *http.Request) (sites []Website, err error) {
	c := appengine.NewContext(r)
	//sites, err := GetAll(r)
	var item *memcache.Item
	var history History
	// Get the item from the memcache
	if item, err = memcache.Get(c, "sites"); err == memcache.ErrCacheMiss {
		q := datastore.NewQuery("website").Order("Name")
		_, err = q.GetAll(c, &sites)
		for i := 0; i < len(sites); i++ {
			sites[i].Status, err = history.GetStatus(r, sites[i].ID)
		}
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
	for i := 0; i < len(sites); i++ {
		sites[i].Status, err = history.GetStatus(r, sites[i].ID)
	}

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
	} else {
		// item does exist, update
		item.Value = sitesdata
		_ = memcache.Set(c, item)
	}
}

func CacheStatusChange(r *http.Request, logs map[int64]History) {
	c := appengine.NewContext(r)
	var item *memcache.Item

	sites, err := GetAll(r)
	for i := 0; i < len(sites); i++ {
		if logentry, ok := logs[sites[i].ID]; ok {
			sites[i].Status = logentry
		}
	}
	item, err = memcache.Get(c, "sites")
	if err != nil {
		log.Fatal(err)
	}
	sitesdata, _ := json.Marshal(sites)
	item.Value = sitesdata
	_ = memcache.Set(c, item)
}

func CacheSiteChange(r *http.Request, site WebsiteSave) {
	c := appengine.NewContext(r)
	var item *memcache.Item
	exists := false

	cacheSite := Website{
		ID:            site.ID,
		Name:          site.Name,
		URL:           site.URL,
		Interval:      site.Interval,
		Monitoring:    site.Monitoring,
		Public:        site.Public,
		EmailInterval: site.EmailInterval,
		LogDays:       site.LogDays,
	}
	cacheSite.Status, _ = history.GetStatus(r, cacheSite.ID)

	sites, err := GetAll(r)
	for i := 0; i < len(sites); i++ {
		if sites[i].ID == site.ID {
			sites[i] = cacheSite
			exists = true
		}
	}
	sitelist := sites
	if !exists {
		sitelist = append(sites, cacheSite)
		sort.Sort(ByName{sitelist})
	}

	if item, err = memcache.Get(c, "sites"); err == memcache.ErrCacheMiss {
		sitesdata, _ := json.Marshal(sitelist)
		item := &memcache.Item{
			Key:   "sites",
			Value: sitesdata,
		}
		_ = memcache.Add(c, item)
	} else {
		sitesdata, _ := json.Marshal(sitelist)
		item.Value = sitesdata
		_ = memcache.Set(c, item)
	}
}

func RemoveFromCache(r *http.Request, siteID int64) {
	c := appengine.NewContext(r)
	var item *memcache.Item

	target := -1
	sites, err := GetAll(r)
	for i := 0; i < len(sites); i++ {
		if sites[i].ID == siteID {
			target = i
		}
	}
	if target > -1 {
		sitelist := make([]Website, len(sites)-1)
		count := 0
		for i := 0; i < len(sites); i++ {
			if sites[i].ID != siteID {
				sitelist[count] = sites[i]
				count++
			}
		}
		// sort	
		sort.Sort(ByName{sitelist})

		if item, err = memcache.Get(c, "sites"); err == memcache.ErrCacheMiss {
			sitesdata, _ := json.Marshal(sitelist)
			item := &memcache.Item{
				Key:   "sites",
				Value: sitesdata,
			}
			_ = memcache.Add(c, item)
		} else {
			sitesdata, _ := json.Marshal(sitelist)
			item.Value = sitesdata
			_ = memcache.Set(c, item)
		}
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
	sites, err := GetAll(r)
	now := time.Now()
	logs := make(map[int64]history.History)
	if err != nil {
		log.Fatal(err)
	}
	if err == nil {
		for i := 0; i < len(sites); i++ {
			dur := time.Duration(sites[i].Interval) * time.Minute

			var status history.History
			if sites[i].Status.Checked.IsZero() {
				status, _ = history.GetStatus(r, sites[i].ID)
			} else {
				status = sites[i].Status
			}

			if now.Sub(status.Checked) >= dur {
				logs[sites[i].ID] = sites[i].Check(r)
			}
		}
	}
	history.SaveLogs(r, logs)
	CacheStatusChange(r, logs)

	return err
}

func (website Website) Get(r *http.Request, key int64) (site Website, sitesave WebsiteSave, err error) {

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
	}
	return site, sitesave, err
}

func (website Website) Delete(r *http.Request) (err error) {
	c := appengine.NewContext(r)
	var keynum int64
	keynum, _ = strconv.ParseInt(r.FormValue("key"), 10, 64)
	k := datastore.NewKey(c, "website", "", keynum, nil)
	err = datastore.Delete(c, k)
	RemoveFromCache(r, keynum)
	return
}

func (website Website) Save(r *http.Request) (err error) {
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

			CacheSiteChange(r, site)
			return err
		}, nil)
		return err
	} else {
		// update website
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			k := datastore.NewKey(c, "website", "", keynum, nil)
			_, site, err := Get(r, keynum)
			if err != nil {
				return err
			}
			site.ID = keynum
			site.Name = name
			site.URL = urlstr
			site.Interval = interval
			site.Monitoring = monitoring
			site.Public = public
			site.LogDays = logdays
			site.EmailInterval = emailInterval

			_, err = datastore.Put(c, k, &site)
			CacheSiteChange(r, site)
			return err
		}, nil)
		return err
	}
	return
}

func (website Website) GetNotifiers(r *http.Request) (notifiers []Notify, err error) {
	notifiers, err = notify.GetAllBySite(r, website.ID)
	return
}

func (website *Website) Check(r *http.Request) History {
	status, code, response := rest.Get(website.URL, r)
	prevStatus := website.Status.Status
	var err error
	if status {
		send := prevStatus == "down"
		website.Status = history.Log(r, website.ID, time.Now(), "up", send, code, response)
		if send {
			err := website.EmailNotifiers(r)
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		send := (prevStatus == "up") || (website.OkToSend(r))
		website.Status = history.Log(r, website.ID, time.Now(), "down", send, code, response)
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
			notifiers[i].Notify(r, website.Name, website.URL, website.Status.Checked, website.Status.Status, website.Status.Code, website.Status.ResponseTime)
		}
	}
	return
}

func (website Website) GetHistory(r *http.Request) (logs []HistoryGroup, err error) {
	logs, err = history.GetHistory(r, website.ID)
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

type ByName struct{ Websites []Website }

func (s ByName) Len() int      { return len(s.Websites) }
func (s ByName) Swap(i, j int) { s.Websites[i], s.Websites[j] = s.Websites[j], s.Websites[i] }

func (s ByName) Less(i, j int) bool { return s.Websites[i].Name < s.Websites[j].Name }
