// Package brand serves the approved Goshtoso v11 identity assets.
package brand

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"
)

//go:embed assets/*.svg
var assetFS embed.FS

//go:embed assets/goshtoso-logo-transparent.svg
var logoSVG []byte

const (
	logoURL = "/brand/goshtoso-logo-transparent.svg"
	iconURL = "/brand/goshtoso-icon-transparent.svg"
)

// Handler serves v11 identity assets below /brand/ after the route prefix is removed.
func Handler() http.Handler {
	assets, err := fs.Sub(assetFS, "assets")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(assets))
}

// Logo returns Goshtoso's approved v11 wordmark inline so its CSS variables
// follow the containing document's Goshtoso .dark state. Charts remains shell text.
func Logo() templ.Component {
	return templ.Raw(`<span class="component-doc-shell__brand-logo component-doc-shell__brand-logo--goshtoso" aria-hidden="true">` + string(logoSVG) + `</span>`)
}

// IconURL is the approved v11 icon used as the product favicon.
func IconURL() string { return iconURL }

// LogoURL is the approved v11 wordmark used as the managed shell baseline.
func LogoURL() string { return logoURL }

// Head applies the v11 surface, ink, and signal contract to Goshtoso's .dark state.
func Head() templ.Component {
	return templ.Raw(`<style>
:root { --araihu-logo-surface: var(--color-surface); --araihu-logo-ink: var(--color-on-surface-strong); --araihu-logo-signal: #c7ff4a; }
.dark { --araihu-logo-surface: var(--color-surface-dark); --araihu-logo-ink: var(--color-on-surface-dark-strong); --araihu-logo-signal: #c7ff4a; }
.component-doc-shell__brand-logo.component-doc-shell__brand-logo--goshtoso,
.component-doc-shell__brand-logo.component-doc-shell__brand-logo--goshtoso svg { width: 7.5rem; height: 2rem; }
.component-doc-shell__brand-logo.component-doc-shell__brand-logo--goshtoso svg { display: block; }
.goshtoso-charts-brand { display: block; width: min(20rem, 100%); }
.goshtoso-charts-brand svg { display: block; width: 100%; height: auto; }
</style>`)
}

// LandingLogo returns the same inline v11 wordmark used in the product header.
func LandingLogo() templ.Component {
	return templ.Raw(`<span class="goshtoso-charts-brand" aria-hidden="true">` + string(logoSVG) + `</span>`)
}
