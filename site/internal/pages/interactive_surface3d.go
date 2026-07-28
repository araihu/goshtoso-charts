package pages

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveSurface3DUpstreamPath     = "examples/surface3d.go"
	interactiveSurface3DUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveSurface3DUpstreamSHA256   = "51ffdac86403c0e6430c0134d1062e5fda9865b7d88061ce4f38129113c455d9"

	interactiveSurface3DBaseFormula = "y = i / 60, x = j / 60, z = sin(x × π) × sin(y × π), with i and j from -60 through 59"
	interactiveSurface3DRoseFormula = "y = i / 10, x = j / 10, z = sin(x² + y²) × x / π, with i and j from -30 through 29"
)

var (
	interactiveSurface3DBasePoints = generateInteractiveSurface3DBasePoints()
	interactiveSurface3DRosePoints = generateInteractiveSurface3DRosePoints()
)

func generateInteractiveSurface3DBasePoints() []interactive.Point3D {
	points := make([]interactive.Point3D, 0, 120*120)
	for i := -60; i < 60; i++ {
		y := float64(i) / 60
		for j := -60; j < 60; j++ {
			x := float64(j) / 60
			points = append(points, interactive.Point3D{
				X: x,
				Y: y,
				Z: math.Sin(x*math.Pi) * math.Sin(y*math.Pi),
			})
		}
	}
	return points
}

func generateInteractiveSurface3DRosePoints() []interactive.Point3D {
	points := make([]interactive.Point3D, 0, 60*60)
	for i := -30; i < 30; i++ {
		y := float64(i) / 10
		for j := -30; j < 30; j++ {
			x := float64(j) / 10
			points = append(points, interactive.Point3D{
				X: x,
				Y: y,
				Z: math.Sin(x*x+y*y) * x / math.Pi,
			})
		}
	}
	return points
}

func interactiveSurface3DDataHash(points []interactive.Point3D) string {
	values := make([][3]float64, len(points))
	for index, point := range points {
		values[index] = [3]float64{point.X, point.Y, point.Z}
	}
	encoded, _ := json.Marshal(values)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func sampleInteractiveSurface3D(label, formula, filename string, points []interactive.Point3D) interactive.Instance {
	return interactive.Surface3D(interactive.Surface3DConfig{
		Label:   label,
		Caption: formula,
		Series: []interactive.Surface3DSeries{{
			Name: "surface3d", Points: append([]interactive.Point3D(nil), points...),
		}},
		VisualRange: &interactive.Surface3DVisualRange{
			Min: -3, Max: 3, Calculable: interactive.Bool(true), Palette: interactive.Surface3DPaletteColdToWarm,
		},
		DataSummary: interactive.Surface3DDataSummary{Formula: formula},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: label},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: filename},
		},
	})
}

func interactiveSurface3DCode() string {
	return `@interactive.Surface3D(interactive.Surface3DConfig{
  Label: "Mathematical surface",
  Series: []interactive.Surface3DSeries{{
    Name: "surface",
    Points: []interactive.Point3D{{X: -1, Y: -1, Z: 0}, {X: 0, Y: -1, Z: 0}},
  }},
  VisualRange: &interactive.Surface3DVisualRange{
    Min: -3, Max: 3, Calculable: interactive.Bool(true),
    Palette: interactive.Surface3DPaletteColdToWarm,
  },
  DataSummary: interactive.Surface3DDataSummary{
    Formula: "z = sin(x × π) × sin(y × π)",
  },
})`
}
