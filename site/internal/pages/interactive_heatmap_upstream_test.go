package pages

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

func TestInteractiveHeatMapCategorySamplePreservesPinnedSourceData(t *testing.T) {
	t.Parallel()
	cfg := sampleInteractiveHeatMap()
	if len(cfg.XAxis) != 24 || len(cfg.YAxis) != 7 || len(cfg.Series) != 1 || len(cfg.Series[0].Data) != 168 {
		t.Fatalf("category heatmap shape = %d x %d, %d series, %d cells", len(cfg.XAxis), len(cfg.YAxis), len(cfg.Series), len(cfg.Series[0].Data))
	}
	if cfg.ValueRange.Min != 0 || cfg.ValueRange.Max != 10 || cfg.ValueRange.Calculable == nil || !*cfg.ValueRange.Calculable {
		t.Fatalf("category value range = %#v", cfg.ValueRange)
	}
	if cfg.SplitArea == nil || !*cfg.SplitArea {
		t.Fatal("category split areas are not enabled")
	}
	missing := 0
	for index, source := range interactiveHeatMapSourceData {
		point := cfg.Series[0].Data[index]
		if point.X != source[1] || point.Y != source[0] || point.Value != float64(source[2]) || point.Missing != (source[2] == 0) {
			t.Fatalf("category cell %d = %#v, source %#v", index, point, source)
		}
		if point.Missing {
			missing++
		}
	}
	if missing != 62 {
		t.Fatalf("category missing cells = %d, want 62", missing)
	}
}

func TestInteractiveHeatMapPinnedSourceInventory(t *testing.T) {
	t.Parallel()
	if interactiveHeatMapUpstreamPath != "examples/heatmap.go" || interactiveHeatMapUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" || interactiveHeatMapUpstreamSHA256 != "c08b194eafa5e02e941ad91f7ff8402448bc77b407cc97903b19063d06dd6f14" {
		t.Fatalf("unexpected HeatMap source identity: %s %s %s", interactiveHeatMapUpstreamPath, interactiveHeatMapUpstreamRevision, interactiveHeatMapUpstreamSHA256)
	}
	if len(interactiveHeatMapUpstreamInventory) != 9 {
		t.Fatalf("HeatMap upstream inventory has %d spans, want 9", len(interactiveHeatMapUpstreamInventory))
	}
	for _, span := range interactiveHeatMapUpstreamInventory {
		if span.Name == "" || span.Lines == "" || len(span.SHA256) != 64 {
			t.Fatalf("invalid HeatMap upstream span: %#v", span)
		}
	}
}

func TestInteractiveHeatMapCalendarSampleIsDeterministicAndPreservesSourceShape(t *testing.T) {
	t.Parallel()
	first := sampleInteractiveCalendarHeatMap()
	second := sampleInteractiveCalendarHeatMap()
	if first.Coordinate != interactive.HeatMapCoordinateCalendar || first.Calendar == nil {
		t.Fatalf("calendar coordinate = %q, calendar = %#v", first.Coordinate, first.Calendar)
	}
	if len(first.Series) != 1 || len(first.Series[0].Data) != 366 {
		t.Fatalf("calendar series shape = %d series, %d cells", len(first.Series), len(first.Series[0].Data))
	}
	if first.ValueRange.Min != 0 || first.ValueRange.Max != 20 {
		t.Fatalf("calendar value range = %#v", first.ValueRange)
	}
	if got := int(first.Calendar.End.Sub(first.Calendar.Start).Hours()/24) + 1; got != 366 {
		t.Fatalf("calendar inclusive span = %d days", got)
	}
	var sequence strings.Builder
	missing := 0
	for index, point := range first.Series[0].Data {
		other := second.Series[0].Data[index]
		if point != other {
			t.Fatalf("calendar cell %d is not deterministic: %#v != %#v", index, point, other)
		}
		wantDate := first.Calendar.Start.AddDate(0, 0, index)
		if !point.Date.Equal(wantDate) {
			t.Fatalf("calendar date %d = %s, want %s", index, point.Date, wantDate)
		}
		if point.Value < 0 || point.Value >= 21 || point.Missing != (point.Value == 0) {
			t.Fatalf("calendar value %d = %#v", index, point)
		}
		if point.Missing {
			missing++
		}
		fmt.Fprintf(&sequence, "%s|%g|%t\n", point.Date.Format("2006-01-02"), point.Value, point.Missing)
	}
	if missing != 30 {
		t.Fatalf("calendar missing cells = %d, want 30", missing)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(sequence.String()))); got != "9a40cbf37efd7a63deedac491f5c22b2dce085bd73ff23bf08aa98cd6d36d431" {
		t.Fatalf("calendar sequence SHA-256 = %s", got)
	}
}

func TestInteractiveHeatMapSnippetsAreSelfContained(t *testing.T) {
	t.Parallel()
	category := interactiveChartHeatMapCode()
	if strings.Contains(category, "dayHours") || !strings.Contains(category, `"12a"`) || !strings.Contains(category, `"11p"`) {
		t.Fatalf("category snippet keeps unresolved labels: %s", category)
	}
	calendar := interactiveCalendarHeatMapCode()
	for _, unresolved := range []string{"Start: start", "End: end", "calendarValues"} {
		if strings.Contains(calendar, unresolved) {
			t.Fatalf("calendar snippet keeps unresolved identifier %q: %s", unresolved, calendar)
		}
	}
	for _, want := range []string{`Add "time"`, "time.Date(2024", "[]interactive.HeatMapData", "Missing: true"} {
		if !strings.Contains(calendar, want) {
			t.Fatalf("calendar snippet missing %q: %s", want, calendar)
		}
	}
}
