# Arai Hu fallback assets

Goshtoso Charts keeps two embedded, repository-owned brand fallbacks synchronized
with immutable [`araihu/assets`](https://github.com/araihu/assets) releases:

- `site/internal/brand/assets/goshtoso-logo-transparent.svg`
- `site/internal/brand/assets/goshtoso-icon-transparent.svg`

The root [`araihu-assets.json`](../araihu-assets.json) manifest pins the release
tag, release commit, archive URL and SHA-256, exact `release.json` SHA-256, and
both allowed source-to-destination mappings. Each mapping also pins its catalog
canonical name and semantic roles. The updater rejects any release where the
canonical name resolves to different path, role, or checksum.

These files are only the no-JavaScript and failure baseline for the documentation
site. Theme CSS, App Shell assets, and presentation runtime remain owned by the
released Goshtoso and Goshtoso App Shell dependencies; this updater does not
vendor them.

## Local update

Download and verify the immutable tar archive separately, then extract it into
a new directory. The Go updater deliberately has no download mode:

```bash
go run ./cmd/araihu-assets-update -release-dir /path/to/verified-release
```

To advance the manifest, provide the complete release identity as one unit:

```bash
go run ./cmd/araihu-assets-update \
  -release-dir /path/to/verified-release \
  -assets-repository araihu/assets \
  -assets-revision 0123456789abcdef0123456789abcdef01234567 \
  -release v1.2.3 \
  -release-url https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz \
  -release-sha256 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
  -release-json-sha256 abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789
```

Only stable `vX.Y.Z` tags and the exact `araihu/assets` release URL shape are
accepted. Source and destination traversal, symlinks, case-folded collisions,
unknown catalog collisions, and checksum mismatches fail before any copy. Run
the same command twice; the second run must report that fallbacks are current.
Every changed output is staged and synced before replacement. Fallbacks are
replaced before the manifest; a mid-apply failure restores replaced files in
reverse order and reports any incomplete rollback operation.

Focused verification:

```bash
go test ./internal/araihuassets ./cmd/araihu-assets-update -count=1
```

## Automation

`.github/workflows/araihu-assets.yml` accepts `repository_dispatch` event
`araihu-assets-released` and guarded manual dispatch. Both use the same six
fields:

- `assets_repository`
- `assets_revision`
- `release`
- `release_url`
- `release_sha256`
- `release_json_sha256`

The workflow validates every field and resolves the release tag to the
dispatched commit before download. It downloads and verifies one archive, runs
the offline updater twice to prove idempotence, and opens or updates
`automation/araihu-assets-vX.Y.Z`. It uses selected-repository GitHub App
secrets `ARAIHU_ASSETS_APP_ID` and `ARAIHU_ASSETS_APP_PRIVATE_KEY`. Existing
`dependencies` and `assets` labels are applied when present. It creates no
labels and never auto-merges.

## Known hardening debt

Archive SHA-256 verification makes the published archive immutable, and the
workflow rejects traversal and link members before extraction. GSC-TD-016
tracks rejecting duplicate member names, case-folded member collisions, and
every non-regular member type before extraction. The offline updater already
rejects release-inventory, catalog, destination, path, symlink, checksum, and
transaction violations.
