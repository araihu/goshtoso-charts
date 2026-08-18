package table

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	chart "github.com/go-analyze/charts"
)

func upstreamConfig() Config {
	return Config{
		Label: "People directory",
		Columns: []Column{
			{Header: "Name", Span: 2}, {Header: "Age", Span: 1}, {Header: "Address", Span: 3},
			{Header: "Tag", Span: 2}, {Header: "Action", Span: 2},
		},
		Rows: [][]string{
			{"John Brown", "32", "New York No. 1 Lake Park", "nice, developer", "Send Mail"},
			{"Jim Green", "42", "London No. 1 Lake Park", "wow", "Send Mail"},
			{"Joe Black", "32", "Sidney No. 1 Lake Park", "cool, teacher", "Send Mail"},
		},
	}
}

func TestTableRendersUpstreamShapeAccessibleFallbackAndSharedControls(t *testing.T) {
	t.Parallel()
	cfg := upstreamConfig()
	cfg.Caption = "Three contacts and available actions."
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "mx-auto"}
	cfg.RootAttrs = templ.Attributes{"id": "people", "data-table-purpose": "directory"}
	cfg.Controls = chartcontrol.Options{Fullscreen: true}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "people-directory"}

	instance := Table(cfg)
	if instance.Kind() != chartcomponents.KindTable {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`<figure class="goshtoso-charts-table goshtoso-charts-palette goshtoso-charts-palette-araihu mx-auto" role="img" aria-label="People directory"`,
		`id="people"`, `data-table-purpose="directory"`, `<svg`, `viewBox="0 0 810 `,
		"John Brown", "32", "New York No. 1 Lake Park", "nice, developer", "Send Mail",
		"Jim Green", "42", "London No. 1 Lake Park", "wow", "Joe Black", "Sidney No. 1 Lake Park", "cool, teacher",
		"Three contacts and available actions.", "Accessible data table", `aria-label="People directory data"`,
		"var(--color-chart-surface", "var(--color-chart-surface-alt", "var(--color-chart-text",
		`data-goshtoso-chart-expand`, `-chart-expand-export"`, `<span class="block">Download SVG</span>`, `<span class="block">Download PNG</span>`,
		`-fullscreen-action`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"rgb(1,1,1)", "rgb(2,2,2)", "rgb(3,3,3)", "rgb(4,4,4)"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains untokenized %q", unwanted)
		}
	}
}

func TestTableMapsWidthsPaddingAlignmentAndCellAppearance(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label: "Market changes", Width: 810,
		Columns: []Column{{Header: "Name"}, {Header: "Price", Align: AlignEnd}, {Header: "Change", Span: 2, Align: AlignEnd}},
		Rows:    [][]string{{"Datadog Inc", "97.32", "-7.49%"}, {"Hashicorp Inc", "28.66", "-9.25%"}, {"Gitlab Inc", "51.63", "+4.32%"}},
		Padding: &Padding{Top: 15, Right: 10, Bottom: 15, Left: 10},
		Colors:  Colors{Surface: "#1c1c20", HeaderBackground: "#505050", HeaderText: "#ffffff", Text: "#ffffff", RowBackgrounds: []string{"#1c1c20"}},
		CellStyle: func(cell Cell) CellAppearance {
			if cell.Column != 2 {
				return CellAppearance{}
			}
			if strings.HasPrefix(cell.Value, "+") {
				return CellAppearance{BackgroundColor: "#b33514"}
			}
			return CellAppearance{BackgroundColor: "#217c32"}
		},
	}
	options := tableOptions(cfg)
	if options.Width != 810 || options.Spans[2] != 2 || options.TextAligns[1] != "right" || options.TextAligns[2] != "right" {
		t.Fatalf("table layout options = %#v", options)
	}
	if options.Padding.Top != 15 || options.Padding.Right != 10 || options.Padding.Bottom != 15 || options.Padding.Left != 10 {
		t.Fatalf("padding = %#v", options.Padding)
	}
	negative := options.CellModifier(tableCell(1, 2, "-7.49%"))
	positive := options.CellModifier(tableCell(3, 2, "+4.32%"))
	if negative.FillColor.R != 33 || negative.FillColor.G != 124 || negative.FillColor.B != 50 {
		t.Fatalf("negative fill = %#v", negative.FillColor)
	}
	if positive.FillColor.R != 179 || positive.FillColor.G != 53 || positive.FillColor.B != 20 {
		t.Fatalf("positive fill = %#v", positive.FillColor)
	}
}

func tableCell(row, column int, text string) chart.TableCell {
	return chart.TableCell{Row: row, Column: column, Text: text}
}

func TestTableDefaultsMatchUpstreamExample(t *testing.T) {
	t.Parallel()
	cfg := upstreamConfig()
	if cfg.width() != 810 {
		t.Fatalf("width = %d, want upstream 810", cfg.width())
	}
	if got := cfg.padding(); got != (Padding{Top: 15, Right: 10, Bottom: 15, Left: 10}) {
		t.Fatalf("padding = %#v", got)
	}
}

func TestTableValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "label", edit: func(c *Config) { c.Label = "" }, want: "label is required"},
		{name: "columns", edit: func(c *Config) { c.Columns = nil }, want: "at least one column"},
		{name: "rows", edit: func(c *Config) { c.Rows = nil }, want: "at least one row"},
		{name: "header", edit: func(c *Config) { c.Columns[0].Header = "" }, want: "column 1 needs a header"},
		{name: "span", edit: func(c *Config) { c.Columns[0].Span = -1 }, want: "span cannot be negative"},
		{name: "alignment", edit: func(c *Config) { c.Columns[0].Align = "justify" }, want: "alignment \"justify\" is unsupported"},
		{name: "row width", edit: func(c *Config) { c.Rows[0] = c.Rows[0][:4] }, want: "row 1 has 4 cells; want 5"},
		{name: "width", edit: func(c *Config) { c.Width = -1 }, want: "width cannot be negative"},
		{name: "font size", edit: func(c *Config) { c.FontSize = -1 }, want: "font size cannot be negative"},
		{name: "padding", edit: func(c *Config) { c.Padding = &Padding{Left: -1} }, want: "padding cannot be negative"},
		{name: "root attr", edit: func(c *Config) { c.RootAttrs = templ.Attributes{"ARIA-Label": "override"} }, want: `root attribute "ARIA-Label" is reserved`},
		{name: "surface color", edit: func(c *Config) { c.Colors.Surface = "var(--unsafe)" }, want: `surface color "var(--unsafe)" is unsupported`},
		{name: "row color", edit: func(c *Config) { c.Colors.RowBackgrounds = []string{"no-such-color"} }, want: `row background 1 color "no-such-color" is unsupported`},
		{name: "cell background", edit: func(c *Config) {
			c.CellStyle = func(Cell) CellAppearance { return CellAppearance{BackgroundColor: "bogus"} }
		}, want: `cell 1,1 background color "bogus" is unsupported`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := upstreamConfig()
			cfg.Columns = append([]Column(nil), cfg.Columns...)
			cfg.Rows = cloneRows(cfg.Rows)
			test.edit(&cfg)
			err := cfg.validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func cloneRows(rows [][]string) [][]string {
	clone := make([][]string, len(rows))
	for index := range rows {
		clone[index] = append([]string(nil), rows[index]...)
	}
	return clone
}
