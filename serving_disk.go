package storage

import "net/http"

// ServeDisk serves files from the given directory.
func ServeDisk(routePath string, dir string) http.Handler {
	return http.StripPrefix(routePath, http.FileServer(http.Dir(dir)))
}
