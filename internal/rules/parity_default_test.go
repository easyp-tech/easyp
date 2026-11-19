package rules_test

// Batch-4 (DEFAULT) parity cells: ENUM_VALUE_PREFIX, ENUM_ZERO_VALUE_SUFFIX,
// FILE_LOWER_SNAKE_CASE, RPC_REQUEST_RESPONSE_UNIQUE, RPC_REQUEST_STANDARD_NAME,
// RPC_RESPONSE_STANDARD_NAME, PACKAGE_VERSION_SUFFIX and SERVICE_SUFFIX join the
// twenty-four already-ported rules under cumulative configs. Every cell asserts
// its expected old-side finding count before the multiset compare (T-04-02),
// and the cells pin main's semantics the lab changed (research inventory rows
// #25-#32; Pitfalls 2, 5; threats T-04-12, T-04-13, T-04-14):
//
//   - ENUM_VALUE_PREFIX uses main's private pascalToUpperSnake (underscore
//     before EVERY uppercase rune, so "HTTPCode" -> "H_T_T_P_CODE") with a bare
//     strings.HasPrefix — acronym enum values issue and a value equal to the
//     enum name passes. The lab's acronym-aware toUpperSnakeCase+"_" (which
//     would pass HTTPCODE_OK and fail the name-only value) is not portable.
//   - ENUM_ZERO_VALUE_SUFFIX checks ONLY the first value for EXACT equality
//     with pascalToUpperSnake(enumName)+"_"+Suffix (default "UNSPECIFIED" via
//     NewLinked's defaultIfEmpty): a later value with a wrong suffix is
//     invisible, and a first value that merely ends with the suffix still
//     issues. The lab's every-zero-value HasSuffix reading is not ported.
//   - FILE_LOWER_SNAKE_CASE matches the FULL filename including ".proto"
//     against main's regexp and reports at the zero Position literal (0:0).
//     The lab's extension-stripping + helper-equality + 1:1 FilePosition is
//     not ported.
//   - RPC_REQUEST_RESPONSE_UNIQUE is PER-FILE uniqueness on the WRITTEN type
//     strings with the first user exempt (later users issue at the rpc
//     keyword position, input==output counts as a duplicate): the same type
//     name in a different file is invisible. The lab's cross-file,
//     qualified-name, report-everyone ModuleRule is not ported.
//   - RPC_REQUEST/RESPONSE_STANDARD_NAME compare the WRITTEN type string
//     EXACTLY against RPCName+"Request"/"Response" or
//     ServiceName+RPCName+"Request"/"Response": a qualified written type
//     fails (the lab's last-dot-segment pascal-normalized compare would pass
//     it), no pascal normalization exists.
//   - PACKAGE_VERSION_SUFFIX uses main's UNANCHORED version regexp: a
//     single-component "v1" passes where the lab's two-component
//     packageHasVersionSuffix helper would fail it.
//   - SERVICE_SUFFIX is HasSuffix at the service keyword position (main
//     Meta.Pos ≡ Decl.Pos), with the Suffix config defaulting to "Service".
//
// Old-side counts below are hand-derived per rule over the cumulative
// 32-name config and probe-confirmed against the live old engine before the
// RED run (see the probe evidence note in 04-05-SUMMARY).

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/config"
)

// defaultUse is every cell's config: the eight batch-4 names plus the
// twenty-four already-ported names, so each cell re-verifies the whole
// ported registry cumulatively (32 rules at this batch's close).
var defaultUse = []string{
	//	minGroup — cumulative re-verification
	"DIRECTORY_SAME_PACKAGE",
	"PACKAGE_DEFINED",
	"PACKAGE_DIRECTORY_MATCH",
	"PACKAGE_SAME_DIRECTORY",
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
	//	basicGroup — semantic batch (cumulative re-verification)
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
	//	defaultGroup — this plan
	"ENUM_VALUE_PREFIX",
	"ENUM_ZERO_VALUE_SUFFIX",
	"FILE_LOWER_SNAKE_CASE",
	"RPC_REQUEST_RESPONSE_UNIQUE",
	"RPC_REQUEST_STANDARD_NAME",
	"RPC_RESPONSE_STANDARD_NAME",
	"PACKAGE_VERSION_SUFFIX",
	"SERVICE_SUFFIX",
}

