package globals

import (
	"flag"
)

var (
	Filepath = func() *string {
		flag.Parse()
		return flag.String("path", "", "path to files")
	}()
)
