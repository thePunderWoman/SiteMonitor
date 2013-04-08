package database

import (
	"../ConnectionString"
	"../mymysql/autorc"
	// "github.com/ziutek/mymysql/mysql"
	_ "../mymysql/thrsafe"
	"log"
	"os"
)

var (
	// MySQL Connection Handler
	Db = autorc.New(ConnectionString.Db_proto, "", ConnectionString.Db_addr, ConnectionString.Db_user, ConnectionString.Db_pass, ConnectionString.Db_name)

	//  Prepared statements would go here
	//  stmt *autorc.Stmt
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
