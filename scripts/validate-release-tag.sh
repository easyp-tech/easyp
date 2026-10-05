#!/usr/bin/env bash
set -euo pipefail

tag=${1:-${GITHUB_REF_NAME:-}}
nightly='^v1\.0\.0-nightly\.[1-9][0-9]{7}\.[1-9][0-9]*$'

if [[ ! $tag =~ $nightly ]]; then
  echo "Unsupported release tag: $tag" >&2
  echo "This branch only publishes v1.0.0-nightly.YYYYMMDD.N (N starts at 1)." >&2
  exit 1
fi

if ! git rev-parse --verify --quiet "refs/tags/$tag^{commit}" > /dev/null; then
  echo "Release tag $tag does not resolve to a commit." >&2
  exit 1
fi

for branch in origin/main origin/v1.0; do
  if git rev-parse --verify --quiet "refs/remotes/$branch^{commit}" > /dev/null &&
     git merge-base --is-ancestor "refs/tags/$tag^{commit}" "refs/remotes/$branch"; then
    printf '%s\n' nightly
    exit 0
  fi
done

echo "Nightly tag $tag is not in origin/v1.0 or origin/main history." >&2
exit 1
