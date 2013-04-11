package models

import (
	"../helpers/database"
	"../helpers/rest"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Website struct {
	ID            int
	Name          string
	URL           string
	Interval      int
	Monitoring    bool
	Status        History
	Public        bool
	EmailInterval int
	Uptime        float32
	LogDays       int
}

func (website Website) IntervalMins() int {
	return website.Interval * website.EmailInterval
}

func (website Website) GetAll() (sites []Website, err error) {
	sel, err := database.GetStatement("getAllWebsitesStmt")
	if err != nil {
		return sites, err
	}

	rows, res, err := sel.Exec()
	if database.MysqlError(err) {
		return sites, err
	}

	if len(rows) > 0 {
		id := res.Map("id")
		name := res.Map("name")
		urlstr := res.Map("URL")
		interval := res.Map("checkInterval")
		monitoring := res.Map("monitoring")
		public := res.Map("public")
		emailInverval := res.Map("emailInterval")
		logDays := res.Map("logDays")

		for _, row := range rows {
			status, _ := GetStatus(row.Int(id))
			site := Website{
				ID:            row.Int(id),
				Name:          row.Str(name),
				URL:           row.Str(urlstr),
				Interval:      row.Int(interval),
				Monitoring:    row.Bool(monitoring),
				Public:        row.Bool(public),
				EmailInterval: row.Int(emailInverval),
				LogDays:       row.Int(logDays),
				Uptime:        GetUptime(row.Int(id)),
				Status:        status,
			}
			sites = append(sites, site)
		}
	}

	return sites, err
}

func CleanLogs() {
	sel, err := database.GetStatement("getAllMonitoringWebsitesStmt")
	if err != nil {
		log.Println(err)
		return
	}

	rows, res, err := sel.Exec()
	if database.MysqlError(err) {
		log.Println(err)
		return
	}

	id := res.Map("id")
	logDays := res.Map("logDays")

	var sites []Website
	for _, row := range rows {
		site := Website{
			ID:      row.Int(id),
			LogDays: row.Int(logDays),
		}
		sites = append(sites, site)
	}

	if err == nil {
		for _, site := range sites {
			ClearOld(site.ID, site.LogDays)
		}
	}
}

func CheckSites(r *http.Request) (err error) {
	s := Website{}
	sites, err := s.GetAll()
	now := time.Now()
	var logs []History
	if err != nil {
		log.Println(err)
		return err
	}
	if err == nil {
		for i := 0; i < len(sites); i++ {
			dur := time.Duration(sites[i].Interval) * time.Minute

			var status History
			if sites[i].Status.Checked.IsZero() {
				status, _ = GetStatus(sites[i].ID)
			} else {
				status = sites[i].Status
			}

			if now.Sub(status.Checked) >= dur {
				logs = append(logs, sites[i].Check(r))
			}
		}
	}
	SaveLogs(logs)
	//CacheStatusChange(r, logs)

	return err
}

func (website Website) Get(id int) (site Website, err error) {
	sel, err := database.GetStatement("getWebsiteByIDStmt")
	if err != nil {
		return site, err
	}

	params := struct {
		ID int
	}{}

	params.ID = id

	sel.Bind(&params)

	row, res, err := sel.ExecFirst()
	if database.MysqlError(err) {
		return site, err
	}

	idval := res.Map("id")
	name := res.Map("name")
	url := res.Map("URL")
	interval := res.Map("checkInterval")
	monitoring := res.Map("monitoring")
	public := res.Map("public")
	emailInterval := res.Map("emailInterval")
	logDays := res.Map("logDays")

	status, _ := GetStatus(row.Int(idval))
	site = Website{
		ID:            row.Int(idval),
		Name:          row.Str(name),
		URL:           row.Str(url),
		Interval:      row.Int(interval),
		Monitoring:    row.Bool(monitoring),
		Public:        row.Bool(public),
		EmailInterval: row.Int(emailInterval),
		LogDays:       row.Int(logDays),
		Uptime:        GetUptime(row.Int(idval)),
		Status:        status,
	}

	return site, err
}

func (website Website) Delete(r *http.Request) (err error) {
	siteID, _ := strconv.Atoi(r.FormValue("key"))
	params := struct {
		SiteID int
	}{}
	params.SiteID = siteID

	del, err := database.GetStatement("deleteNotifierStmt")
	if err != nil {
		return err
	}

	del.Bind(&params)
	_, _, err = del.Exec()
	if err != nil {
		return err
	}

	del, err = database.GetStatement("deleteHistoryBySiteStmt")
	if err != nil {
		return err
	}

	del.Bind(&params)
	_, _, err = del.Exec()
	if err != nil {
		return err
	}

	del, err = database.GetStatement("deleteWebsiteByIDStmt")
	if err != nil {
		return err
	}

	del.Bind(&params)
	_, _, err = del.Exec()
	if err != nil {
		return err
	}

	return
}

func (website Website) Save(r *http.Request) (err error) {
	name := r.FormValue("name")
	urlstr := r.FormValue("url")
	interval, err := strconv.Atoi(r.FormValue("interval"))
	emailInterval, err := strconv.Atoi(r.FormValue("emailinterval"))
	logdays, err := strconv.Atoi(r.FormValue("logdays"))
	if err != nil || logdays < 1 {
		logdays = 1
	}
	var monitoring bool
	var public bool
	if r.FormValue("monitoring") == "" {
		monitoring = false
	} else {
		monitoring = true
	}
	if r.FormValue("public") == "" {
		public = false
	} else {
		public = true
	}

	if strings.TrimSpace(name) == "" || strings.TrimSpace(urlstr) == "" || err != nil || interval < 5 || logdays < 1 {
		err = errors.New("Name and URL are required. Interval must be an integer greater than 5. Log Days kept must be greater than 1.")
		return
	}

	siteID, err := strconv.Atoi(r.FormValue("siteID"))

	if err != nil {
		// new Website
		params := struct {
			Name          string
			URL           string
			Interval      int
			Monitoring    bool
			Public        bool
			EmailInterval int
			LogDays       int
		}{}

		params.Name = name
		params.URL = urlstr
		params.Interval = interval
		params.Monitoring = monitoring
		params.Public = public
		params.EmailInterval = emailInterval
		params.LogDays = logdays

		ins, err := database.GetStatement("insertWebsiteStmt")

		if err != nil {
			return err
		}

		ins.Bind(&params)
		_, _, err = ins.Exec()

		return err
	} else {
		// new Website
		params := struct {
			Name          string
			URL           string
			Interval      int
			Monitoring    bool
			Public        bool
			EmailInterval int
			LogDays       int
			SiteID        int
		}{}

		params.Name = name
		params.URL = urlstr
		params.Interval = interval
		params.Monitoring = monitoring
		params.Public = public
		params.EmailInterval = emailInterval
		params.LogDays = logdays
		params.SiteID = siteID

		upd, err := database.GetStatement("updateWebsiteStmt")
		if err != nil {
			return err
		}

		upd.Bind(&params)
		_, _, err = upd.Exec()
		return err
	}
	return
}

func (website Website) GetNotifiers(r *http.Request) (notifiers []Notify, err error) {
	n := Notify{}
	notifiers, err = n.GetAllBySite(website.ID)
	return
}

func (website *Website) Check(r *http.Request) History {
	status, code, response := rest.Get(website.URL, r)
	prevStatus := website.Status.Status
	var err error
	if status {
		send := prevStatus == "down"
		website.Status = Log(website.ID, time.Now(), "up", send, code, response)
		if send {
			err := website.EmailNotifiers(r)
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		send := (prevStatus == "up") || (website.OkToSend(r))
		website.Status = Log(website.ID, time.Now(), "down", send, code, response)
		if send {
			err = website.EmailNotifiers(r)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return website.Status
}

func (website Website) EmailNotifiers(r *http.Request) (err error) {
	notifiers, err := website.GetNotifiers(r)
	if err == nil {
		for i := 0; i < len(notifiers); i++ {
			notifiers[i].Notify(website.Name, website.URL, website.Status.Checked, website.Status.Status, website.Status.Code, website.Status.ResponseTime)
		}
	}
	return
}

func (website Website) GetHistory(r *http.Request) (logs []HistoryGroup, err error) {
	logs, err = GetHistory(website.ID)
	return
}

func (website Website) OkToSend(r *http.Request) bool {
	lastChange, err := GetLastEmail(website.ID)
	if err != nil {
		return true
	}
	sinceLast := time.Now().Sub(lastChange.Checked).Minutes()
	dur := (time.Duration(website.Interval*website.EmailInterval) * time.Minute).Minutes()
	return sinceLast > dur
}

type ByName struct{ Websites []Website }

func (s ByName) Len() int      { return len(s.Websites) }
func (s ByName) Swap(i, j int) { s.Websites[i], s.Websites[j] = s.Websites[j], s.Websites[i] }

func (s ByName) Less(i, j int) bool { return s.Websites[i].Name < s.Websites[j].Name }
