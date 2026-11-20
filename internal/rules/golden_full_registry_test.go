package rules_test

// The Gate-1 compatibility golden (05-01): the byte-frozen old engine's
// complete observable output over the Phase-4 full-registry corpus, frozen as
// committed literals while the oracle still existed (05-01 captured it
// pre-flip; the Phase 5 flip deleted the old engine in the same commit that
// re-pointed this drive, so there is no oracle left to capture from — the
// literals ARE the compatibility contract).
//
// TestLintGoldenFullRegistry is THE golden drive: the linked engine must
// reproduce the committed literals exactly — every field including Message
// and Position (Filename/Offset), in the engine's own D-01 order.
//
// Anti-vacuity and purity guards on the drive:
//   - the pinned 73-finding count is asserted BEFORE the golden compare
//     (T-04-02 pattern);
//   - zero findings named COMPILE — the corpus must keep compiling under the
//     engine's full type resolution; an uncompiling fixture routes the run to
//     COMPILE and vacates the rule-finding golden (Pitfall 7);
//   - the coverage assertion requires every registry rule to fire at least
//     once (a silent rule is a corpus hole, not a pass);
//   - the D-01 totality guard halts on any five-term tie — a tie would make
//     the frozen order engine-nondeterministic (flagged assumption 3), so it
//     must never be papered over.
//
// The corpus map below is re-hosted VERBATIM from the Phase-4 parity corpus
// (same 13 file keys, same contents, same comments): the parity file died
// with its oracle in the flip commit, and the corpus — with its declared rpc
// payload types and directive placements — lives on here unchanged. Do not
// "improve" any fixture.
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/rules"
)

// goldenCorpusUse is the golden config: every registered rule, exactly like
// the Phase-4 closure cell the corpus was captured from.
var goldenCorpusUse = rules.AllRuleNames()

// goldenFindingCount pins the captured finding total, asserted before the
// golden compare on both drives (T-04-02 anti-vacuity).
const goldenFindingCount = 73

