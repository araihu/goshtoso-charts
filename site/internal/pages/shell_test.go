package pages

import "testing"

func TestShellConfigUsesReleasedBrandAndLockedThemeHooks(t *testing.T) {
	t.Parallel()

	config := shellConfigForVersion("development")
	if !config.Brand.HideName {
		t.Error("shell brand name remains visible beside the wordmark")
	}
	if config.Brand.Badge == nil || config.Brand.Badge.Label != "dev" || config.Brand.Badge.Href != "" {
		t.Fatalf("development brand badge = %#v", config.Brand.Badge)
	}
	if config.Appearance.DefaultTheme != "araihu" || !config.Appearance.DisableThemeSelector || !config.Appearance.PersistPreferences {
		t.Fatalf("locked appearance = %#v", config.Appearance)
	}
}

func TestShellConfigLinksExactReleaseBadge(t *testing.T) {
	t.Parallel()

	badge := shellConfigForVersion("v0.0.1").Brand.Badge
	if badge == nil {
		t.Fatal("release brand badge is nil")
	}
	if badge.Label != "v0.0.1" || badge.AriaLabel != "Goshtoso Charts release v0.0.1" || badge.Href != "https://github.com/araihu/goshtoso-charts/releases/tag/v0.0.1" {
		t.Fatalf("release brand badge = %#v", badge)
	}
	if badge := shellConfigForVersion("commit-deadbeef").Brand.Badge; badge != nil {
		t.Fatalf("non-release brand badge = %#v, want nil", badge)
	}
}
