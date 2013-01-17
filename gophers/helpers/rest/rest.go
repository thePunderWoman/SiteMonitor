package rest

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"net/http"
)

func Get(url string, r *http.Request) (status bool) {

	status = false
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return status
	}

	t := &urlfetch.Transport{Context: appengine.NewContext(r)}

	trip, err := t.RoundTrip(req)
	if err != nil || trip.StatusCode != 200 {
		return status
	}

	defer trip.Body.Close()

	var buf bytes.Buffer
	buf.ReadFrom(trip.Body)

	if buf.Len() > 0 {
		status = true
	}
	return
}