// goldenCorpusFiles is the broad compiling corpus: 13 files, 9 packages,
// every batch's edge classes represented. Identifier names are chosen so the
// intended rule fires; unintended co-firings are fine — the old side is the
// oracle and the count pins the total.
var goldenCorpusFiles = map[string]string{
	// Packageless -> PACKAGE_DEFINED at the zero Position literal; invisible
	// to every group rule.
	"pkgless/orphan.proto": `
syntax = "proto3";
`,

	// Same directory, different packages: second.proto issues
	// DIRECTORY_SAME_PACKAGE (first exempt); both issue
	// PACKAGE_DIRECTORY_MATCH (dir_a != samed/v1, dir_a != other/v1).
	"dir_a/v1/first.proto": `
syntax = "proto3";

package samed.v1;
`,
	"dir_a/v1/second.proto": `
syntax = "proto3";

package other.v1;
`,

	// samed.v1 already lives in dir_a: third.proto issues
	// PACKAGE_SAME_DIRECTORY (first directory exempt) and
	// PACKAGE_DIRECTORY_MATCH (dir_b != samed/v1).
	"dir_b/third.proto": `
syntax = "proto3";

package samed.v1;
`,

	// First of the opts.v1 pair: every option value here is the group's
	// first-exempt value for all seven PACKAGE_SAME_* rules.
	"opts/v1/alpha.proto": `
syntax = "proto3";

package opts.v1;

option java_package = "com.alpha";
option java_multiple_files = true;
option go_package = "github.com/alpha/v1";
option csharp_namespace = "Alpha.V1";
option php_namespace = "Alpha\\V1";
option ruby_package = "Alpha::V1";
option swift_prefix = "Alpha";

message AlphaMsg {}
`,

	// Second of the pair: every compared value differs from alpha.proto's —
	// all seven PACKAGE_SAME_* rules issue on this file (raw written
	// constants; swift_prefix reads ruby_package, so its own identical
	// swift_prefix value stays clean while the ruby_package difference
	// issues).
	"opts/v1/beta.proto": `
syntax = "proto3";

package opts.v1;

option java_package = "com.beta";
option java_multiple_files = false;
option go_package = "github.com/beta/v1";
option csharp_namespace = "Beta.V1";
option php_namespace = "Beta\\V1";
option ruby_package = "Beta::V1";
option swift_prefix = "Alpha";

message BetaMsg {}
`,

	// The casing/semantic teeth file. NOTE: the declarations that must fire
	// COMMENT_* deliberately carry NO preceding comment (a comment here
	// would document them); dir_only_msg carries ONLY a suppression
	// directive, which documents it for COMMENT_MESSAGE while
	// MESSAGE_PASCAL_CASE stays centrally suppressed.
	"teeth/v1/teeth.proto": `
syntax = "proto3";

package teeth.v1;

import public "dep/v1/dep.proto";
import weak "wdep/v1/wdep.proto";
import "dep/v1/unused_dep.proto";

enum ABC {
  ABC_FIRST = 0;
  _LEADING = 1;
}

enum ZeroEnum {
  ZeroEnum_00 = 00;
  ZeroEnum_OK = 1;
}

enum AliasEnum {
  option allow_alias = false;

  AliasEnum_FIRST = 0;
}

message badMessage {
  string _field = 1;

  oneof oneof1 {
    string inner_field = 2;
  }
}

// easyp:disable MESSAGE_PASCAL_CASE
message dir_only_msg {}

// The rpc payload messages below are declared so the corpus COMPILES under
// the linked engine's full type resolution (the old engine is syntax-only
// and never resolves rpc types — undefined ones would route the linked side
// to COMPILE and make parity vacuous). Each carries its own documenting
// comment and a PascalCase name, so all of them stay clean for every rule.

// request of get_thing
message GetReq {}

// response of get_thing
message GetResp {}

// request of DoStream
message DoStreamRequest {}

// response of DoStream
message DoStreamResponse {}

// request of Dup and DupAgain
message DupRequest {}

// response of Dup
message DupResponse {}

// response of DupAgain
message DupAgainResponse {}

message UsesDep {
  dep.v1.Outer dep_field = 1;
  wdep.v1.WDepMsg other = 2;
}

service bad_service {
  rpc get_thing(GetReq) returns (GetResp);
}

service StreamService {
  rpc DoStream(stream DoStreamRequest) returns (stream DoStreamResponse);

  rpc Dup(DupRequest) returns (DupResponse);

  rpc DupAgain(DupRequest) returns (DupAgainResponse);
}
`,

	// Referenced (used) public-import payload.
	"dep/v1/dep.proto": `
syntax = "proto3";

package dep.v1;

message Outer {}
`,

	// Referenced (used) weak-import payload.
	"wdep/v1/wdep.proto": `
syntax = "proto3";

package wdep.v1;

message WDepMsg {}
`,

	// Never referenced: the plain import in teeth.proto issues IMPORT_USED.
	"dep/v1/unused_dep.proto": `
syntax = "proto3";

package dep.v1;

message UnusedMsg {}
`,

	// Full-filename check: InvalidName.proto issues FILE_LOWER_SNAKE_CASE at
	// the zero Position literal (directory matches the package, so nothing
	// else co-fires here).
	"inv/v1/InvalidName.proto": `
syntax = "proto3";

package inv.v1;
`,

	// Uppercase package component: PACKAGE_LOWER_SNAKE_CASE (digit-free
	// but for the 'P'); the directory mismatches, so
	// PACKAGE_DIRECTORY_MATCH co-fires; the v2 component satisfies
	// PACKAGE_VERSION_SUFFIX.
	"casing_pkg/teeth2.proto": `
syntax = "proto3";

package badPkg.v2;
`,

	// Version suffix: the single-component lowercase package has no version
	// token -> PACKAGE_VERSION_SUFFIX; the directory matches and the name
	// passes the lower-snake regexp, so nothing else co-fires here.
	"no_version/nover.proto": `
syntax = "proto3";

package no_version;
`,
}

