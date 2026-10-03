#!/bin/sh
# Offline Taskfile regression checks. No real installer, Go tool or Docker is run.
set -eu
repo=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
task_bin=$(command -v task) || { echo 'Task 3 is required on PATH.' >&2; exit 1; }
work=$(mktemp -d)
work=$(CDPATH= cd -- "$work" && pwd -P)
trap 'rm -rf "$work"' EXIT HUP INT TERM
fixture="$work/repo with spaces and 'quotes'"
trace_bin="$work/trace bin"
mkdir -p "$fixture/bin" "$fixture/nested/dir" "$trace_bin" "$work/tmp dir"
cp "$repo/Taskfile.yml" "$repo/.mockery.yaml" "$repo/Dockerfile" "$repo/go.mod" "$repo/go.sum" "$fixture/"
cp "$fixture/go.mod" "$work/go.mod.before"
cp "$fixture/go.sum" "$work/go.sum.before"

# An allowlist PATH prevents regressions from reaching real network/system tools.
for utility in sh cat cmp mkdir mktemp rm; do
    ln -s "$(command -v "$utility")" "$trace_bin/$utility"
done
ln -s "$task_bin" "$trace_bin/task"
cat > "$work/trace-tool" <<'SH'
#!/bin/sh
set -eu
tool=${0##*/}
entry="$tool cwd=<$PWD>"
for arg do entry="$entry <$arg>"; done
printf '%s\n' "$entry" >> "$TRACE_DIR/$tool.log"
test "$PWD" = "$FIXTURE" || { echo 'Tool ran outside repository root' >&2; exit 1; }
case "$tool" in
    go)
        case "$*" in
            'mod download') ;;
            'install gotest.tools/gotestsum@v1.13.0'|'install github.com/vektra/mockery/v2@v2.53.7')
                test "$GOBIN" = "$FIXTURE/bin" ;;
            *) echo "Unexpected Go command: $*" >&2; exit 1 ;;
        esac ;;
    golangci-lint)
        test "$#" = 2 && test "$1" = run && test "$2" = ./...
        exit "${GO_LINT_EXIT:-0}" ;;
    gotestsum)
        test "$*" = '--format pkgname -- -coverprofile=coverage.out -race -count=1 ./...' ;;
    mockery)
        test "$#" = 6 && test "$1" = --name && test "$3" = --dir && test "$5" = --output
        test "$6" = "$4/mocks"
        case "$2:$4" in
            Rule:./internal/core|CurrentProjectGitWalker:./internal/core|Console:./internal/adapters/console|Example:'./interface dir') ;;
            *) echo "Unexpected mock owner: $*" >&2; exit 1 ;;
        esac ;;
    docker)
        case "$1" in
            run)
                test "$*" = 'run --rm -i ghcr.io/hadolint/hadolint:v2.12.1-beta'
                cmp - "$FIXTURE/Dockerfile"
                exit "${DOCKER_LINT_EXIT:-0}" ;;
            build)
                test "$#" = 6 && test "$2" = -f && test "$3" = "$FIXTURE/Dockerfile"
                test "$4" = -t && test "$6" = "$FIXTURE"
                case "$5" in easyp:local|easyp:smoke|easyp:alias-smoke) ;; *) exit 1 ;; esac ;;
            *) echo "Unexpected Docker action: $*" >&2; exit 1 ;;
        esac ;;
    curl)
        test "$#" = 4 && test "$1" = -fsSL
        test "$2" = https://raw.githubusercontent.com/golangci/golangci-lint/v2.14.0/install.sh
        test "$3" = -o
        cat > "$4" <<'INSTALLER'
#!/bin/sh
set -eu
test "$#" = 3 && test "$1" = -b && test "$2" = "$FIXTURE/bin" && test "$3" = v2.14.0
printf '%s\n' installed >> "$TRACE_DIR/installer.log"
exit "${INSTALLER_EXIT:-0}"
INSTALLER
        exit "${CURL_EXIT:-0}" ;;
    *) echo "Unexpected tool: $tool" >&2; exit 1 ;;
esac
SH
chmod +x "$work/trace-tool"
for utility in go docker curl; do
    ln -s "$work/trace-tool" "$trace_bin/$utility"
done
for utility in golangci-lint gotestsum mockery; do
    ln -s "$work/trace-tool" "$fixture/bin/$utility"
done

