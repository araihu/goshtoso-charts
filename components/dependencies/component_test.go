package dependencies_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/assets"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/dependencies"
)

func TestDependenciesDefaultsToVendoredRuntime(t *testing.T) {
	t.Parallel()

	out := render(t, context.Background(), dependencies.Dependencies())
	if !strings.Contains(out, `src="`+assets.RuntimeURL+`"`) {
		t.Fatalf("Dependencies() missing local runtime %q\n%s", assets.RuntimeURL, out)
	}
	if strings.Contains(out, "https://") {
		t.Fatalf("Dependencies() made CDN the default\n%s", out)
	}
	if strings.Contains(out, " defer") {
		t.Fatalf("runtime must load before inline chart initializers\n%s", out)
	}
}

func TestDependenciesCDNIsExplicitAndPinned(t *testing.T) {
	t.Parallel()

	out := render(t, context.Background(), dependencies.Dependencies(dependencies.WithCDN()))
	for _, want := range []string{
		`src="` + assets.RuntimeCDNURL + `"`,
		`integrity="` + assets.RuntimeCDNIntegrity + `"`,
		`crossorigin="anonymous"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Dependencies(WithCDN()) missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, assets.RuntimeURL) {
		t.Fatalf("CDN option retained local runtime URL\n%s", out)
	}
}

func TestDependenciesAllowsApplicationOwnedLocalPath(t *testing.T) {
	t.Parallel()

	out := render(t, context.Background(), dependencies.Dependencies(
		dependencies.WithLocalURL("/static/charts/runtime.js"),
	))
	if !strings.Contains(out, `src="/static/charts/runtime.js"`) {
		t.Fatalf("custom local path missing\n%s", out)
	}
	if strings.Contains(out, assets.RuntimeURL) {
		t.Fatalf("custom local path retained default\n%s", out)
	}
}

func TestDependenciesAllowsApplicationOwnedCDN(t *testing.T) {
	t.Parallel()

	out := render(t, context.Background(), dependencies.Dependencies(
		dependencies.WithCDNURL("https://cdn.example.test/charts/runtime.js", "sha384-custom"),
	))
	for _, want := range []string{
		`src="https://cdn.example.test/charts/runtime.js"`,
		`integrity="sha384-custom"`,
		`crossorigin="anonymous"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("custom CDN dependency missing %q\n%s", want, out)
		}
	}
}

func TestDependenciesPropagatesTemplNonce(t *testing.T) {
	t.Parallel()

	ctx := templ.WithNonce(context.Background(), "chart-nonce")
	out := render(t, ctx, dependencies.Dependencies())
	if !strings.Contains(out, `nonce="chart-nonce"`) {
		t.Fatalf("dependency script missing templ nonce\n%s", out)
	}
}

func TestDependenciesZeroValueKeepsDefaults(t *testing.T) {
	t.Parallel()

	out := render(t, context.Background(), dependencies.Instance{})
	if !strings.Contains(out, `src="`+assets.RuntimeURL+`"`) {
		t.Fatalf("zero-value Instance missing default runtime\n%s", out)
	}
}

func TestDependenciesHasStableKind(t *testing.T) {
	t.Parallel()

	if got := dependencies.Dependencies().Kind(); got != chartcomponents.KindDependencies {
		t.Fatalf("Kind() = %q, want %q", got, chartcomponents.KindDependencies)
	}
}

func render(t *testing.T, ctx context.Context, component templ.Component) string {
	t.Helper()
	var output bytes.Buffer
	if err := component.Render(ctx, &output); err != nil {
		t.Fatalf("render component: %v", err)
	}
	return output.String()
}
