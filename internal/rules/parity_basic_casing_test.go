package rules_test

// Batch-2 (BASIC casing) parity cells: the ten casing/import rules join the
// four MINIMAL rules under cumulative configs. Every cell asserts its
// expected old-side finding count before the multiset compare (T-04-02), and
// the cells pin main's semantics the lab changed (research Pitfall 2 and the
// position mapping): main's regexps reject acronym/digit/underscore names the
// casing.go helpers accept, positions stay on the declaration keyword
// (Decl.Pos, never NamePos), and walks stay shallow (top-level declarations
// plus main's one-level enum exception — the etalon's recursive Visitor would
// over-report).
//
// Execution note (probe evidence, 2026-11-19): a digit-led rpc name ("2Fast")
// is rejected by BOTH parsers at the rpc-name token ("expected rpc method
// name"), so a compiling corpus cannot carry that edge; RPC_PASCAL_CASE's
// regexp semantics ride the shared get_thing finding here plus the carried
// test literals.

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

// basicCasingUse is every cell's config: the ten batch-2 names plus the four
// MINIMAL names, so each cell re-verifies the whole ported registry
// cumulatively (14 rules at this batch's close).
var basicCasingUse = []string{
	//	minGroup — cumulative re-verification
	"DIRECTORY_SAME_PACKAGE",
	"PACKAGE_DEFINED",
	"PACKAGE_SAME_DIRECTORY",
	"PACKAGE_DIRECTORY_MATCH",
	//	basicGroup casing batch (this plan)
	"ENUM_PASCAL_CASE",
	"ENUM_VALUE_UPPER_SNAKE_CASE",
	"FIELD_LOWER_SNAKE_CASE",
	"IMPORT_NO_PUBLIC",
	"IMPORT_NO_WEAK",
	"MESSAGE_PASCAL_CASE",
	"ONEOF_LOWER_SNAKE_CASE",
	"PACKAGE_LOWER_SNAKE_CASE",
	"RPC_PASCAL_CASE",
	"SERVICE_PASCAL_CASE",
}

// posKey is one (rule, line, column) literal a cell pins on the old side;
// the parity compare then forces the new engine to reproduce it exactly.
type posKey struct {
	rule string
	line int
	col  int
}

// oldFindings returns the sorted old-engine fingerprint for one fixture tree
// (assertBatchParity's old side, exposed so a cell can pin exact position
// literals before the both-engines compare).
func oldFindings(t *testing.T, cfg config.LintConfig, files map[string]string) []core.IssueInfo {
	t.Helper()

	root := t.TempDir()
	for name, src := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(src), 0o644))
	}

	oldRules, oldIgnoreOnly, err := rules.New(cfg)
	require.NoError(t, err)
	oldApp := core.New(core.Options{
		Logger:              logger.NewNop(),
		Rules:               oldRules,
		IgnoreOnly:          oldIgnoreOnly,
		AllowCommentIgnores: true,
		KnownLintRules:      rules.AllRuleNames(),
	})
	oldIssues, err := oldApp.Lint(t.Context(), fs.NewFSWalker(root, "."))
	require.NoError(t, err)

	return fingerprint(oldIssues)
}

