package pages

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
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
