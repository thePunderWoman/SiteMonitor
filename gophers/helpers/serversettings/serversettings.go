package serversettings

import (
	"appengine"
	"appengine/datastore"
	//"log"
	"net/http"
	"strconv"
)

type Setting struct {
	Server   string
	Email    string
	SSL      bool
	Username string
	Password string
	Port     int
}

func Get(r *http.Request) (setting *Setting, err error) {
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, "settings", "emailSetting", 0, nil)
	s := new(Setting)
	err = datastore.Get(c, key, s)

	return s, err
}

func Save(r *http.Request) (err error) {
	c := appengine.NewContext(r)

	server := r.FormValue("server")
	email := r.FormValue("email")
	username := r.FormValue("username")
	SSL, err := strconv.ParseBool(r.FormValue("ssl"))
	if err != nil {
		SSL = false
	}
	password := r.FormValue("password")
	port, err := strconv.Atoi(r.FormValue("port"))
	if err != nil {
		port = 0
	}

	key := datastore.NewKey(c, "settings", "emailSetting", 0, nil)

	// new Notify
	settings := Setting{
		Server:   server,
		Email:    email,
		SSL:      SSL,
		Username: username,
		Password: password,
		Port:     port,
	}

	_, err = datastore.Put(c, key, &settings)

	return err
}
