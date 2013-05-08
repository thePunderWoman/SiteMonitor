package database

import (
	"errors"
	"expvar"
	"github.com/ziutek/mymysql/autorc"
	_ "github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/thrsafe"
)

// History prepared statements
var (
	getSiteHistoryStmt = "select * from History where siteID = ? order by checked desc"

	getSiteHistoryUpCountStmt = `select COUNT(*) from History 
								where siteID = ? AND status = 'up'`

	getSiteHistoryDownCountStmt = `select COUNT(*) from History 
								where siteID = ? AND Status = 'down'`
	getSiteUptimeStmt = `select (select COUNT(*) FROM History WHERE status = 'up' AND siteID = ?) AS upcount, 
								(select COUNT(*) FROM History WHERE siteID = ?) AS total`

	getLastEmailStmt = `select * from History WHERE siteID = ? and emailed = 1 order by checked desc limit 1`
	insertLogStmt    = `insert into History (siteID,checked,status,emailed,code,responseTime) VALUES (?,UTC_TIMESTAMP(),?,?,?,?)`
	clearOldLogsStmt = `delete from History WHERE siteID = ? and checked < ?`
)

// Notify Prepared Statements
var (
	getAllNotifiersStmt = `select * from Notify
						where siteID = ?
						order by name`
	getNotifierByIDStmt = `select * from Notify
						   where id = ?`
	deleteNotifierStmt = `delete from Notify where id = ?`
	insertNotifierStmt = `insert into Notify (siteID,name,email) VALUES (?,?,?)`
)

// Settings Prepared Statements
var (
	getSettingsStmt = `select * from Setting
							limit 1`

	getSettingsIDStmt = `select id from Setting limit 1`

	insertSettingsStmt = `insert into Setting (server,email,requireSSL,username,password,port) VALUES (?,?,?,?,?,?)`
	updateSettingsStmt = `update Setting set server = ?, email = ?, requireSSL = ?, username = ?, password = ?, port = ? WHERE id = ?`
)

// Website Prepared Statements
var (
	getAllWebsitesStmt           = "select * from Website order by name"
	getAllMonitoringWebsitesStmt = "select * from Website where monitoring = 1"
	getWebsiteByIDStmt           = "select * from Website where id = ?"
	deleteWebsiteByIDStmt        = "delete from Website where id = ?"
	deleteHistoryBySiteStmt      = "delete from History where siteID = ?"
	deleteNotifiersBySiteStmt    = "delete from Notify where siteID = ?"
	insertWebsiteStmt            = "INSERT INTO Website (name, URL, checkinterval, monitoring, public, emailInterval, logDays) VALUES (?,?,?,?,?,?,?)"
	updateWebsiteStmt            = "update Website set name = ?, URL = ?, checkinterval = ?, monitoring = ?, public = ?, emailInterval = ?, logDays = ? WHERE id = ?"
)

// Create map of all statements
var (
	Statements map[string]*autorc.Stmt
)

