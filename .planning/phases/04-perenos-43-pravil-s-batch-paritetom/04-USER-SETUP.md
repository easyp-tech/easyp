# Phase 04 — User Setup: compiler-lab etalon (private repo, read-only)

**Status: VERIFIED COMPLETE (checked by executor 2026-11-19 at 04-01 start)**

## Why

All 41 rule ports of Phase 4 port FROM the compiler-lab etalon (the post-flip tree where rules already run on protoast). The etalon is a private repo; a `/tmp` copy is the only supported access path.

## Environment checklist

| Item | Required value | Verified |
|------|----------------|----------|
| Etalon checkout | `/tmp/gsd-research/compiler-lab/easyp-cli` | yes (04-01) |
| Etalon SHA | `08cef10909c0722f7042ba17b5c11951f9c6b6f6` | yes (04-01, `git -C /tmp/gsd-research/compiler-lab rev-parse HEAD`) |
| Access | read-only pull access to `easyp-tech/compiler-lab` | yes (existing clone) |

## If `/tmp` is wiped before a later batch plan (04-02..04-07)

1. Re-clone with your GitHub credentials: `gh repo clone easyp-tech/compiler-lab /tmp/gsd-research/compiler-lab`
2. Pin: `git -C /tmp/gsd-research/compiler-lab checkout 08cef10`
3. Verify: `git -C /tmp/gsd-research/compiler-lab rev-parse HEAD` starts with `08cef10`

The executor cannot re-clone itself if private-repo access is missing — that condition stops the affected plan at its precondition check and asks for this setup.

## Verification command

```sh
git -C /tmp/gsd-research/compiler-lab rev-parse HEAD
# expect: 08cef10909c0722f7042ba17b5c11951f9c6b6f6
```
