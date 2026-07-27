// Package server serves the Goshtoso Charts component demo site.
package server

import (
	"net/http"

	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
	"github.com/araihu/goshtoso-charts/site/internal/pages"
	"github.com/araihu/goshtoso/assets"
)

// New returns the demo site's HTTP handler.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.Handle("GET /componentdocshell/assets/", shellassets.Handler())
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		render(writer, request, pages.OverviewPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/heartbeat", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.HeartbeatPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/line", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.LinePage(isFragment(request)))
	})
	mux.HandleFunc("GET /examples/status-page", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.StatusPageExample(isFragment(request)))
	})
	return mux
}

func isFragment(request *http.Request) bool {
	return request.Header.Get("HX-Request") == "true"
}

func render(writer http.ResponseWriter, request *http.Request, page pages.Page) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.Render(request.Context(), writer); err != nil {
		http.Error(writer, "render demo", http.StatusInternalServerError)
	}
}
