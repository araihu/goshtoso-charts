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
	if !strings.Contains(out, `src="`+assets.WordCloudRuntimeURL+`"`) {
		t.Fatalf("Dependencies() missing local word-cloud runtime %q\n%s", assets.WordCloudRuntimeURL, out)
	}
	if strings.Index(out, assets.RuntimeURL) > strings.Index(out, assets.WordCloudRuntimeURL) {
		t.Fatalf("word-cloud runtime loaded before core runtime\n%s", out)
	}
	if !strings.Contains(out, `src="`+assets.LiquidRuntimeURL+`"`) {
		t.Fatalf("Dependencies() missing local liquid runtime %q\n%s", assets.LiquidRuntimeURL, out)
	}
	if strings.Index(out, assets.WordCloudRuntimeURL) > strings.Index(out, assets.LiquidRuntimeURL) {
		t.Fatalf("liquid runtime loaded before word-cloud runtime\n%s", out)
	}
	for _, url := range []string{assets.ChinaMapURL, assets.GuangdongMapURL} {
		if !strings.Contains(out, `src="`+url+`"`) {
			t.Errorf("Dependencies() missing local map resource %q", url)
		}
		if strings.Index(out, assets.RuntimeURL) > strings.Index(out, url) {
			t.Errorf("map resource loaded before core runtime: %q", url)
		}
	}
	if strings.Index(out, assets.LiquidRuntimeURL) > strings.Index(out, assets.ChinaMapURL) || strings.Index(out, assets.ChinaMapURL) > strings.Index(out, assets.GuangdongMapURL) {
		t.Fatalf("extension and map resources are not in stable dependency order\n%s", out)
	}
}

func TestDependenciesCDNIsExplicitAndPinned(t *testing.T) {
	t.Parallel()

	out := render(t, context.Background(), dependencies.Dependencies(dependencies.WithCDN()))
	for _, want := range []string{
		`src="` + assets.RuntimeCDNURL + `"`,
		`integrity="` + assets.RuntimeCDNIntegrity + `"`,
		`src="` + assets.WordCloudRuntimeCDNURL + `"`,
		`integrity="` + assets.WordCloudRuntimeCDNIntegrity + `"`,
		`src="` + assets.LiquidRuntimeCDNURL + `"`,
		`integrity="` + assets.LiquidRuntimeCDNIntegrity + `"`,
		`src="` + assets.ChinaMapCDNURL + `"`,
		`integrity="` + assets.ChinaMapCDNIntegrity + `"`,
		`src="` + assets.GuangdongMapCDNURL + `"`,
		`integrity="` + assets.GuangdongMapCDNIntegrity + `"`,
		`crossorigin="anonymous"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Dependencies(WithCDN()) missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, assets.RuntimeURL) {
		t.Fatalf("CDN option retained local runtime URL\n%s", out)
	}
	if strings.Contains(out, assets.WordCloudRuntimeURL) {
		t.Fatalf("CDN option retained local word-cloud runtime URL\n%s", out)
	}
	if strings.Contains(out, assets.LiquidRuntimeURL) {
		t.Fatalf("CDN option retained local liquid runtime URL\n%s", out)
	}
	if !(strings.Index(out, assets.RuntimeCDNURL) < strings.Index(out, assets.WordCloudRuntimeCDNURL) &&
		strings.Index(out, assets.WordCloudRuntimeCDNURL) < strings.Index(out, assets.LiquidRuntimeCDNURL) &&
		strings.Index(out, assets.LiquidRuntimeCDNURL) < strings.Index(out, assets.ChinaMapCDNURL) &&
		strings.Index(out, assets.ChinaMapCDNURL) < strings.Index(out, assets.GuangdongMapCDNURL)) {
		t.Fatalf("CDN dependencies not ordered core, word-cloud, liquid, china map, guangdong map\n%s", out)
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
	if !strings.Contains(out, `src="`+assets.WordCloudRuntimeURL+`"`) {
		t.Fatalf("custom core path removed local word-cloud runtime\n%s", out)
	}
	if !strings.Contains(out, `src="`+assets.LiquidRuntimeURL+`"`) {
		t.Fatalf("custom core path removed local liquid runtime\n%s", out)
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
	if !strings.Contains(out, `src="`+assets.WordCloudRuntimeURL+`"`) {
		t.Fatalf("custom core CDN removed local word-cloud runtime\n%s", out)
	}
}

func TestDependenciesPropagatesTemplNonce(t *testing.T) {
	t.Parallel()

	ctx := templ.WithNonce(context.Background(), "chart-nonce")
	out := render(t, ctx, dependencies.Dependencies())
	if !strings.Contains(out, `nonce="chart-nonce"`) {
		t.Fatalf("dependency script missing templ nonce\n%s", out)
	}
	if strings.Count(out, `nonce="chart-nonce"`) != 5 {
		t.Fatalf("dependency scripts did not all receive templ nonce\n%s", out)
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
