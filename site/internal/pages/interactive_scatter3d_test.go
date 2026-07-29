package pages

import "testing"

func TestInteractiveScatter3DSourcePinAndDeterministicPoints(t *testing.T) {
	t.Parallel()
	if interactiveScatter3DUpstreamPath != "examples/scatter3d.go" ||
		interactiveScatter3DUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveScatter3DUpstreamSHA256 != "cf654926b96edca762bd3d1280d0ce6ef7a7affc8a63dcaf2f5207b09b216d8c" {
		t.Fatal("Scatter3D upstream source pin changed")
	}
	if len(interactiveScatter3DBasicPoints) != 80 {
		t.Fatalf("point count = %d, want 80", len(interactiveScatter3DBasicPoints))
	}
	for index, point := range interactiveScatter3DBasicPoints {
		for name, value := range map[string]float64{"x": point.X, "y": point.Y, "z": point.Z} {
			if value < 0 || value > 99 || value != float64(int(value)) {
				t.Fatalf("point %d %s = %v, want integer in [0,99]", index, name, value)
			}
		}
	}
	const wantHash = "f01d67d48ef648c1b32e94ae2940e38b5c66037219516ab9d204e43819cc482e"
	if got := interactiveScatter3DPointHash(interactiveScatter3DBasicPoints); got != wantHash {
		t.Fatalf("deterministic point hash = %s, want %s", got, wantHash)
	}
}
