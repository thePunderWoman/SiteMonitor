package history

import (
	"appengine"
	"appengine/datastore"
	//"errors"
	"log"
	"math"
	"net/http"
	"time"
)

type History struct {
	ID      int64
	SiteID  int64
	Checked time.Time
	Status  string
	Emailed bool
	Percent float32
}

func GetHistory(r *http.Request, siteID int64, page int, perpage int) (logentries []History, entries int, err error) {
	c := appengine.NewContext(r)
	offset := (page - 1) * perpage
	q := datastore.NewQuery("history").Filter("SiteID =", siteID).Order("-Checked").Limit(perpage).Offset(offset)
	q2 := datastore.NewQuery("history").Filter("SiteID =", siteID)
	entries, _ = q2.Count(c)
	for t := q.Run(c); ; {
		var x History
		_, err := t.Next(&x)
		if err == datastore.Done || err != nil {
			break
		}
		logentries = append(logentries, x)
	}

	return logentries, entries, err
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

func Uptime(r *http.Request, siteID int64, status string) (uptime float32) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("history").Filter("SiteID =", siteID)
	qUp := datastore.NewQuery("history").Filter("SiteID =", siteID).Filter("Status =", "up")
	total, _ := q.Count(c)
	uptotal, _ := qUp.Count(c)

	// Take into account current log entry
	total += 1
	if status == "up" {
		uptotal += 1
	}
	
	// Run uptime calculation
	uptime = (float32(uptotal) / float32(total)) * 100.0
	// Check for errors
	if math.IsNaN(float64(uptime)) {
		uptime = 0
	}
	return uptime
}

func Log(r *http.Request, siteID int64, checked time.Time, status string, emailed bool) (logentry History) {
	logentry = History{
		SiteID:  siteID,
		Checked: checked,
		Status:  status,
		Emailed: emailed,
		Percent: Uptime(r, siteID, status),
	}
	return
}

func SaveLogs(r *http.Request, logs map[int64]History) {
	c := appengine.NewContext(r)
	toPut := make([]History, 0)
	keys := make([]*datastore.Key, 0)
	for siteID, logentry := range logs {
		parentKey := datastore.NewKey(c, "website", "", siteID, nil)
		newKey := datastore.NewIncompleteKey(c, "history", parentKey)
		keys = append(keys, newKey)
		toPut = append(toPut, logentry)
	}
	_, err := datastore.PutMulti(c, keys, toPut)
	if err != nil {
		log.Println(err)
	}
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
