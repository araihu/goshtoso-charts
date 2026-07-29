package interactive

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCartesianSnapshotHasRendererNeutralJSONContract(t *testing.T) {
	t.Parallel()
	payload, err := json.Marshal(CartesianSnapshot{
		Categories: []string{"12:00", "12:01"},
		Series:     []CartesianSnapshotSeries{{Name: "Healthy", Values: []float64{1, 0}}},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if got, want := string(payload), `{"categories":["12:00","12:01"],"series":[{"name":"Healthy","values":[1,0]}]}`; got != want {
		t.Fatalf("snapshot JSON = %s, want %s", got, want)
	}
}

func TestBarRendersSSELiveDataContract(t *testing.T) {
	t.Parallel()
	instance := Bar(BarConfig{
		Label: "Live availability", XAxis: []string{"12:00"},
		Series: []BarSeries{{Name: "Healthy", Data: []BarData{{Value: 1}}}},
		Live:   &LiveData{URL: "/examples/live-availability/events", Event: "chart"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`data-goshtoso-charts-live-url="/examples/live-availability/events"`,
		`data-goshtoso-charts-live-event="chart"`,
		`data-goshtoso-charts-live-shape="cartesian"`,
		`data-goshtoso-charts-live-runtime`, `new EventSource`,
		`payload.categories`, `payload.series`, `configuredByName`, `existing.id`,
		`animationDurationUpdate: 0`, `series: updates`, `source.close()`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("live bar markup missing %q", want)
		}
	}
}

func TestLiveRuntimePreservesConfiguredSeriesIdentity(t *testing.T) {
	t.Parallel()
	for _, unwanted := range []string{`replaceMerge`, `type: existing.type`} {
		if strings.Contains(liveRuntimeMarkup, unwanted) {
			t.Errorf("live runtime must not contain %q", unwanted)
		}
	}
	for _, want := range []string{
		`configured.length !== payload.series.length`,
		`configuredByName.has(configuredSeries.name)`,
		`configuredByName.get(snapshotSeries.name)`,
		`seen.has(snapshotSeries.name)`,
		`id: existing.id`,
		`data: snapshotSeries.values`,
		`chart.setOption({`,
		`xAxis: [{ data: payload.categories }]`,
	} {
		if !strings.Contains(liveRuntimeMarkup, want) {
			t.Errorf("identity-preserving live runtime missing %q", want)
		}
	}
	if strings.Contains(liveRuntimeMarkup, `axisLabel`) {
		t.Error("live snapshot merge must preserve configured x-axis label settings")
	}
}

func TestLineSupportsSameLiveDataPrimitive(t *testing.T) {
	t.Parallel()
	instance := Line(LineConfig{
		Label: "Live latency", XAxis: []string{"12:00"},
		Series: []LineSeries{{Name: "Latency", Data: []LineData{{Value: 42}}}},
		Live:   &LiveData{URL: "/latency/events"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{`data-goshtoso-charts-live-url="/latency/events"`, `data-goshtoso-charts-live-event="message"`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("live line markup missing %q", want)
		}
	}
}

func TestLiveDataRequiresURL(t *testing.T) {
	t.Parallel()
	instance := Bar(BarConfig{
		Label: "Live availability", XAxis: []string{"12:00"},
		Series: []BarSeries{{Name: "Healthy", Data: []BarData{{Value: 1}}}},
		Live:   &LiveData{},
	})
	var output bytes.Buffer
	err := instance.Render(context.Background(), &output)
	if err == nil || err.Error() != "live data URL is required" {
		t.Fatalf("Render() error = %v, want live data URL is required", err)
	}
}
