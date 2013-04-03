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
	code = 500
	retries := 5
	i := 0
	var err error
	var trip *http.Response

	for i < retries {
		started = time.Now()
		trip, err = RunRequest(url, r)
		if err != nil || trip.StatusCode != 200 {
			i += 1
		} else {
			i = 5
		}
	}
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

func RunRequest(url string, r *http.Request) (t *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return t, err
	}

	tpt := &urlfetch.Transport{Context: appengine.NewContext(r)}
	t, err = tpt.RoundTrip(req)
	return

}
