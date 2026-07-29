package pages

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestInteractiveLineTimeSourcePinAndContract(t *testing.T) {
	t.Parallel()
	if interactiveLineTimeUpstreamPath != "examples/line.go (lineTime and generateLineItemsTwoAxis)" || interactiveLineTimeUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" {
		t.Fatal("temporal Line upstream source pin changed")
	}
	cfg := sampleInteractiveLineTime()
	if len(cfg.TimeAxis.Values) != 50 || cfg.TimeAxis.Minimum.Format("2006-01-02") != "2025-01-01" || cfg.TimeAxis.Values[0].Format("2006-01-02") != "2025-01-31" || cfg.TimeAxis.SplitNumber != 0 {
		t.Fatalf("temporal axis = %#v", cfg.TimeAxis)
	}
	if cfg.Options.Title.Text != "temporal X axis" || cfg.Options.Title.Subtitle != "time.Date as X axis values" || *cfg.Options.YAxis.Min != 0 || *cfg.Options.YAxis.Max != 200 || cfg.Options.Tooltip.Trigger != "axis" {
		t.Fatalf("temporal options = %#v", cfg.Options)
	}
	for _, point := range cfg.Series[0].Data {
		if point.Value < 100 || point.Value > 119 {
			t.Fatalf("value %v outside source domain", point.Value)
		}
	}
	if got, again := deterministicLineTimeData(50), deterministicLineTimeData(50); len(got) != len(again) || got[0] != again[0] || got[49] != again[49] {
		t.Fatal("temporal data generator is not deterministic")
	}
}

func TestInteractiveLineTimePageRendersOneLineComponentWithEvidence(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := InteractiveLinePage(false).Render(context.Background(), &output); err != nil {
		t.Fatalf("render line page: %v", err)
	}
	for _, want := range []string{"Temporal X-axis treatment", "Exact time and values", "2025-01-31T00:00:00Z", "UTC timestamps.", "Show change across an ordered sequence."} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("line page missing %q", want)
		}
	}
}
