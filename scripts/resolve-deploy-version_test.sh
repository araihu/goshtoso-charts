#!/usr/bin/env bash

set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
resolver="$script_dir/resolve-deploy-version.sh"
source_sha=0123456789abcdef0123456789abcdef01234567

if got=$("$resolver" branch main "$source_sha"); then
	if [ "$got" != "commit-0123456789ab" ]; then
		echo "main version = $got, want commit-0123456789ab" >&2
		exit 1
	fi
else
	echo "main version resolution failed" >&2
	exit 1
fi

for release in v0.0.0 v1.2.3 v1.2.3-rc.1+build.7; do
	if got=$("$resolver" tag "$release" "$source_sha"); then
		if [ "$got" != "$release" ]; then
			echo "tag version = $got, want $release" >&2
			exit 1
		fi
	else
		echo "tag version resolution failed: $release" >&2
		exit 1
	fi
done

assert_rejected() {
	local description=$1
	shift
	if "$resolver" "$@" >/dev/null 2>&1; then
		echo "accepted invalid deployment identity: $description" >&2
		exit 1
	fi
}

assert_rejected "non-main branch" branch feature/example "$source_sha"
assert_rejected "short source SHA" branch main 0123456789ab
assert_rejected "uppercase source SHA" branch main 0123456789ABCDEF0123456789abcdef01234567
assert_rejected "invalid release tag" tag v1.2 "$source_sha"
assert_rejected "empty tag source SHA" tag v1.2.3 ""
assert_rejected "arbitrary tag source SHA" tag v1.2.3 not-a-sha
assert_rejected "uppercase tag source SHA" tag v1.2.3 0123456789ABCDEF0123456789abcdef01234567
assert_rejected "newline tag source SHA" tag v1.2.3 $'0123456789abcdef0123456789abcdef01234567\npoison'
assert_rejected "unsupported ref type" pull_request main "$source_sha"
