#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
	echo "usage: validate-release-version.sh VERSION" >&2
	exit 2
fi

# SemVer 2.0.0 with the release-specific required v prefix. Numeric identifiers
# cannot contain leading zeroes; pre-release and build identifiers cannot be
# empty.
release_version_pattern='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?(\+([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?$'

[[ "$1" =~ $release_version_pattern ]]
