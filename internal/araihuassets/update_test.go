package araihuassets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestUpdateCopiesAllowlistedBrandFallbacksInStableOrderAndIsIdempotent(t *testing.T) {
	repoRoot := t.TempDir()
	releaseRoot := t.TempDir()
	release := writeReleaseFixture(t, releaseRoot)
	manifest := fixtureManifest(release)
	writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), manifest)

	result, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
	if err != nil {
		t.Fatal(err)
	}
	wantChanged := []string{
		"site/internal/brand/assets/goshtoso-icon-transparent.svg",
		"site/internal/brand/assets/goshtoso-logo-transparent.svg",
	}
	if !slices.Equal(result.Changed, wantChanged) {
		t.Fatalf("changed paths = %q, want stable order %q", result.Changed, wantChanged)
	}
	for _, mapping := range manifest.Mappings {
		got, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mapping.Destination)))
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join(releaseRoot, filepath.FromSlash(mapping.Source)))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatalf("%s bytes differ from allowlisted source %s", mapping.Destination, mapping.Source)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "unlisted.txt")); !os.IsNotExist(err) {
		t.Fatalf("unlisted release file copied: %v", err)
	}

	second, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Changed) != 0 {
		t.Fatalf("second update changed %q, want clean", second.Changed)
	}
}

func TestUpdateStagesAllWritesAndRollsBackReleaseUpgradeFailure(t *testing.T) {
	repoRoot := t.TempDir()
	releaseRoot := t.TempDir()
	release := writeReleaseFixture(t, releaseRoot)
	oldManifest := fixtureManifest(release)
	oldManifest.AssetsRevision = strings.Repeat("1", 40)
	oldManifest.Release = "v1.2.2"
	oldManifest.ReleaseURL = "https://github.com/araihu/assets/releases/download/v1.2.2/araihu-assets-v1.2.2.tar.gz"
	oldManifest.ReleaseSHA256 = strings.Repeat("2", 64)
	oldManifest.ReleaseJSONSHA256 = strings.Repeat("3", 64)
	writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), oldManifest)
	writeOldFallbacks(t, repoRoot, oldManifest.Mappings)

	paths := []string{"araihu-assets.json"}
	for _, mapping := range oldManifest.Mappings {
		paths = append(paths, mapping.Destination)
	}
	before := snapshotFiles(t, repoRoot, paths)
	newIdentity := fixtureManifest(release)
	sawAllStaged := false
	opts := Options{
		RepoRoot:    repoRoot,
		ReleaseRoot: releaseRoot,
		Identity: &ReleaseIdentity{
			AssetsRepository:  newIdentity.AssetsRepository,
			AssetsRevision:    newIdentity.AssetsRevision,
			Release:           newIdentity.Release,
			ReleaseURL:        newIdentity.ReleaseURL,
			ReleaseSHA256:     newIdentity.ReleaseSHA256,
			ReleaseJSONSHA256: newIdentity.ReleaseJSONSHA256,
		},
		beforeReplace: func(index int, _ string) error {
			if index == 0 {
				sawAllStaged = countStagedFiles(t, repoRoot) == len(paths)
			}
			if index == 2 {
				return errors.New("injected replacement failure")
			}
			return nil
		},
	}
	_, err := Update(opts)
	var applyErr *ApplyError
	if !errors.As(err, &applyErr) {
		t.Fatalf("Update() error = %v, want *ApplyError", err)
	}
	if !sawAllStaged {
		t.Fatal("replacement began before every fallback and manifest write was staged")
	}
	if !applyErr.RollbackComplete() {
		t.Fatalf("rollback errors = %v, want complete rollback", applyErr.RollbackErrors)
	}
	if applyErr.FailedPath != "araihu-assets.json" || len(applyErr.AppliedPaths) != 2 {
		t.Fatalf("failure state = %#v, want manifest failure after two fallback replacements", applyErr)
	}
	if slices.Contains(applyErr.AppliedPaths, "araihu-assets.json") {
		t.Fatalf("manifest applied before fallback failure: %q", applyErr.AppliedPaths)
	}
	assertFilesMatch(t, repoRoot, paths, before)
	if got := countStagedFiles(t, repoRoot); got != 0 {
		t.Fatalf("staged files after rollback = %d, want 0", got)
	}

	opts.beforeReplace = nil
	result, err := Update(opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changed) != len(paths) {
		t.Fatalf("successful upgrade changed %q, want every fallback plus manifest", result.Changed)
	}
	if result.Changed[len(result.Changed)-1] != "araihu-assets.json" {
		t.Fatalf("upgrade order = %q, want manifest last", result.Changed)
	}
}