// golden is the compact literal constructor for one committed finding. Every
// field is explicit — Path, RuleName, Position (Filename, Offset, Line,
// Column), SourceName, Message — so the golden block is reviewable one
// finding per line, in the D-01 order of the corpus output.
func golden(path, rule, filename string, offset, line, column int, source, message string) core.IssueInfo {
	return core.IssueInfo{
		Path: path,
		Issue: core.Issue{
			Position:   core.Position{Filename: filename, Offset: offset, Line: line, Column: column},
			SourceName: source,
			Message:    message,
			RuleName:   rule,
		},
	}
}

// goldenIssues is the committed golden: all 73 findings of the byte-frozen
// old engine over the corpus below, every field populated, stored in the
// engine's D-01 order (path, rule, line, column, source). Captured from the
// old engine while the oracle still exists (05-01 Task 1); the linked engine
// must reproduce it exactly, message for message (05-01 Task 2).
var goldenIssues = []core.IssueInfo{
	golden("casing_pkg/teeth2.proto", "FILE_LOWER_SNAKE_CASE", "casing_pkg/teeth2.proto", 0, 0, 0, "casing_pkg/teeth2.proto", "file name should be lower_snake_case.proto"),
	golden("casing_pkg/teeth2.proto", "PACKAGE_DIRECTORY_MATCH", "", 21, 4, 1, "casing_pkg/teeth2.proto", "package \"badPkg.v2\" does not match module-relative directory \"casing_pkg\"; expected \"casing_pkg\""),
	golden("casing_pkg/teeth2.proto", "PACKAGE_LOWER_SNAKE_CASE", "", 21, 4, 1, "badPkg.v2", "package name should be lower_snake_case"),
	golden("dep/v1/dep.proto", "COMMENT_MESSAGE", "", 38, 6, 1, "Outer", "message comments must not be empty"),
	golden("dep/v1/unused_dep.proto", "COMMENT_MESSAGE", "", 38, 6, 1, "UnusedMsg", "message comments must not be empty"),
	golden("dir_a/v1/first.proto", "PACKAGE_DIRECTORY_MATCH", "", 21, 4, 1, "dir_a/v1/first.proto", "package \"samed.v1\" does not match module-relative directory \"dir_a/v1\"; expected \"dir_a.v1\""),
	golden("dir_a/v1/second.proto", "DIRECTORY_SAME_PACKAGE", "", 21, 4, 1, "other.v1", "all files in the same directory must have the same package name"),
	golden("dir_a/v1/second.proto", "PACKAGE_DIRECTORY_MATCH", "", 21, 4, 1, "dir_a/v1/second.proto", "package \"other.v1\" does not match module-relative directory \"dir_a/v1\"; expected \"dir_a.v1\""),
	golden("dir_b/third.proto", "PACKAGE_DIRECTORY_MATCH", "", 21, 4, 1, "dir_b/third.proto", "package \"samed.v1\" does not match module-relative directory \"dir_b\"; expected \"dir_b\""),
	golden("dir_b/third.proto", "PACKAGE_SAME_DIRECTORY", "", 21, 4, 1, "samed.v1", "different proto files in the same package should be in the same directory"),
	golden("inv/v1/InvalidName.proto", "FILE_LOWER_SNAKE_CASE", "inv/v1/InvalidName.proto", 0, 0, 0, "inv/v1/InvalidName.proto", "file name should be lower_snake_case.proto"),
	golden("no_version/nover.proto", "PACKAGE_VERSION_SUFFIX", "", 21, 4, 1, "no_version", "package name should have a version suffix"),
	golden("opts/v1/alpha.proto", "COMMENT_MESSAGE", "", 293, 14, 1, "AlphaMsg", "message comments must not be empty"),
	golden("opts/v1/beta.proto", "COMMENT_MESSAGE", "", 289, 14, 1, "BetaMsg", "message comments must not be empty"),
	golden("opts/v1/beta.proto", "PACKAGE_SAME_CSHARP_NAMESPACE", "", 151, 9, 1, "\"Beta.V1\"", "different proto files in the same package should have the same csharp_namespace"),
	golden("opts/v1/beta.proto", "PACKAGE_SAME_GO_PACKAGE", "", 109, 8, 1, "\"github.com/beta/v1\"", "all files in the same package must have the same go_package name"),
	golden("opts/v1/beta.proto", "PACKAGE_SAME_JAVA_MULTIPLE_FILES", "", 73, 7, 1, "false", "all files in the same package must have the same java_multiple_files option"),
	golden("opts/v1/beta.proto", "PACKAGE_SAME_JAVA_PACKAGE", "", 39, 6, 1, "\"com.beta\"", "all files in the same package must have the same java_package option"),
	golden("opts/v1/beta.proto", "PACKAGE_SAME_PHP_NAMESPACE", "", 188, 10, 1, "\"Beta\\\\V1\"", "all files in the same package must have the same php_namespace option"),
	golden("opts/v1/beta.proto", "PACKAGE_SAME_RUBY_PACKAGE", "", 223, 11, 1, "\"Beta::V1\"", "all files in the same package must have the same ruby_package option"),
	golden("opts/v1/beta.proto", "PACKAGE_SAME_SWIFT_PREFIX", "", 223, 11, 1, "\"Beta::V1\"", "all files in the same package must have the same swift_prefix option"),
	golden("pkgless/orphan.proto", "PACKAGE_DEFINED", "pkgless/orphan.proto", 0, 0, 0, "pkgless/orphan.proto", "package should be defined"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM", "", 143, 10, 1, "ABC", "enum comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM", "", 190, 15, 1, "ZeroEnum", "enum comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM", "", 248, 20, 1, "AliasEnum", "enum comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM_VALUE", "", 156, 11, 3, "ABC_FIRST", "enum value comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM_VALUE", "", 173, 12, 3, "_LEADING", "enum value comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM_VALUE", "", 208, 16, 3, "ZeroEnum_00", "enum value comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM_VALUE", "", 228, 17, 3, "ZeroEnum_OK", "enum value comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ENUM_VALUE", "", 298, 23, 3, "AliasEnum_FIRST", "enum value comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_FIELD", "", 345, 27, 3, "_field", "field comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_FIELD", "", 1221, 65, 3, "dep_field", "field comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_FIELD", "", 1251, 66, 3, "other", "field comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_MESSAGE", "", 322, 26, 1, "badMessage", "message comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_MESSAGE", "", 1201, 64, 1, "UsesDep", "message comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_ONEOF", "", 367, 29, 3, "oneof1", "oneof comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_RPC", "", 1305, 70, 3, "get_thing", "rpc comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_RPC", "", 1375, 74, 3, "DoStream", "rpc comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_RPC", "", 1450, 76, 3, "Dup", "rpc comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_RPC", "", 1496, 78, 3, "DupAgain", "rpc comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_SERVICE", "", 1281, 69, 1, "bad_service", "service comments must not be empty"),
	golden("teeth/v1/teeth.proto", "COMMENT_SERVICE", "", 1349, 73, 1, "StreamService", "service comments must not be empty"),
	golden("teeth/v1/teeth.proto", "ENUM_FIRST_VALUE_ZERO", "", 208, 16, 3, "00", "enum first value must be zero"),
	golden("teeth/v1/teeth.proto", "ENUM_NO_ALLOW_ALIAS", "", 248, 20, 1, "AliasEnum", "enum must not allow alias"),
	golden("teeth/v1/teeth.proto", "ENUM_PASCAL_CASE", "", 143, 10, 1, "ABC", "enum name must be in PascalCase"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_PREFIX", "", 156, 11, 3, "ABC_FIRST", "enum value prefix is not valid"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_PREFIX", "", 173, 12, 3, "_LEADING", "enum value prefix is not valid"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_PREFIX", "", 208, 16, 3, "ZeroEnum_00", "enum value prefix is not valid"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_PREFIX", "", 228, 17, 3, "ZeroEnum_OK", "enum value prefix is not valid"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_PREFIX", "", 298, 23, 3, "AliasEnum_FIRST", "enum value prefix is not valid"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_UPPER_SNAKE_CASE", "", 173, 12, 3, "_LEADING", "enum value must be in UPPER_SNAKE_CASE"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_UPPER_SNAKE_CASE", "", 208, 16, 3, "ZeroEnum_00", "enum value must be in UPPER_SNAKE_CASE"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_UPPER_SNAKE_CASE", "", 228, 17, 3, "ZeroEnum_OK", "enum value must be in UPPER_SNAKE_CASE"),
	golden("teeth/v1/teeth.proto", "ENUM_VALUE_UPPER_SNAKE_CASE", "", 298, 23, 3, "AliasEnum_FIRST", "enum value must be in UPPER_SNAKE_CASE"),
	golden("teeth/v1/teeth.proto", "ENUM_ZERO_VALUE_SUFFIX", "", 156, 11, 3, "ABC_FIRST", "enum zero value suffix is not valid"),
	golden("teeth/v1/teeth.proto", "ENUM_ZERO_VALUE_SUFFIX", "", 208, 16, 3, "ZeroEnum_00", "enum zero value suffix is not valid"),
	golden("teeth/v1/teeth.proto", "ENUM_ZERO_VALUE_SUFFIX", "", 298, 23, 3, "AliasEnum_FIRST", "enum zero value suffix is not valid"),
	golden("teeth/v1/teeth.proto", "FIELD_LOWER_SNAKE_CASE", "", 345, 27, 3, "_field", "message field should be lower_snake_case"),
	golden("teeth/v1/teeth.proto", "IMPORT_NO_PUBLIC", "", 40, 6, 1, "\"dep/v1/dep.proto\"", "import should not be public"),
	golden("teeth/v1/teeth.proto", "IMPORT_NO_WEAK", "", 74, 7, 1, "\"wdep/v1/wdep.proto\"", "import should not be weak"),
	golden("teeth/v1/teeth.proto", "IMPORT_USED", "", 108, 8, 1, "\"dep/v1/unused_dep.proto\"", "import is not used"),
	golden("teeth/v1/teeth.proto", "MESSAGE_PASCAL_CASE", "", 322, 26, 1, "badMessage", "message name should be PascalCase"),
	golden("teeth/v1/teeth.proto", "ONEOF_LOWER_SNAKE_CASE", "", 367, 29, 3, "oneof1", "oneof name should be lower_snake_case"),
	golden("teeth/v1/teeth.proto", "RPC_NO_CLIENT_STREAMING", "", 1375, 74, 3, "DoStream", "client streaming RPCs are not allowed"),
	golden("teeth/v1/teeth.proto", "RPC_NO_SERVER_STREAMING", "", 1375, 74, 3, "DoStream", "server streaming RPCs are not allowed"),
	golden("teeth/v1/teeth.proto", "RPC_PASCAL_CASE", "", 1305, 70, 3, "get_thing", "RPC names should be PascalCase"),
	golden("teeth/v1/teeth.proto", "RPC_REQUEST_RESPONSE_UNIQUE", "", 1496, 78, 3, "DupRequest", "request and response types must be unique across all RPCs"),
	golden("teeth/v1/teeth.proto", "RPC_REQUEST_STANDARD_NAME", "", 1305, 70, 3, "GetReq", "rpc request should have suffix 'Request'"),
	golden("teeth/v1/teeth.proto", "RPC_REQUEST_STANDARD_NAME", "", 1496, 78, 3, "DupRequest", "rpc request should have suffix 'Request'"),
	golden("teeth/v1/teeth.proto", "RPC_RESPONSE_STANDARD_NAME", "", 1305, 70, 3, "GetResp", "rpc response should have suffix 'Response'"),
	golden("teeth/v1/teeth.proto", "SERVICE_PASCAL_CASE", "", 1281, 69, 1, "bad_service", "service names must be PascalCase"),
	golden("teeth/v1/teeth.proto", "SERVICE_SUFFIX", "", 1281, 69, 1, "bad_service", "service name should have suffix"),
	golden("wdep/v1/wdep.proto", "COMMENT_MESSAGE", "", 39, 6, 1, "WDepMsg", "message comments must not be empty"),
}

// writeGoldenCorpus materializes the committed corpus into a fresh temp root
// per drive, mirroring the parity harness's fixture writer.
func writeGoldenCorpus(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, src := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(src), 0o644))
	}
	return root
}

