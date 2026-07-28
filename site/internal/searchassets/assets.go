// Package searchassets serves the local documentation search runtime.
package searchassets

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/*.js
var files embed.FS

// Handler serves the documentation search runtime.
func Handler() http.Handler {
	assets, err := fs.Sub(files, "assets")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(assets))
}
