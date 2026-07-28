// Package server serves the Goshtoso Charts component demo site.
package server

import (
	"net/http"

	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
	"github.com/araihu/goshtoso-charts/site/internal/brand"
	"github.com/araihu/goshtoso-charts/site/internal/echartsassets"
	"github.com/araihu/goshtoso-charts/site/internal/pages"
	"github.com/araihu/goshtoso-charts/site/internal/searchassets"
	"github.com/araihu/goshtoso/assets"
)

// New returns the demo site's HTTP handler.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.Handle("GET /brand/", http.StripPrefix("/brand/", brand.Handler()))
	mux.Handle("GET /charts/echarts/", http.StripPrefix("/charts/echarts/", echartsassets.Handler()))
	mux.Handle("GET /componentdocshell/assets/", shellassets.Handler())
	mux.Handle("GET /search/assets/", http.StripPrefix("/search/assets/", searchassets.Handler()))
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		render(writer, request, pages.OverviewPage(isFragment(request)))
	})
	mux.HandleFunc("GET /attributions", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.AttributionsPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/heartbeat", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.HeartbeatPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/line", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.LinePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/bar", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.BarPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/pie", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.PiePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/bar", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveBarPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/line", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveLinePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/scatter", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveScatterPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/effect-scatter", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/components/interactive/scatter", http.StatusPermanentRedirect)
	})
	mux.HandleFunc("GET /components/interactive/pie", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractivePiePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/radar", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveRadarPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/heatmap", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveHeatMapPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/boxplot", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveBoxPlotPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/gauge", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveGaugePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/funnel", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveFunnelPage(isFragment(request)))
	})
	for _, component := range []string{"bar", "line", "scatter", "pie", "radar", "heatmap", "boxplot", "gauge", "funnel"} {
		mux.HandleFunc("GET /components/echarts/"+component, func(writer http.ResponseWriter, request *http.Request) {
			http.Redirect(writer, request, "/components/interactive/"+component, http.StatusPermanentRedirect)
		})
	}
	mux.HandleFunc("GET /components/echarts/effect-scatter", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/components/interactive/scatter", http.StatusPermanentRedirect)
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
