package pages

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestLineCoverageUsesOneCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()

	const canonicalPath = "../../../docs/upstream-example-coverage.md"
	data, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)

	staticExamples := []string{
		"line_chart-1-basic", "line_chart-2-symbols", "line_chart-3-smooth",
		"line_chart-4-mark", "line_chart-5-area", "line_chart-6-stacked",
		"line_chart-7-boundary_gap", "line_chart-8-dual_y_axis",
		"line_chart-9-custom", "line_chart-10-gradient_labels",
	}
	for _, name := range staticExamples {
		if count := strings.Count(ledger, "examples/1-Painter/"+name+"/main.go"); count != 1 {
			t.Errorf("static coverage row %q occurs %d times, want 1", name, count)
		}
	}

	for _, entry := range interactiveLineUpstreamCoverage() {
		row := "| `" + entry.Name + "` |"
		if count := strings.Count(ledger, row); count != 1 {
			t.Errorf("interactive coverage row %q occurs %d times, want 1", entry.Name, count)
		}
	}
	if !strings.Contains(ledger, "| `lineOverlap` | Unsupported |") {
		t.Error("lineOverlap must remain explicitly unsupported in canonical coverage ledger")
	}

	if _, err := os.Stat("../../../docs/upstream-coverage.md"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy duplicate coverage ledger still exists or is unreadable: %v", err)
	}
	for _, attribution := range chartAttributions {
		if !strings.Contains(attribution.UsedFor, "Line") {
			continue
		}
		if !strings.Contains(attribution.UsedFor, "docs/upstream-example-coverage.md") {
			t.Errorf("%s Line attribution does not point to canonical coverage ledger", attribution.Name)
		}
	}
}

func TestInteractiveBarCoverageUsesCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()
	const canonicalPath = "../../../docs/upstream-example-coverage.md"
	data, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)
	for _, entry := range interactiveBarUpstreamCoverage() {
		row := "| `" + entry.Name + "` |"
		if count := strings.Count(ledger, row); count != 1 {
			t.Errorf("interactive Bar coverage row %q occurs %d times, want 1", entry.Name, count)
		}
	}
	if !strings.Contains(ledger, "| `barOverlap` | Unsupported |") {
		t.Error("barOverlap must remain explicitly unsupported in canonical coverage ledger")
	}
	for _, source := range interactiveBarSupplementarySources() {
		for _, want := range []string{"`" + source.Path + "`", "`" + source.SHA256 + "`"} {
			if count := strings.Count(ledger, want); count != 1 {
				t.Errorf("supplementary source evidence %q occurs %d times, want 1", want, count)
			}
		}
	}
	if count := strings.Count(ledger, "`"+interactiveBarUpstreamPath+"`"); count != 1 {
		t.Errorf("canonical interactive Bar source path occurs %d times, want 1", count)
	}
}
