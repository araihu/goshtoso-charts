package bar_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"regexp"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

var chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestBarPreservesLegacyRenderContract(t *testing.T) {
	t.Parallel()
	showLabels := true
	cfg := interactivebar.Config{
		Label:       "Quarterly revenue",
		Caption:     "Revenue by product.",
		XAxis:       []string{"Q1", "Q2"},
		Orientation: interactivebar.OrientationHorizontal,
		Zoom:        &interactivebar.Zoom{Mode: interactivebar.ZoomSlider, StartPercent: 10, EndPercent: 80},
		Series: []interactivebar.Series{{
			Name: "Hardware",
			Data: []interactivebar.Data{{Value: 12}, {Value: 18}},
			References: interactivebar.References{
				Points: []interactivebar.PointReference{
					{Name: "Maximum", Statistic: interactivebar.StatisticMaximum},
					{Name: "Q1 target", Coordinate: &interactivebar.Coordinate{Category: "Q1", Value: 12}},
				},
				Lines:      []interactivebar.GuideReference{{Name: "Average", Statistic: interactivebar.StatisticAverage}},
				ShowLabels: &showLabels,
			},
		}},
		Options: interactive.ChartOptions{Animation: interactive.Bool(false)},
	}

	var legacyConfig interactive.BarConfig = cfg
	var canonicalConfig interactivebar.Config = legacyConfig
	canonical := interactivebar.Bar(canonicalConfig)
	legacy := interactive.Bar(legacyConfig)
	if canonical.Kind() != chartcomponents.KindInteractiveBar || canonical.Kind() != legacy.Kind() {
		t.Fatalf("canonical Kind() = %q, legacy Kind() = %q", canonical.Kind(), legacy.Kind())
	}

	canonicalMarkup := render(t, canonical)
	legacyMarkup := render(t, legacy)
	if canonicalMarkup != legacyMarkup {
		t.Fatalf("canonical render differs from legacy render\ncanonical: %s\nlegacy: %s", canonicalMarkup, legacyMarkup)
	}
	digest := sha256.Sum256([]byte(canonicalMarkup))
	if got, want := hex.EncodeToString(digest[:]), "f9270acadd88778afa13f78b99855026a90b07d069cb574583c1a67472c55d0e"; got != want {
		t.Fatalf("normalized render SHA-256 = %s, want %s", got, want)
	}
}

func TestBarPreservesLegacyValidation(t *testing.T) {
	t.Parallel()
	invalid := interactivebar.Config{
		Label:       "Revenue",
		XAxis:       []string{"Q1"},
		Series:      []interactivebar.Series{{Name: "Hardware", Data: []interactivebar.Data{{Value: 12}}}},
		Orientation: interactivebar.Orientation("diagonal"),
	}

	canonicalError := renderError(interactivebar.Bar(invalid))
	legacyError := renderError(interactive.Bar(invalid))
	if canonicalError != `bar chart orientation "diagonal" is not supported` {
		t.Fatalf("canonical validation error = %q", canonicalError)
	}
	if canonicalError != legacyError {
		t.Fatalf("canonical validation error = %q, legacy = %q", canonicalError, legacyError)
	}
}

func TestBarExportsConciseChartSpecificConstants(t *testing.T) {
	t.Parallel()
	_ = []interactivebar.Orientation{interactivebar.OrientationVertical, interactivebar.OrientationHorizontal}
	_ = []interactivebar.ZoomMode{interactivebar.ZoomInside, interactivebar.ZoomSlider}
	_ = []interactivebar.Statistic{interactivebar.StatisticMinimum, interactivebar.StatisticMaximum, interactivebar.StatisticAverage}
}

func TestBarSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()

	wantPackage := "github.com/araihu/goshtoso-charts/components/interactive/bar"
	barTypes := []reflect.Type{
		reflect.TypeOf(interactivebar.Config{}),
		reflect.TypeOf(interactivebar.Series{}),
		reflect.TypeOf(interactivebar.Data{}),
		reflect.TypeOf(interactivebar.Orientation("")),
		reflect.TypeOf(interactivebar.ZoomMode("")),
		reflect.TypeOf(interactivebar.Zoom{}),
		reflect.TypeOf(interactivebar.Statistic("")),
		reflect.TypeOf(interactivebar.Coordinate{}),
		reflect.TypeOf(interactivebar.PointReference{}),
		reflect.TypeOf(interactivebar.GuideReference{}),
		reflect.TypeOf(interactivebar.References{}),
	}
	for _, barType := range barTypes {
		if got := barType.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", barType, got, wantPackage)
		}
	}

	if reflect.TypeOf(interactive.BarConfig{}) != reflect.TypeOf(interactivebar.Config{}) {
		t.Fatal("interactive.BarConfig is not an exact alias of bar.Config")
	}
	if reflect.TypeOf(interactive.BarSeries{}) != reflect.TypeOf(interactivebar.Series{}) {
		t.Fatal("interactive.BarSeries is not an exact alias of bar.Series")
	}
	if reflect.TypeOf(interactive.BarData{}) != reflect.TypeOf(interactivebar.Data{}) {
		t.Fatal("interactive.BarData is not an exact alias of bar.Data")
	}
}

var (
	_ func(interactivebar.Config) chart.Instance = interactivebar.Bar
	_ func(interactivebar.Config) chart.Instance = interactive.Bar
	_ func(interactive.BarConfig) chart.Instance = interactivebar.Bar
)

func render(t *testing.T, instance interactive.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	match := chartIDPattern.FindStringSubmatch(markup)
	if len(match) != 2 {
		t.Fatalf("rendered markup lacks chart ID: %s", markup)
	}
	return strings.ReplaceAll(markup, match[1], "CHARTID")
}

func renderError(instance interactive.Instance) string {
	var output bytes.Buffer
	err := instance.Render(context.Background(), &output)
	if err == nil {
		return ""
	}
	return err.Error()
}
