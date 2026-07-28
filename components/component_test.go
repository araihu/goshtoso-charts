package components

import "testing"

func TestAllKinds(t *testing.T) {
	t.Parallel()
	want := []Kind{KindLineChart, KindBarChart, KindPieChart, KindScatterChart, KindRadarChart, KindCandlestickChart, KindInteractiveBar, KindInteractiveLine, KindInteractiveScatter, KindInteractivePie, KindInteractiveRadar, KindInteractiveHeatMap, KindInteractiveBoxPlot, KindInteractiveGauge, KindInteractiveFunnel, KindInteractiveGraph, KindInteractiveSankey, KindInteractiveTree, KindInteractiveSunburst, KindInteractiveTreemap}
	got := AllKinds()
	if len(got) != len(want) {
		t.Fatalf("AllKinds() = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("AllKinds()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
	got[0] = "changed"
	if AllKinds()[0] != KindLineChart {
		t.Fatal("AllKinds() leaked internal slice")
	}
}
