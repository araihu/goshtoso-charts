package heartbeat

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
)

func TestHeartbeatRendersSemanticStatusHistory(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)
	instance := Heartbeat(Config{
		Label: "API availability over last three checks",
		Points: []Point{
			{At: base, State: StateUp, Latency: 42 * time.Millisecond},
			{At: base.Add(time.Minute), State: StateDegraded},
			{At: base.Add(2 * time.Minute), State: StateDown},
		},
	})
	if instance.Kind() != chartcomponents.KindHeartbeat {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindHeartbeat)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<svg", "var(--color-success)", "var(--color-warning)", "var(--color-danger)", "2026-07-27T12:00:00Z: up (42ms)", "3 checks: 1 up, 1 degraded, 1 down, 0 unknown."} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q:\n%s", want, markup)
		}
	}
}

func TestHeartbeatSupportsNoData(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := Heartbeat(Config{Label: "API availability"}).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(output.String(), "No monitoring data in this period.") {
		t.Fatalf("no-data state missing: %s", output.String())
	}
}

func TestHeartbeatRejectsInvalidPointOrder(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)
	err := Heartbeat(Config{Label: "API availability", Points: []Point{
		{At: base.Add(time.Minute), State: StateUp},
		{At: base, State: StateDown},
	}}).Render(context.Background(), &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "chronological order") {
		t.Fatalf("Render() error = %v, want chronological-order error", err)
	}
}

func TestHeartbeatRejectsTooManyPoints(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)
	points := make([]Point, maxPoints+1)
	for index := range points {
		points[index] = Point{At: base.Add(time.Duration(index) * time.Minute), State: StateUp}
	}
	err := Heartbeat(Config{Label: "API availability", Points: points}).Render(context.Background(), &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "maximum") {
		t.Fatalf("Render() error = %v, want maximum-points error", err)
	}
}