// requireTotalOrder is the D-01 totality guard (flagged assumption 3): no two
// adjacent findings may share all five sort terms. A five-term tie means the
// engine's order between the pair is nondeterministic and must not be frozen;
// if this fires, the capture STOPs — surface the tie, never paper over it.
func requireTotalOrder(t *testing.T, issues []core.IssueInfo) {
	t.Helper()

	for i := 1; i < len(issues); i++ {
		a, b := issues[i-1], issues[i]
		if a.Path == b.Path && a.RuleName == b.RuleName &&
			a.Position.Line == b.Position.Line && a.Position.Column == b.Position.Column &&
			a.SourceName == b.SourceName {
			t.Fatalf("D-01 five-term tie between findings %d and %d: %+v vs %+v — order is engine-nondeterministic and must not be frozen", i, i+1, a, b)
		}
	}
}

// requireGoldenPurity asserts the shared anti-vacuity guards before the
// golden compare: zero COMPILE findings (the corpus compiles under the linked
// engine's full type resolution), the pinned 73 count, and the coverage
// invariant — every one of the 41 registry names fires at least once.
func requireGoldenPurity(t *testing.T, got []core.IssueInfo) {
	t.Helper()

	for _, iss := range got {
		require.NotEqual(t, core.CompileRuleName, iss.RuleName,
			"the corpus must compile under the linked engine's full type resolution: a COMPILE finding vacates the rule-finding golden (Pitfall 7)")
	}
	require.Len(t, got, goldenFindingCount, "pinned old-engine finding count, asserted BEFORE the golden compare")

	fired := make(map[string]bool, len(got))
	for _, iss := range got {
		fired[iss.RuleName] = true
	}
	var missing []string
	for _, name := range rules.AllRuleNames() {
		if !fired[name] {
			missing = append(missing, name)
		}
	}
	require.Empty(t, missing, "every rule must fire at least once on the full-registry corpus")
}

