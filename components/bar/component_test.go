package bar

import (
	"bytes"
	"context"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
)

func TestBarRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Bar(Config{
		Label: "Deployments by environment", Caption: "This week.", Labels: []string{"Development", "Production"},
		Series: []Series{{Name: "Successful", Values: []float64{14, 8}}, {Name: "Failed", Values: []float64{1, 2}}}, Stacked: true,
	})
	if instance.Kind() != chartcomponents.KindBarChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindBarChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<figure class=\"goshtoso-charts-bar\" role=\"img\" aria-label=\"Deployments by environment\"", "<svg", "This week.", "var(--goshtoso-charts-primary)", "var(--goshtoso-charts-surface)"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, "<script") {
		t.Errorf("SSR chart unexpectedly contains script")
	}
}

func TestBarRejectsMisalignedSeries(t *testing.T) {
	t.Parallel()
	_, err := renderSVG(Config{Label: "Deployments", Labels: []string{"Development", "Production"}, Series: []Series{{Name: "Successful", Values: []float64{14}}}})
	if err == nil || !strings.Contains(err.Error(), "has 1 values; need 2") {
		t.Fatalf("renderSVG() error = %v, want alignment error", err)
	}
}
