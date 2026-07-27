package line

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLineRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Line(Config{
		Label:   "Weekly signups",
		Caption: "Seven-day trend",
		Labels:  []string{"Mon", "Tue", "Wed"},
		Series: []Series{{Name: "Signups", Values: []float64{12, 18, 15}}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<figure role=\"img\" aria-label=\"Weekly signups\"", "<svg", "Seven-day trend"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q:\n%s", want, markup)
		}
	}
	if strings.Contains(markup, "<script") {
		t.Errorf("SSR chart unexpectedly contains script: %s", markup)
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
