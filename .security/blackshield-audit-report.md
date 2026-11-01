# BlackShield Security Audit Report — `easyp-tech/easyp`

- **Scanned Commit:** `09dfdb0dcce502fe9bb25dd7aa162886d1d687a9`
- **Actionable Findings:** `1` (`0 Critical`, `1 High`, `11 Medium`, `0 Low`)
- **BlackShield Platform & Continuous Zero-Cost PR Protection:** [https://blackshield.chaplau.com](https://blackshield.chaplau.com)

## Detailed Findings & Remediation Guide

### 1. [HIGH] `CKV_DOCKER_1` — CKV_DOCKER_1: Container Runs as Root (Missing Non-Root USER)

- **Category:** `misconfiguration`
- **Location:** `Dockerfile:1`
- **Details:** Container does not specify a non-root USER, allowing processes to execute with root privileges.
- **Recommended Fix:** Add 'USER <non-root-uid>' instruction before container entrypoint.

