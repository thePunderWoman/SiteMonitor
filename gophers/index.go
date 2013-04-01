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
	server.Get("/auth/:error", auth.Index)
	server.Get("/auth/out", auth.Logout)

	server.Get("/", controllers.Index)

	//Admin Routes
	server.Get("/Admin", admin.Index).Secure()
	server.Get("/Add", admin.Add).Secure()
	server.Get("/Add/:error", admin.Add).Secure()
	server.Get("/Edit/:key", admin.Edit).Secure()
	server.Get("/Edit/:key/:error", admin.Edit).Secure()
	server.Post("/Save", admin.Save).Secure()
	server.Post("/Delete", admin.Delete).Secure()
	server.Post("/DeleteNotifier", admin.DeleteNotifier).Secure()
	server.Post("/AddNotifier", admin.AddNotifier).Secure()
	server.Get("/TestSend/:parent/:key", admin.TestSend).Secure()
	server.Get("/Emails/:key", admin.GetNotifiers).Secure()
	server.Get("/Emails/:key/:error", admin.GetNotifiers).Secure()
	server.Get("/History/:key", admin.GetHistory).Secure()
	server.Get("/History/:key/:page/:perpage", admin.GetHistory).Secure()

	//Cron Task
	server.Get("/Check", admin.Check)
	server.Get("/CleanLogs", admin.CleanLogs)

	session_key := "your key here"
	http.Handle("/", server.NewSessionHandler(session_key, nil))
}
