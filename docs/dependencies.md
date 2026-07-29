# Interactive chart dependencies

Static vector charts need no browser runtime. Interactive charts need the
version-matched scripts emitted by `components/dependencies`. The dependency
set includes the word-cloud and liquid-gauge extensions plus pinned geographic
resources. They load in stable order after core: word cloud, liquid, national
geometry, then regional geometry. Every public interactive component shares
one document-level dependency contract.

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

`WithCDN` uses pinned 5.4.3 core, 2.1.0 word-cloud, 3.1.0 liquid-gauge, and
2.0.9 three-dimensional runtime URLs with SHA-384 Subresource Integrity and
`crossorigin="anonymous"`. Brazil geometry remains on its pinned embedded paths:
IBGE publishes source geography rather than safe executable registration artifacts.
Allow the CDN origin in `script-src` when enforcing CSP.

Applications that mirror assets may use `WithLocalURL`. Applications that own
a different CDN may use `WithCDNURL(url, integrity)`; keep its runtime version
compatible with the Go renderer adapter.

Implementation asset names stay confined to this dependency package, the asset
package, and dependency documentation. Public chart configs remain
renderer-neutral.

Map and Geo geometry resources load after core, word-cloud, and liquid-gauge scripts. Local
paths, immutable source revisions, SHA-256 values, licenses, and the deliberate
CDN exception live in `assets/NOTICE.md` and `assets/assets.go`.
