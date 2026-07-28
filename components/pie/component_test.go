package pie

import (
	"bytes"
	"context"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
)

func TestPieRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Pie(Config{
		Label: "Deployments by status", Caption: "This week.",
		Slices: []Slice{{Name: "Successful", Value: 14}, {Name: "Failed", Value: 2}},
	})
	if instance.Kind() != chartcomponents.KindPieChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindPieChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<figure class=\"goshtoso-charts-pie goshtoso-charts-palette goshtoso-charts-palette-auto\" role=\"img\" aria-label=\"Deployments by status\"", "<svg", "This week.", "var(--goshtoso-charts-series-1)", "var(--goshtoso-charts-surface)", "var(--font-paragraph), sans-serif"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, "<script") {
		t.Errorf("SSR chart unexpectedly contains script")
	}
}

func TestPieRejectsNegativeSlice(t *testing.T) {
	t.Parallel()
	if _, err := renderSVG(Config{Label: "Deployments", Slices: []Slice{{Name: "Successful", Value: -1}}}); err == nil {
		t.Fatal("renderSVG() error = nil, want validation error")
	}
}

func TestPieRendersNoDataState(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := Pie(Config{Label: "Deployment outcomes"}).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(output.String(), "No data in this period.") {
		t.Fatalf("rendered markup missing explicit no-data caption: %s", output.String())
	}
}
