package rest

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"net/http"
	"time"
)

func Get(url string, r *http.Request) (status bool, code int, response float64) {

	status = false
	started := time.Now()
	req, err := http.NewRequest("GET", url, nil)
	code = 500
	if err != nil {
		response = float64(time.Now().Sub(started).Nanoseconds()) / float64(1000000)
		return status, code, response
	}

	t := &urlfetch.Transport{Context: appengine.NewContext(r)}

	trip, err := t.RoundTrip(req)
	response = float64(time.Now().Sub(started).Nanoseconds()) / float64(1000000)
	code = trip.StatusCode
	if err != nil || trip.StatusCode != 200 {
		return status, code, response
	}

	defer trip.Body.Close()

	var buf bytes.Buffer
	buf.ReadFrom(trip.Body)

	if buf.Len() > 0 {
		status = true
	} else {
		code = 500
	}
	return status, code, response
}
