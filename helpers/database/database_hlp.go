package database

import (
	"github.com/ziutek/mymysql/autorc"
	_ "github.com/ziutek/mymysql/thrsafe"
	"log"
	"os"
)

var (
	// MySQL Connection Handler
	Db = autorc.New(db_proto, "", db_addr, db_user, db_pass, db_name)
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
