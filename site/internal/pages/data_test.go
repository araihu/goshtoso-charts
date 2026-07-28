package pages

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/araihu/goshtoso-charts/components/interactive"
	"github.com/araihu/goshtoso-charts/components/pie"
)

func TestDoughnutSampleMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if doughnutUpstreamPath != "examples/1-Painter/doughnut_chart-1-basic/main.go" ||
		doughnutUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		doughnutUpstreamSHA256 != "b97bca2322e90e2f03ab49aa77f683d0c58e027846b939e5a61100602dad1ebf" {
		t.Fatalf("doughnut upstream source = %s@%s SHA-256 %s", doughnutUpstreamPath, doughnutUpstreamRevision, doughnutUpstreamSHA256)
	}
	cfg := sampleDoughnutChart()
	if cfg.Label != "Doughnut Chart" || cfg.Title.Text != "Doughnut Chart" || cfg.Title.Subtitle != "(Fake Data)" {
		t.Fatalf("doughnut titles = label %q title %#v", cfg.Label, cfg.Title)
	}
	if cfg.Width != 600 || cfg.Height != 400 || cfg.InnerRadiusPercent != 60 {
		t.Fatalf("doughnut geometry = %dx%d inner %v", cfg.Width, cfg.Height, cfg.InnerRadiusPercent)
	}
	if cfg.Padding.Top != 20 || cfg.Padding.Right != 20 || cfg.Padding.Bottom != 20 || cfg.Padding.Left != 20 {
		t.Fatalf("doughnut padding = %#v", cfg.Padding)
	}
	if cfg.Legend.LeftPercent != 80 || cfg.Legend.VerticalPlacement != pie.VerticalPlacementBottom ||
		cfg.Legend.Orientation != pie.LegendVertical || cfg.Legend.FontSize != 10 {
		t.Fatalf("doughnut legend = %#v", cfg.Legend)
	}
	names := make([]string, len(cfg.Slices))
	values := make([]float64, len(cfg.Slices))
	for index, slice := range cfg.Slices {
		names[index], values[index] = slice.Name, slice.Value
	}
	if want := []string{"Search Engine", "Direct", "Email", "Union Ads", "Video Ads"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("doughnut names = %v, want %v", names, want)
	}
	if want := []float64{1048, 735, 580, 484, 300}; !reflect.DeepEqual(values, want) {
		t.Fatalf("doughnut values = %v, want %v", values, want)
	}
}

func TestHorizontalBarMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if horizontalBarUpstreamPath != "examples/1-Painter/horizontal_bar_chart-1-basic/main.go" ||
		horizontalBarUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" ||
		horizontalBarUpstreamSHA256 != "735240dd8433bd2494ae019f272840a8ff2fcf5572166b78269e23cbff7111a0" {
		t.Fatalf("horizontal bar upstream source = %s@%s SHA-256 %s", horizontalBarUpstreamPath, horizontalBarUpstreamRevision, horizontalBarUpstreamSHA256)
	}
	cfg := sampleHorizontalWorldPopulation()
	if cfg.Title != "World Population" || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("title/dimensions drifted: %q %dx%d", cfg.Title, cfg.Width, cfg.Height)
	}
	if cfg.Padding.Top != 20 || cfg.Padding.Right != 40 || cfg.Padding.Bottom != 20 || cfg.Padding.Left != 20 {
		t.Fatalf("padding drifted: %#v", cfg.Padding)
	}
	wantLabels := []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"}
	wantNames := []string{"2011", "2012"}
	wantValues := [][]float64{{10, 30, 50, 70, 90, 110, 130}, {20, 40, 60, 80, 100, 120, 140}}
	if !reflect.DeepEqual(cfg.Labels, wantLabels) || len(cfg.Series) != 2 {
		t.Fatalf("category/series shape drifted: labels %v series %d", cfg.Labels, len(cfg.Series))
	}
	for index, series := range cfg.Series {
		if series.Name != wantNames[index] || !reflect.DeepEqual(series.Values, wantValues[index]) {
			t.Fatalf("series %d = %q %v, want %q %v", index, series.Name, series.Values, wantNames[index], wantValues[index])
		}
	}
}

