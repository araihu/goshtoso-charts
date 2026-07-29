#!/usr/bin/env bash

set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
validator="$script_dir/validate-release-version.sh"

for version in \
	'v0.0.0' \
	'v1.2.3' \
	'v1.2.3-0' \
	'v1.2.3-rc.1' \
	'v1.2.3-01a' \
	'v1.2.3+build.01' \
	'v1.2.3-rc.1+build.7'; do
	"$validator" "$version"
done

for version in \
	'1.2.3' \
	'v01.2.3' \
	'v1.02.3' \
	'v1.2.03' \
	'v1.2.3-01' \
	'v1.2.3-foo..bar' \
	'v1.2.3-.foo' \
	'v1.2.3-foo.' \
	'v1.2.3+' \
	'v1.2.3+foo..bar'; do
	if "$validator" "$version"; then
		echo "accepted invalid release version: $version" >&2
		exit 1
	fi
done
