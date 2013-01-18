package history

import (
	"appengine"
	"appengine/datastore"
	//"errors"
	"log"
	"net/http"
	"time"
)

type History struct {
	ID      int64
	SiteID  int64
	Checked time.Time
	Status  string
}

func GetHistory(r *http.Request, siteID int64, page int, perpage int) (logentries []History, err error) {
	c := appengine.NewContext(r)
	offset := (page - 1) * perpage
	q := datastore.NewQuery("history").Filter("SiteID =", siteID).Order("-Checked").Limit(perpage).Offset(offset)

	for t := q.Run(c); ; {
		var x History
		_, err := t.Next(&x)
		if err == datastore.Done || err != nil {
			break
		}
		logentries = append(logentries, x)
	}

	return logentries, err
}

func GetStatus(r *http.Request, siteID int64) (status History, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("history").Filter("SiteID =", siteID).Order("-Checked").Limit(1)
	for t := q.Run(c); ; {
		_, err := t.Next(&status)
		if err == datastore.Done || err != nil {
			break
		}
	}
	return status, err
}

func Log(r *http.Request, siteID int64, checked time.Time, status string) (logentry History, err error) {
	c := appengine.NewContext(r)

	parentKey := datastore.NewKey(c, "website", "", siteID, nil)

	// new Notify
	logentry = History{
		SiteID:  siteID,
		Checked: checked,
		Status:  status,
	}

	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "history", parentKey), &logentry)

	if err == nil {
		logentry.ID = key.IntID()
		key, err = datastore.Put(c, key, &logentry)
	}

	return logentry, err
}

func ClearOld(r *http.Request, siteID int64, days int) {
	c := appengine.NewContext(r)
	//parentKey := datastore.NewKey(c, "website", "", siteID, nil)
	deleteBefore := time.Now().AddDate(0, 0, -days)
	q := datastore.NewQuery("history").Filter("SiteID =", siteID).Filter("Checked <", deleteBefore).KeysOnly()
	keys, err := q.GetAll(c, nil)
	err = datastore.DeleteMulti(c, keys)
	if err != nil {
		log.Println(err)
	}
}
