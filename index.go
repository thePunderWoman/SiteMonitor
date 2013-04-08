package main

import (
	"./controllers"
	"./controllers/admin"
	"./controllers/auth"
	"./helpers/plate"
	"flag"
	"log"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":8080", "http listen address")

	CorsHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		return
	}
)

const (
	port = "80"
)

func main() {
	log.Println("Initializing application")
	flag.Parse()
	server := plate.NewServer("doughboy")
	plate.DefaultAuthHandler = auth.AuthHandler

	server.AddFilter(CorsHandler)

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
	server.Get("/TestSend/:key", admin.TestSend).Secure()
	server.Get("/Emails/:key", admin.GetNotifiers).Secure()
	server.Get("/Emails/:key/:error", admin.GetNotifiers).Secure()
	server.Get("/History/:key", admin.GetHistory).Secure()

	//Setting Routes
	server.Get("/Settings", admin.Settings).Secure()
	server.Get("/Settings/:error", admin.Settings).Secure()
	server.Post("/Settings", admin.SaveSettings).Secure()

	//Cron Task
	server.Get("/Check", admin.Check)
	server.Get("/CleanLogs", admin.CleanLogs)

	server.Static("/", "static")

	http.Handle("/", server)

	log.Println("Server running on port " + *listenAddr)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
