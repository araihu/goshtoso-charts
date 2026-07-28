// Package echartsassets serves the pinned local ECharts runtime.
package echartsassets

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/echarts.min.js
var files embed.FS

// Handler serves ECharts without any third-party runtime origin.
func Handler() http.Handler {
	assets, err := fs.Sub(files, "assets")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(assets))
}
