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
	Emailed bool
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

func GetLastEmail(r *http.Request, siteID int64) (logentry History, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("history").Filter("SiteID =", siteID).Filter("Emailed =", true).Order("-Checked").Limit(1)
	for t := q.Run(c); ; {
		_, err := t.Next(&logentry)
		if err == datastore.Done || err != nil {
			break
		}
	}
	return
}

func Uptime(r *http.Request, siteID int64) (uptime float32) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("history").Filter("SiteID =", siteID)
	qUp := datastore.NewQuery("history").Filter("SiteID =", siteID).Filter("Status =", "up")
	total, _ := q.Count(c)
	uptotal, _ := qUp.Count(c)
	uptime = (float32(uptotal) / float32(total)) * 100.0
	return uptime
}

func Log(r *http.Request, siteID int64, checked time.Time, status string, emailed bool) (logentry History, err error) {
	c := appengine.NewContext(r)

	parentKey := datastore.NewKey(c, "website", "", siteID, nil)

	// new Notify
	logentry = History{
		SiteID:  siteID,
		Checked: checked,
		Status:  status,
		Emailed: emailed,
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
