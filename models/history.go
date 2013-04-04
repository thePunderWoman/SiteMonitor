package models

import (
	"../helpers/database"
	//"errors"
	"log"
	"math"
	"time"
)

var (
	UTC, _ = time.LoadLocation("UTC")

	getSiteHistoryStmt = `select * from History
						where siteID = ?
						order by checked desc`

	getSiteHistoryUpCountStmt = `select COUNT(*) from History 
								where siteID = ? AND status = 'up'`

	getSiteHistoryDownCountStmt = `select COUNT(*) from History 
								where siteID = ? AND Status = 'down'`
	getSiteUptimeStmt = `select (select COUNT(*) FROM History WHERE status = 'up' AND siteID = ?) AS upcount, 
								(select COUNT(*) FROM History WHERE siteID = ?) AS total`

	getLastEmailStmt = `select * from History WHERE siteID = ? and emailed = 1 order by checked desc limit 1`
	insertLogStmt    = `insert into History (siteID,checked,status,emailed,code,responseTime) VALUES (?,CURRENT_TIMESTAMP,?,?,?,?)`
	clearOldLogsStmt = `delete from History WHERE siteID = ? and checked < ?`
)

type History struct {
	ID           int
	SiteID       int
	Checked      time.Time
	Status       string
	Emailed      bool
	Code         int
	ResponseTime float64
}

type HistoryGroup struct {
	Status string
	Start  time.Time
	End    time.Time
	Logs   []History
}

func GetHistory(siteID int) (loggroups []HistoryGroup, err error) {

	sel, err := database.Db.Prepare(getSiteHistoryStmt)
	if err != nil {
		return loggroups, err
	}

	params := struct {
		SiteID int
	}{}

	params.SiteID = siteID
	sel.Bind(&params)

	rows, res, err := sel.Exec()
	if database.MysqlError(err) {
		return loggroups, err
	}

	id := res.Map("id")
	sID := res.Map("siteID")
	checked := res.Map("checked")
	status := res.Map("status")
	emailed := res.Map("emailed")
	code := res.Map("code")
	responseTime := res.Map("responseTime")

	var logs []History
	for _, row := range rows {
		history := History{
			ID:           row.Int(id),
			SiteID:       row.Int(sID),
			Checked:      row.Time(checked, UTC),
			Status:       row.Str(status),
			Emailed:      row.Bool(emailed),
			Code:         row.Int(code),
			ResponseTime: row.Float(responseTime),
		}
		logs = append(logs, history)
	}

	var group HistoryGroup
	for i, entry := range logs {
		//if group != nil {
		if group.Logs != nil && len(group.Logs) > 0 {
			//logs exist
			// check if status matches
			if entry.Status == group.Status {
				// append to Logs list
				group.Logs = append(group.Logs, entry)
			} else {
				// status change
				group.Start = entry.Checked
				if i != 0 {
					group.Start = logs[i-1].Checked
				}
				loggroups = append(loggroups, group)
				group = HistoryGroup{
					Status: entry.Status,
					End:    entry.Checked,
					Logs:   make([]History, 0),
				}
				if i != 0 {
					group.End = logs[i-1].Checked
				}
				group.Logs = append(group.Logs, entry)
			}
		} else {
			// no logs
			group.Status = entry.Status
			group.End = entry.Checked
			group.Logs = make([]History, 0)
			group.Logs = append(group.Logs, entry)
		}

		if i == len(logs)-1 {
			group.Start = entry.Checked
			loggroups = append(loggroups, group)
		}

	}
	return loggroups, err
}

