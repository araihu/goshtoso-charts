package buildinfo

import "testing"

func TestSiteVersionDefaultsToDevelopment(t *testing.T) {
	if got, want := SiteVersion(), "development"; got != want {
		t.Fatalf("SiteVersion() = %q, want %q", got, want)
	}
}