// TestLinkedParityBasicCasing pins old-vs-new parity for the BASIC casing
// batch. The requirePorted guard is the RED assertion: it fails listing the
// unported rules while the package still compiles, and goes green as the
// batch's ports land.
func TestLinkedParityBasicCasing(t *testing.T) {
	// RED guard — must stay the first statement of the test body.
	requirePorted(t,
		"ENUM_PASCAL_CASE",
		"ENUM_VALUE_UPPER_SNAKE_CASE",
		"FIELD_LOWER_SNAKE_CASE",
		"IMPORT_NO_PUBLIC",
		"IMPORT_NO_WEAK",
		"MESSAGE_PASCAL_CASE",
		"ONEOF_LOWER_SNAKE_CASE",
		"PACKAGE_LOWER_SNAKE_CASE",
		"RPC_PASCAL_CASE",
		"SERVICE_PASCAL_CASE",
	)

	t.Parallel()

	cfg := config.LintConfig{Use: basicCasingUse}

	t.Run("casing_edges", func(t *testing.T) {
		t.Parallel()

		// One compiling corpus exercising every divergence edge: acronym enum
		// ABC and single-char message A and acronym service AB fail main's
		// regexps while the lab's toPascalCase equality passes them; oneof1 and
		// foo1bar fail main's regexps while toLowerSnakeCase keeps digits;
		// _LEADING, _field and oneTwo fail main's regexps outright. Their
		// well-named counterparts produce nothing, so the hand-derived old-side
		// count is 9; the lab helpers would produce a DIFFERENT count on the
		// same corpus (this cell is the Pitfall-2 teeth).
		files := map[string]string{
			"enums/acronym.proto": `
syntax = "proto3";

package enums;

enum ABC {
  ABC_UNSPECIFIED = 0;
}
`,
			"enums/values.proto": `
syntax = "proto3";

package enums;

enum GoodEnum {
  GOOD = 0;
  _LEADING = 1;
}
`,
			"fields/bad.proto": `
syntax = "proto3";

package fields;

message GoodMsg {
  string _field = 1;
  string oneTwo = 2;
  string good_field = 3;
}
`,
			"msgs/single.proto": `
syntax = "proto3";

package msgs;

message A {}
`,
			"oneofs/digit.proto": `
syntax = "proto3";

package oneofs;

message GoodMsg2 {
  oneof oneof1 {
    string val = 1;
  }

  oneof good_one {
    string val2 = 2;
  }
}
`,
			"foo1bar/decl.proto": `
syntax = "proto3";

package foo1bar;

message Mx {}
`,
			"goodpkg/decl.proto": `
syntax = "proto3";

package goodpkg;

message Mx {}
`,
			"rpcs/bad.proto": `
syntax = "proto3";

package rpcs;

message Req {}
message Res {}

service GoodService {
  rpc get_thing(Req) returns (Res);
  rpc GetThing(Req) returns (Res);
}
`,
			"services/acronym.proto": `
syntax = "proto3";

package services;

message Req2 {}
message Res2 {}

service AB {
  rpc GetThing(Req2) returns (Res2);
}
`,
		}
		assertBatchParity(t, cfg, files, 9)
	})

	t.Run("keyword_positions", func(t *testing.T) {
		t.Parallel()

		// One file with badly-named top-level message, service, rpc, field and
		// enum at known columns: message/service at column 1, rpc at column 3
		// (keyword), field at the type token (column 3), enum at its keyword,
		// enum value at its name start. A name-token-position port shifts the
		// columns by the keyword width (message 1->9, rpc 3->7, field 3->10,
		// service 1->9) and fails the parity compare; the old-side literals are
		// pinned explicitly so the count cannot pass by coincidence.
		files := map[string]string{
			"pos/one.proto": `
syntax = "proto3";

package pos;

enum bad_top_enum {
  bad_value = 0;
}

message Req {}
message Res {}

message bad_position_msg {
  string BadFieldName = 1;
}

service bad_service {
  rpc get_thing(Req) returns (Res);
}
`,
		}
		old := oldFindings(t, cfg, files)
		require.Len(t, old, 6, "old-side finding count before the multiset compare")
		for _, want := range []posKey{
			{"ENUM_PASCAL_CASE", 6, 1},
			{"ENUM_VALUE_UPPER_SNAKE_CASE", 7, 3},
			{"MESSAGE_PASCAL_CASE", 13, 1},
			{"FIELD_LOWER_SNAKE_CASE", 14, 3},
			{"SERVICE_PASCAL_CASE", 17, 1},
			{"RPC_PASCAL_CASE", 18, 3},
		} {
			require.Contains(t, keys(old), want, "old-side keyword-position literal %v must reach the fingerprint", want)
		}
		assertBatchParity(t, cfg, files, 6)
	})

	t.Run("shallow_walk", func(t *testing.T) {
		t.Parallel()

		// Badly-named declarations at nesting depths main never walks: a nested
		// message inside a well-named message, an enum inside the NESTED
		// message, a oneof and a plain field of the nested message, and a
		// badly-named map field. Main reports NONE of them (shallow shapes:
		// top-level messages' plain fields and oneofs, top-level enums plus
		// enums of TOP-LEVEL messages); the etalon's recursive Visitor walk
		// would report each, so the asserted zero is the main shape, not a
		// vacuous pass.
		files := map[string]string{
			"shallow/one.proto": `
syntax = "proto3";

package shallow;

message Wrapper {
  message bad_nested {
    enum bad_enum {
      BAD_VALUE = 0;
    }

    oneof bad_oneof1 {
      string badOneofField = 1;
    }

    string badFieldName = 2;
    map<string, int32> badMapName = 3;
  }
}
`,
		}
		assertBatchParity(t, cfg, files, 0)
	})

	t.Run("top_level_message_enum", func(t *testing.T) {
		t.Parallel()

		// A badly-named enum inside a TOP-LEVEL message: ENUM_PASCAL_CASE and
		// ENUM_VALUE_UPPER_SNAKE_CASE DO report it (main's +1 level) — this
		// pins the exact boundary between the shallow shape and the one-level
		// enum exception, at the enum's keyword column (3) and the value's
		// column (5).
		files := map[string]string{
			"toplevel/one.proto": `
syntax = "proto3";

package toplevel;

enum GoodTopEnum {
  GOOD_TOP = 0;
}

message Holder {
  enum bad_holder_enum {
    bad_value = 0;
  }
}
`,
		}
		assertBatchParity(t, cfg, files, 2)
	})

	t.Run("import_flags", func(t *testing.T) {
		t.Parallel()

		// import public and import weak declarations: IMPORT_NO_PUBLIC and
		// IMPORT_NO_WEAK report at the import keyword (column 1) with the
		// quoted location as the source name; positions are test-verified equal
		// to main's Meta.Pos (Wave-0 probe, Import kind). The imports resolve
		// inside the corpus, so both engines lint the same linked tree.
		files := map[string]string{
			"imp/lib.proto": `
syntax = "proto3";

package imp;

message Lib {}
`,
			"imp/lib2.proto": `
syntax = "proto3";

package imp;

message Lib2 {}
`,
			"imp/main.proto": `
syntax = "proto3";

package imp;

import public "imp/lib.proto";
import weak "imp/lib2.proto";

message Main {}
`,
		}
		assertBatchParity(t, cfg, files, 2)
	})
}

// keys flattens a fingerprint to the (rule, line, col) literals a cell pins.
func keys(fps []core.IssueInfo) []posKey {
	out := make([]posKey, 0, len(fps))
	for _, fp := range fps {
		out = append(out, posKey{rule: fp.RuleName, line: fp.Position.Line, col: fp.Position.Column})
	}
	return out
}
