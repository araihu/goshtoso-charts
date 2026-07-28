package pages

import (
	"math/rand"
	"reflect"
	"testing"
)

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