// Prepare all MySQL statements
func PrepareAll() error {

	Statements = make(map[string]*autorc.Stmt, 0)

	if !Db.Raw.IsConnected() {
		Db.Raw.Connect()
	}

	// Start History Statements
	getSiteHistoryPrepared, err := Db.Prepare(getSiteHistoryStmt)
	if err != nil {
		return err
	}
	Statements["getSiteHistoryStmt"] = getSiteHistoryPrepared

	getSiteHistoryUpCountPrepared, err := Db.Prepare(getSiteHistoryUpCountStmt)
	if err != nil {
		return err
	}
	Statements["getSiteHistoryUpCountStmt"] = getSiteHistoryUpCountPrepared

	getSiteHistoryDownCountPrepared, err := Db.Prepare(getSiteHistoryDownCountStmt)
	if err != nil {
		return err
	}
	Statements["getSiteHistoryDownCountStmt"] = getSiteHistoryDownCountPrepared

	getSiteUptimePrepared, err := Db.Prepare(getSiteUptimeStmt)
	if err != nil {
		return err
	}
	Statements["getSiteUptimeStmt"] = getSiteUptimePrepared

	getLastEmailPrepared, err := Db.Prepare(getLastEmailStmt)
	if err != nil {
		return err
	}
	Statements["getLastEmailStmt"] = getLastEmailPrepared

	insertLogPrepared, err := Db.Prepare(insertLogStmt)
	if err != nil {
		return err
	}
	Statements["insertLogStmt"] = insertLogPrepared

	clearOldLogsPrepared, err := Db.Prepare(clearOldLogsStmt)
	if err != nil {
		return err
	}
	Statements["clearOldLogsStmt"] = clearOldLogsPrepared

	// End History Statements

	// Start Notify Statements

	getAllNotifiersPrepared, err := Db.Prepare(getAllNotifiersStmt)
	if err != nil {
		return err
	}
	Statements["getAllNotifiersStmt"] = getAllNotifiersPrepared

	getNotifierByIDPrepared, err := Db.Prepare(getNotifierByIDStmt)
	if err != nil {
		return err
	}
	Statements["getNotifierByIDStmt"] = getNotifierByIDPrepared

	deleteNotifierPrepared, err := Db.Prepare(deleteNotifierStmt)
	if err != nil {
		return err
	}
	Statements["deleteNotifierStmt"] = deleteNotifierPrepared

	insertNotifierPrepared, err := Db.Prepare(insertNotifierStmt)
	if err != nil {
		return err
	}
	Statements["insertNotifierStmt"] = insertNotifierPrepared

	// End Notify Statements

	// Start Setting Statements

	getSettingsPrepared, err := Db.Prepare(getSettingsStmt)
	if err != nil {
		return err
	}
	Statements["getSettingsStmt"] = getSettingsPrepared

	getSettingsIDPrepared, err := Db.Prepare(getSettingsIDStmt)
	if err != nil {
		return err
	}
	Statements["getSettingsIDStmt"] = getSettingsIDPrepared

	insertSettingsPrepared, err := Db.Prepare(insertSettingsStmt)
	if err != nil {
		return err
	}
	Statements["insertSettingsStmt"] = insertSettingsPrepared

	updateSettingsPrepared, err := Db.Prepare(updateSettingsStmt)
	if err != nil {
		return err
	}
	Statements["updateSettingsStmt"] = updateSettingsPrepared

	// End Setting Statements

	// Start Website Statements

	getAllWebsitesPrepared, err := Db.Prepare(getAllWebsitesStmt)
	if err != nil {
		return err
	}
	Statements["getAllWebsitesStmt"] = getAllWebsitesPrepared

	getAllMonitoringWebsitesPrepared, err := Db.Prepare(getAllMonitoringWebsitesStmt)
	if err != nil {
		return err
	}
	Statements["getAllMonitoringWebsitesStmt"] = getAllMonitoringWebsitesPrepared

	getWebsiteByIDPrepared, err := Db.Prepare(getWebsiteByIDStmt)
	if err != nil {
		return err
	}
	Statements["getWebsiteByIDStmt"] = getWebsiteByIDPrepared

	deleteWebsiteByIDPrepared, err := Db.Prepare(deleteWebsiteByIDStmt)
	if err != nil {
		return err
	}
	Statements["deleteWebsiteByIDStmt"] = deleteWebsiteByIDPrepared

	deleteHistoryBySitePrepared, err := Db.Prepare(deleteHistoryBySiteStmt)
	if err != nil {
		return err
	}
	Statements["deleteHistoryBySiteStmt"] = deleteHistoryBySitePrepared

	deleteNotifiersBySitePrepared, err := Db.Prepare(deleteNotifiersBySiteStmt)
	if err != nil {
		return err
	}
	Statements["deleteNotifiersBySiteStmt"] = deleteNotifiersBySitePrepared

	insertWebsitePrepared, err := Db.Prepare(insertWebsiteStmt)
	if err != nil {
		return err
	}
	Statements["insertWebsiteStmt"] = insertWebsitePrepared

	updateWebsitePrepared, err := Db.Prepare(updateWebsiteStmt)
	if err != nil {
		return err
	}
	Statements["updateWebsiteStmt"] = updateWebsitePrepared

	return nil
}

func GetStatement(key string) (stmt *autorc.Stmt, err error) {
	stmt, ok := Statements[key]
	if !ok {
		qry := expvar.Get(key)
		if qry == nil {
			err = errors.New("Invalid query reference")
		} else {
			stmt, err = Db.Prepare(qry.String())
		}
	}
	return

}
