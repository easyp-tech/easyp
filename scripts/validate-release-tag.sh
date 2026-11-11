#!/usr/bin/env bash
# Decides whether a pushed tag may be published, and on which channel.
# Prints "stable" or "nightly"; anything else exits non-zero before the
# release job touches registry credentials.
#
#   stable   v1.X.Y                       from origin/main only
#   nightly  v1.X.Y-nightly.YYYYMMDD.N    from origin/main or origin/v1.0
#
# v0 releases belong to the v0 sources and their own pipeline.
set -euo pipefail

tag=${1:-${GITHUB_REF_NAME:-}}
number='(0|[1-9][0-9]*)'
stable="^v1\\.${number}\\.${number}\$"
nightly="^v1\\.${number}\\.${number}-nightly\\.[1-9][0-9]{7}\\.[1-9][0-9]*\$"

if [[ $tag =~ $stable ]]; then
  channel=stable
  branches=(origin/main)
elif [[ $tag =~ $nightly ]]; then
  channel=nightly
  branches=(origin/main origin/v1.0)
else
  echo "Unsupported release tag: $tag" >&2
  echo "This branch publishes v1.X.Y and v1.X.Y-nightly.YYYYMMDD.N (N starts at 1)." >&2
  exit 1
fi

if ! git rev-parse --verify --quiet "refs/tags/$tag^{commit}" > /dev/null; then
  echo "Release tag $tag does not resolve to a commit." >&2
  exit 1
fi

for branch in "${branches[@]}"; do
  if git rev-parse --verify --quiet "refs/remotes/$branch^{commit}" > /dev/null &&
     git merge-base --is-ancestor "refs/tags/$tag^{commit}" "refs/remotes/$branch"; then
    printf '%s\n' "$channel"
    exit 0
  fi
done

if [[ $channel == stable ]]; then
  echo "Stable tag $tag is not in origin/main history." >&2
else
  echo "Nightly tag $tag is not in origin/v1.0 or origin/main history." >&2
fi
exit 1
