# Interactive chart dependencies

Static vector charts need no browser runtime. Interactive charts need the
version-matched script emitted by `components/dependencies`.

## Recommended: embedded local runtime

Mount the public asset handler directly at its default prefix. `Handler`
removes the prefix itself; adding another `http.StripPrefix` causes 404s.

```go
import (
	"net/http"

	chartassets "github.com/araihu/goshtoso-charts/assets"
)

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET "+chartassets.Prefix, chartassets.Handler())
	return mux
}
```

Emit the dependency in the document head, before any interactive chart output:

```templ
import "github.com/araihu/goshtoso-charts/components/dependencies"

templ Layout() {
	<html>
		<head>
			@dependencies.Dependencies()
		</head>
	</html>
}
```

This is the default. It serves a versioned, vendored runtime and makes no
third-party request. The script intentionally loads synchronously because the
interactive chart markup contains inline initialization that needs the runtime
to exist first. A nonce supplied with `templ.WithNonce` is copied to the tag.

## Explicit CDN delivery

Opt in only when third-party delivery is acceptable:

```templ
@dependencies.Dependencies(dependencies.WithCDN())
```

`WithCDN` uses a pinned 5.4.3 jsDelivr URL with SHA-384 Subresource Integrity
and `crossorigin="anonymous"`. It does not silently fall back to local assets.
Allow the CDN origin in `script-src` when enforcing CSP.

Applications that mirror assets may use `WithLocalURL`. Applications that own
a different CDN may use `WithCDNURL(url, integrity)`; keep its runtime version
compatible with the Go renderer adapter.

Implementation asset names stay confined to this dependency package, the asset
package, and dependency documentation. Public chart configs remain
renderer-neutral.
