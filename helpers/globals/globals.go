package globals

import (
	"flag"
)

var (
	Filepath   = GetGlobal("path")
	ListenAddr = flag.String("http", ":8080", "http listen address")
)

func SetGlobals() {
	flag.Parse()
}

func GetGlobal(k string) string {
	var flagdata string
	flag.StringVar(&flagdata, k, "", "path to files")
	return flagdata
}
