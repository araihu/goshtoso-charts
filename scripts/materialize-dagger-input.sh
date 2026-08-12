#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 2 ]; then
	echo "usage: materialize-dagger-input.sh TYPE OUTPUT" >&2
	exit 2
fi

input_type=$1
output=$2
mkdir -p -- "$(dirname -- "$output")"

case "$input_type" in
assets)
	jq -n \
		--arg assets_repository "${ASSETS_REPOSITORY:?}" \
		--arg assets_revision "${ASSETS_REVISION:?}" \
		--arg release "${RELEASE:?}" \
		--arg release_url "${RELEASE_URL:?}" \
		--arg release_sha256 "${RELEASE_SHA256:?}" \
		--arg release_json_sha256 "${RELEASE_JSON_SHA256:?}" \
		--arg source_repository "${GITHUB_REPOSITORY:?}" \
		'{
			assets_repository: $assets_repository,
			assets_revision: $assets_revision,
			release: $release,
			release_url: $release_url,
			release_sha256: $release_sha256,
			release_json_sha256: $release_json_sha256,
			source_repository: $source_repository
		}' >"$output"
	;;
deploy)
	jq -n \
		--arg ref_type "${REF_TYPE:?}" \
		--arg ref_name "${REF_NAME:?}" \
		--arg source_sha "${SOURCE_SHA:?}" \
		--arg source_run_id "${SOURCE_RUN_ID:?}" \
		--arg source_repository "${GITHUB_REPOSITORY:?}" \
		'{
			ref_type: $ref_type,
			ref_name: $ref_name,
			source_sha: $source_sha,
			source_run_id: $source_run_id,
			source_repository: $source_repository
		}' >"$output"
	;;
*)
	echo "unsupported Dagger input type: $input_type" >&2
	exit 2
	;;
esac