run_task() {
    (cd "$fixture/nested/dir" && env -i PATH="$trace_bin" TMPDIR="$work/tmp dir" LC_ALL=C \
        FIXTURE="$fixture" TRACE_DIR="$work" \
        GO_LINT_EXIT="${GO_LINT_EXIT:-0}" DOCKER_LINT_EXIT="${DOCKER_LINT_EXIT:-0}" \
        CURL_EXIT="${CURL_EXIT:-0}" INSTALLER_EXIT="${INSTALLER_EXIT:-0}" \
        "$task_bin" "$@")
}
pass() {
    if ! run_task "$@" > "$work/task.log" 2>&1; then
        cat "$work/task.log" >&2
        echo "Expected task to succeed: $*" >&2
        exit 1
    fi
}
fail() {
    if run_task "$@" > "$work/task.log" 2>&1; then
        echo "Expected task to fail: $*" >&2
        exit 1
    fi
}
contains() {
    grep -F -- "$2" "$1" > /dev/null || { cat "$1" >&2; echo "Missing: $2" >&2; exit 1; }
}

pass --list
pass lint:go
pass lint:docker
pass lint
pass test
pass mocks
pass mock NAME=Example 'DIR=./interface dir'
# Aggregate subprocess tasks must preserve explicit local-tool overrides.
alternate_bin="$fixture/alternate tools with 'quotes'"
mv "$fixture/bin" "$alternate_bin"
pass lint "LOCAL_BIN=$alternate_bin"
pass test "LOCAL_BIN=$alternate_bin"
pass mocks "LOCAL_BIN=$alternate_bin"
mv "$alternate_bin" "$fixture/bin"
contains "$work/mockery.log" '<Console> <--dir> <./internal/adapters/console>'
contains "$work/mockery.log" '<Rule> <--dir> <./internal/core>'
contains "$work/mockery.log" '<CurrentProjectGitWalker> <--dir> <./internal/core>'

# Each aggregate must attempt both checks, including when the first one fails.
for failing_linter in go docker; do
    : > "$work/golangci-lint.log"
    : > "$work/docker.log"
    if test "$failing_linter" = go; then
        GO_LINT_EXIT=23; export GO_LINT_EXIT
    else
        DOCKER_LINT_EXIT=24; export DOCKER_LINT_EXIT
    fi
    fail lint
    contains "$work/golangci-lint.log" '<run> <./...>'
    contains "$work/docker.log" '<run> <--rm> <-i>'
    unset GO_LINT_EXIT DOCKER_LINT_EXIT
done

for spec in 'golangci-lint lint:go install_linters' 'gotestsum test install_gotestsum' 'mockery mocks install_mockery'; do
    set -- $spec
    mv "$fixture/bin/$1" "$fixture/bin/$1.saved"
    fail "$2"
    contains "$work/task.log" "task $3"
    mv "$fixture/bin/$1.saved" "$fixture/bin/$1"
done
mv "$trace_bin/docker" "$trace_bin/docker.saved"
fail lint:docker
contains "$work/task.log" 'Docker is required'
fail docker:build
contains "$work/task.log" 'Docker is required'
mv "$trace_bin/docker.saved" "$trace_bin/docker"

pass init
contains "$work/go.log" '<mod> <download>'
contains "$work/installer.log" installed
cp "$work/installer.log" "$work/installer.before"
CURL_EXIT=22; export CURL_EXIT
fail install_linters
unset CURL_EXIT
cmp "$work/installer.before" "$work/installer.log"
INSTALLER_EXIT=25; export INSTALLER_EXIT
fail install_linters
unset INSTALLER_EXIT

: > "$work/docker.log"
pass docker
pass docker:build DOCKER_IMAGE=easyp:smoke
pass docker DOCKER_IMAGE=easyp:alias-smoke
contains "$work/docker.log" '<-t> <easyp:local>'
contains "$work/docker.log" '<-t> <easyp:smoke>'
contains "$work/docker.log" '<-t> <easyp:alias-smoke>'
for obsolete in docker_base docker_lint docker_push; do
    fail "$obsolete"
done
if grep -E '<(push|prune|system|tag)>' "$work/docker.log"; then
    echo 'Docker helper attempted a forbidden action' >&2
    exit 1
fi
cmp "$work/go.mod.before" "$fixture/go.mod"
cmp "$work/go.sum.before" "$fixture/go.sum"
test ! -e "$fixture/nested/dir/bin"
test ! -e "$fixture/easyp"
printf '%s\n' 'Dev tools: offline dispatch, quoting, failures, local Docker builds and manifest checks passed.'
