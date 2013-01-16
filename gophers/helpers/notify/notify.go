package notify

import (
	"appengine"
	"appengine/datastore"
	"errors"
	//"gophers/helpers/serversettings"
	"gophers/helpers/website"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Notify struct {
	ID       int64
	ParentID int64
	Name     string
	Email    string
}

func GetAllBySite(r *http.Request, parentID int64) (notifiers []Notify, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("notify").Filter("ParentID =", parentID).Order("Name")

	notifiers = make([]Notify, 0)
	_, err = q.GetAll(c, &notifiers)

	return notifiers, err
}

func Delete(r *http.Request) (err error) {
	c := appengine.NewContext(r)
	var keynum, parentkeynum int64
	keynum, _ = strconv.ParseInt(r.FormValue("key"), 10, 64)
	parentkeynum, _ = strconv.ParseInt(r.FormValue("parent"), 10, 64)
	parentKey := datastore.NewKey(c, "website", "", parentkeynum, nil)
	k := datastore.NewKey(c, "notify", "", keynum, parentKey)
	err = datastore.Delete(c, k)
	return
}

func Save(r *http.Request) (err error) {
	c := appengine.NewContext(r)

	name := r.FormValue("name")
	email := r.FormValue("email")

	var keynum int64
	keynum, _ = strconv.ParseInt(r.FormValue("parentID"), 10, 64)
	parentKey := datastore.NewKey(c, "website", "", keynum, nil)

	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" {
		err = errors.New("Name and Email are required.")
		return
	}

	// new Notify
	notifiee := Notify{
		Name:     name,
		Email:    email,
		ParentID: keynum,
	}

	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "notify", parentKey), &notifiee)

	if err == nil {
		notifiee.ID = key.IntID()
		key, err = datastore.Put(c, key, &notifiee)
	}

	return err

}

func (notifier *Notify) Notify(r *http.Request, w *website.Website) {
	// email person that site is down
	log.Println("Emailed " + notifier.Name + " that " + w.Name + " is down.")
}
