#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
	echo "usage: reject-archive-links.sh ARCHIVE" >&2
	exit 2
fi

# Do not use grep --quiet here. An early match can close the pipe while tar is
# still writing; under pipefail, tar's SIGPIPE status would make the if false.
if tar --list --verbose --gzip --file "$1" | grep --extended-regexp '^[lh]' >/dev/null; then
	echo "release archive contains a link" >&2
	exit 1
fi
