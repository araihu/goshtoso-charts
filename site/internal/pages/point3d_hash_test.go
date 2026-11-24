package pages

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/araihu/goshtoso-charts/components/interactive"
)

// Go's trigonometric functions can differ by a few ULPs across architectures.
// Hash ordered triples in billionths, matching the browser's dataset check.
func checkPoint3DHash(t *testing.T, name string, points []interactive.Point3D) {
	t.Helper()
	golden, err := os.ReadFile("testdata/point3d-hashes.json")
	if err != nil {
		t.Fatal(err)
	}
	var hashes map[string]string
	if err := json.Unmarshal(golden, &hashes); err != nil {
		t.Fatal(err)
	}
	triples := make([][3]int64, len(points))
	for i, point := range points {
		for axis, coordinate := range [3]float64{point.X, point.Y, point.Z} {
			if math.IsNaN(coordinate) || math.IsInf(coordinate, 0) {
				t.Fatalf("%s point %d axis %d is nonfinite: %g", name, i, axis, coordinate)
			}
			triples[i][axis] = int64(math.Round(coordinate * 1e9))
		}
	}
	encoded, err := json.Marshal(triples)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(encoded)
	if got, want := hex.EncodeToString(sum[:]), hashes[name]; got != want {
		t.Fatalf("%s point hash = %s, want %s", name, got, want)
	}
}
