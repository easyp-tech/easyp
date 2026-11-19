package rules_test

// Batch-3 (BASIC semantic) parity cells: ENUM_FIRST_VALUE_ZERO, ENUM_NO_ALLOW_ALIAS,
// IMPORT_USED and the seven option-comparison ModuleRules join the fourteen
// already-ported rules under cumulative configs. Every cell asserts its
// expected old-side finding count before the multiset compare (T-04-02), and
// the cells pin main's semantics the lab changed (research inventory rows #5,
// #6, #12, #16-22; Pitfalls 5, 10):
//
//   - ENUM_FIRST_VALUE_ZERO compares the first value's number AS WRITTEN against
//     the literal "0" (string compare): "00" and "0x0" issue, where the lab's
//     numeric parse-and-compare passes them; only the first value is checked;
//     the two-level walk (top-level enums plus enums of top-level messages)
//     stops there — an enum inside a NESTED message reports nothing.
//   - ENUM_NO_ALLOW_ALIAS issues for ANY enum-level option named allow_alias,
//     regardless of its value, at the ENUM keyword position (the lab checks
//     Value=="true" at the option position — not ported).
//   - IMPORT_USED walks the file's written references only (stateless syntax
//     walk; LinkedFileInfo.Resolved is never read): a bare unqualified name is
//     matched against imports of the checking file's own package, a qualified
//     name against imports of the written package prefix, and existInProto
//     sees only the imported file's TOP-LEVEL names — a qualified nested
//     reference (which resolution would mark used) stays unused and issues.
//   - The seven option rules compare option constants AS WRITTEN (raw
//     opt.Value, never optionString-unquoted), skip files without the option
//     entirely, exempt the first value per package, and report later differing
//     files at their own option positions.
//   - PACKAGE_SAME_SWIFT_PREFIX compares the literal ruby_package option —
//     main's bug, pinned bug-for-bug (T-04-09); "fixing" it here would break
//     Phase-7 corpus parity.
//
// Execution-note probe evidence (2026-11-19, TestZZProbe runs, deleted before
// commit):
//   - the easypc/protoast parser accepts the enum value numbers "00" and "0x0"
//     (octal/hex integer literals), single-quoted option literals ('x'), and
//     adjacent string literals ("a" "b" — protoast merges them to "ab"; main's
//     Constant basis was not probed for this edge, so adjacent literals stay an
//     observation-only note, not a parity cell);
//   - the REAL testdata/import_used corpus does not link (protoc reports the
//     undefined extend target import_used.thrd_party.MethodOptions, the
//     undefined MethodOption, the undeclared extension number and the
//     duplicate TestService), so the engine-level IMPORT_USED cells use
//     compiling same-package variants; the real fixtures stay in
//     import_used_test.go, which drives the linked rule directly over
//     parseLinkedFile-built infos without the linking step;
//   - every cell corpus below links cleanly (a zero-rules linked-engine run
//     produced no COMPILE issues).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/rules"
)

// basicSemanticUse is every cell's config: the ten batch-3 names plus the
// fourteen already-ported names, so each cell re-verifies the whole ported
// registry cumulatively (24 rules at this batch's close).
var basicSemanticUse = []string{
	//	minGroup — cumulative re-verification
	"DIRECTORY_SAME_PACKAGE",
	"PACKAGE_DEFINED",
	"PACKAGE_SAME_DIRECTORY",
	"PACKAGE_DIRECTORY_MATCH",
	//	basicGroup — casing batch (cumulative re-verification)
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
	//	basicGroup — semantic batch (this plan)
	"ENUM_FIRST_VALUE_ZERO",
	"ENUM_NO_ALLOW_ALIAS",
	"IMPORT_USED",
	"PACKAGE_SAME_CSHARP_NAMESPACE",
	"PACKAGE_SAME_GO_PACKAGE",
	"PACKAGE_SAME_JAVA_MULTIPLE_FILES",
	"PACKAGE_SAME_JAVA_PACKAGE",
	"PACKAGE_SAME_PHP_NAMESPACE",
	"PACKAGE_SAME_RUBY_PACKAGE",
	"PACKAGE_SAME_SWIFT_PREFIX",
}

