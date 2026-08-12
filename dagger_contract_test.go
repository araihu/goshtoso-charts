package charts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDaggerTypeScriptRuntimePackageContract(t *testing.T) {
	type packageManifest struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	type lockPackage struct {
		Resolved     string            `json:"resolved"`
		Link         bool              `json:"link"`
		Dependencies map[string]string `json:"dependencies"`
	}
	type packageLock struct {
		LockfileVersion int                    `json:"lockfileVersion"`
		Packages        map[string]lockPackage `json:"packages"`
	}

	root := ".dagger"
	var manifest packageManifest
	readDaggerJSON(t, filepath.Join(root, "package.json"), &manifest)
	if got := manifest.Dependencies["@dagger.io/dagger"]; got != "./sdk" {
		t.Fatalf("Dagger SDK dependency = %q, want generated local ./sdk", got)
	}
	if got := manifest.Dependencies["typescript"]; got != "^6.0.3" {
		t.Fatalf("TypeScript runtime dependency = %q, want ^6.0.3", got)
	}
	if _, ok := manifest.DevDependencies["typescript"]; ok {
		t.Fatal("TypeScript is a devDependency, so npm --omit=dev removes a runtime requirement")
	}

	var lock packageLock
	readDaggerJSON(t, filepath.Join(root, "package-lock.json"), &lock)
	if lock.LockfileVersion != 3 {
		t.Fatalf("Dagger npm lockfile version = %d, want 3", lock.LockfileVersion)
	}
	lockedRoot := lock.Packages[""]
	if got := lockedRoot.Dependencies["@dagger.io/dagger"]; got != "./sdk" {
		t.Fatalf("locked Dagger SDK dependency = %q, want ./sdk", got)
	}
	if got := lockedRoot.Dependencies["typescript"]; got != "^6.0.3" {
		t.Fatalf("locked TypeScript runtime dependency = %q, want ^6.0.3", got)
	}
	lockedSDK := lock.Packages["node_modules/@dagger.io/dagger"]
	if !lockedSDK.Link || lockedSDK.Resolved != "sdk" {
		t.Fatalf("locked Dagger SDK is not the generated local link: %+v", lockedSDK)
	}
}

func TestDaggerProvidesPinnedJQForPayloadRegressions(t *testing.T) {
	data, err := os.ReadFile(".dagger/src/index.ts")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, required := range []string{
		"ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486",
		`.file("/jq")`,
		`.withFile("/usr/local/bin/jq", jq, { permissions: 0o755 })`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("Dagger module misses pinned jq runtime contract %q", required)
		}
	}
}

func TestPullRequestsCannotMountPersistentCaches(t *testing.T) {
	data, err := os.ReadFile(".dagger/src/index.ts")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	guard := `if (partition === "fork" || partition === "internal") {
      return project
    }`
	guardIndex := strings.Index(source, guard)
	cacheIndex := strings.Index(source, ".withMountedCache(")
	if guardIndex < 0 || cacheIndex < 0 || guardIndex > cacheIndex {
		t.Fatal("pull-request cache guard must return before every persistent cache mount")
	}
}

func readDaggerJSON(t *testing.T, path string, target any) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(contents, target); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}
