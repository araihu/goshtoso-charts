#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 3 ]; then
	echo "usage: resolve-deploy-version.sh REF_TYPE REF_NAME SOURCE_SHA" >&2
	exit 2
fi

ref_type=$1
ref_name=$2
source_sha=$3
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

if [[ ! "$source_sha" =~ ^[0-9a-f]{40}$ ]]; then
	echo "source SHA is not a full lowercase Git SHA-1" >&2
	exit 1
fi

case "$ref_type" in
tag)
	if ! "$script_dir/validate-release-version.sh" "$ref_name"; then
		echo "ref tag is not a supported semantic version" >&2
		exit 1
	fi
	printf '%s\n' "$ref_name"
	;;
branch)
	if [ "$ref_name" != "main" ]; then
		echo "ref branch is not deployable: $ref_name" >&2
		exit 1
	fi
	printf 'commit-%s\n' "${source_sha:0:12}"
	;;
*)
	echo "unsupported ref type: $ref_type" >&2
	exit 1
	;;
esac
