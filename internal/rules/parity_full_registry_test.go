package rules_test

// Phase-4 closure gate (04-07 Task 3): the FULL-REGISTRY parity cell. One
// broad compiling fixture corpus — messages/enums/services/rpcs/imports/
// options across multiple files and packages — exercising every batch's edge
// classes at least once, driven through assertBatchParity with
// cfg.Use = rules.AllRuleNames(), so all 41 rules run old-vs-new and the
// sorted (rule, path, line, col) multisets must match exactly.
//
// Two anti-vacuity layers before the compare:
//   - the coverage assertion requires EVERY one of the 41 rule names to fire
//     at least once on the corpus (a silent rule is a coverage hole, not a
//     pass);
//   - wantOldCount pins the probe-confirmed total old-side finding count
//     (T-04-02), asserted inside assertBatchParity before the multiset
//     compare.
//
// Corpus map (rule -> file):
//   - MINIMAL: orphan.proto (PACKAGE_DEFINED 0:0), dir_a first/second +
//     dir_b third (DIRECTORY_SAME_PACKAGE, PACKAGE_SAME_DIRECTORY,
//     PACKAGE_DIRECTORY_MATCH)
//   - BASIC casing: teeth.proto (enum ABC acronym + _LEADING value, oneof1,
//     badMessage/_field/get_thing/bad_service, badPkg.v2)
//   - BASIC semantic: ZeroEnum (=00 octal, string compare), AliasEnum
//     (allow_alias=false), unused_dep import (IMPORT_USED), opts.v1 pair
//     (all seven PACKAGE_SAME_* on beta, swift reads ruby_package)
//   - DEFAULT: ZeroEnum/_LEADING (prefix algorithm), ABC first value (zero
//     suffix), InvalidName.proto (full-filename 0:0), Dup request reuse
//     (per-file uniqueness), GetReq/GetResp (written-name compares),
//     no_version (version suffix), bad_service (service suffix)
//   - COMMENTS: every declaration in teeth/dep/wdep files is undocumented,
//     and dir_only_msg carries ONLY a suppression directive — documenting
//     for COMMENT_MESSAGE while MESSAGE_PASCAL_CASE stays centrally
//     suppressed (the 04-06 interplay pin, at full-registry scale)
//   - UNARY_RPC: DoStream (stream on both sides -> both rules)

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/rules"
)

// fullRegistryUse is the closure gate's config: every registered rule.
var fullRegistryUse = rules.AllRuleNames()

// fullRegistryFiles is the broad compiling corpus: 13 files, 9 packages,
// every batch's edge classes represented. Identifier names are chosen so the
// intended rule fires; unintended co-firings are fine — the old side is the
// oracle and the count pins the total.
var fullRegistryFiles = map[string]string{
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

// TestLinkedParityFullRegistry is the Phase-4 closure gate: all 41 rules,
// one broad corpus, old-vs-new parity under the full-registry config.
func TestLinkedParityFullRegistry(t *testing.T) {
	t.Parallel()

	cfg := config.LintConfig{Use: fullRegistryUse}

	// Coverage assertion: every rule must fire at least once on the corpus.
	old := oldFindings(t, cfg, fullRegistryFiles)
	fired := make(map[string]bool, len(old))
	for _, iss := range old {
		fired[iss.RuleName] = true
	}
	var missing []string
	for _, name := range rules.AllRuleNames() {
		if !fired[name] {
			missing = append(missing, name)
		}
	}
	require.Empty(t, missing, "every rule must fire at least once on the full-registry corpus")

	assertBatchParity(t, cfg, fullRegistryFiles, 73)
}
