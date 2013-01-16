package notify

import (
	"appengine"
	"appengine/datastore"
	//"bytes"
	//"crypto/tls"
	"errors"
	"gophers/helpers/serversettings"
	"log"
	//"net"
	"net/http"
	"net/smtp"
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

type plainAuth struct {
	identity, username, password string
	host                         string
}

func PlainAuth(identity, username, password, host string) smtp.Auth {
	return &plainAuth{identity, username, password, host}
}

func (a *plainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
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

func (notifier *Notify) Notify(r *http.Request, name string, url string, lastChecked time.Time) {
	// email person that site is down
	//log.Println("Emailed " + notifier.Name + " that " + name + " is down.")
	settings, err := serversettings.Get(r)
	fullserver := settings.Server + ":" + strconv.Itoa(settings.Port)

	if err != nil {
		log.Fatal(err)
	}
	layout := "Mon, 01/02/06, 3:04PM MST"
	Local, _ := time.LoadLocation("US/Central")
	localChecked := lastChecked.In(Local).Format(layout)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + name + " Site Outage Warning!\n"
	msg := []byte(subject + mime + "<html><body><p>Hello " + notifier.Name + "</p><p>The " + name + " website at " + url + " is down as of " + localChecked + ".</p></body></html>")

	// Set up authentication information.
	auth := PlainAuth(
		"",
		settings.Username,
		settings.Password,
		settings.Server,
	)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err = smtp.SendMail(
		fullserver,
		auth,
		settings.Email,
		[]string{notifier.Email},
		msg,
	)
	if err != nil {
		log.Println(err)
	}
}