func GetStatus(siteID int) (history History, err error) {
	sel, err := database.Db.Prepare(getSiteHistoryStmt)
	if err != nil {
		return history, err
	}

	row, res, err := sel.ExecFirst()
	if database.MysqlError(err) {
		return history, err
	}

	id := res.Map("id")
	sID := res.Map("siteID")
	checked := res.Map("checked")
	status := res.Map("status")
	emailed := res.Map("emailed")
	code := res.Map("code")
	responseTime := res.Map("responseTime")

	if err != nil { // Must be something wrong with the db, lets bail
		return history, err
	} else if row != nil { // populate history object
		history = History{
			ID:           row.Int(id),
			SiteID:       row.Int(sID),
			Checked:      row.Time(checked, UTC),
			Status:       row.Str(status),
			Emailed:      row.Bool(emailed),
			Code:         row.Int(code),
			ResponseTime: row.Float(responseTime),
		}
	}
	return history, err
}

func GetLastEmail(siteID int) (logentry History, err error) {
	sel, err := database.Db.Prepare(getLastEmailStmt)
	if err != nil {
		return logentry, err
	}

	params := struct {
		SiteID *int
	}{}

	params.SiteID = &siteID

	sel.Bind(&params)

	row, res, err := sel.ExecFirst()
	if database.MysqlError(err) {
		return logentry, err
	}

	id := res.Map("id")
	sID := res.Map("siteID")
	checked := res.Map("checked")
	status := res.Map("status")
	emailed := res.Map("emailed")
	code := res.Map("code")
	responseTime := res.Map("responseTime")

	if err != nil { // Must be something wrong with the db, lets bail
		return logentry, err
	} else if row != nil { // populate history object
		logentry = History{
			ID:           row.Int(id),
			SiteID:       row.Int(sID),
			Checked:      row.Time(checked, UTC),
			Status:       row.Str(status),
			Emailed:      row.Bool(emailed),
			Code:         row.Int(code),
			ResponseTime: row.Float(responseTime),
		}
	}
	return logentry, err
}

func GetUptime(siteID int) (uptime float32) {
	uptime = 0
	sel, err := database.Db.Prepare(getSiteUptimeStmt)
	if err != nil {
		return uptime
	}

	params := struct {
		ID1 *int
		ID2 *int
	}{}

	params.ID1 = &siteID
	params.ID2 = &siteID

	sel.Bind(&params)

	row, res, err := sel.ExecFirst()
	if database.MysqlError(err) {
		return uptime
	}

	up := row.Int(res.Map("upcount"))
	total := row.Int(res.Map("total"))

	// Run uptime calculation
	uptime = (float32(up) / float32(total)) * 100.0
	// Check for errors
	if math.IsNaN(float64(uptime)) {
		uptime = 0
	}
	return uptime
}

func Log(siteID int, checked time.Time, status string, emailed bool, code int, resptime float64) (logentry History) {
	logentry = History{
		SiteID:       siteID,
		Checked:      checked,
		Status:       status,
		Emailed:      emailed,
		Code:         code,
		ResponseTime: resptime,
	}
	return
}

func SaveLogs(logs []History) {
	c := make(chan int)
	for _, logentry := range logs {
		go func(log History, ch chan int) {
			log.Save()
			ch <- 1
		}(logentry, c)
	}
}

func (entry *History) Save() {
	ins, err := database.Db.Prepare(insertLogStmt)
	if err != nil {
		log.Fatal(err)
	}

	params := struct {
		SiteID       int
		Status       string
		Emailed      bool
		Code         int
		ResponseTime float64
	}{}

	params.SiteID = entry.SiteID
	params.Status = entry.Status
	params.Emailed = entry.Emailed
	params.Code = entry.Code
	params.ResponseTime = entry.ResponseTime

	ins.Bind(&params)
	_, _, err = ins.Exec()
}

func ClearOld(siteID int, days int) {
	del, err := database.Db.Prepare(clearOldLogsStmt)
	if err != nil {
		log.Fatal(err)
	}
	deleteBefore := time.Now().AddDate(0, 0, -days)
	params := struct {
		SiteID  int
		Checked time.Time
	}{}

	params.SiteID = siteID
	params.Checked = deleteBefore

	del.Bind(&params)
	_, _, err = del.Exec()
	if err != nil {
		log.Fatal(err)
	}
}
