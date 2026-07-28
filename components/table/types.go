// Package table renders accessible server-side SVG data tables.
package table

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var supportedColor = regexp.MustCompile(`(?i)^(?:#[0-9a-f]{3,4}|#[0-9a-f]{6}|#[0-9a-f]{8}|rgba?\(\s*[0-9.]+\s*,\s*[0-9.]+\s*,\s*[0-9.]+(?:\s*,\s*[0-9.]+)?\s*\))$`)

// Alignment controls renderer-neutral horizontal cell alignment.
type Alignment string

const (
	// AlignStart uses the reading-direction start edge.
	AlignStart Alignment = ""
	// AlignCenter centers cell content.
	AlignCenter Alignment = "center"
	// AlignEnd uses the reading-direction end edge.
	AlignEnd Alignment = "end"
)

// Column describes one table column. Span is a relative width weight and defaults to one.
type Column struct {
	Header string
	Span   int
	Align  Alignment
}

// Padding controls cell insets in pixels.
type Padding struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// Colors overrides semantic table colors with CSS hex or rgb/rgba values.
// Empty fields use Goshtoso chart-theme tokens.
type Colors struct {
	Surface          string
	HeaderBackground string
	HeaderText       string
	Text             string
	RowBackgrounds   []string
}

// Cell describes a data cell passed to a renderer-neutral styling function.
type Cell struct {
	Row    int
	Column int
	Value  string
}

// CellAppearance customizes one data cell. Empty fields preserve row and text defaults.
type CellAppearance struct {
	BackgroundColor string
	TextColor       string
}

// CellStyler returns presentation for one data cell.
type CellStyler func(Cell) CellAppearance

// Config describes an SSR SVG table.
type Config struct {
	Label   string
	Caption string
	Columns []Column
	Rows    [][]string
	Width   int
	// Padding defaults to the upstream example's 15px vertical and 10px horizontal insets.
	Padding   *Padding
	FontSize  float64
	Colors    Colors
	CellStyle CellStyler
	Style     charttheme.Style
	RootAttrs templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen and collapse default off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("table label is required")
	}
	if len(cfg.Columns) == 0 {
		return fmt.Errorf("table needs at least one column")
	}
	if len(cfg.Rows) == 0 {
		return fmt.Errorf("table needs at least one row")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("table width cannot be negative")
	}
	if cfg.FontSize < 0 {
		return fmt.Errorf("table font size cannot be negative")
	}
	if cfg.Padding != nil && (cfg.Padding.Top < 0 || cfg.Padding.Right < 0 || cfg.Padding.Bottom < 0 || cfg.Padding.Left < 0) {
		return fmt.Errorf("table padding cannot be negative")
	}
	for index, column := range cfg.Columns {
		if strings.TrimSpace(column.Header) == "" {
			return fmt.Errorf("table column %d needs a header", index+1)
		}
		if column.Span < 0 {
			return fmt.Errorf("table column %q span cannot be negative", column.Header)
		}
		switch column.Align {
		case AlignStart, AlignCenter, AlignEnd:
		default:
			return fmt.Errorf("table column %q alignment %q is unsupported", column.Header, column.Align)
		}
	}
	for rowIndex, row := range cfg.Rows {
		if len(row) != len(cfg.Columns) {
			return fmt.Errorf("table row %d has %d cells; want %d", rowIndex+1, len(row), len(cfg.Columns))
		}
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("table root attribute %q is reserved", attribute)
			}
		}
	}
	for name, color := range map[string]string{
		"surface": cfg.Colors.Surface, "header background": cfg.Colors.HeaderBackground,
		"header text": cfg.Colors.HeaderText, "text": cfg.Colors.Text,
	} {
		if err := validateColor(name, color); err != nil {
			return err
		}
	}
	for index, color := range cfg.Colors.RowBackgrounds {
		if err := validateColor(fmt.Sprintf("row background %d", index+1), color); err != nil {
			return err
		}
	}
	if cfg.CellStyle != nil {
		for rowIndex, row := range cfg.Rows {
			for columnIndex, value := range row {
				appearance := cfg.CellStyle(Cell{Row: rowIndex, Column: columnIndex, Value: value})
				if err := validateColor(fmt.Sprintf("cell %d,%d background", rowIndex+1, columnIndex+1), appearance.BackgroundColor); err != nil {
					return err
				}
				if err := validateColor(fmt.Sprintf("cell %d,%d text", rowIndex+1, columnIndex+1), appearance.TextColor); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateColor(name, value string) error {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "transparent") {
		return nil
	}
	if !supportedColor.MatchString(value) {
		return fmt.Errorf("table %s color %q is unsupported", name, value)
	}
	return nil
}

func (cfg Config) width() int {
	if cfg.Width > 0 {
		return cfg.Width
	}
	return 810
}

func (cfg Config) padding() Padding {
	if cfg.Padding != nil {
		return *cfg.Padding
	}
	return Padding{Top: 15, Right: 10, Bottom: 15, Left: 10}
}
