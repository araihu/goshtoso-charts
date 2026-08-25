package pages

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso-charts/site/internal/brand"
	goshtosoassets "github.com/araihu/goshtoso/assets"
)

func TestReleasedAppShellDependenciesArePinnedWithoutOverrides(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	module := string(content)
	for _, dependency := range []struct {
		path    string
		version string
	}{
		{path: "github.com/araihu/goshtoso", version: "v0.2.6"},
		{path: "github.com/araihu/goshtoso-app-shells", version: "v0.1.7"},
	} {
		if !moduleRequiresVersion(module, dependency.path, dependency.version) {
			t.Errorf("released consumer dependency %s is not pinned to exact %s", dependency.path, dependency.version)
		}
		if moduleReplacesPath(module, dependency.path) {
			t.Errorf("released consumer dependency %s uses a replace override", dependency.path)
		}
	}
}

func TestModuleReplacesPathRecognizesDirectiveForms(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		module string
		path   string
		want   bool
	}{
		{
			name:   "single line",
			module: "replace example.com/target => ../target\n",
			path:   "example.com/target",
			want:   true,
		},
		{
			name: "block",
			module: `replace (
	example.com/other => ../other
	example.com/target => ../target
)
`,
			path: "example.com/target",
			want: true,
		},
		{
			name:   "versioned block entry",
			module: "replace (\n\texample.com/target v1.2.3 => ../target\n)\n",
			path:   "example.com/target",
			want:   true,
		},
		{
			name:   "different left hand module",
			module: "replace example.com/other => ../other\n",
			path:   "example.com/target",
		},
		{
			name:   "target appears only on right hand side",
			module: "replace (\n\texample.com/other => example.com/target\n)\n",
			path:   "example.com/target",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := moduleReplacesPath(test.module, test.path); got != test.want {
				t.Errorf("moduleReplacesPath() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestReleasedShellKeepsChartsIdentityAndCurrentNavigation(t *testing.T) {
	t.Parallel()
	cfg := shellConfigForVersion("development")
	if cfg.Brand.Name != "Charts" || cfg.Brand.HomeURL != "/" || cfg.Brand.Logo != nil || !cfg.Brand.HideName {
		t.Fatalf("brand contract = %#v", cfg.Brand)
	}
	if cfg.Brand.ManagedLogo == nil || cfg.Brand.ManagedLogo.URL != brand.LogoURL() || cfg.Brand.ManagedLogo.Alt != "Goshtoso" || cfg.Brand.ManagedLogo.Width != 120 || cfg.Brand.ManagedLogo.Height != 32 || !cfg.Brand.ManageFavicon || cfg.Brand.FaviconURL != brand.IconURL() {
		t.Fatalf("managed brand contract = %#v", cfg.Brand)
	}
	if !cfg.Appearance.DisableThemeSelector || !cfg.Appearance.PersistPreferences {
		t.Fatalf("parent appearance contract = %#v", cfg.Appearance)
	}
	channel := cfg.Interactions.PresentationChannel
	if channel == nil {
		t.Fatal("Charts is not enrolled in the released presentation channel")
	}
	if channel.RuntimeURL != seasonalAssetsRuntimeURL || channel.ChannelURL != seasonalAssetsChannelURL || channel.Integrity != seasonalAssetsRuntimeSRI || channel.UseCampaignLabel != "Use seasonal appearance" || channel.UseBaselineLabel != "Use standard appearance" {
		t.Fatalf("presentation channel contract = %#v", channel)
	}

	wantItems := []string{
		"getting-started|Getting started|/",
		"chart-modes|Static and interactive|/docs/chart-modes",
		"chart-controls|Chart controls|/docs/chart-controls",
		"theme-playground|Theme playground|/docs/theme-playground",
		"attributions|Attributions|/attributions",
	}
	gotItems := make([]string, 0, len(cfg.Navigation.Items))
	for _, item := range cfg.Navigation.Items {
		gotItems = append(gotItems, item.ID+"|"+item.Label+"|"+item.Href)
	}
	if !reflect.DeepEqual(gotItems, wantItems) {
		t.Fatalf("top-level navigation = %#v, want %#v", gotItems, wantItems)
	}
	wantSections := []string{
		"Static / Vector", "Interactive / Cartesian", "Interactive / 3D",
		"Interactive / Statistical", "Interactive / Geographic",
		"Interactive / Relationships", "Examples",
	}
	gotSections := make([]string, 0, len(cfg.Navigation.Sections))
	for _, section := range cfg.Navigation.Sections {
		gotSections = append(gotSections, section.Title)
		for _, item := range section.Items {
			identity := strings.ToLower(item.ID + " " + item.Label + " " + item.Href)
			if strings.Contains(identity, "assets") || strings.Contains(identity, "seasonal") || strings.Contains(identity, "campaign") {
				t.Errorf("demo-only navigation enrollment = %#v", item)
			}
		}
	}
	if !reflect.DeepEqual(gotSections, wantSections) {
		t.Fatalf("navigation sections = %#v, want %#v", gotSections, wantSections)
	}
}

func TestReleasedBrandRendersDevelopmentAndStableReleaseBadges(t *testing.T) {
	t.Parallel()
	development := shellConfigForVersion("development").Brand.Badge
	if development == nil || development.Label != "dev" || development.AriaLabel != "Development build" || development.Href != "" {
		t.Fatalf("development badge = %#v", development)
	}

	commitBuild := shellConfigForVersion("commit-0ba2442b339d").Brand.Badge
	if commitBuild == nil || commitBuild.Label != "commit-0ba2442b339d" || commitBuild.AriaLabel != "Goshtoso Charts commit build commit-0ba2442b339d" || commitBuild.Href != "" {
		t.Fatalf("commit build badge = %#v", commitBuild)
	}

	release := shellConfigForVersion("v0.1.1").Brand.Badge
	if release == nil || release.Label != "v0.1.1" || release.AriaLabel != "Goshtoso Charts release v0.1.1" || release.Href != "https://github.com/araihu/goshtoso-charts/releases/tag/v0.1.1" {
		t.Fatalf("release badge = %#v", release)
	}

	for _, version := range []string{
		"release-candidate",
		"commit-0ba2442b339",
		"commit-0BA2442B339D",
		"v0.1.1-beta.1",
		"v0.1.1/../../unexpected",
		`v0.1.1" onclick="alert(1)`,
	} {
		if got := shellConfigForVersion(version).Brand.Badge; got != nil {
			t.Errorf("unsafe or malformed version %q rendered badge %#v", version, got)
		}
	}
}

func TestCommitBuildBadgeRendersAsNonLink(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	err := componentdocshell.Layout(shellConfigForVersion("commit-0ba2442b339d"), componentdocshell.Page{
		Title:   "Getting Started",
		Content: templ.NopComponent,
	}).Render(context.Background(), &output)
	if err != nil {
		t.Fatal(err)
	}

	body := output.String()
	want := `<span class="component-doc-shell__brand-badge" aria-label="Goshtoso Charts commit build commit-0ba2442b339d">commit-0ba2442b339d</span>`
	if !strings.Contains(body, want) {
		t.Fatalf("commit build badge missing from rendered shell: %q", body)
	}
	if strings.Contains(body, `<a class="component-doc-shell__brand-badge"`) {
		t.Fatalf("commit build badge must not render as a link: %q", body)
	}
}

func TestReleasedShellUsesCanonicalStylesAndKeepsThemePlaygroundIsolated(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	err := componentdocshell.Layout(shellConfigForVersion("development"), componentdocshell.Page{
		Title:   "Getting Started",
		Active:  "getting-started",
		Content: templ.NopComponent,
	}).Render(context.Background(), &output)
	if err != nil {
		t.Fatal(err)
	}
	body := output.String()
	for _, want := range []string{
		`class="component-doc-shell__managed-logo"`,
		`data-asset-brand="logo"`,
		`src="/brand/goshtoso-logo-transparent.svg"`,
		`alt="Goshtoso"`,
		`width="120"`,
		`height="32"`,
		`data-asset-brand="icon"`,
		`href="/"`,
		`aria-current="page"`,
		`Getting started`,
		`/componentdocshell/assets/shell.css?v=`,
		goshtosoassets.FirstPartyBundleURL,
		`class="component-doc-shell__brand-badge"`,
		`data-campaign-toggle`,
		`data-use-campaign-label="Use seasonal appearance"`,
		`data-use-baseline-label="Use standard appearance"`,
		`src="` + seasonalAssetsRuntimeURL + `"`,
		`data-channel="` + seasonalAssetsChannelURL + `"`,
		`integrity="` + seasonalAssetsRuntimeSRI + `"`,
		`crossorigin="anonymous"`,
		`defer`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("released shell missing %q", want)
		}
	}
	for _, forbidden := range []string{
		`component-doc-shell__brand-name`,
		`id="componentdocshell-theme-trigger"`,
		`href="/assets"`,
		goshtosoassets.ActionGroupURL,
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("released shell contains forbidden parent concern %q", forbidden)
		}
	}
	if strings.Count(body, `/componentdocshell/assets/shell.css?v=`) != 1 {
		t.Fatalf("canonical shell stylesheet count = %d", strings.Count(body, `/componentdocshell/assets/shell.css?v=`))
	}

	frame := renderThemePlayground(t, ThemePlaygroundFrame())
	if !strings.Contains(frame, `id="theme-playground-theme-trigger"`) {
		t.Fatal("isolated Theme Playground picker is missing")
	}
	if strings.Contains(frame, `component-doc-shell__header`) {
		t.Fatal("isolated Theme Playground contains parent App Shell")
	}
}

func TestSiteDoesNotForkAppShellLayoutCSS(t *testing.T) {
	t.Parallel()
	root := os.DirFS("../..")
	err := fs.WalkDir(root, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == "node_modules" {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_templ.go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		extension := filepath.Ext(path)
		if extension != ".go" && extension != ".templ" && extension != ".css" {
			return nil
		}
		content, err := fs.ReadFile(root, path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), ".component-doc-shell__main") {
			t.Errorf("%s forks App Shell main layout CSS", path)
		}
		if filepath.Base(path) == "shell.css" {
			t.Errorf("%s copies App Shell stylesheet", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func moduleRequiresVersion(module, path, version string) bool {
	for _, line := range strings.Split(module, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == path && fields[1] == version {
			return true
		}
	}
	return false
}

func moduleReplacesPath(module, path string) bool {
	inReplaceBlock := false
	for _, line := range strings.Split(module, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if inReplaceBlock {
			if fields[0] == ")" {
				inReplaceBlock = false
				continue
			}
			if replacementEntryTargets(fields, path) {
				return true
			}
			continue
		}
		if len(fields) >= 2 && fields[0] == "replace" && fields[1] == "(" {
			inReplaceBlock = true
			continue
		}
		if fields[0] == "replace" && replacementEntryTargets(fields[1:], path) {
			return true
		}
	}
	return false
}

func replacementEntryTargets(fields []string, path string) bool {
	return len(fields) >= 3 && fields[0] == path &&
		(fields[1] == "=>" || len(fields) >= 4 && fields[2] == "=>")
}
