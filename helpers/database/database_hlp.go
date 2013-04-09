package database

import (
	"../mymysql/autorc"
	// "github.com/ziutek/mymysql/mysql"
	_ "../mymysql/thrsafe"
	"log"
	"os"
)

var (
	// MySQL Connection Handler
	Db = autorc.New("tcp", "", "curtsql.cloudapp.net:3306", "sitemonitor", "S1teM0nitor", "SiteMonitor")
)

func MysqlError(err error) (ret bool) {
	ret = (err != nil)
	if ret {
		log.Println("MySQL error: ", err)
	}
	return
}

func MysqlErrExit(err error) {
	if MysqlError(err) {
		os.Exit(1)
	}
}