// TestLintGoldenFullRegistry drives THE engine over the committed golden:
// the full compare including Message, Position (Filename/Offset/Line/Column),
// SourceName and the D-01 order. The drive was captured pre-flip as the
// linked parity drive and re-pointed by the flip commit; the old drive that
// froze the literals died with the old engine in the same commit.
//
// The engine output is compared as emitted: Lint already applies the D-01
// total sort, so re-sorting here would mask an engine-order regression. The
// totality guard runs on the engine's own order — a five-term tie means the
// engine's stable sort had nothing to order the pair by and the frozen
// sequence would be nondeterministic.
func TestLintGoldenFullRegistry(t *testing.T) {
	t.Parallel()

	cfg := config.LintConfig{Use: goldenCorpusUse}
	root := writeGoldenCorpus(t, goldenCorpusFiles)

	lintRules, ignoreOnly, err := rules.New(cfg)
	require.NoError(t, err)
	app := core.New(core.Options{
		Logger:              logger.NewNop(),
		Rules:               lintRules,
		IgnoreOnly:          ignoreOnly,
		AllowCommentIgnores: true,
		KnownLintRules:      rules.AllRuleNames(),
	})
	got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))
	require.NoError(t, err)

	requireGoldenPurity(t, got)
	requireTotalOrder(t, got)

	require.Equal(t, goldenIssues, got, "the linked engine must reproduce the old-captured golden message-for-message, in the D-01 order")
}
