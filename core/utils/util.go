package utils

import (
	"fmt"
	"os"
	"strings"
)

func GetReqNamePath(req string) (name string, path string) {
	restlerPath := os.Getenv("RESTLER_PATH")

	if strings.Contains(req, "/") {
		_paths := strings.Split(req, "/")
		name = _paths[len(_paths)-1]
		path = fmt.Sprintf("%s/%s", restlerPath, strings.Join(_paths[:len(_paths)-1], "/"))
		return
	}

	name = req
	path = restlerPath
	return
}
