package heartbeat

import (
	"context"
	"fmt"
	"io"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

const chartWidth = 600

// Instance is a renderable availability heartbeat chart.
type Instance struct {
	cfg Config
}

// Heartbeat returns a renderable availability history chart.
func Heartbeat(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a heartbeat chart.
func (Instance) Kind() chartcomponents.Kind {
	return chartcomponents.KindHeartbeat
}

// Render writes the chart markup without client-side runtime or hydration.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	if err := instance.cfg.validate(); err != nil {
		return err
	}
	return heartbeatTemplate(instance.cfg, makeRenderPoints(instance.cfg.Points, instance.cfg.Style), instance.summary()).Render(ctx, writer)
}

type renderPoint struct {
	X     int
	Width int
	Fill  string
	Title string
}

func makeRenderPoints(points []Point, style charttheme.Style) []renderPoint {
	if len(points) == 0 {
		return nil
	}
	gap := 1
	available := chartWidth - gap*(len(points)-1)
	width := max(1, available/len(points))
	result := make([]renderPoint, 0, len(points))
	for index, point := range points {
		result = append(result, renderPoint{
			X:     index * (width + gap),
			Width: width,
			Fill:  point.fill(style),
			Title: point.title(),
		})
	}
	return result
}

func (point Point) fill(style charttheme.Style) string {
	switch point.State {
	case StateUp:
		return stateColor(style, 0, "var(--color-success)")
	case StateDegraded:
		return stateColor(style, 1, "var(--color-warning)")
	case StateDown:
		return stateColor(style, 2, "var(--color-danger)")
	default:
		return stateColor(style, 3, "var(--color-outline)")
	}
}

func stateColor(style charttheme.Style, index int, fallback string) string {
	if index < len(style.Colors) && style.Colors[index] != "" {
		return style.Colors[index]
	}
	return fallback
}

func (point Point) title() string {
	title := fmt.Sprintf("%s: %s", point.At.UTC().Format(time.RFC3339), point.State)
	if point.Latency > 0 {
		title += fmt.Sprintf(" (%s)", point.Latency.Round(time.Millisecond))
	}
	return title
}

func (instance Instance) summary() string {
	if instance.cfg.Caption != "" {
		return instance.cfg.Caption
	}
	if len(instance.cfg.Points) == 0 {
		return "No monitoring data in this period."
	}
	counts := map[State]int{}
	for _, point := range instance.cfg.Points {
		counts[point.State]++
	}
	return fmt.Sprintf("%d checks: %d up, %d degraded, %d down, %d unknown.", len(instance.cfg.Points), counts[StateUp], counts[StateDegraded], counts[StateDown], counts[StateUnknown])
}

var _ chartcomponents.Component = Instance{}