func TestViolinSamplesAreDeterministicAndPreserveUpstreamGenerator(t *testing.T) {
	t.Parallel()
	first, second := sampleDistributionShapes(), sampleDistributionShapes()
	if !reflect.DeepEqual(first.Series, second.Series) {
		t.Fatal("fixed LCG seed did not reproduce violin samples")
	}
	if first.Title != "Distribution Shapes" || first.Width != 1200 || first.Height != 800 || first.Density.Points != 80 {
		t.Fatalf("sample config = title %q, %dx%d, %d points", first.Title, first.Width, first.Height, first.Density.Points)
	}
	wantNames := []string{"Normal", "Right Skewed", "Bimodal", "Tight"}
	wantFirst := []float64{51.632672269835695, 34.032704172611375, 81.03976781672685, 44.259116252823056}
	if len(first.Series) != len(wantNames) {
		t.Fatalf("series count = %d", len(first.Series))
	}
	for index, series := range first.Series {
		if series.Name != wantNames[index] || len(series.Samples) != 200 {
			t.Fatalf("series %d = %q with %d samples", index, series.Name, len(series.Samples))
		}
		if math.Abs(series.Samples[0]-wantFirst[index]) > 1e-12 {
			t.Errorf("series %q first sample = %.15f, want %.15f", series.Name, series.Samples[0], wantFirst[index])
		}
		if !series.Marks.Mean || !series.Marks.Median || !reflect.DeepEqual(series.Statistics.Quantiles, []float64{.25, .75}) {
			t.Errorf("series %q statistics = marks %#v, summary %#v", series.Name, series.Marks, series.Statistics)
		}
	}
}

func TestBasicFunnelDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	cfg := sampleBasicFunnel()
	wantLabels := []string{"Show", "Click", "Visit", "Inquiry", "Order", "Pay", "Cancel"}
	wantValues := []float64{100, 80, 60, 40, 20, 10, 2}
	if cfg.Title != "Funnel" || cfg.Width != 0 || cfg.Height != 0 || cfg.Options.Legend.Padding.Left != 100 {
		t.Fatalf("title/dimensions/legend geometry drifted: %#v", cfg)
	}
	if len(cfg.Stages) != len(wantLabels) {
		t.Fatalf("stage count = %d, want %d", len(cfg.Stages), len(wantLabels))
	}
	for index, stage := range cfg.Stages {
		if stage.Label != wantLabels[index] || stage.Value != wantValues[index] {
			t.Fatalf("stage %d = (%q, %g), want (%q, %g)", index, stage.Label, stage.Value, wantLabels[index], wantValues[index])
		}
	}
}

func TestInteractiveCandlestickDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if interactiveCandlestickUpstreamPath != "examples/kline.go" {
		t.Fatalf("upstream path = %q", interactiveCandlestickUpstreamPath)
	}
	if interactiveCandlestickUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" {
		t.Fatalf("upstream revision = %q", interactiveCandlestickUpstreamRevision)
	}
	if len(interactiveCandlestickUpstreamData) != 88 {
		t.Fatalf("candlestick datum count = %d, want 88", len(interactiveCandlestickUpstreamData))
	}
	hash := sha256.New()
	for _, datum := range interactiveCandlestickUpstreamData {
		fmt.Fprintf(hash, "%s|%g|%g|%g|%g\n", datum.Category, datum.Candle.Open, datum.Candle.Close, datum.Candle.Low, datum.Candle.High)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "03fd8007530e739fbfce31cbe3e3e2f59e174ddaece06a7a144a30ec225f3c4f" {
		t.Fatalf("normalized pinned upstream OHLC SHA-256 = %s", got)
	}
}

