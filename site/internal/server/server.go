// Package server serves the Goshtoso Charts component demo site.
package server

import (
	"net/http"

	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
	chartassets "github.com/araihu/goshtoso-charts/assets"
	"github.com/araihu/goshtoso-charts/site/internal/brand"
	"github.com/araihu/goshtoso-charts/site/internal/pages"
	"github.com/araihu/goshtoso-charts/site/internal/searchassets"
	"github.com/araihu/goshtoso/assets"
)

// New returns the demo site's HTTP handler.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.Handle("GET "+chartassets.Prefix, chartassets.Handler())
	mux.Handle("GET /brand/", http.StripPrefix("/brand/", brand.Handler()))
	mux.Handle("GET /componentdocshell/assets/", shellassets.Handler())
	mux.Handle("GET /search/assets/", http.StripPrefix("/search/assets/", searchassets.Handler()))
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		render(writer, request, pages.GettingStartedPage(isFragment(request)))
	})
	mux.HandleFunc("GET /attributions", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.AttributionsPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/heartbeat", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/examples/live-availability", http.StatusPermanentRedirect)
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
	mux.HandleFunc("GET /components/scatter", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.ScatterPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/radar", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.RadarPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/candlestick", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.CandlestickPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/funnel", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.FunnelPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/heatmap", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.HeatMapPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/table", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.TablePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/violin", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.ViolinPage(isFragment(request)))
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
	mux.HandleFunc("GET /components/interactive/candlestick", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveCandlestickPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/gauge", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveGaugePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/funnel", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveFunnelPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/graph", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveGraphPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/sankey", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveSankeyPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/tree", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveTreePage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/sunburst", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveSunburstPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/treemap", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveTreemapPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/parallel", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveParallelPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/theme-river", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveThemeRiverPage(isFragment(request)))
	})
	mux.HandleFunc("GET /components/interactive/word-cloud", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.InteractiveWordCloudPage(isFragment(request)))
	})
	for _, component := range []string{"bar", "line", "scatter", "pie", "radar", "heatmap", "boxplot", "candlestick", "gauge", "funnel", "graph", "sankey", "tree", "word-cloud"} {
		mux.HandleFunc("GET /components/echarts/"+component, func(writer http.ResponseWriter, request *http.Request) {
			http.Redirect(writer, request, "/components/interactive/"+component, http.StatusPermanentRedirect)
		})
	}
	mux.HandleFunc("GET /components/echarts/effect-scatter", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/components/interactive/scatter", http.StatusPermanentRedirect)
	})
	mux.HandleFunc("GET /examples/status-page", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/examples/live-availability", http.StatusPermanentRedirect)
	})
	mux.HandleFunc("GET /examples/live-availability", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.LiveAvailabilityExample(isFragment(request)))
	})
	mux.HandleFunc("GET /examples/live-availability/events", liveAvailabilityEvents)
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
