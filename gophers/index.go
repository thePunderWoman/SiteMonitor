package main

import (
	"gophers/controllers"
	"gophers/controllers/admin"
	"gophers/controllers/auth"
	"gophers/plate"
	"net/http"
)

func init() {
	server := plate.NewServer("doughboy")
	plate.DefaultAuthHandler = auth.AuthHandler

	server.Get("/", controllers.Index)

	//Admin Routes
	server.Get("/Admin", admin.Index).Secure()

	session_key := "your key here"
	http.Handle("/", server.NewSessionHandler(session_key, nil))
}