func TestInteractiveMapDatasetsAndVariantsMatchPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if interactiveMapUpstreamPath != "examples/map.go" || interactiveMapUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" {
		t.Fatalf("map upstream source = %s@%s", interactiveMapUpstreamPath, interactiveMapUpstreamRevision)
	}
	if interactiveMapUpstreamSHA256 != "3b59b5cb7ed392f3fa436d51fd420704ab2e82e439c95b226d35d12b913cf9da" {
		t.Fatalf("map upstream SHA-256 = %s", interactiveMapUpstreamSHA256)
	}
	wantBase := []string{"北京", "上海", "广东", "辽宁", "山东", "山西", "陕西", "新疆", "内蒙古"}
	wantGuangdong := []string{"深圳市", "广州市", "湛江市", "汕头市", "东莞市", "佛山市", "云浮市", "肇庆市", "梅州市"}
	for name, test := range map[string]struct {
		regions []interactive.MapRegion
		want    []string
	}{"base": {interactiveMapBaseRegions, wantBase}, "Guangdong": {interactiveMapGuangdongRegions, wantGuangdong}} {
		if len(test.regions) != len(test.want) {
			t.Fatalf("%s region count = %d, want %d", name, len(test.regions), len(test.want))
		}
		for index, region := range test.regions {
			if region.Name != test.want[index] || region.Value < 0 || region.Value >= 150 {
				t.Errorf("%s region %d = (%q, %g), want name %q and upstream [0,150) value domain", name, index, region.Name, region.Value, test.want[index])
			}
		}
	}
	variants := sampleInteractiveMaps()
	wantVariants := []interactive.MapVariant{interactive.MapVariantBasic, interactive.MapVariantLabels, interactive.MapVariantScale, interactive.MapVariantRegional, interactive.MapVariantTheme}
	if len(variants) != len(wantVariants) {
		t.Fatalf("map variant count = %d, want %d", len(variants), len(wantVariants))
	}
	for index := range variants {
		if variants[index].variant != wantVariants[index] {
			t.Errorf("map variant %d = %q, want %q", index, variants[index].variant, wantVariants[index])
		}
	}
}

func TestInteractiveGeoDatasetsAndVariantsMatchPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	if interactiveGeoUpstreamPath != "examples/geo.go" || interactiveGeoUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" {
		t.Fatalf("geo upstream source = %s@%s", interactiveGeoUpstreamPath, interactiveGeoUpstreamRevision)
	}
	if interactiveGeoUpstreamSHA256 != "3a6dbe86c34e5ea478b1dea5430c10cac9f7c4905264e12fc37654f0f5d4550a" {
		t.Fatalf("geo upstream SHA-256 = %s", interactiveGeoUpstreamSHA256)
	}
	wantNational := []interactive.GeoPoint{
		{Name: "北京", Longitude: 116.40, Latitude: 39.90, Value: 81},
		{Name: "上海", Longitude: 121.47, Latitude: 31.23, Value: 27},
		{Name: "重庆", Longitude: 106.55, Latitude: 29.56, Value: 47},
		{Name: "武汉", Longitude: 114.31, Latitude: 30.52, Value: 59},
		{Name: "台湾", Longitude: 121.30, Latitude: 25.03, Value: 18},
		{Name: "香港", Longitude: 114.17, Latitude: 22.28, Value: 63},
	}
	wantGuangdong := []interactive.GeoPoint{
		{Name: "汕头", Longitude: 116.69, Latitude: 23.39, Value: 12},
		{Name: "深圳", Longitude: 114.07, Latitude: 22.62, Value: 76},
		{Name: "广州", Longitude: 113.23, Latitude: 23.16, Value: 41},
	}
	if !reflect.DeepEqual(interactiveGeoNationalPoints, wantNational) {
		t.Fatalf("national points = %#v", interactiveGeoNationalPoints)
	}
	if !reflect.DeepEqual(interactiveGeoGuangdongPoints, wantGuangdong) {
		t.Fatalf("Guangdong points = %#v", interactiveGeoGuangdongPoints)
	}
	for _, points := range [][]interactive.GeoPoint{interactiveGeoNationalPoints, interactiveGeoGuangdongPoints} {
		for _, point := range points {
			if point.Value < 0 || point.Value >= 100 {
				t.Errorf("fixed literal value %q=%g outside upstream [0,100) domain", point.Name, point.Value)
			}
		}
	}
	variants := sampleInteractiveGeos()
	if len(variants) != 2 || variants[0].name != "effect-scatter" || variants[1].name != "scatter" {
		t.Fatalf("geo variants = %#v", variants)
	}
}

