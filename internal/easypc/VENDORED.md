# Vendored: easypc compiler

This tree is a verbatim copy-in of the `easypc` compiler from the private
`easyp-tech/compiler-lab` repository. It is sealed: see the prohibition below.

- Upstream repository: `easyp-tech/compiler-lab` (private)
- Branch: `main` — the former `easyp-cli` branch was merged under the
  `easyp-cli/` directory in this SHA, so "branch `easyp-cli`" resolves to
  main + directory
- Directory in upstream: `easyp-cli/internal/easypc`
  (`internal/protoast` in this repository was vendored from the sibling
  `easyp-cli/internal/protoast`; protoast is the maintained adaptation seam
  and is NOT sealed)
- Pinned SHA: `08cef10909c0722f7042ba17b5c11951f9c6b6f6`
- Vendored: 2026-11-17
- Import paths: already `github.com/easyp-tech/easyp/internal/easypc/...` —
  no rewrite was performed at copy-in (the one-time rewrite exemption is
  therefore exhausted)
- Edits inside this tree are FORBIDDEN (see the "Do not" section in
  AGENTS.md). Compiler fixes go upstream in compiler-lab and arrive here by
  re-sync.
- Lint findings inside this tree are upstream's: it is excluded via
  `.golangci.yml` (`linters.exclusions.paths`), so golangci-lint stays green
  without touching vendored code. Findings anywhere else remain actionable.

## Re-sync procedure

1. Clone compiler-lab and check out the new pinned SHA:
   `gh repo clone easyp-tech/compiler-lab /tmp/compiler-lab && git -C /tmp/compiler-lab checkout <new SHA>`
2. Diff `easyp-cli/internal/easypc` against this tree and review the delta.
3. Copy the upstream tree over `internal/easypc`.
4. Re-run the vendoring gates: `go build ./...`,
   `go test -count=1 ./internal/easypc/... ./internal/protoast/...`,
   `go mod tidy && git diff --exit-code go.mod go.sum`,
   `bin/golangci-lint run ./...`.
5. Update this file's SHA and vendored date.

## protoast divergence (re-apply on every sync)

`internal/protoast` was vendored from the same pinned SHA and directory
(`easyp-cli/internal/protoast`). Unlike this tree, protoast is NOT sealed: it
is the maintained adaptation seam where easyp's output contract meets the
vendored compiler. It carries one deliberate divergence from upstream:

- `build.go` `pos()` computes 1-based columns **counted in runes, tab = +1**
  (the legacy go-protoparser output contract that every lint issue position
  uses) instead of upstream protoc's 1-based byte columns with tab stops of 8.
  Line numbering is untouched. Proven by `internal/protoast/position_test.go`
  (D-07a) and `internal/core/position_parity_test.go` (D-07b).
- `protoast_test.go` pins the tab-as-1 expectations (enum value column 2,
  number column 13) with rune-convention wording instead of upstream's
  tab-8 assertions.

A future compiler-lab re-sync MUST re-apply this divergence after copying:
after the copy step, re-apply the `pos()` rune-column computation in
`build.go`, the position wording in `protoast.go`'s package doc, and the tab
expectations in `protoast_test.go`, then re-run both D-07 tests. The easypc
tree itself stays verbatim.
