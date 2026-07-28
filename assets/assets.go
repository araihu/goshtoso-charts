// Package assets serves versioned browser assets used by Goshtoso Charts.
package assets

import (
	"embed"
	"net/http"
)

const (
	// Prefix is the default HTTP mount path consumed by components/dependencies.
	Prefix = "/charts/assets/"
	// RuntimeURL is the versioned path of the embedded interactive-chart runtime.
	RuntimeURL = Prefix + "js/runtime/echarts/5.4.3/echarts.min.js"
	// RuntimeCDNURL is the pinned opt-in CDN source for the same runtime version.
	RuntimeCDNURL = "https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"
	// RuntimeCDNIntegrity authenticates bytes served by RuntimeCDNURL.
	RuntimeCDNIntegrity = "sha384-BQKzmHvQLMCAnL3UtDBA1Al5tFjsCz1wrMlIUA1wkzo14DYkRWjywW+p9pCj0cwd"
)

//go:embed js/runtime/echarts/5.4.3/echarts.min.js
var files embed.FS

// Handler serves embedded assets at Prefix. Mount it directly at Prefix; the
// handler removes that prefix itself.
func Handler() http.Handler {
	return http.StripPrefix(Prefix, http.FileServer(http.FS(files)))
}
