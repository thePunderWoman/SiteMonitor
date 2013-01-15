package notify

import (
/*"appengine"
"appengine/datastore"
"errors"
//"log"
"net/http"
"strconv"
"strings"
"time"*/
)

type Notify struct {
	ID       int64
	parentID int64
	Name     string
	Email    string
}

/*func GetAllBySite(r *http.Request) (sites map[int64]Website) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("website").Order("Name")

	//var sites []QueryResult
	sites = make(map[int64]Website, 0)
	for t := q.Run(c); ; {
		var x Website
		key, err := t.Next(&x)

		// Just had to switch this to check before you attempt to do an assignment
		if err == datastore.Done || err != nil {
			break
		}
		// Also, you can key that array using the IntID() function
		// of a *Key property. This will return and int64
		// and you can use this value later to quiery the
		// datastore.
		sites[key.IntID()] = x

	}
	return
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

}*/
