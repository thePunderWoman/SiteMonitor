package auth

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"encoding/json"
	"gophers/plate"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Customer struct {
	UserID    int
	Username  string
	Email     string
	Fname     string
	Lname     string
	Website   string
	Phone     string
	Fax       string
	IsAdmin   int
	Comments  string
	IsActive  int
	SuperUser int
	IsDealer  int
	Photo     string
}

func AuthHandler(w http.ResponseWriter, r *http.Request) bool {
	session := plate.Session.Get(r)

	userID, _ := session["user"].(int)
	if userID == 0 {
		http.Redirect(w, r, "/Auth", http.StatusFound)
	}

	return true
}

func Index(w http.ResponseWriter, r *http.Request) {
	var err error
	var tmpl plate.Template

	params := r.URL.Query()
	error := params.Get(":error")
	server := plate.NewServer()

	tmpl, err = server.Template(w)

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Message"] = strings.ToTitle(error)
	tmpl.Layout = "templates/admin/layout.html"
	tmpl.Template = "templates/auth/in.html"
	tmpl.DisplayTemplate()
}

func Login(w http.ResponseWriter, r *http.Request) {
	session := plate.Session.Get(r)

	username := r.FormValue("username")
	password := r.FormValue("password")
	cust, err := LoadUser(username, password, r, w)

	if err != nil || cust == nil {
		log.Println("hit error")
		http.Redirect(w, r, "/login/Failed to log you into the system", http.StatusFound)
	} else {
		session["user"] = cust.UserID
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session := plate.Session.Get(r)
	delete(session, "user")
	http.Redirect(w, r, "/login", http.StatusFound)
}

func LoadUser(u string, p string, r *http.Request, w http.ResponseWriter) (c *Customer, err error) {
	// send off post request to http://api.curtmfg.com/User/GetUser
	// with the following params:
	// username: username or email for a user
	// password: password of a user
	// API key: key for the internal account

	// Create post data
	values := make(url.Values)
	values.Set("username", u)
	values.Set("password", p)
	values.Set("key", "8aee0620-412e-47fc-900a-947820ea1c1d")

	// Encode post data and make request
	b := strings.NewReader(values.Encode())
	req, _ := http.NewRequest("POST", "https://api.curtmfg.com/user/getuser", b)

	// Set up Tansport using app engines urlfetch service
	t := &urlfetch.Transport{Context: appengine.NewContext(r)}

	// Roundtrip our request to the api and handle response
	r2, err := t.RoundTrip(req)
	if err != nil {
		return
	}
	if r2.StatusCode != 200 {
		return
	}

	// Set up close of body data - save memory
	defer r2.Body.Close()

	// Parse our response data int a buffer 
	// then unmarshal the json into our Customer struct
	var buf bytes.Buffer
	buf.ReadFrom(r2.Body)

	err = json.Unmarshal(buf.Bytes(), &c)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
