#!/usr/bin/env bash

set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
materializer="$script_dir/materialize-dagger-input.sh"
test_dir=$(mktemp -d "${TMPDIR:-/tmp}/materialize-dagger-input.XXXXXX")
trap 'rm -rf -- "$test_dir"' EXIT

injection=$'v1.2.3\n--github-token=env://ATTACKER_TOKEN; $(touch '"$test_dir"$'/never-run)'

ASSETS_REPOSITORY='araihu/assets' \
	ASSETS_REVISION='0123456789abcdef0123456789abcdef01234567' \
	RELEASE="$injection" \
	RELEASE_URL='https://example.invalid/a b;$(false)' \
	RELEASE_SHA256='0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef' \
	RELEASE_JSON_SHA256='abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789' \
	GITHUB_REPOSITORY='araihu/goshtoso-charts' \
	"$materializer" assets "$test_dir/assets.json"

jq -e --arg expected "$injection" '.release == $expected' "$test_dir/assets.json" >/dev/null
jq -e '.release_url == "https://example.invalid/a b;$(false)"' "$test_dir/assets.json" >/dev/null
jq -e 'keys == ["assets_repository", "assets_revision", "release", "release_json_sha256", "release_sha256", "release_url", "source_repository"]' "$test_dir/assets.json" >/dev/null

REF_TYPE='tag' \
	REF_NAME="$injection" \
	SOURCE_SHA='0123456789abcdef0123456789abcdef01234567' \
	SOURCE_RUN_ID='12345' \
	GITHUB_REPOSITORY='araihu/goshtoso-charts' \
	"$materializer" deploy "$test_dir/deploy.json"

jq -e --arg expected "$injection" '.ref_name == $expected' "$test_dir/deploy.json" >/dev/null
jq -e 'keys == ["ref_name", "ref_type", "source_repository", "source_run_id", "source_sha"]' "$test_dir/deploy.json" >/dev/null
test ! -e "$test_dir/never-run"

if "$materializer" unknown "$test_dir/unknown.json" >/dev/null 2>&1; then
	echo "unsupported input type unexpectedly passed" >&2
	exit 1
fi