// prefixFiles pins the prefix algorithm (threat T-04-12). Cell plan:
//
//	httpcode.proto — enum HTTPCode { HTTPCODE_OK = 0; OTHER = 1; }: both
//	  values issue under main's "H_T_T_P_CODE" prefix (no value starts with
//	  it; the lab's acronym-aware helper would pass HTTPCODE_OK — the tooth).
//	  Co-firing on this file (counted): ENUM_PASCAL_CASE ("HTTPCode" fails
//	  main's regexp) and ENUM_ZERO_VALUE_SUFFIX (first value not
//	  H_T_T_P_CODE_UNSPECIFIED).
//	nameonly.proto — enum Fob { FOB = 0; }: the name-only value passes main's
//	  bare HasPrefix (the lab requires the trailing "_" — would issue — the
//	  tooth); ENUM_ZERO_VALUE_SUFFIX co-fires once.
//	prefix_ok.proto — enum Foa { FOA_BAR = 0; }: clean for the prefix rule;
//	  ENUM_ZERO_VALUE_SUFFIX co-fires once.
//	msg_enum.proto — enum Baz inside a TOP-LEVEL message: the two-level walk
//	  reaches it, the name-only value BAZ passes the bare HasPrefix (+1
//	  level); ENUM_ZERO_VALUE_SUFFIX co-fires once.
//	nested_enum.proto — enum inside a NESTED message: main's shallow
//	  two-level walk never reaches it — nothing reports.
//
// Hand-derived old-side count: 4 + 1 + 1 + 1 + 0 = 7.
var prefixFiles = map[string]string{
	"pref.v1/httpcode.proto": `
syntax = "proto3";

package pref.v1;

enum HTTPCode {
  HTTPCODE_OK = 0;
  OTHER = 1;
}
`,
	"pref.v1/nameonly.proto": `
syntax = "proto3";

package pref.v1;

enum Fob {
  FOB = 0;
}
`,
	"pref.v1/prefix_ok.proto": `
syntax = "proto3";

package pref.v1;

enum Foa {
  FOA_BAR = 0;
}
`,
	"pref.v1/msg_enum.proto": `
syntax = "proto3";

package pref.v1;

message Holder {
  enum Baz {
    BAZ = 0;
  }
}
`,
	"pref.v1/nested_enum.proto": `
syntax = "proto3";

package pref.v1;

message Outer {
  message Inner {
    enum QuxEnum {
      QUXENUM_ZERO = 0;
    }
  }
}
`,
}

// zeroValueFiles pin first-value exact equality (threat T-04-12/15). Cell plan:
//
//	clean.proto — first value FOO_UNSPECIFIED exactly matches
//	  pascalToUpperSnake("Foo")+"_UNSPECIFIED": clean (the second value's
//	  suffix is never checked).
//	wrong_first.proto — first value FOB_OTHER issues at the first value's
//	  position; the second value's wrong suffix adds NOTHING (main checks
//	  the first value only — an every-value port would report two).
//	second_clean.proto — first value FOE_UNSPECIFIED clean, second value
//	  FOE_WEIRD wrong: nothing reports (first-only).
//	exact_eq.proto — first value BARX_UNSPECIFIED ENDS with "_UNSPECIFIED"
//	  (a HasSuffix port would pass it) but is not EXACTLY
//	  "BAR_UNSPECIFIED": issues — the exact-equality tooth.
//
// Hand-derived old-side count: 0 + 1 + 0 + 1 = 2.
var zeroValueFiles = map[string]string{
	"zvs.v1/clean.proto": `
syntax = "proto3";

package zvs.v1;

enum Foo {
  FOO_UNSPECIFIED = 0;
  FOO_OTHER = 1;
}
`,
	"zvs.v1/wrong_first.proto": `
syntax = "proto3";

package zvs.v1;

enum Fob {
  FOB_OTHER = 0;
  FOB_UNEXPECTED = 1;
}
`,
	"zvs.v1/second_clean.proto": `
syntax = "proto3";

package zvs.v1;

enum Foe {
  FOE_UNSPECIFIED = 0;
  FOE_WEIRD = 1;
}
`,
	"zvs.v1/exact_eq.proto": `
syntax = "proto3";

package zvs.v1;

enum Bar {
  BARX_UNSPECIFIED = 0;
}
`,
}

