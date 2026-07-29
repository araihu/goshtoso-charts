package table

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	chart "github.com/go-analyze/charts"
)

var (
	tokenSurface    = chart.ColorRGB(1, 1, 1)
	tokenSurfaceAlt = chart.ColorRGB(2, 2, 2)
	tokenText       = chart.ColorRGB(3, 3, 3)
	tokenTextStrong = chart.ColorRGB(4, 4, 4)
)

// Instance is a renderable server-side table.
type Instance struct{ cfg Config }

// Table returns a server-side SVG table.
func Table(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a static table.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindTable }

// Render writes an accessible figure, SSR SVG, and adjacent semantic HTML table.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	content := tableTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, content).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	painter, err := chart.TableOptionRenderDirect(tableOptions(cfg))
	if err != nil {
		return "", fmt.Errorf("render table: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode table SVG: %w", err)
	}
	return tokenizedSVG(string(data)), nil
}

func tableOptions(cfg Config) chart.TableChartOption {
	columns := len(cfg.Columns)
	header := make([]string, columns)
	spans := make([]int, columns)
	alignments := make([]string, columns)
	for index, column := range cfg.Columns {
		header[index] = column.Header
		spans[index] = column.Span
		if spans[index] == 0 {
			spans[index] = 1
		}
		switch column.Align {
		case AlignCenter:
			alignments[index] = chart.AlignCenter
		case AlignEnd:
			alignments[index] = chart.AlignRight
		}
	}
	padding := cfg.padding()
	options := chart.TableChartOption{
		OutputFormat:          chart.ChartOutputSVG,
		Width:                 cfg.width(),
		Header:                header,
		Data:                  cfg.Rows,
		Spans:                 spans,
		TextAligns:            alignments,
		Padding:               chart.Box{Top: padding.Top, Right: padding.Right, Bottom: padding.Bottom, Left: padding.Left},
		BackgroundColor:       colorOrDefault(cfg.Colors.Surface, tokenSurface),
		HeaderBackgroundColor: colorOrDefault(cfg.Colors.HeaderBackground, tokenSurfaceAlt),
		HeaderFontColor:       colorOrDefault(cfg.Colors.HeaderText, tokenTextStrong),
		FontStyle:             chart.FontStyle{FontSize: cfg.FontSize, FontColor: colorOrDefault(cfg.Colors.Text, tokenText)},
	}
	if len(cfg.Colors.RowBackgrounds) == 0 {
		options.RowBackgroundColors = []chart.Color{tokenSurface, tokenSurfaceAlt}
	} else {
		options.RowBackgroundColors = make([]chart.Color, len(cfg.Colors.RowBackgrounds))
		for index, color := range cfg.Colors.RowBackgrounds {
			options.RowBackgroundColors[index] = colorOrDefault(color, tokenSurface)
		}
	}
	if cfg.CellStyle != nil {
		options.CellModifier = func(cell chart.TableCell) chart.TableCell {
			if cell.Row == 0 {
				return cell
			}
			appearance := cfg.CellStyle(Cell{Row: cell.Row - 1, Column: cell.Column, Value: cell.Text})
			if strings.TrimSpace(appearance.BackgroundColor) != "" {
				cell.FillColor = parseColor(appearance.BackgroundColor)
			}
			if strings.TrimSpace(appearance.TextColor) != "" {
				cell.FontStyle.FontColor = parseColor(appearance.TextColor)
			}
			return cell
		}
	}
	return options
}

func colorOrDefault(value string, fallback chart.Color) chart.Color {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return parseColor(value)
}

func parseColor(value string) chart.Color {
	if strings.EqualFold(strings.TrimSpace(value), "transparent") {
		return chart.ColorTransparent
	}
	return chart.ParseColor(strings.TrimSpace(value))
}

func tokenizedSVG(svg string) string {
	return strings.NewReplacer(
		"rgb(1,1,1)", "var(--color-chart-surface, var(--color-surface, #ffffff))",
		"rgb(2,2,2)", "var(--color-chart-surface-alt, var(--color-surface-alt, #f3f4f6))",
		"rgb(3,3,3)", "var(--color-chart-text, var(--color-on-surface, #111827))",
		"rgb(4,4,4)", "var(--color-chart-text-strong, var(--color-on-surface-strong, #111827))",
		"'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif",
	).Replace(svg)
}

var _ chartcomponents.Component = Instance{}
