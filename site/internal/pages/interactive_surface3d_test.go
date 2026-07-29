package pages

import (
	"math"
	"testing"
)

func TestInteractiveSurface3DSourcePinAndExactData(t *testing.T) {
	t.Parallel()
	if interactiveSurface3DUpstreamPath != "examples/surface3d.go" ||
		interactiveSurface3DUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveSurface3DUpstreamSHA256 != "51ffdac86403c0e6430c0134d1062e5fda9865b7d88061ce4f38129113c455d9" {
		t.Fatal("Surface3D upstream source pin changed")
	}
	if len(interactiveSurface3DBasePoints) != 14400 {
		t.Fatalf("base point count = %d, want 14400", len(interactiveSurface3DBasePoints))
	}
	if len(interactiveSurface3DRosePoints) != 3600 {
		t.Fatalf("rose point count = %d, want 3600", len(interactiveSurface3DRosePoints))
	}
	baseFirst, baseLast := interactiveSurface3DBasePoints[0], interactiveSurface3DBasePoints[len(interactiveSurface3DBasePoints)-1]
	if baseFirst.X != -1 || baseFirst.Y != -1 || math.Abs(baseFirst.Z) > 1e-15 {
		t.Fatalf("base first point = %#v", baseFirst)
	}
	if baseLast.X != float64(59)/60 || baseLast.Y != float64(59)/60 {
		t.Fatalf("base last point = %#v", baseLast)
	}
	roseFirst, roseLast := interactiveSurface3DRosePoints[0], interactiveSurface3DRosePoints[len(interactiveSurface3DRosePoints)-1]
	if roseFirst.X != -3 || roseFirst.Y != -3 {
		t.Fatalf("rose first point = %#v", roseFirst)
	}
	if roseLast.X != 2.9 || roseLast.Y != 2.9 {
		t.Fatalf("rose last point = %#v", roseLast)
	}
	const wantBaseHash = "84dc950f2c5fb6d80b289da7568bf7f119e5f0a372cff23350b43bc31ac7d6e1"
	const wantRoseHash = "c0c847473affc08312baf3226c5f011d6070c3a3fb93a7c71ac330642829e0da"
	if got := interactiveSurface3DDataHash(interactiveSurface3DBasePoints); got != wantBaseHash {
		t.Fatalf("base data hash = %s, want %s", got, wantBaseHash)
	}
	if got := interactiveSurface3DDataHash(interactiveSurface3DRosePoints); got != wantRoseHash {
		t.Fatalf("rose data hash = %s, want %s", got, wantRoseHash)
	}
}
