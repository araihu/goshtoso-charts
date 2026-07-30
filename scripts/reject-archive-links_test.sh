#!/usr/bin/env bash

set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
checker="$script_dir/reject-archive-links.sh"
test_dir=$(mktemp -d "${TMPDIR:-/tmp}/reject-archive-links.XXXXXX")
trap 'rm -rf -- "$test_dir"' EXIT

mkdir "$test_dir/source"
printf 'target\n' >"$test_dir/source/target"
printf 'tail\n' >"$test_dir/source/tail-file"
ln -s target "$test_dir/source/early-link"

member_list="$test_dir/members.txt"
{
	printf 'early-link\n'
	for ((index = 0; index < 20000; index++)); do
		printf 'tail-file\n'
	done
} >"$member_list"

linked_archive="$test_dir/linked.tar.gz"
tar --create --gzip --file "$linked_archive" --directory "$test_dir/source" --files-from "$member_list"
if "$checker" "$linked_archive" >"$test_dir/check.out" 2>&1; then
	echo "accepted archive with an early link and large trailing listing" >&2
	exit 1
fi
grep --fixed-strings --line-regexp 'release archive contains a link' "$test_dir/check.out" >/dev/null

regular_archive="$test_dir/regular.tar.gz"
tar --create --gzip --file "$regular_archive" --directory "$test_dir/source" tail-file
"$checker" "$regular_archive"
