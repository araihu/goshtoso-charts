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
		{path: "github.com/araihu/goshtoso", version: "v0.1.1"},
		{path: "github.com/araihu/goshtoso-app-shells", version: "v0.1.0"},
	} {
		if !moduleRequiresVersion(module, dependency.path, dependency.version) {
			t.Errorf("released consumer dependency %s is not pinned to exact %s", dependency.path, dependency.version)
		}
		if moduleReplacesPath(module, dependency.path) {
			t.Errorf("released consumer dependency %s uses a replace override", dependency.path)
		}
	}
}

func TestReleasedShellKeepsChartsIdentityAndCurrentNavigation(t *testing.T) {
	t.Parallel()
	cfg := shellConfigForVersion("development")
	if cfg.Brand.Name != "Charts" || cfg.Brand.HomeURL != "/" || cfg.Brand.Logo == nil || !cfg.Brand.HideName {
		t.Fatalf("brand contract = %#v", cfg.Brand)
	}
	if !cfg.Appearance.DisableThemeSelector || cfg.Appearance.PersistPreferences {
		t.Fatalf("parent appearance contract = %#v", cfg.Appearance)
	}
	if cfg.Interactions.PresentationChannel != nil {
		t.Fatal("Charts enrolled in a demo-only presentation channel")
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

func TestReleasedHeaderActionsRenderDevelopmentAndStableReleaseBadges(t *testing.T) {
	t.Parallel()
	development := renderHeaderActions(t, shellConfigForVersion("development"))
	for _, want := range []string{`data-site-version`, `>dev<`} {
		if !strings.Contains(development, want) {
			t.Errorf("development badge missing %q in %q", want, development)
		}
	}
	if strings.Contains(development, "<a ") {
		t.Fatalf("development badge must not link: %q", development)
	}

	release := renderHeaderActions(t, shellConfigForVersion("v0.1.1"))
	for _, want := range []string{
		`data-site-version`,
		`>v0.1.1<`,
		`href="https://github.com/araihu/goshtoso-charts/releases/tag/v0.1.1"`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		`aria-label="Goshtoso Charts release v0.1.1"`,
	} {
		if !strings.Contains(release, want) {
			t.Errorf("release badge missing %q in %q", want, release)
		}
	}

	for _, version := range []string{
		"release-candidate",
		"v0.1.1-beta.1",
		"v0.1.1/../../unexpected",
		`v0.1.1" onclick="alert(1)`,
	} {
		if got := strings.TrimSpace(renderHeaderActions(t, shellConfigForVersion(version))); got != "" {
			t.Errorf("unsafe or malformed version %q rendered badge %q", version, got)
		}
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
		`component-doc-shell__brand-logo--goshtoso`,
		`href="/"`,
		`aria-current="page"`,
		`Getting started`,
		`/componentdocshell/assets/shell.css?v=`,
		goshtosoassets.FirstPartyBundleURL,
		`data-site-version`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("released shell missing %q", want)
		}
	}
	for _, forbidden := range []string{
		`component-doc-shell__brand-name`,
		`id="componentdocshell-theme-trigger"`,
		`data-campaign-toggle`,
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

func renderHeaderActions(t *testing.T, cfg componentdocshell.Config) string {
	t.Helper()
	if cfg.HeaderActions == nil {
		return ""
	}
	var output bytes.Buffer
	if err := cfg.HeaderActions.Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	return output.String()
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
	for _, line := range strings.Split(module, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "replace" && fields[1] == path {
			return true
		}
	}
	return false
}