func TestUpdateRequiresExactCatalogCanonicalNameRolesAndHash(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, releaseRoot string, manifest *Manifest)
		want   string
	}{
		{
			name: "canonical name",
			mutate: func(_ *testing.T, _ string, manifest *Manifest) {
				manifest.Mappings[0].CanonicalName = "missing-icon"
			},
			want: `catalog canonicalName "missing-icon" not found`,
		},
		{
			name: "role",
			mutate: func(_ *testing.T, _ string, manifest *Manifest) {
				manifest.Mappings[1].Artwork = "icon"
			},
			want: `catalog role artwork = "logo", want "icon"`,
		},
		{
			name: "catalog hash",
			mutate: func(t *testing.T, releaseRoot string, _ *Manifest) {
				t.Helper()
				catalogPath := filepath.Join(releaseRoot, "catalog.json")
				contents, err := os.ReadFile(catalogPath)
				if err != nil {
					t.Fatal(err)
				}
				contents[1] = ' '
				if err := os.WriteFile(catalogPath, contents, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: "catalog.json SHA-256",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			releaseRoot := t.TempDir()
			release := writeReleaseFixture(t, releaseRoot)
			manifest := fixtureManifest(release)
			test.mutate(t, releaseRoot, &manifest)
			writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), manifest)

			_, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Update() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestUpdateRejectsTraversalSymlinksAndDestinationCollisions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, repoRoot, releaseRoot string, manifest *Manifest)
		want   string
	}{
		{
			name: "source traversal",
			mutate: func(_ *testing.T, _, _ string, manifest *Manifest) {
				manifest.Mappings[0].Source = "../goshtoso-icon.svg"
			},
			want: "unsafe source",
		},
		{
			name: "destination traversal",
			mutate: func(_ *testing.T, _, _ string, manifest *Manifest) {
				manifest.Mappings[0].Destination = "../goshtoso-icon.svg"
			},
			want: "unsafe destination",
		},
		{
			name: "case folded destination collision",
			mutate: func(_ *testing.T, _, _ string, manifest *Manifest) {
				manifest.Mappings[1].Destination = strings.ToUpper(manifest.Mappings[0].Destination)
			},
			want: "destination collision",
		},
		{
			name: "source symlink",
			mutate: func(t *testing.T, _, releaseRoot string, manifest *Manifest) {
				t.Helper()
				target := filepath.Join(releaseRoot, filepath.FromSlash(manifest.Mappings[0].Source))
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("../../unlisted.txt", target); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
		{
			name: "destination parent symlink",
			mutate: func(t *testing.T, repoRoot, _ string, _ *Manifest) {
				t.Helper()
				if err := os.MkdirAll(filepath.Join(repoRoot, "site", "internal"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(t.TempDir(), filepath.Join(repoRoot, "site", "internal", "brand")); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			releaseRoot := t.TempDir()
			release := writeReleaseFixture(t, releaseRoot)
			manifest := fixtureManifest(release)
			test.mutate(t, repoRoot, releaseRoot, &manifest)
			writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), manifest)

			_, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Update() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestUpdateRequiresExactReleaseJSONAndSelectedFileHashes(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, releaseRoot string, manifest *Manifest)
		want   string
	}{
		{
			name: "release json identity",
			mutate: func(_ *testing.T, _ string, manifest *Manifest) {
				manifest.ReleaseJSONSHA256 = strings.Repeat("0", 64)
			},
			want: "release.json SHA-256",
		},
		{
			name: "selected file bytes",
			mutate: func(t *testing.T, releaseRoot string, manifest *Manifest) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(releaseRoot, filepath.FromSlash(manifest.Mappings[1].Source)), []byte("<svg>evil</svg>\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: "SHA-256",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			releaseRoot := t.TempDir()
			release := writeReleaseFixture(t, releaseRoot)
			manifest := fixtureManifest(release)
			test.mutate(t, releaseRoot, &manifest)
			writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), manifest)

			_, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Update() error = %v, want %q", err, test.want)
			}
		})
	}
}

func fixtureManifest(release fixtureRelease) Manifest {
	return Manifest{
		SchemaVersion:     1,
		AssetsRepository:  "araihu/assets",
		AssetsRevision:    strings.Repeat("a", 40),
		Release:           "v1.2.3",
		ReleaseURL:        "https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz",
		ReleaseSHA256:     strings.Repeat("b", 64),
		ReleaseJSONSHA256: release.releaseJSONSHA256,
		Mappings: []Mapping{
			{
				Source: "icons/brand/goshtoso-icon-adaptive-transparent-optical.svg", Destination: "site/internal/brand/assets/goshtoso-icon-transparent.svg",
				CanonicalName: "goshtoso-icon-adaptive-transparent-optical", Namespace: "brand", Product: "goshtoso",
				Artwork: "icon", Appearance: "adaptive", Surface: "transparent", Framing: "optical", Format: "svg",
			},
			{
				Source: "brand/goshtoso/logo/adaptive-transparent-optical.svg", Destination: "site/internal/brand/assets/goshtoso-logo-transparent.svg",
				CanonicalName: "goshtoso-logo-adaptive-transparent-optical", Namespace: "brand", Product: "goshtoso",
				Artwork: "logo", Appearance: "adaptive", Surface: "transparent", Framing: "optical", Format: "svg",
			},
		},
	}
}

