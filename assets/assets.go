// Package assets serves versioned browser assets used by Goshtoso Charts.
package assets

import (
	"embed"
	"net/http"
)

const (
	// Prefix is the default HTTP mount path consumed by components/dependencies.
	Prefix = "/charts/assets/"
	// RuntimeURL is the versioned path of the embedded interactive-chart runtime.
	RuntimeURL = Prefix + "js/runtime/echarts/5.4.3/echarts.min.js"
	// RuntimeCDNURL is the pinned opt-in CDN source for the same runtime version.
	RuntimeCDNURL = "https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js"
	// RuntimeCDNIntegrity authenticates bytes served by RuntimeCDNURL.
	RuntimeCDNIntegrity = "sha384-BQKzmHvQLMCAnL3UtDBA1Al5tFjsCz1wrMlIUA1wkzo14DYkRWjywW+p9pCj0cwd"
	// WordCloudRuntimeURL is the versioned local word-cloud extension path.
	WordCloudRuntimeURL = Prefix + "js/runtime/word-cloud/2.1.0/runtime.min.js"
	// WordCloudRuntimeCDNURL is the pinned opt-in CDN source for the word-cloud extension.
	WordCloudRuntimeCDNURL = "https://cdn.jsdelivr.net/npm/echarts-wordcloud@2.1.0/dist/echarts-wordcloud.min.js"
	// WordCloudRuntimeCDNIntegrity authenticates bytes served by WordCloudRuntimeCDNURL.
	WordCloudRuntimeCDNIntegrity = "sha384-U1KEY0DDCF4Dq6Yx1J+EZ5Hnj8X5bMn52OAcJB8C4OiAWeU4iJhJ/Tv5KhTqu8zZ"
	// LiquidRuntimeURL is the versioned local liquid-gauge extension path.
	LiquidRuntimeURL = Prefix + "js/runtime/liquid/3.1.0/runtime.min.js"
	// LiquidLicenseURL is the bundled license for the liquid-gauge extension.
	LiquidLicenseURL = Prefix + "js/runtime/liquid/3.1.0/LICENSE.md"
	// LiquidRuntimeCDNURL is the pinned opt-in CDN source for the liquid-gauge extension.
	LiquidRuntimeCDNURL = "https://cdn.jsdelivr.net/npm/echarts-liquidfill@3.1.0/dist/echarts-liquidfill.min.js"
	// LiquidRuntimeCDNIntegrity authenticates bytes served by LiquidRuntimeCDNURL.
	LiquidRuntimeCDNIntegrity = "sha384-+LS91q88WjMob2zpAaAPTyASiqV4HPo9zJHsEwjcukMZevj//sFrxBXdAHe1t2CL"
	// ThreeDRuntimeURL is the versioned local three-dimensional chart extension path.
	ThreeDRuntimeURL = Prefix + "js/runtime/three-d/2.0.9/runtime.min.js"
	// ThreeDLicenseURL is the bundled BSD-3-Clause license for the extension.
	ThreeDLicenseURL = Prefix + "js/runtime/three-d/2.0.9/LICENSE"
	// ThreeDRuntimeCDNURL is the pinned opt-in CDN source for the extension.
	ThreeDRuntimeCDNURL = "https://cdn.jsdelivr.net/npm/echarts-gl@2.0.9/dist/echarts-gl.min.js"
	// ThreeDRuntimeCDNIntegrity authenticates bytes served by ThreeDRuntimeCDNURL.
	ThreeDRuntimeCDNIntegrity = "sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl"
	// ChinaMapURL is the versioned local resource for national geometry.
	ChinaMapURL = Prefix + "js/maps/41f247b1cbb6/china.js"
	// ChinaMapCDNURL is the commit-pinned opt-in CDN source for national geometry.
	ChinaMapCDNURL = "https://cdn.jsdelivr.net/gh/go-echarts/go-echarts-assets@41f247b1cbb649b029a2d3fffb04f469de372aa7/assets/maps/china.js"
	// ChinaMapCDNIntegrity authenticates bytes served by ChinaMapCDNURL.
	ChinaMapCDNIntegrity = "sha384-qwEZxzbtfuBsHahOge6aHnLsYt6NBGcOFoTctegFtOU3h/mWm8PYtRbJ19Xa6B5I"
	// GuangdongMapURL is the versioned local resource for Guangdong geometry.
	GuangdongMapURL = Prefix + "js/maps/41f247b1cbb6/guangdong.js"
	// GuangdongMapCDNURL is the commit-pinned opt-in CDN source for Guangdong geometry.
	GuangdongMapCDNURL = "https://cdn.jsdelivr.net/gh/go-echarts/go-echarts-assets@41f247b1cbb649b029a2d3fffb04f469de372aa7/assets/maps/guangdong.js"
	// GuangdongMapCDNIntegrity authenticates bytes served by GuangdongMapCDNURL.
	GuangdongMapCDNIntegrity = "sha384-Q7MOpZeBbcPxI3hKHud73/Z1PjvChsn12B3IN6NqOj08KXRF1IU2D7LvaY16uV4w"
	// ControlRuntimeURL is the versioned shared chart-controls runtime.
	ControlRuntimeURL = Prefix + "js/controls/2/controls.js"
)

//go:embed js/runtime/echarts/5.4.3/echarts.min.js js/runtime/word-cloud/2.1.0/runtime.min.js js/runtime/liquid/3.1.0/runtime.min.js js/runtime/liquid/3.1.0/LICENSE.md js/runtime/three-d/2.0.9/runtime.min.js js/runtime/three-d/2.0.9/LICENSE js/maps/41f247b1cbb6/*.js js/controls/1/controls.js js/controls/2/controls.js
var files embed.FS

// Handler serves embedded assets at Prefix. Mount it directly at Prefix; the
// handler removes that prefix itself.
func Handler() http.Handler {
	return http.StripPrefix(Prefix, http.FileServer(http.FS(files)))
}
