package notify

import (
	"appengine"
	"appengine/datastore"
	"errors"
	"gophers/helpers/email"
	"gophers/helpers/serversettings"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

func Get(r *http.Request) (notify Notify, err error) {
	c := appengine.NewContext(r)
	params := r.URL.Query()
	var keynum int64
	keynum, _ = strconv.ParseInt(params.Get(":key"), 10, 64)
	parentkeynum, _ := strconv.ParseInt(params.Get(":parent"), 10, 64)
	parentKey := datastore.NewKey(c, "website", "", parentkeynum, nil)
	key := datastore.NewKey(c, "notify", "", keynum, parentKey)

	err = datastore.Get(c, key, &notify)

	return notify, err
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

func (notifier *Notify) Notify(r *http.Request, name string, url string, lastChecked time.Time, template string) {
	serversettings, err := serversettings.Get(r)
	if err != nil {
		log.Fatal(err)
	}
	settings := email.Settings{
		Server:   serversettings.Server,
		Email:    serversettings.Email,
		SSL:      serversettings.SSL,
		Username: serversettings.Username,
		Password: serversettings.Password,
		Port:     serversettings.Port,
	}
	layout := "Mon, 01/02/06, 3:04PM MST"
	Local, _ := time.LoadLocation("US/Central")
	localChecked := lastChecked.In(Local).Format(layout)

	var tos []string
	var subject string
	var body string

	if template == "up" {
		tos = []string{notifier.Email}
		subject = name + " Site Restored Notification"
		body = "<html><body><p>Hello " + notifier.Name + "</p><p>The " + name + " website at " + url + " is restored as of " + localChecked + ".</p></body></html>"
	} else {
		tos = []string{notifier.Email}
		subject = name + " Site Outage Warning!"
		body = "<html><body><p>Hello " + notifier.Name + "</p><p>The " + name + " website at " + url + " is down as of " + localChecked + ".</p></body></html>"
	}
	email.Send(settings, tos, subject, body, true)
}
