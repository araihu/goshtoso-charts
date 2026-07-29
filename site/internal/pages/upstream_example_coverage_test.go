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

func TestStaticBarCoverageUsesCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)
	if !strings.Contains(ledger, "## Static/vector Bar") || !strings.Contains(ledger, staticBarUpstreamRevision) {
		t.Fatal("canonical ledger is missing the pinned static Bar section or revision")
	}
	for _, entry := range staticBarUpstreamCoverage() {
		if count := strings.Count(ledger, "`"+entry.Path+"`"); count != 1 {
			t.Errorf("static Bar coverage row %q occurs %d times, want 1", entry.Path, count)
		}
		if !strings.Contains(ledger, "`"+entry.SHA256+"`") {
			t.Errorf("static Bar coverage row %q is missing SHA-256 %s", entry.Path, entry.SHA256)
		}
	}
	if !strings.Contains(ledger, "eleven dedicated") || !strings.Contains(ledger, "nine distinct") {
		t.Error("static Bar ledger must state both file and distinct-behavior coverage counts")
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

func TestInteractivePieCoverageUsesCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)
	if !strings.Contains(ledger, "## Interactive Pie") || !strings.Contains(ledger, interactivePieUpstreamRevision) || !strings.Contains(ledger, interactivePieUpstreamSHA256) {
		t.Fatal("canonical ledger is missing the pinned interactive Pie section")
	}
	for _, entry := range interactivePieUpstreamCoverage() {
		row := "| `" + entry.Name + "` | Example |"
		if count := strings.Count(ledger, row); count != 1 {
			t.Errorf("interactive Pie coverage row %q occurs %d times, want 1", entry.Name, count)
		}
	}
	if !strings.Contains(ledger, "all nine upstream Pie behavior functions") || strings.Contains(ledger, "pieWithDispatchAction` | Unsupported") {
		t.Error("interactive Pie ledger must record exhaustive nine-of-nine coverage")
	}
	for _, source := range interactivePieSupplementarySources() {
		for _, want := range []string{"`" + source.Path + "`", "`" + source.SHA256 + "`"} {
			if count := strings.Count(ledger, want); count != 1 {
				t.Errorf("shared Pie layout source evidence %q occurs %d times, want 1", want, count)
			}
		}
	}
}

func TestInteractiveScatterCoverageUsesCanonicalLedgerWithoutLostEntries(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read canonical upstream coverage ledger: %v", err)
	}
	ledger := string(data)
	for _, want := range []string{
		"## Interactive Scatter", interactiveScatterUpstreamRevision,
		interactiveScatterUpstreamPath, interactiveScatterUpstreamSHA256,
		interactiveEffectScatterUpstreamPath, interactiveEffectScatterUpstreamSHA256,
		"all five upstream behavior functions", "renderer-neutral `interactive.Scatter` component",
	} {
		if !strings.Contains(ledger, want) {
			t.Errorf("canonical ledger missing interactive Scatter evidence %q", want)
		}
	}
	for _, entry := range interactiveScatterUpstreamCoverage() {
		if count := strings.Count(ledger, "| `"+entry.Name+"` | Example |"); count != 1 {
			t.Errorf("interactive Scatter behavior row %q occurs %d times, want 1", entry.Name, count)
		}
	}
	for _, function := range interactiveScatterSourceFunctions() {
		if !strings.Contains(ledger, "`"+function.Name+"`") {
			t.Errorf("interactive Scatter function inventory missing %q", function.Name)
		}
		if count := strings.Count(ledger, "`"+function.SHA256+"`"); count != 1 {
			t.Errorf("interactive Scatter function hash %q occurs %d times, want 1", function.SHA256, count)
		}
	}
	for _, phrase := range []string{"visual", "maps or pieces", "dataset transforms", "data zoom", "statistical references"} {
		if !strings.Contains(ledger, phrase) {
			t.Errorf("interactive Scatter scope boundary missing %q", phrase)
		}
	}
}
