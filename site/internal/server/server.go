// Package server serves the Goshtoso Charts component demo site.
package server

import (
	"net/http"

	"github.com/araihu/goshtoso-charts/site/internal/pages"
	"github.com/araihu/goshtoso-charts/site/internal/siteassets"
	"github.com/araihu/goshtoso/assets"
)

// New returns the demo site's HTTP handler.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /site.css", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = writer.Write([]byte(siteassets.CSS))
	})
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		render(writer, request, pages.OverviewPage())
	})
	mux.HandleFunc("GET /components/heartbeat", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.HeartbeatPage())
	})
	mux.HandleFunc("GET /components/line", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.LinePage())
	})
	mux.HandleFunc("GET /examples/status-page", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.StatusPageExample())
	})
	return mux
}

func render(writer http.ResponseWriter, request *http.Request, page pages.Page) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.Render(request.Context(), writer); err != nil {
		http.Error(writer, "render demo", http.StatusInternalServerError)
	}
}
