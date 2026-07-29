package pages

import (
	"math"
	"testing"
)

func TestInteractiveLine3DSourcePinAndExactData(t *testing.T) {
	t.Parallel()
	if interactiveLine3DUpstreamPath != "examples/line3d.go" ||
		interactiveLine3DUpstreamRevision != "bda428480a82d6d77ebb9fa939cf8d52528453dd" ||
		interactiveLine3DUpstreamSHA256 != "1f8367a05db06bfe657bfb8cec1b843878ae205d8686e6c68a420b1caec8a7b4" {
		t.Fatal("Line3D upstream source pin changed")
	}
	const wantFormula = "t = i / 1000; x = (1 + 0.25 × cos(75 × t)) × cos(t); y = (1 + 0.25 × cos(75 × t)) × sin(t); z = t + 2 × sin(75 × t), with i from 0 through 24999"
	if interactiveLine3DFormula != wantFormula {
		t.Fatalf("formula = %q, want %q", interactiveLine3DFormula, wantFormula)
	}
	if len(interactiveLine3DPoints) != 25000 {
		t.Fatalf("point count = %d, want 25000", len(interactiveLine3DPoints))
	}
	first, last := interactiveLine3DPoints[0], interactiveLine3DPoints[len(interactiveLine3DPoints)-1]
	if first.X != 1.25 || first.Y != 0 || first.Z != 0 {
		t.Fatalf("first point = %#v", first)
	}
	tValue := 24.999
	radius := 1 + .25*math.Cos(75*tValue)
	wantLast := [3]float64{radius * math.Cos(tValue), radius * math.Sin(tValue), tValue + 2*math.Sin(75*tValue)}
	if last.X != wantLast[0] || last.Y != wantLast[1] || last.Z != wantLast[2] {
		t.Fatalf("last point = %#v, want %#v", last, wantLast)
	}
	const wantHash = "63f356cee0db8603edec10d54e8aec4f5eba291e8c99b961119c9543fd6c63a4"
	if got := interactiveLine3DDataHash(interactiveLine3DPoints); got != wantHash {
		t.Fatalf("data hash = %s, want %s", got, wantHash)
	}
	domains := [6]float64{math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)}
	for _, point := range interactiveLine3DPoints {
		domains[0] = math.Min(domains[0], point.X)
		domains[1] = math.Max(domains[1], point.X)
		domains[2] = math.Min(domains[2], point.Y)
		domains[3] = math.Max(domains[3], point.Y)
		domains[4] = math.Min(domains[4], point.Z)
		domains[5] = math.Max(domains[5], point.Z)
	}
	wantDomains := [6]float64{
		-1.2489045114273476, 1.25,
		-1.2497258288146875, 1.2497201051428937,
		-1.9368409642920035, 26.985899643193864,
	}
	if domains != wantDomains {
		t.Fatalf("coordinate domains = %#v, want %#v", domains, wantDomains)
	}
}

func TestInteractiveLine3DVariantsPreserveSourceTreatments(t *testing.T) {
	t.Parallel()
	base := sampleInteractiveLine3D("basic line3d example", "basic-line3d-example", false)
	rotating := sampleInteractiveLine3D("auto rotating", "auto-rotating", true)
	if base.Kind() != rotating.Kind() {
		t.Fatalf("variant kinds = %q and %q", base.Kind(), rotating.Kind())
	}
}
