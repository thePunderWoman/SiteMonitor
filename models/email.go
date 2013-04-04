package models

import (
	"errors"
	"log"
	"net/smtp"
	"strconv"
)

type Settings struct {
	Server   string
	Email    string
	SSL      bool
	Username string
	Password string
	Port     int
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

func Send(settings Settings, tos []string, subject string, body string, html bool) {
	fullserver := settings.Server + ":" + strconv.Itoa(settings.Port)
	mimetype := "text/plain"
	if html {
		mimetype = "text/html"
	}
	mime := "MIME-version: 1.0;\nContent-Type: " + mimetype + "; charset=\"UTF-8\";\n\n"
	subject = "Subject: " + subject + "\n"
	msg := []byte(subject + mime + body)

	// Set up authentication information.
	auth := PlainAuth(
		"",
		settings.Username,
		settings.Password,
		settings.Server,
	)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		fullserver,
		auth,
		settings.Email,
		tos,
		msg,
	)
	if err != nil {
		log.Println(err)
	}
}

/*func Send(r *http.Request, tos []string, subject string, body string, html bool) {
	c := appengine.NewContext(r)

	msg := &mail.Message{
		Sender:  "CURT Site Monitor <status@curtmfg.com>",
		ReplyTo: "CURT Site Monitor <websupport@curtmfg.com>",
		To:      tos,
		Subject: subject,
	}
	if html {
		msg.HTMLBody = body
	} else {
		msg.Body = body
	}

	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Couldn't send email: %v", err)
	}

}*/
