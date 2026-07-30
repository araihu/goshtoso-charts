// Command araihu-assets-update refreshes repository-owned embedded brand
// fallbacks from an already downloaded and extracted araihu/assets release.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/araihu/goshtoso-charts/internal/araihuassets"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "araihu-assets-update:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("araihu-assets-update", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", ".", "repository root")
	releaseDir := flags.String("release-dir", "", "verified, extracted araihu/assets release directory")
	manifest := flags.String("manifest", araihuassets.DefaultManifestPath, "repository-relative manifest path")
	repository := flags.String("assets-repository", "", "release repository identity")
	revision := flags.String("assets-revision", "", "release commit identity")
	release := flags.String("release", "", "stable release tag")
	releaseURL := flags.String("release-url", "", "immutable release archive URL")
	releaseSHA256 := flags.String("release-sha256", "", "release archive SHA-256")
	releaseJSONSHA256 := flags.String("release-json-sha256", "", "release.json SHA-256")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if *releaseDir == "" {
		return errors.New("-release-dir is required")
	}
	identityValues := []string{*repository, *revision, *release, *releaseURL, *releaseSHA256, *releaseJSONSHA256}
	provided := 0
	for _, value := range identityValues {
		if value != "" {
			provided++
		}
	}
	var identity *araihuassets.ReleaseIdentity
	if provided != 0 {
		if provided != len(identityValues) {
			return errors.New("release identity flags must be provided together")
		}
		identity = &araihuassets.ReleaseIdentity{
			AssetsRepository:  *repository,
			AssetsRevision:    *revision,
			Release:           *release,
			ReleaseURL:        *releaseURL,
			ReleaseSHA256:     *releaseSHA256,
			ReleaseJSONSHA256: *releaseJSONSHA256,
		}
	}

	result, err := araihuassets.Update(araihuassets.Options{
		RepoRoot:     *root,
		ReleaseRoot:  *releaseDir,
		ManifestPath: *manifest,
		Identity:     identity,
	})
	if err != nil {
		return err
	}
	if len(result.Changed) == 0 {
		fmt.Fprintln(stdout, "Arai Hu fallback assets already current")
		return nil
	}
	for _, changed := range result.Changed {
		fmt.Fprintln(stdout, changed)
	}
	return nil
}
