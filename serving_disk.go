package storage

import "net/http"

func ServeDisk(routePath string, dir string) http.Handler {
	return http.StripPrefix(routePath, http.FileServer(http.Dir(dir)))
}
