package globals

import (
	"flag"
)

func GetGlobal(k string) string {
	var filePath string
	flag.StringVar(&filePath, "path", "", "path to files")
	flag.Parse()
	switch k {
	case "Filepath":
		return filePath
	}
	return ""
}
