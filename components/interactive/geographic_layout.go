package interactive

import (
	"strconv"
	"strings"
)

// geographicLayout is private renderer input shared by Map and Geo. GeoJSON
// coordinates use their source aspect instead of the renderer's historical
// 0.75 default. layoutCenter/layoutSize preserve that aspect while sizing from
// the chart's shorter dimension during page, modal, and fullscreen resizes.
type geographicLayout struct {
	centerY string
	size    string
}

func resolvedGeographicLayout(hasTitle, hasScale bool) geographicLayout {
	if hasScale {
		// Keep a full horizontal visual scale below the geographic plot.
		return geographicLayout{centerY: "46%", size: "82%"}
	}
	if hasTitle {
		return geographicLayout{centerY: "54%", size: "88%"}
	}
	return geographicLayout{centerY: "50%", size: "92%"}
}

func geographicLayoutReplacements(mapName string, hasTitle, hasScale bool) []scriptReplacement {
	layout := resolvedGeographicLayout(hasTitle, hasScale)
	// Geo series reference the coordinate system rather than repeating its map
	// name; Map has one map-bearing series. The target is unique in both charts.
	target := `"map":` + strconv.Quote(mapName)
	return []scriptReplacement{{
		Old: target,
		New: target + `,"aspectScale":1,"layoutCenter":["50%",` + strconv.Quote(layout.centerY) + `],"layoutSize":` + strconv.Quote(layout.size),
	}}
}

func hasChartTitle(options ChartOptions) bool {
	return options.Title != nil && (strings.TrimSpace(options.Title.Text) != "" || strings.TrimSpace(options.Title.Subtitle) != "")
}