// fileNameFiles pin the full-filename check at 0:0 (threat T-04-12). Cell plan:
//
//	InvalidName.proto, aB.proto — uppercase in the full filename fails the
//	  regexp: FILE_LOWER_SNAKE_CASE at the zero Position literal (path, 0, 0).
//	good_name.proto, good_name2.proto — lower snake with separators and a
//	  digit in a LATER component (after a separator): clean.
//	v1alpha.proto — a digit directly after the leading lowercase run fails
//	  main's regexp on the FULL name: issues. This is also the main-vs-lab
//	  tooth: the lab strips ".proto" and compares the stem against its own
//	  lower-snake form ("v1alpha" is already lower snake), so the lab would
//	  pass the file that main fails. (The verdict is the old engine's — the
//	  live oracle — probe-confirmed before the RED run.)
//
// Hand-derived old-side count: 1 + 1 + 0 + 0 + 1 = 3.
var fileNameFiles = map[string]string{
	"fn.v1/InvalidName.proto": `
syntax = "proto3";

package fn.v1;
`,
	"fn.v1/aB.proto": `
syntax = "proto3";

package fn.v1;
`,
	"fn.v1/good_name.proto": `
syntax = "proto3";

package fn.v1;
`,
	"fn.v1/good_name2.proto": `
syntax = "proto3";

package fn.v1;
`,
	"fn.v1/v1alpha.proto": `
syntax = "proto3";

package fn.v1;
`,
}

// rpcUniqueFiles pin per-file uniqueness on written names with first-user
// exempt (threat T-04-13). Cell plan:
//
//	ua.proto (package ru.v1) — rpc B re-uses ReqT: ONE unique finding at
//	  rpc B's keyword position with "ReqT" as the source (the first user is
//	  exempt — the lab's report-everyone Finish would report rpc A too); rpc
//	  D(OneX) returns (OneX): the response occurrence duplicates the request
//	  within the file — ONE more unique finding at rpc D's position. The
//	  standard-name rules co-fire on every rpc whose written types are not
//	  RPCName+"Request"/"Response" or ServiceName+RPCName+"..." (counted: 3
//	  request + 3 response findings).
//	uc.proto (package rw.v1) — the SAME written name ReqT in a DIFFERENT
//	  file: NO unique finding (main is per-file; a cross-file port would
//	  issue — the tooth). The standard-name rules co-fire once each.
//
// Hand-derived old-side count: 8 + 2 = 10.
var rpcUniqueFiles = map[string]string{
	"ru.v1/ua.proto": `
syntax = "proto3";

package ru.v1;

message ReqT {}
message ResU {}
message ResV {}
message OneX {}

service UniqueService {
  rpc A(ReqT) returns (ResU);
  rpc B(ReqT) returns (ResV);
  rpc D(OneX) returns (OneX);
}
`,
	"rw.v1/uc.proto": `
syntax = "proto3";

package rw.v1;

message ReqT {}
message ResW {}

service SoloService {
  rpc C(ReqT) returns (ResW);
}
`,
}

// standardNameFiles pin the exact written-name compares (threat T-04-14).
// Cell plan:
//
//	nq.proto — rpc GetFoo(.pkg.v1.GetFooRequest) returns (GetFooReply):
//	  BOTH standard-name rules issue — the written request type is
//	  ".pkg.v1.GetFooRequest" (qualified), which equals neither
//	  "GetFooRequest" nor "ProbeServiceGetFooRequest" (the lab's
//	  last-dot-segment pascal-normalized compare would pass it — the tooth),
//	  and "GetFooReply" matches neither response alternative. Both findings
//	  sit at the rpc keyword position (pinned by the fingerprint).
//	ok.proto — the unqualified exact form (GetFooRequest == RPCName+
//	  "Request") and the service-qualified form (SecondServiceGetBarResponse
//	  == ServiceName+RPCName+"Response") are both clean; the import is used
//	  by both rpc references, so IMPORT_USED stays silent; all four written
//	  type names are distinct, so the unique rule stays silent.
//
// Hand-derived old-side count: 2 + 0 = 2.
var standardNameFiles = map[string]string{
	"pkg.v1/nq.proto": `
syntax = "proto3";

package pkg.v1;

message GetFooRequest {}
message GetFooResponse {}
message GetFooReply {}

service ProbeService {
  rpc GetFoo(.pkg.v1.GetFooRequest) returns (GetFooReply);
}
`,
	"pkg.v1/ok.proto": `
syntax = "proto3";

package pkg.v1;

import "pkg.v1/nq.proto";

message GetBarRequest {}
message GetBarResponse {}
message SecondServiceGetBarResponse {}

service SecondService {
  rpc GetFoo(GetFooRequest) returns (GetFooResponse);
  rpc GetBar(GetBarRequest) returns (SecondServiceGetBarResponse);
}
`,
}

