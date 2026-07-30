package chart_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
	interactiveboxplot "github.com/araihu/goshtoso-charts/components/interactive/boxplot"
	interactivecandlestick "github.com/araihu/goshtoso-charts/components/interactive/candlestick"
	interactiveheatmap "github.com/araihu/goshtoso-charts/components/interactive/heatmap"
	interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"
	interactivepie "github.com/araihu/goshtoso-charts/components/interactive/pie"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

const uninitializedRenderError = "interactive chart label is required"

type contextKey struct{}

type probeComponent struct {
	kind          chartcomponents.Kind
	markup        string
	err           error
	renders       int
	contextValue  string
	receivedWrite io.Writer
}

func (component *probeComponent) Kind() chartcomponents.Kind {
	if component == nil {
		panic("typed-nil component Kind called")
	}
	return component.kind
}

func (component *probeComponent) Render(ctx context.Context, writer io.Writer) error {
	if component == nil {
		panic("typed-nil component Render called")
	}
	component.renders++
	component.contextValue, _ = ctx.Value(contextKey{}).(string)
	component.receivedWrite = writer
	if component.err != nil {
		return component.err
	}
	_, err := io.WriteString(writer, component.markup)
	return err
}

func TestNewInstanceDelegatesComponentContract(t *testing.T) {
	t.Parallel()

	delegate := &probeComponent{
		kind:   chartcomponents.Kind("custom-chart"),
		markup: `<figure data-chart="custom"></figure>`,
	}
	instance := chart.NewInstance(delegate)

	if got := instance.Kind(); got != delegate.kind {
		t.Fatalf("Kind() = %q, want %q", got, delegate.kind)
	}
	ctx := context.WithValue(context.Background(), contextKey{}, "consumer-context")
	var output bytes.Buffer
	if err := instance.Render(ctx, &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := output.String(); got != delegate.markup {
		t.Fatalf("Render() output = %q, want %q", got, delegate.markup)
	}
	if delegate.renders != 1 {
		t.Fatalf("delegate renders = %d, want 1", delegate.renders)
	}
	if delegate.contextValue != "consumer-context" {
		t.Fatalf("delegate context value = %q, want consumer-context", delegate.contextValue)
	}
	if delegate.receivedWrite != &output {
		t.Fatal("Render() did not preserve writer identity")
	}
}

func TestNewInstancePreservesDelegateError(t *testing.T) {
	t.Parallel()

	want := errors.New("custom render failed")
	instance := chart.NewInstance(&probeComponent{
		kind: chartcomponents.Kind("custom-chart"),
		err:  want,
	})

	if got := instance.Render(context.Background(), io.Discard); got != want {
		t.Fatalf("Render() error = %v, want exact delegate error %v", got, want)
	}
}

func TestInstanceNilAndZeroContract(t *testing.T) {
	t.Parallel()

	var typedNil *probeComponent
	tests := map[string]chart.Instance{
		"zero":      {},
		"nil":       chart.NewInstance(nil),
		"typed nil": chart.NewInstance(typedNil),
	}
	for name, instance := range tests {
		name, instance := name, instance
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := instance.Kind(); got != "" {
				t.Fatalf("Kind() = %q, want empty kind", got)
			}
			var output bytes.Buffer
			err := instance.Render(context.Background(), &output)
			if err == nil || err.Error() != uninitializedRenderError {
				t.Fatalf("Render() error = %v, want %q", err, uninitializedRenderError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %q for uninitialized instance", output.String())
			}
		})
	}
}

func TestNewInstanceIsIdempotentForInstance(t *testing.T) {
	t.Parallel()

	delegate := &probeComponent{kind: chartcomponents.Kind("custom-chart"), markup: "ok"}
	once := chart.NewInstance(delegate)
	twice := chart.NewInstance(once)

	if twice != once {
		t.Fatal("NewInstance(existing Instance) nested instead of returning same instance")
	}
	var output bytes.Buffer
	if err := twice.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if delegate.renders != 1 || output.String() != "ok" {
		t.Fatalf("double-wrapped render = (%d calls, %q), want (1 call, %q)", delegate.renders, output.String(), "ok")
	}
}

