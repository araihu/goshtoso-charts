// Package buildinfo exposes immutable metadata injected into the demo binary.
package buildinfo

// siteVersion is replaced at build time by the deployment image. Plain local
// builds intentionally retain "development" so an unversioned checkout is not
// presented as a released site.
var siteVersion = "development"

// SiteVersion returns the Charts version built into this demo binary.
func SiteVersion() string {
	return siteVersion
}
