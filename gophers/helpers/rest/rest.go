package rest

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"net/http"
)

func Get(url string, r *http.Request) (status bool) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	t := &urlfetch.Transport{Context: appengine.NewContext(r)}

	trip, err := t.RoundTrip(req)
	if err != nil {
		return
	}

	defer trip.Body.Close()

	var buf bytes.Buffer
	buf.ReadFrom(trip.Body)

	status = false
	if buf.Len() > 0 {
		status = true
	}
	return
}