func TestInstanceAliasesRetainCanonicalTypeIdentity(t *testing.T) {
	t.Parallel()

	wantPackage := "github.com/araihu/goshtoso-charts/components/chart"
	types := map[string]reflect.Type{
		"chart":                   reflect.TypeOf(chart.Instance{}),
		"interactive":             reflect.TypeOf(interactive.Instance{}),
		"interactive/bar":         reflect.TypeOf(interactivebar.Instance{}),
		"interactive/boxplot":     reflect.TypeOf(interactiveboxplot.Instance{}),
		"interactive/candlestick": reflect.TypeOf(interactivecandlestick.Instance{}),
		"interactive/heatmap":     reflect.TypeOf(interactiveheatmap.Instance{}),
		"interactive/line":        reflect.TypeOf(interactiveline.Instance{}),
		"interactive/pie":         reflect.TypeOf(interactivepie.Instance{}),
		"interactive/scatter":     reflect.TypeOf(interactivescatter.Instance{}),
	}
	for name, instanceType := range types {
		if got := instanceType.PkgPath(); got != wantPackage {
			t.Errorf("%s Instance PkgPath() = %q, want canonical %q", name, got, wantPackage)
		}
		if instanceType != reflect.TypeOf(chart.Instance{}) {
			t.Errorf("%s Instance is not identical to chart.Instance", name)
		}
	}
}

var (
	_ chart.Instance                  = interactive.Instance{}
	_ interactive.Instance            = chart.Instance{}
	_ chart.Instance                  = interactivebar.Instance{}
	_ interactivebar.Instance         = chart.Instance{}
	_ chart.Instance                  = interactiveboxplot.Instance{}
	_ interactiveboxplot.Instance     = chart.Instance{}
	_ chart.Instance                  = interactivecandlestick.Instance{}
	_ interactivecandlestick.Instance = chart.Instance{}
	_ chart.Instance                  = interactiveheatmap.Instance{}
	_ interactiveheatmap.Instance     = chart.Instance{}
	_ chart.Instance                  = interactiveline.Instance{}
	_ interactiveline.Instance        = chart.Instance{}
	_ chart.Instance                  = interactivepie.Instance{}
	_ interactivepie.Instance         = chart.Instance{}
	_ chart.Instance                  = interactivescatter.Instance{}
	_ interactivescatter.Instance     = chart.Instance{}

	_ func(chartcomponents.Component) chart.Instance     = chart.NewInstance
	_ func(interactive.BarConfig) chart.Instance         = interactive.Bar
	_ func(interactive.BoxPlotConfig) chart.Instance     = interactive.BoxPlot
	_ func(interactive.LineConfig) chart.Instance        = interactive.Line
	_ func(interactive.ScatterConfig) chart.Instance     = interactive.Scatter
	_ func(interactive.CandlestickConfig) chart.Instance = interactive.Candlestick
	_ func(interactive.HeatMapConfig) chart.Instance     = interactive.HeatMap
	_ func(interactive.PieConfig) chart.Instance         = interactive.Pie
	_ func(interactivebar.Config) chart.Instance         = interactivebar.Bar
	_ func(interactiveboxplot.Config) chart.Instance     = interactiveboxplot.BoxPlot
	_ func(interactivecandlestick.Config) chart.Instance = interactivecandlestick.Candlestick
	_ func(interactiveheatmap.Config) chart.Instance     = interactiveheatmap.HeatMap
	_ func(interactiveline.Config) chart.Instance        = interactiveline.Line
	_ func(interactivepie.Config) chart.Instance         = interactivepie.Pie
	_ func(interactivescatter.Config) chart.Instance     = interactivescatter.Scatter
)
