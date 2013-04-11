package models

import (
	"../helpers/database"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Notify struct {
	ID       int
	ParentID int
	Name     string
	Email    string
}

func (n Notify) GetAllBySite(parentID int) (notifiers []Notify, err error) {
	sel, err := database.GetStatement("getAllNotifiersStmt")
	if err != nil {
		return notifiers, err
	}

	params := struct {
		SiteID int
	}{}

	params.SiteID = parentID

	sel.Bind(&params)

	rows, res, err := sel.Exec()
	if database.MysqlError(err) {
		return notifiers, err
	}

	id := res.Map("id")
	sID := res.Map("siteID")
	name := res.Map("name")
	email := res.Map("email")

	for _, row := range rows {
		notifier := Notify{
			ID:       row.Int(id),
			ParentID: row.Int(sID),
			Name:     row.Str(name),
			Email:    row.Str(email),
		}
		notifiers = append(notifiers, notifier)
	}

	return notifiers, err
}

func (n Notify) Get(r *http.Request) (notify Notify, err error) {
	qparams := r.URL.Query()
	id, _ := strconv.Atoi(qparams.Get(":key"))
	sel, err := database.GetStatement("getNotifierByIDStmt")
	if err != nil {
		return notify, err
	}

	params := struct {
		ID int
	}{}

	params.ID = id

	sel.Bind(&params)

	row, res, err := sel.ExecFirst()
	if database.MysqlError(err) {
		return notify, err
	}

	idval := res.Map("id")
	sID := res.Map("siteID")
	name := res.Map("name")
	email := res.Map("email")

	notify = Notify{
		ID:       row.Int(idval),
		ParentID: row.Int(sID),
		Name:     row.Str(name),
		Email:    row.Str(email),
	}

	return notify, err
}

func (n Notify) Delete(r *http.Request) (err error) {
	id, _ := strconv.Atoi(r.FormValue("key"))
	del, err := database.GetStatement("deleteNotifierStmt")
	if err != nil {
		return err
	}

	params := struct {
		ID int
	}{}

	params.ID = id

	del.Bind(&params)

	_, _, err = del.Exec()
	if database.MysqlError(err) {
		return err
	}

	return
}

func (n Notify) Save(r *http.Request) (err error) {
	name := r.FormValue("name")
	email := r.FormValue("email")

	siteID, _ := strconv.Atoi(r.FormValue("parentID"))

	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" {
		err = errors.New("Name and Email are required.")
		return
	}

	ins, err := database.GetStatement("insertNotifierStmt")
	if err != nil {
		return err
	}

	// new Notify
	params := struct {
		SiteID int
		Name   string
		Email  string
	}{}

	params.SiteID = siteID
	params.Name = name
	params.Email = email

	ins.Bind(&params)

	_, _, err = ins.Exec()
	if database.MysqlError(err) {
		return err
	}

	return err

}

func (notifier *Notify) Notify(name string, url string, lastChecked time.Time, template string, code int, responseTime float64) {
	ssettings := Setting{}
	serversettings, err := ssettings.Get()
	if err != nil {
		log.Println(err)
		return
	}
	settings := Settings{
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
		body = "<html><body><p>Hello " + notifier.Name + "</p><p>The " + name + " website at " + url + " is down as of " + localChecked + ". It responded with an error code of " + strconv.Itoa(code) + " and a response time of " + strconv.FormatFloat(responseTime, 'g', -1, 64) + " ms.</p></body></html>"
	}
	Send(settings, tos, subject, body, true)
}
