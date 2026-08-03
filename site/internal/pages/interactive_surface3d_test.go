package pages

import (
	"math"
	"testing"

	"github.com/araihu/goshtoso-charts/components/interactive"
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
	if len(interactiveSurface3DHeartPoints) != interactiveSurface3DHeartRows*interactiveSurface3DHeartColumns {
		t.Fatalf("heart point count = %d, want %d", len(interactiveSurface3DHeartPoints), interactiveSurface3DHeartRows*interactiveSurface3DHeartColumns)
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
	const wantHeartHash = "54bdc4db1027b9b8abd55867a6981458b9a857bbe074021f1f761af33245784b"
	if got := interactiveSurface3DDataHash(interactiveSurface3DHeartPoints); got != wantHeartHash {
		t.Fatalf("heart data hash = %s, want %s", got, wantHeartHash)
	}
	for column := 0; column < interactiveSurface3DHeartColumns; column++ {
		first := interactiveSurface3DHeartPoints[column]
		last := interactiveSurface3DHeartPoints[(interactiveSurface3DHeartRows-1)*interactiveSurface3DHeartColumns+column]
		if math.Abs(first.X-last.X) > 1e-12 || math.Abs(first.Y-last.Y) > 1e-12 || math.Abs(first.Z-last.Z) > 1e-12 {
			t.Fatalf("heart column %d outline seam is open: first=%#v last=%#v", column, first, last)
		}
	}
	front := interactiveSurface3DHeartPoints[0]
	back := interactiveSurface3DHeartPoints[interactiveSurface3DHeartColumns-1]
	for row := 1; row < interactiveSurface3DHeartRows; row++ {
		rowFront := interactiveSurface3DHeartPoints[row*interactiveSurface3DHeartColumns]
		rowBack := interactiveSurface3DHeartPoints[(row+1)*interactiveSurface3DHeartColumns-1]
		if math.Abs(front.X-rowFront.X) > 1e-12 || math.Abs(front.Y-rowFront.Y) > 1e-12 || math.Abs(front.Z-rowFront.Z) > 1e-12 {
			t.Fatalf("heart front pole is open at row %d: first=%#v row=%#v", row, front, rowFront)
		}
		if math.Abs(back.X-rowBack.X) > 1e-12 || math.Abs(back.Y-rowBack.Y) > 1e-12 || math.Abs(back.Z-rowBack.Z) > 1e-12 {
			t.Fatalf("heart back pole is open at row %d: first=%#v row=%#v", row, back, rowBack)
		}
	}
}

func TestInteractiveSurface3DHeartHasRecognizableFrontSilhouette(t *testing.T) {
	t.Parallel()
	point := func(row, column int) interactive.Point3D {
		return interactiveSurface3DHeartPoints[row*interactiveSurface3DHeartColumns+column]
	}

	frontOutline := interactiveSurface3DHeartColumns / 2
	cleft := point(0, frontOutline)
	rightLobe := point(5, frontOutline)
	bottom := point(interactiveSurface3DHeartRows/2, frontOutline)
	leftLobe := point(interactiveSurface3DHeartRows-1-5, frontOutline)

	if math.Abs(cleft.X) > 1e-10 || math.Abs(bottom.X) > 1e-10 {
		t.Fatalf("heart center landmarks are not centered: cleft=%#v bottom=%#v", cleft, bottom)
	}
	if rightLobe.X < 3 || leftLobe.X > -3 {
		t.Fatalf("heart lacks separated upper lobes: right=%#v left=%#v", rightLobe, leftLobe)
	}
	if math.Abs(rightLobe.X+leftLobe.X) > 1e-10 || math.Abs(rightLobe.Z-leftLobe.Z) > 1e-10 {
		t.Fatalf("heart upper lobes are not bilaterally symmetric: right=%#v left=%#v", rightLobe, leftLobe)
	}
	const wantCleftDepth = 4.678937306500523
	if got := rightLobe.Z - cleft.Z; math.Abs(got-wantCleftDepth) > 0.01 {
		t.Fatalf("heart central cleft depth = %f, want %f", got, wantCleftDepth)
	}
	if cleft.Z-bottom.Z < 15 {
		t.Fatalf("heart does not taper to a lower point: cleft=%#v bottom=%#v", cleft, bottom)
	}
}
