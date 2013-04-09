package auth

import (
	"../../helpers/plate"
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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
	Comments  string
	IsActive  int
	SuperUser int
	IsDealer  int
	Photo     string
}

func AuthHandler(w http.ResponseWriter, r *http.Request) bool {
	cook, err := r.Cookie("user")
	userID := 0
	if err == nil && cook != nil {
		userID, err = strconv.Atoi(cook.Value)
		if err != nil {
			userID = 0
		}
	}
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
	error, _ = url.QueryUnescape(error)
	server := plate.NewServer()

	tmpl, err = server.Template(w)

	if err != nil {
		plate.Serve404(w, err.Error())
		return
	}

	tmpl.Bag["Error"] = strings.ToTitle(error)
	tmpl.Template = "templates/auth/in.html"
	tmpl.DisplayTemplate()
}

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	cust, err := LoadUser(username, password, r, w)

	if err != nil || cust == nil {
		urlpath := "/auth/" + url.QueryEscape("Failed to log you into the system")
		http.Redirect(w, r, urlpath, http.StatusFound)
	} else {
		cook := http.Cookie{
			Name:    "user",
			Value:   strconv.Itoa(cust.UserID),
			Expires: time.Now().AddDate(0, 1, 0),
		}
		http.SetCookie(w, &cook)
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// expire cookie
	cook, err := r.Cookie("user")

	if err == nil {
		cook.Expires = time.Now().AddDate(0, 0, -1)
		http.SetCookie(w, cook)
	}
	http.Redirect(w, r, "/auth", http.StatusFound)
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

	t, err := http.PostForm("https://api.curtmfg.com/user/getuser", values)

	if err != nil {
		return
	}
	if t.StatusCode != 200 {
		return
	}

	// Set up close of body data - save memory
	defer t.Body.Close()

	// Parse our response data int a buffer 
	// then unmarshal the json into our Customer struct
	var buf bytes.Buffer
	buf.ReadFrom(t.Body)
	err = json.Unmarshal(buf.Bytes(), &c)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
