package models

import (
	//"log"
	"../helpers/database"
	"net/http"
	"strconv"
)

var (
	getSettingsStmt = `select * from Setting
							limit 1`

	getSettingsIDStmt = `select id from Setting limit 1`

	insertSettingsStmt = `insert into Setting (server,email,SSL,username,password,port) VALUES (?,?,?,?,?,?)`
	updateSettingsStmt = `update Setting set server = ?, email = ?, SSL = ?, username = ?, password = ?, port = ? WHERE id = ?`
)

type Setting struct {
	id       int
	Server   string
	Email    string
	SSL      bool
	Username string
	Password string
	Port     int
}

func (s Setting) Get() (setting Setting, err error) {
	qry, err := database.Db.Prepare(getSettingsStmt)
	if err != nil {
		return setting, err
	}

	row, res, err := qry.ExecFirst()
	if database.MysqlError(err) {
		return setting, err
	} else if row == nil {
		return setting, nil
	}

	ID := res.Map("id")
	server := res.Map("server")
	email := res.Map("email")
	ssl := res.Map("SSL")
	username := res.Map("username")
	password := res.Map("password")
	port := res.Map("port")

	setting = Setting{
		id:       row.Int(ID),
		Server:   row.Str(server),
		Email:    row.Str(email),
		SSL:      row.Bool(ssl),
		Username: row.Str(username),
		Password: row.Str(password),
		Port:     row.Int(port),
	}

	return setting, err
}

func (s Setting) Save(r *http.Request) (err error) {

	// check if there's a row already
	qry, err := database.Db.Prepare(getSettingsIDStmt)
	if err != nil {
		return err
	}

	row, res, err := qry.ExecFirst()
	if database.MysqlError(err) {
		return err
	} else if row == nil {
		return nil
	}

	settingID := row.Int(res.Map("id"))

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

	if settingID == 0 {
		// check if there's a row already
		ins, err := database.Db.Prepare(insertSettingsStmt)
		if err != nil {
			return err
		}

		params := struct {
			Server   *string
			Email    *string
			SSL      *bool
			Username *string
			Password *string
			Port     *int
		}{}

		params.Server = &server
		params.Email = &email
		params.SSL = &SSL
		params.Username = &username
		params.Password = &password
		params.Port = &port

		ins.Bind(&params)

		_, _, err = ins.Exec()
	} else {
		// check if there's a row already
		upd, err := database.Db.Prepare(updateSettingsStmt)
		if err != nil {
			return err
		}

		params := struct {
			Server   *string
			Email    *string
			SSL      *bool
			Username *string
			Password *string
			Port     *int
			ID       *int
		}{}

		params.Server = &server
		params.Email = &email
		params.SSL = &SSL
		params.Username = &username
		params.Password = &password
		params.Port = &port
		params.ID = &settingID

		upd.Bind(&params)

		_, _, err = upd.Exec()
	}

	return err
}
