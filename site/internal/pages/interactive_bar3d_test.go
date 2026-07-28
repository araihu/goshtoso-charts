package pages

import (
	"slices"
	"testing"
)

func TestInteractiveBar3DSourcePinAndExactData(t *testing.T) {
	t.Parallel()
	if interactiveBar3DUpstreamPath != "examples/bar3d.go" ||
		interactiveBar3DUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveBar3DUpstreamSHA256 != "110b3b85f2528d76eb8271b64f1facd81a974e30ecc0dd77319d5a409ff64275" {
		t.Fatal("Bar3D upstream source pin changed")
	}
	wantHours := []string{"12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a", "12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p"}
	wantDays := []string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"}
	if !slices.Equal(interactiveBar3DHours, wantHours) || !slices.Equal(interactiveBar3DDays, wantDays) {
		t.Fatal("Bar3D categorical axis order changed")
	}
	cells := interactiveBar3DCells()
	if len(cells) != 168 {
		t.Fatalf("cell count = %d, want 168", len(cells))
	}
	for index, cell := range cells {
		source := interactiveBar3DSourceCells[index]
		if cell.XIndex != source[1] || cell.YIndex != source[0] || cell.Value != float64(source[2]) {
			t.Fatalf("cell %d does not preserve upstream X/Y swap: %#v from %#v", index, cell, source)
		}
	}
	const wantHash = "580250773b8f88507e97adbdf56b90d3b0f6e5cb13e7e13a3b8e7c7f377e8e94"
	if got := interactiveBar3DDataHash(cells); got != wantHash {
		t.Fatalf("exact data hash = %s, want %s", got, wantHash)
	}
}