type fixtureRelease struct {
	releaseJSONSHA256 string
}

func writeReleaseFixture(t *testing.T, root string) fixtureRelease {
	t.Helper()
	files := map[string][]byte{
		"brand/goshtoso/logo/adaptive-transparent-optical.svg":       []byte("<svg>logo</svg>\n"),
		"icons/brand/goshtoso-icon-adaptive-transparent-optical.svg": []byte("<svg>icon</svg>\n"),
		"unlisted.txt": []byte("do not copy\n"),
	}
	catalog := map[string]any{
		"schemaVersion":    1,
		"release":          "v1.2.3",
		"identityRevision": 11,
		"assets": []map[string]any{
			catalogFixtureAsset("goshtoso-icon-adaptive-transparent-optical", "icons/brand/goshtoso-icon-adaptive-transparent-optical.svg", "icon", files),
			catalogFixtureAsset("goshtoso-logo-adaptive-transparent-optical", "brand/goshtoso/logo/adaptive-transparent-optical.svg", "logo", files),
		},
	}
	catalogBytes := marshalJSON(t, catalog)
	files["catalog.json"] = catalogBytes

	for name, contents := range files {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	inventory := make([]map[string]any, 0, len(files))
	for name, contents := range files {
		inventory = append(inventory, map[string]any{
			"path": name, "sha256": fixtureDigest(contents), "size": len(contents),
		})
	}
	releaseJSON := map[string]any{
		"schemaVersion":    1,
		"release":          "v1.2.3",
		"identityRevision": 11,
		"runtimeVersion":   1,
		"catalogSha256":    fixtureDigest(catalogBytes),
		"themesSha256":     strings.Repeat("c", 64),
		"campaignsSha256":  strings.Repeat("d", 64),
		"files":            inventory,
	}
	releaseBytes := marshalJSON(t, releaseJSON)
	if err := os.WriteFile(filepath.Join(root, "release.json"), releaseBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	return fixtureRelease{releaseJSONSHA256: fixtureDigest(releaseBytes)}
}

func catalogFixtureAsset(name, namePath, artwork string, files map[string][]byte) map[string]any {
	return map[string]any{
		"canonicalName": name,
		"namespace":     "brand",
		"path":          namePath,
		"product":       "goshtoso",
		"artwork":       artwork,
		"appearance":    "adaptive",
		"surface":       "transparent",
		"framing":       "optical",
		"format":        "svg",
		"dimensions":    map[string]any{"viewBox": "0 0 16 16"},
		"spriteSymbol":  "",
		"colorBehavior": "protected",
		"license":       "Arai Hu Brand Terms",
		"source":        "fixture",
		"sha256":        fixtureDigest(files[namePath]),
	}
}

func writeOldFallbacks(t *testing.T, root string, mappings []Mapping) {
	t.Helper()
	for index, mapping := range mappings {
		target := filepath.Join(root, filepath.FromSlash(mapping.Destination))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(fmt.Sprintf("old-%d\n", index)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func snapshotFiles(t *testing.T, root string, paths []string) map[string][]byte {
	t.Helper()
	snapshot := make(map[string][]byte, len(paths))
	for _, name := range paths {
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		snapshot[name] = contents
	}
	return snapshot
}

func assertFilesMatch(t *testing.T, root string, paths []string, want map[string][]byte) {
	t.Helper()
	for _, name := range paths {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want[name]) {
			t.Fatalf("%s changed after rollback\ngot:  %q\nwant: %q", name, got, want[name])
		}
	}
}

func countStagedFiles(t *testing.T, root string) int {
	t.Helper()
	count := 0
	if err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(entry.Name(), ".araihu-assets-") && strings.HasSuffix(entry.Name(), ".tmp") {
			count++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return count
}

func writeJSON(t *testing.T, name string, value any) {
	t.Helper()
	if err := os.WriteFile(name, marshalJSON(t, value), 0o644); err != nil {
		t.Fatal(err)
	}
}

func marshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(contents, '\n')
}

func fixtureDigest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
