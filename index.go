package main

import (
	"./controllers"
	"./controllers/admin"
	"./controllers/auth"
	"./helpers/database"
	"./helpers/globals"
	"./helpers/plate"
	"log"
	"net/http"
)

var (
	CorsHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		return
	}
	AuthHandler = func(w http.ResponseWriter, r *http.Request) {
		auth.AuthHandler(w, r)
		return
	}
)

const (
	port = "80"
)

func main() {
	err := database.PrepareAll()
	if err != nil {
		log.Fatal(err)
	}

	globals.SetGlobals()
	server := plate.NewServer("doughboy")

	server.AddFilter(CorsHandler)

	//Auth Routes
	server.Get("/auth", auth.Index)
	server.Post("/auth", auth.Login)
	server.Get("/auth/:error", auth.Index)
	server.Get("/logout", auth.Logout)

	server.Get("/", controllers.Index)

	//Admin Routes
	server.Get("/Admin", admin.Index).AddFilter(AuthHandler)
	server.Get("/Add", admin.Add).AddFilter(AuthHandler)
	server.Get("/Add/:error", admin.Add).AddFilter(AuthHandler)
	server.Get("/Edit/:key", admin.Edit).AddFilter(AuthHandler)
	server.Get("/Edit/:key/:error", admin.Edit).AddFilter(AuthHandler)
	server.Post("/Save", admin.Save).AddFilter(AuthHandler)
	server.Post("/Delete", admin.Delete).AddFilter(AuthHandler)
	server.Post("/DeleteNotifier", admin.DeleteNotifier).AddFilter(AuthHandler)
	server.Post("/AddNotifier", admin.AddNotifier).AddFilter(AuthHandler)
	server.Get("/TestSend/:key", admin.TestSend).AddFilter(AuthHandler)
	server.Get("/Emails/:key", admin.GetNotifiers).AddFilter(AuthHandler)
	server.Get("/Emails/:key/:error", admin.GetNotifiers).AddFilter(AuthHandler)
	server.Get("/History/:key", admin.GetHistory).AddFilter(AuthHandler)

	//Setting Routes
	server.Get("/Settings", admin.Settings).AddFilter(AuthHandler)
	server.Get("/Settings/:error", admin.Settings).AddFilter(AuthHandler)
	server.Post("/Settings", admin.SaveSettings).AddFilter(AuthHandler)

	//Cron Task
	server.Get("/Check", admin.Check)
	server.Get("/CleanLogs", admin.CleanLogs)

	server.Static("/", *globals.Filepath+"static")

	http.Handle("/", server)

	log.Println("Server running on port " + *globals.ListenAddr)

	log.Fatal(http.ListenAndServe(*globals.ListenAddr, nil))
}
