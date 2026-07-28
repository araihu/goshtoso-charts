package components

import "testing"

func TestAllKinds(t *testing.T) {
	t.Parallel()
	want := []Kind{KindHeartbeat, KindLineChart, KindBarChart, KindPieChart, KindInteractiveBar, KindInteractiveLine, KindInteractiveScatter, KindInteractivePie, KindInteractiveRadar, KindInteractiveHeatMap, KindInteractiveBoxPlot, KindInteractiveGauge, KindInteractiveFunnel}
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
	if AllKinds()[0] != KindHeartbeat {
		t.Fatal("AllKinds() leaked internal slice")
	}
}
