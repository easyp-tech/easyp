---
name: go-code-style
description: "Use when writing, reviewing, or refactoring Go code or creating packages in the github.com/easyp-tech/easyp CLI repository."
argument-hint: "Describe the Go code you are writing or reviewing"
---

# Go Code Style — EasyP CLI

Apply these conventions to <code>github.com/easyp-tech/easyp</code>, the Protocol Buffers CLI toolkit. Start with [AGENTS.md](../../../AGENTS.md), [.spec/README.md](../../../.spec/README.md), and the mandatory [agent rules](../../../.spec/agent-rules.md). Verify relevant source before changing behavior; older code may not follow every current convention.

## Errors and Resource Cleanup

- Wrap propagated call failures with <code>fmt.Errorf("&lt;callee&gt;: %w", err)</code>. Use only the called function or method name, without a package or receiver prefix.
- For <code>os.Open</code>, use <code>"Open: %w"</code>; for <code>source.Fetch</code>, use <code>"Fetch: %w"</code>; for <code>c.protoInfoRead</code>, use <code>"protoInfoRead: %w"</code>.
- Use <code>errors.Is</code> for sentinel identity and <code>errors.As</code> for typed error details. Reuse errors in their owning packages rather than inventing duplicates.
- Never use a bare <code>defer resource.Close()</code> or discard the close error. Wrap the deferred call and log or handle its error. Propagate finalization failures when they affect successful output.
- Log a failure where it is handled or terminates an operation; lower layers normally wrap and return it.

Error ownership follows actual source:

| Owner | Examples |
|-------|----------|
| [internal/core/core.go](../../../internal/core/core.go) | <code>ErrInvalidRule</code>, <code>ErrRepositoryDoesNotExist</code>, <code>ErrEmptyInputFiles</code> |
| [internal/core/dom.go](../../../internal/core/dom.go) | <code>OpenImportFileError</code>, <code>GitRefNotFoundError</code> |
| [internal/modules/immutable_versions.go](../../../internal/modules/immutable_versions.go) | <code>ErrLockedVersionChanged</code> |
| [internal/config/v1/legacy_detection.go](../../../internal/config/v1/legacy_detection.go) | <code>ErrLegacyConfiguration</code> |
| [internal/migration](../../../internal/migration) | Contextual planning and apply errors |

See [.spec/ERRORS.md](../../../.spec/ERRORS.md) and the relevant CLI handler for reporting and exit behavior. There is no central domain-to-gRPC-status mapper. Follow [cmd/easyp/main.go](../../../cmd/easyp/main.go) and the handler's actual return/exit path instead of assigning a universal exit code to a sentinel.

This complete, generic standalone example illustrates wrapping and cleanup; it is not an existing EasyP helper:

~~~go
package example

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// ReadText reads a file and reports any close failure to the logger.
func ReadText(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Open: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			slog.Error("Close", "error", err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("ReadAll: %w", err)
	}
	return string(data), nil
}
~~~

## Assignments, Imports, and Comments

- Assign existing variables and check errors on separate lines. A block-scoped short declaration with <code>:=</code> in an <code>if</code> initializer is allowed.
- Put comments above control flow, never inline on <code>if</code>, <code>for</code>, or <code>return</code> lines.
- Group imports as standard library, third-party, then project packages, separated by blank lines. Project imports start with <code>github.com/easyp-tech/easyp</code>.
- Use English comments and godoc. Every exported symbol needs a comment starting with its name. Put a package comment in one file per package.
- Format Go code with <code>gofmt</code>. Import grouping and tag spellings are repository conventions: [.golangci.yml](../../../.golangci.yml) explicitly enables <code>staticcheck</code>, not <code>gci</code> or <code>tagliatelle</code>. Do not infer enabled linters from these conventions.

## Naming and Package Boundaries

| Responsibility | Source to follow |
|----------------|------------------|
| CLI handlers | [internal/api](../../../internal/api), implementing <code>Handler.Command() *cli.Command</code> |
| Process entry and command registration | [cmd/easyp/main.go](../../../cmd/easyp/main.go) |
| Lint/breaking engines and low-level plugin execution | [internal/core](../../../internal/core) |
| Dependency resolution and repositories | [internal/modules](../../../internal/modules) |
| Generation orchestration | [internal/generation](../../../internal/generation) |
| v1 configuration and schema source | [internal/config/v1](../../../internal/config/v1) |
| Lint rules and tests | [internal/rules](../../../internal/rules), colocated <code>&lt;rule&gt;.go</code> and <code>&lt;rule&gt;_test.go</code> |

Use exported domain types and adapter implementations where required by their consumers; keep local configuration structs unexported. Existing public models such as <code>core.Options</code> and <code>v1.Policy</code> stay exported. Follow neighboring filenames and colocate tests as <code>&lt;file&gt;_test.go</code>.

Reserve zero for new enum types with <code>_ = iota</code> so an unset value is not silently valid. Preserve established public configuration spellings and formats when extending existing types.

## Interfaces and Context

- Put <code>context.Context</code> first in methods that need it, and <code>error</code> last in results. Preserve existing contracts that do not take a context.
- Define small interfaces in the consuming package; prefer one method when sufficient. Do not prefix interface names with <code>I</code>.
- Actual contracts include <code>core.Rule</code> and <code>core.CurrentProjectGitWalker</code> in [dom.go](../../../internal/core/dom.go), <code>modules.Source</code> in [resolve.go](../../../internal/modules/resolve.go), and repository/cache interfaces in [repository.go](../../../internal/modules/repository.go).
- <code>Console</code> belongs to [internal/adapters/console/new.go](../../../internal/adapters/console/new.go). Use the current owner when implementing or mocking it.
- CLI actions receive <code>*cli.Context</code> from urfave/cli v2; pass <code>ctx.Context</code> to context-aware operations. Keep cancellation and mutable state scoped to the operation. Choose concurrency from the actual engine/adapter contract.

## Struct Tags

Follow the existing model's tags. v1 configuration types live in <code>internal/config/v1</code>, shared engine configuration in <code>internal/config</code>. Use the established YAML/JSON keys, usually <code>snake_case</code>, and retain hyphenated public keys such as <code>linters-settings</code> and <code>exclude-rules</code>. Avoid adding serialization tags to internal-only types without a consumer. Never hand-edit generated protobuf code or schema JSON.

## Quick Checklist

- [ ] Error wraps contain the callee name only, with <code>%w</code> and no package/receiver prefix.
- [ ] Existing-variable assignment and error checking are separate; deferred close errors are handled.
- [ ] Imports use the repository module path and the three conventional groups.
- [ ] Exported symbols have English godoc; control-flow comments are on separate lines.
- [ ] Errors and interfaces stay with their actual owners; new enums reserve zero.
- [ ] Public tag spellings are preserved; lint claims match the current configuration.
- [ ] Tests follow [go-testing](../go-testing/SKILL.md); CLI changes follow the legacy-named [epctl-commands](../epctl-commands/SKILL.md).