// TestLinkedParityBasicSemantic pins old-vs-new parity for the BASIC semantic
// batch. The requirePorted guard is the RED assertion: it fails listing the
// unported rules while the package still compiles, and goes green as the
// batch's ports land.
func TestLinkedParityBasicSemantic(t *testing.T) {
	// RED guard — must stay the first statement of the test body.
	requirePorted(t,
		"ENUM_FIRST_VALUE_ZERO",
		"ENUM_NO_ALLOW_ALIAS",
		"IMPORT_USED",
		"PACKAGE_SAME_CSHARP_NAMESPACE",
		"PACKAGE_SAME_GO_PACKAGE",
		"PACKAGE_SAME_JAVA_MULTIPLE_FILES",
		"PACKAGE_SAME_JAVA_PACKAGE",
		"PACKAGE_SAME_PHP_NAMESPACE",
		"PACKAGE_SAME_RUBY_PACKAGE",
		"PACKAGE_SAME_SWIFT_PREFIX",
	)

	t.Parallel()

	cfg := config.LintConfig{Use: basicSemanticUse}

	t.Run("first_value_string_compare", func(t *testing.T) {
		t.Parallel()

		// Main's string compare against the literal "0": "00" (octal zero) and
		// "0x0" (hex zero) ISSUE, where the lab's numeric parse-and-compare
		// passes both — this is the string-vs-numeric tooth. The literal "0" is
		// clean, the second value "5" is never checked (first only), an enum
		// inside a TOP-LEVEL message behaves the same (main's +1 level), and an
		// enum inside a NESTED message reports nothing (the two-level walk
		// stops there). Hand-derived old-side count: 3.
		files := map[string]string{
			"fv/zero.proto": `
syntax = "proto3";

package fv;

enum GoodEnum {
  GOOD = 00;
}
`,
			"fv/hex.proto": `
syntax = "proto3";

package fv;

enum HexEnum {
  HEX_ZERO = 0x0;
}
`,
			"fv/ok.proto": `
syntax = "proto3";

package fv;

enum OkEnum {
  OK_ZERO = 0;
  FIVE = 5;
}
`,
			"fv/holder.proto": `
syntax = "proto3";

package fv;

message Holder {
  enum BadHolderEnum {
    BAD = 00;
  }
}
`,
			"fv/nested.proto": `
syntax = "proto3";

package fv;

message Outer {
  message Inner {
    enum BadInnerEnum {
      BAD = 00;
    }
  }
}
`,
		}
		assertBatchParity(t, cfg, files, 3)
	})

	t.Run("allow_alias_any_value", func(t *testing.T) {
		t.Parallel()

		// Main issues for ANY enum-level option named allow_alias, regardless
		// of its value (false included), at the ENUM keyword position with the
		// enum name as the source; the lab's Value=="true"-at-option-position
		// reading would report neither file. The optionless enum is clean.
		// Hand-derived old-side count: 2.
		files := map[string]string{
			"aa/false_alias.proto": `
syntax = "proto3";

package aa;

enum FalseAliasEnum {
  option allow_alias = false;
  FALSE_ZERO = 0;
}
`,
			"aa/true_alias.proto": `
syntax = "proto3";

package aa;

enum TrueAliasEnum {
  option allow_alias = true;
  TRUE_ZERO = 0;
}
`,
			"aa/no_alias.proto": `
syntax = "proto3";

package aa;

enum NoAliasEnum {
  NO_ZERO = 0;
}
`,
		}
		assertBatchParity(t, cfg, files, 2)
	})

	t.Run("import_used_verdicts", func(t *testing.T) {
		t.Parallel()

		// Compiling same-package variant of the testdata/import_used corpus
		// (see the probe note in the header): a field-type reference to an
		// imported TOP-LEVEL message and an rpc request/response reference to
		// imported messages mark the import used (both files clean), a file
		// whose written names match nothing leaves the import unused —
		// IMPORT_USED at the import keyword with the quoted location as the
		// source. Hand-derived old-side count: 1.
		files := map[string]string{
			"imp/lib.proto": `
syntax = "proto3";

package imp;

message Used1 {}
message Used2 {}

enum UsedEnum {
  USED_ENUM_ZERO = 0;
}
`,
			"imp/used.proto": `
syntax = "proto3";

package imp;

import "imp/lib.proto";

message Caller {
  Used1 one = 1;
  Used2 two = 2;
}
`,
			"imp/rpc_used.proto": `
syntax = "proto3";

package imp;

import "imp/lib.proto";

service SvcX {
  rpc Call(Used1) returns (Used2);
}
`,
			"imp/not_used.proto": `
syntax = "proto3";

package imp;

import "imp/lib.proto";

message Lonely {}
`,
		}
		assertBatchParity(t, cfg, files, 1)
	})

	t.Run("syntax_walk_only", func(t *testing.T) {
		t.Parallel()

		// The verdicts ride the syntax walk, never LinkedFileInfo.Resolved: a
		// field-type reference to an imported top-level message is clean, while
		// a qualified reference into an imported NESTED message
		// (LibOuter.LibNested — which name resolution resolves and a
		// resolution-based port would mark used) does not match existInProto's
		// top-level-only reading and ISSUES. Hand-derived old-side count: 1.
		files := map[string]string{
			"impwalk/lib.proto": `
syntax = "proto3";

package impwalk;

message Used1 {}

message LibOuter {
  message LibNested {}
}
`,
			"impwalk/field_used.proto": `
syntax = "proto3";

package impwalk;

import "impwalk/lib.proto";

message Holder2 {
  Used1 inner = 1;
}
`,
			"impwalk/nested_qualified.proto": `
syntax = "proto3";

package impwalk;

import "impwalk/lib.proto";

message Outer3 {
  LibOuter.LibNested inner = 1;
}
`,
		}
		assertBatchParity(t, cfg, files, 1)
	})

	t.Run("option_groups_first_exempt", func(t *testing.T) {
		t.Parallel()

		// One package across three files (a=x, b=y, c=x for every option) plus
		// one optionless file in the same package: exactly the second file
		// issues, once per rule, at its own option position (first value per
		// package exempt; the optionless file contributes nothing — main's
		// skip, where the lab's add-always would give it a distinct empty
		// value and 7 extra findings). All seven option rules run; the
		// swift_prefix finding sits at b's ruby_package position, which is the
		// bug pin, not a mistake. Hand-derived old-side count: 7.
		files := map[string]string{
			"samepkg/a.proto": `
syntax = "proto3";

package samepkg;

option csharp_namespace = "x";
option go_package = "x";
option java_multiple_files = true;
option java_package = "x";
option php_namespace = "x";
option ruby_package = "x";
option swift_prefix = "x";

message Am {}
`,
			"samepkg/b.proto": `
syntax = "proto3";

package samepkg;

option csharp_namespace = "y";
option go_package = "y";
option java_multiple_files = false;
option java_package = "y";
option php_namespace = "y";
option ruby_package = "y";
option swift_prefix = "y";

message Bm {}
`,
			"samepkg/c.proto": `
syntax = "proto3";

package samepkg;

option csharp_namespace = "x";
option go_package = "x";
option java_multiple_files = true;
option java_package = "x";
option php_namespace = "x";
option ruby_package = "x";
option swift_prefix = "x";

message Cm {}
`,
			"samepkg/d.proto": `
syntax = "proto3";

package samepkg;

message Dm {}
`,
		}
		assertBatchParity(t, cfg, files, 7)
	})

	t.Run("option_written_compare", func(t *testing.T) {
		t.Parallel()

		// Main compares the constants AS WRITTEN: "x" (double-quoted) and 'x'
		// (single-quoted) are different constants, so the later file issues —
		// an optionString-unquoting port would see the same unquoted value and
		// pass it (Pitfall 10). Hand-derived old-side count: 1.
		files := map[string]string{
			"w/a.proto": `
syntax = "proto3";

package w;

option go_package = "x";

message Am {}
`,
			"w/c.proto": `
syntax = "proto3";

package w;

option go_package = 'x';

message Cm {}
`,
		}
		assertBatchParity(t, cfg, files, 1)
	})

	t.Run("swift_prefix_bug_pin", func(t *testing.T) {
		t.Parallel()

		// T-04-09 bug-for-bug pin: PACKAGE_SAME_SWIFT_PREFIX compares the
		// literal ruby_package option (main's copy-paste bug). Corpus 1 —
		// differing ruby_package values: the pinned rule issues (alongside
		// PACKAGE_SAME_RUBY_PACKAGE, at the same ruby_package position);
		// corpus 2 — differing swift_prefix values with matching ruby_package:
		// the pinned rule stays silent. Hand-derived old-side counts: 2, then 0.
		rubyDiffers := map[string]string{
			"bug/a.proto": `
syntax = "proto3";

package bug;

option ruby_package = "a.ruby";

message Am {}
`,
			"bug/b.proto": `
syntax = "proto3";

package bug;

option ruby_package = "b.ruby";

message Bm {}
`,
		}
		assertBatchParity(t, cfg, rubyDiffers, 2)

		swiftDiffers := map[string]string{
			"bugswift/a.proto": `
syntax = "proto3";

package bugswift;

option swift_prefix = "pre.a";

message Am {}
`,
			"bugswift/b.proto": `
syntax = "proto3";

package bugswift;

option swift_prefix = "pre.b";

message Bm {}
`,
		}
		assertBatchParity(t, cfg, swiftDiffers, 0)
	})

	t.Run("packageless_files_ignored", func(t *testing.T) {
		t.Parallel()

		// Files without a package that set options: main returns early, so
		// none of the seven option rules reports (the only findings are
		// PACKAGE_DEFINED's own packageless cells). A port that ignores the
		// package guard would collect the diverging go_package values into a
		// ""-keyed group and report.
		files := map[string]string{
			"p1.proto": `
syntax = "proto3";

option go_package = "one";

message P1x {}
`,
			"p2.proto": `
syntax = "proto3";

option go_package = "two";

message P2x {}
`,
		}
		assertBatchParity(t, cfg, files, 2)

		// And none of the findings belongs to the seven option rules.
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
		require.NotEmpty(t, oldIssues)
		for _, is := range oldIssues {
			require.False(t, strings.HasPrefix(is.RuleName, "PACKAGE_SAME_"),
				"option rules must ignore packageless files, got %s at %s", is.RuleName, is.Path)
		}
	})
}
