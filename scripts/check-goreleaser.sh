#!/usr/bin/env bash
# goreleaser check, accepting exactly one deprecation: brews.
#
# GoReleaser deprecates brews (formulas) in favour of homebrew_casks, but
# easyp-tech/homebrew-tap ships a formula that macOS and Linux users already
# upgrade through. Moving it to a cask is its own migration; until then this
# deprecation is expected and any other finding still fails.
set -uo pipefail

goreleaser=${GORELEASER:-goreleaser}
output=$(NO_COLOR=1 "$goreleaser" check "$@" 2>&1)
status=$?
printf '%s\n' "$output"
[[ $status -eq 0 ]] && exit 0
# CI forces colour regardless of NO_COLOR; match on the plain text.
output=$(sed -E $'s/\x1b\\[[0-9;]*m//g' <<< "$output")

deprecated=$(grep -oE 'DEPRECATED: +[a-z_.]+' <<< "$output" | awk '{print $2}' | sort -u)
if [[ $deprecated == brews ]] &&
   grep -q 'configuration is valid, but uses deprecated properties' <<< "$output"; then
  echo "goreleaser check: only the expected brews deprecation; accepted."
  exit 0
fi
exit "$status"