// versionFiles pin the unanchored version regexp (threat T-04-12). Cell plan —
// one package per directory (the directory IS the package, so
// PACKAGE_DIRECTORY_MATCH stays clean everywhere):
//
//	example.v1, example.v1test2, example.v2alpha1, v1 — all match main's
//	  unanchored regexp: clean. "v1" is the tooth — the lab's two-component
//	  packageHasVersionSuffix helper would fail the single component;
//	  PACKAGE_LOWER_SNAKE_CASE co-fires on "v1" (a digit cannot follow the
//	  leading lowercase run without a separator — counted).
//	example, example.foo — no v<digits> anywhere: PACKAGE_VERSION_SUFFIX
//	  issues.
//
// Hand-derived old-side count: 0 + 1 + 1 + 0 + 0 + 1 = 3.
var versionFiles = map[string]string{
	"example.v1/p.proto": `
syntax = "proto3";

package example.v1;
`,
	"example/p.proto": `
syntax = "proto3";

package example;
`,
	"v1/p.proto": `
syntax = "proto3";

package v1;
`,
	"example.v1test2/p.proto": `
syntax = "proto3";

package example.v1test2;
`,
	"example.v2alpha1/p.proto": `
syntax = "proto3";

package example.v2alpha1;
`,
	"example.foo/p.proto": `
syntax = "proto3";

package example.foo;
`,
}

// serviceSuffixFiles pin HasSuffix at the service keyword position with the
// default Suffix config ("Service" via NewLinked's defaultIfEmpty). Cell plan:
//
//	svc_ok.proto — service UserService ends with "Service": clean.
//	svc_bad.proto — service User does not: SERVICE_SUFFIX at the service
//	  keyword position, column 1 (main Meta.Pos ≡ Decl.Pos — the lab's
//	  name-token column 9 is forbidden).
//
// Hand-derived old-side count: 0 + 1 = 1.
var serviceSuffixFiles = map[string]string{
	"ss.v1/svc_ok.proto": `
syntax = "proto3";

package ss.v1;

service UserService {}
`,
	"ss.v1/svc_bad.proto": `
syntax = "proto3";

package ss.v1;

service User {}
`,
}

// TestLinkedParityDefault pins old-vs-new parity for the DEFAULT batch. The
// requirePorted guard is the RED assertion: it fails listing the unported
// rules while the package still compiles, and goes green as the batch's ports
// land.
func TestLinkedParityDefault(t *testing.T) {
	// RED guard — must stay the first statement of the test body.
	requirePorted(t,
		"ENUM_VALUE_PREFIX",
		"ENUM_ZERO_VALUE_SUFFIX",
		"FILE_LOWER_SNAKE_CASE",
		"RPC_REQUEST_RESPONSE_UNIQUE",
		"RPC_REQUEST_STANDARD_NAME",
		"RPC_RESPONSE_STANDARD_NAME",
		"PACKAGE_VERSION_SUFFIX",
		"SERVICE_SUFFIX",
	)

	t.Parallel()

	cfg := config.LintConfig{Use: defaultUse}

	t.Run("prefix_algorithm", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, prefixFiles, 7)
	})

	t.Run("zero_value_suffix", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, zeroValueFiles, 2)
	})

	t.Run("file_name_full", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, fileNameFiles, 3)
	})

	t.Run("rpc_unique_per_file", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, rpcUniqueFiles, 10)
	})

	t.Run("standard_name_written", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, standardNameFiles, 2)
	})

	t.Run("version_suffix", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, versionFiles, 3)
	})

	t.Run("service_suffix", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, serviceSuffixFiles, 1)
	})

	// Standing constraint (b), mechanized at batch level: a batch-4 finding
	// is silenced ONLY by the central filter (buf:lint:ignore attached above
	// the declaration) — the ported rules never consult comments themselves,
	// and the suppression parity holds old-vs-new (0 suppressed, 1 not).
	t.Run("central_filter_suppressible", func(t *testing.T) {
		t.Parallel()

		suppressed := map[string]string{
			"sup.v1/svc_bad.proto": `
syntax = "proto3";

package sup.v1;

// buf:lint:ignore SERVICE_SUFFIX
service User {}
`,
		}
		assertBatchParity(t, cfg, suppressed, 0)

		unsuppressed := map[string]string{
			"sup.v1/svc_bad.proto": `
syntax = "proto3";

package sup.v1;

service User {}
`,
		}
		assertBatchParity(t, cfg, unsuppressed, 1)
	})
}
