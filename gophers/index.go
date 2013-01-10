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

	//Auth Routes
	server.Get("/auth", auth.Index)
	server.Post("/auth", auth.Login)
	server.Get("/auth/out", auth.Logout)

	server.Get("/", controllers.Index)

	//Admin Routes
	server.Get("/Admin", admin.Index).Secure()

	session_key := "your key here"
	http.Handle("/", server.NewSessionHandler(session_key, nil))
}
