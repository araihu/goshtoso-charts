package line

import (
	"bytes"
	"context"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestLineSupportsSharedControlsAndExport(t *testing.T) {
	t.Parallel()
	instance := Line(Config{
		Label:    "Latency",
		Labels:   []string{"Mon"},
		Series:   []Series{{Name: "p95", Values: []float64{12}}},
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "latency"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`data-goshtoso-chart-control="fullscreen"`,
		`data-goshtoso-chart-control="collapse"`,
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export-menu`,
		`>SVG</button>`, `>PNG</button>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("markup missing %q", want)
		}
	}
}

func TestLineRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Line(Config{
		Label:   "Weekly signups",
		Caption: "Seven-day trend",
		Labels:  []string{"Mon", "Tue", "Wed"},
		Series:  []Series{{Name: "Signups", Values: []float64{12, 18, 15}}},
		Style:   charttheme.Style{Palette: charttheme.PalettePastel, Colors: []string{"#123456"}, Class: "ring-2"},
	})
	if instance.Kind() != chartcomponents.KindLineChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindLineChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<figure class=\"goshtoso-charts-line goshtoso-charts-palette goshtoso-charts-palette-pastel ring-2\" role=\"img\" aria-label=\"Weekly signups\"", "<svg", "Seven-day trend", "#123456", "var(--color-chart-surface)", "var(--font-paragraph), sans-serif"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q:\n%s", want, markup)
		}
	}
	if strings.Contains(markup, "echarts.init") {
		t.Errorf("SSR chart unexpectedly contains interactive renderer initialization: %s", markup)
	}
}

func TestLineEscapesProgrammaticSeriesColors(t *testing.T) {
	instance := Line(Config{
		Label:  "Safe chart",
		Labels: []string{"one"},
		Series: []Series{{Name: "value", Values: []float64{1}}},
		Style:  charttheme.Style{Colors: []string{`red" onload="alert(1)`}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	if strings.Contains(markup, `fill="red" onload=`) {
		t.Fatalf("programmatic color escaped its SVG attribute: %s", markup)
	}
	if !strings.Contains(markup, `red&#34; onload=&#34;alert(1)`) {
		t.Fatalf("escaped programmatic color missing from SVG: %s", markup)
	}
}

func TestLineRejectsMisalignedSeries(t *testing.T) {
	t.Parallel()
	_, err := renderSVG(Config{
		Label:  "Weekly signups",
		Labels: []string{"Mon", "Tue"},
		Series: []Series{{Name: "Signups", Values: []float64{12}}},
	})
	if err == nil || !strings.Contains(err.Error(), "has 1 values; need 2") {
		t.Fatalf("renderSVG() error = %v, want value alignment error", err)
	}
}