func TestDenseScatterValuesAreDeterministicAndPreserveUpstreamDistribution(t *testing.T) {
	t.Parallel()
	first := denseScatterValues(rand.New(rand.NewSource(20260728)), 3, 1000, 10)
	second := denseScatterValues(rand.New(rand.NewSource(20260728)), 3, 1000, 10)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fixed local seed did not reproduce dense data")
	}
	if len(first) != 3 {
		t.Fatalf("series count = %d", len(first))
	}
	for seriesIndex, series := range first {
		if len(series) != 1000 {
			t.Fatalf("series %d category count = %d", seriesIndex, len(series))
		}
		for index, samples := range series {
			want := 1
			if index > 0 && index%2 == 0 {
				want++
			}
			if index > 0 && index%10 == 0 {
				want++
			}
			if len(samples) != want {
				t.Fatalf("series %d category %d samples = %d, want %d", seriesIndex, index, len(samples), want)
			}
			if index > 0 {
				previous := series[index-1][0]
				minimum, maximum := previous*.9, previous*1.1
				for _, sample := range samples {
					if sample < minimum || sample > maximum {
						t.Fatalf("series %d category %d value %f outside 10%% walk [%f,%f]", seriesIndex, index, sample, minimum, maximum)
					}
				}
			}
		}
	}
}

func TestThemeRiverDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	streams := sampleThemeRiverStreams()
	wantNames := []string{"DQ", "TY", "SS", "QG", "SY", "DD"}
	wantValues := [][]float64{
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{35, 36, 37, 22, 24, 26, 34, 21, 18, 45, 32, 35, 30, 28, 27, 26, 15, 30, 35, 42, 42},
		{21, 25, 27, 23, 24, 21, 35, 39, 40, 36, 33, 43, 40, 34, 28, 26, 37, 41, 46, 47, 41},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 4, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 4, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
	}
	if len(streams) != len(wantNames) {
		t.Fatalf("stream count = %d, want %d", len(streams), len(wantNames))
	}
	for streamIndex, stream := range streams {
		if stream.Name != wantNames[streamIndex] {
			t.Fatalf("stream %d name = %q, want %q", streamIndex, stream.Name, wantNames[streamIndex])
		}
		if len(stream.Points) != 21 {
			t.Fatalf("stream %q point count = %d", stream.Name, len(stream.Points))
		}
		for pointIndex, point := range stream.Points {
			wantDate := time.Date(2015, time.November, 8+pointIndex, 0, 0, 0, 0, time.UTC)
			if !point.Time.Equal(wantDate) || point.Value != wantValues[streamIndex][pointIndex] {
				t.Fatalf("stream %q point %d = (%s, %g), want (%s, %g)", stream.Name, pointIndex, point.Time, point.Value, wantDate, wantValues[streamIndex][pointIndex])
			}
		}
	}
}

func TestWordCloudDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	words := sampleWordCloudWords()
	want := []struct {
		name  string
		value float64
	}{
		{"Sam S Club", 10000}, {"Macys", 6181}, {"Amy Schumer", 4386}, {"Jurassic World", 4055},
		{"Charter Communications", 2467}, {"Chick Fil A", 2244}, {"Planet Fitness", 1898},
		{"Pitch Perfect", 1484}, {"Express", 1689}, {"Home", 1112}, {"Johnny Depp", 985},
		{"Lena Dunham", 847}, {"Lewis Hamilton", 582}, {"KXAN", 555}, {"Mary Ellen Mark", 550},
		{"Farrah Abraham", 462}, {"Rita Ora", 366}, {"Serena Williams", 282},
		{"NCAA baseball tournament", 273}, {"Point Break", 265},
	}
	if len(words) != len(want) {
		t.Fatalf("word count = %d, want %d", len(words), len(want))
	}
	for index, word := range words {
		if word.Name != want[index].name || word.Value != want[index].value {
			t.Fatalf("word %d = (%q, %g), want (%q, %g)", index, word.Name, word.Value, want[index].name, want[index].value)
		}
	}
}

func TestLiquidGaugeDataMechanicallyMatchesPinnedUpstreamExample(t *testing.T) {
	t.Parallel()
	data := sampleLiquidGaugeData()
	want := []float64{.3, .4, .5}
	if len(data) != len(want) {
		t.Fatalf("liquid reading count = %d, want %d", len(data), len(want))
	}
	for index := range want {
		if data[index].Name != fmt.Sprintf("Wave %d", index+1) || data[index].Value != want[index] {
			t.Fatalf("liquid reading %d = (%q, %g), want (Wave %d, %g)", index, data[index].Name, data[index].Value, index+1, want[index])
		}
	}
}
