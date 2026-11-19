package rules_test

// Batch-5 (COMMENTS) parity cells: COMMENT_ENUM, COMMENT_ENUM_VALUE,
// COMMENT_FIELD, COMMENT_MESSAGE, COMMENT_ONEOF, COMMENT_RPC and
// COMMENT_SERVICE join the thirty-two already-ported rules under cumulative
// configs. Every cell asserts its expected old-side finding count before the
// multiset compare (T-04-02), and the cells pin main's presence semantics the
// lab changed (research inventory rows #33-#39; Pitfall 3; threats T-04-16,
// T-04-17; LINT-PARITY-02's motivating case):
//
//   - The presence predicate is MAIN'S: a declaration is undocumented iff no
//     comment at all precedes it. go-protoparser's ParseComments attaches ALL
//     preceding comments (including across blank lines) to element.Comments
//     and keeps the same-line trailing comment in the separate InlineComment
//     field [VERIFIED: go-protoparser v4.14.2 parser/message.go parse loop +
//     parser/declaration.go], so main's len(Comments)==0 maps exactly to
//     protoast HasCommentAbove()==false (Leading+Detached empty). The lab's
//     leadingCommentDocuments helper (directive-only => undocumented,
//     detached => undocumented) is NOT portable — it would fabricate findings
//     on suppression-heavy corpora.
//   - Walk shapes are main's shallow shapes: top-level messages' direct
//     children only (fields/oneofs), top-level enums plus enums of top-level
//     messages, top-level services' RPCs. Nested declarations are never
//     reported; an enum inside a TOP-LEVEL message IS reported (main's +1
//     level).
//   - Proto2 group fields never enter any walk: go-protoparser's unordered
//     model keeps them in MessageBody.Groups, protoast keeps them in
//     Message.Groups [VERIFIED: both sources], so CommentField needs no group
//     guard (research A4 resolved at source).
//   - The directive-only cell is the pivotal LINT-PARITY-02 interplay pin: a
//     comment consisting ONLY of "easyp:disable MESSAGE_PASCAL_CASE" (a rule
//     other than the COMMENT_* verdicts, so the central filter does not mask
//     the comment verdict) both suppresses MESSAGE_PASCAL_CASE centrally
//     (both engines) and documents the declaration for COMMENT_MESSAGE (main:
//     len(Comments)>0). The two behaviors coexist in one fixture.
//
// Old-side counts below are hand-derived per rule over the cumulative 39-name
// config and probe-confirmed against the live old engine before the RED run
// (see the probe evidence note in 04-06-SUMMARY).

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/config"
)

// commentsUse is every cell's config: the seven batch-5 names plus the
// thirty-two already-ported names, so each cell re-verifies the whole ported
// registry cumulatively (39 rules at this batch's close). UNARY_RPC rules
// arrive in batch 6.
var commentsUse = []string{
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
	//	defaultGroup (cumulative re-verification)
	"ENUM_VALUE_PREFIX",
	"ENUM_ZERO_VALUE_SUFFIX",
	"FILE_LOWER_SNAKE_CASE",
	"RPC_REQUEST_RESPONSE_UNIQUE",
	"RPC_REQUEST_STANDARD_NAME",
	"RPC_RESPONSE_STANDARD_NAME",
	"PACKAGE_VERSION_SUFFIX",
	"SERVICE_SUFFIX",
	//	commentsGroup — this plan
	"COMMENT_ENUM",
	"COMMENT_ENUM_VALUE",
	"COMMENT_FIELD",
	"COMMENT_MESSAGE",
	"COMMENT_ONEOF",
	"COMMENT_RPC",
	"COMMENT_SERVICE",
}

// presenceFiles pin the basic presence semantics: every documented
// declaration stays clean, every undocumented one issues at its keyword
// position. All names are rule-clean for the other 32 rules (exact standard
// rpc names, PascalCase identifiers, lower-snake fields, UPPER_SNAKE enum
// values, <NAME>_UNSPECIFIED zero values, "Service" suffix) so the COMMENT_*
// verdicts are the only findings. Oneof member fields are deliberately
// undocumented: they are outside BOTH engines' field walks (main iterates
// MessageBody.Fields; the port iterates Message.Fields), so they never fire
// COMMENT_FIELD — the undocumented member proves the boundary.
//
// Hand-derived old-side count: 1 (COMMENT_ENUM UndocEnum) + 2
// (COMMENT_ENUM_VALUE both UNDOC values) + 1 (COMMENT_MESSAGE UndocMessage)
// + 1 (COMMENT_FIELD undoc_field) + 1 (COMMENT_ONEOF undoc_choice) + 1
// (COMMENT_SERVICE UndocService) + 1 (COMMENT_RPC UndocMethod) = 8.
var presenceFiles = map[string]string{
	"pb.v1/presence.proto": `
syntax = "proto3";

package pb.v1;

// documented enum
enum DocEnum {
  // documented value
  DOC_ENUM_UNSPECIFIED = 0;
  // documented value
  DOC_ENUM_OTHER = 1;
}

enum UndocEnum {
  UNDOC_ENUM_UNSPECIFIED = 0;
  UNDOC_ENUM_OTHER = 1;
}

// documented request
message DocMethodRequest {}
// documented response
message DocMethodResponse {}
// documented request
message UndocMethodRequest {}
// documented response
message UndocMethodResponse {}

// documented message
message DocMessage {
  // documented field
  string doc_field = 1;

  // documented oneof
  oneof doc_choice {
    // documented oneof member
    string doc_member = 2;
  }
}

message UndocMessage {
  string undoc_field = 1;

  oneof undoc_choice {
    string undoc_member = 3;
  }
}

// documented service
service DocService {
  // documented rpc
  rpc DocMethod(DocMethodRequest) returns (DocMethodResponse);
}

service UndocService {
  rpc UndocMethod(UndocMethodRequest) returns (UndocMethodResponse);
}
`,
}

// directedFiles is the pivotal LINT-PARITY-02 cell: the message's ONLY
// preceding comment is "easyp:disable MESSAGE_PASCAL_CASE" — a directive for
// a DIFFERENT rule, so the central suppression machinery owns it and the
// COMMENT_* verdict is unmasked. Main reads the declaration as DOCUMENTED
// (len(Comments)>0): no finding anywhere. The lab's directive-filtering
// helper would call it undocumented and fabricate a COMMENT_MESSAGE finding —
// the tooth this cell kills.
//
// The undirected sibling corpus (same file without the comment) proves the
// directive actually does central work in BOTH engines: without it the
// corpus produces 2 findings (MESSAGE_PASCAL_CASE on some_message +
// COMMENT_MESSAGE), with it exactly 0 — suppression AND documentation
// coexist in one fixture pair.
//
// Hand-derived old-side counts: directed 0, undirected 2.
var directedFiles = map[string]string{
	"dir.v1/directive.proto": `
syntax = "proto3";

package dir.v1;

// easyp:disable MESSAGE_PASCAL_CASE
message some_message {}
`,
}

var undirectedFiles = map[string]string{
	"dir.v1/directive.proto": `
syntax = "proto3";

package dir.v1;

message some_message {}
`,
}

// detachedFiles pin the Leading+Detached composition of the predicate: a
// comment separated from the declaration by a blank line is protoast
// Detached, but go-protoparser's ParseComments still attaches it to
// Comments — main calls the declaration DOCUMENTED (no finding). The lab's
// leadingCommentDocuments reads Leading only and would call it undocumented —
// the tooth.
//
// Hand-derived old-side count: 0.
var detachedFiles = map[string]string{
	"det.v1/detached.proto": `
syntax = "proto3";

package det.v1;

// this comment is detached from the message below

message DetachedMessage {}
`,
}

// trailingFiles pin the exclusion of inline trailing comments: a same-line
// comment lands in the separate InlineComment field (go-protoparser) /
// Trailing (protoast), never in Comments — main calls the declaration
// UNDOCUMENTED and issues. Covers the message, enum and enum-value walks.
//
// Hand-derived old-side count: 1 (COMMENT_MESSAGE TrailMessage) + 1
// (COMMENT_ENUM TrailEnum) + 1 (COMMENT_ENUM_VALUE TRAIL_ENUM_UNSPECIFIED) = 3.
var trailingFiles = map[string]string{
	"trail.v1/trailing.proto": `
syntax = "proto3";

package trail.v1;

message TrailMessage {} // trailing only

enum TrailEnum {
  TRAIL_ENUM_UNSPECIFIED = 0; // trailing only
}
`,
}

// shallowFiles pin main's shallow walk shapes. nested.proto: a NESTED
// message with undocumented message/enum/field/oneof children two levels
// deep — nothing reports (main never recurses past one level for enums and
// never into nested messages at all). toplevel_enum.proto: an enum inside a
// TOP-LEVEL message IS reported by COMMENT_ENUM and COMMENT_ENUM_VALUE (main's
// +1 level) — the exact boundary.
//
// Hand-derived old-side count: 0 (nested.proto) + 2 (toplevel_enum.proto:
// COMMENT_ENUM HolderEnum + COMMENT_ENUM_VALUE HOLDER_ENUM_UNSPECIFIED) = 2.
var shallowFiles = map[string]string{
	"sh.v1/nested.proto": `
syntax = "proto3";

package sh.v1;

// documented outer
message OuterHolder {
  message NestedInner {
    string nested_field = 1;

    oneof nested_oneof {
      string nested_member = 2;
    }

    enum NestedEnum {
      NESTED_ENUM_UNSPECIFIED = 0;
    }
  }
}
`,
	"sh.v1/toplevel_enum.proto": `
syntax = "proto3";

package sh.v1;

// documented holder
message EnumHolder {
  enum HolderEnum {
    HOLDER_ENUM_UNSPECIFIED = 0;
  }
}
`,
}

// groupFiles pin the proto2 group behavior (research A4): the group field
// GrpFoo and its body field grp_field are outside BOTH engines' field walks
// (main MessageBody.Fields / port Message.Fields; groups are listed
// separately in both models), so CommentField reports NOTHING for them —
// no "!field.Group" guard is needed or present (the lab's guard is not
// ported). The holder message is documented so the cell isolates the group
// question.
//
// Hand-derived old-side count: 0.
var groupFiles = map[string]string{
	"grp.v1/group_fixture.proto": `
syntax = "proto2";

package grp.v1;

// documented holder
message GroupHolder {
  optional group GrpFoo = 1 {
    optional string grp_field = 2;
  }
}
`,
}

// TestLinkedParityComments pins old-vs-new parity for the COMMENTS batch. The
// requirePorted guard is the RED assertion: it fails listing the unported
// rules while the package still compiles, and goes green as the batch's ports
// land.
func TestLinkedParityComments(t *testing.T) {
	// RED guard — must stay the first statement of the test body.
	requirePorted(t,
		"COMMENT_ENUM",
		"COMMENT_ENUM_VALUE",
		"COMMENT_FIELD",
		"COMMENT_MESSAGE",
		"COMMENT_ONEOF",
		"COMMENT_RPC",
		"COMMENT_SERVICE",
	)

	t.Parallel()

	cfg := config.LintConfig{Use: commentsUse}

	t.Run("presence_basics", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, presenceFiles, 8)
	})

	t.Run("directive_only_documents", func(t *testing.T) {
		t.Parallel()

		// The directive documents the declaration for the COMMENT_* verdict
		// AND suppresses MESSAGE_PASCAL_CASE centrally — 0 findings total.
		assertBatchParity(t, cfg, directedFiles, 0)

		// The undirected sibling: the same declaration without the comment
		// produces the casing finding + the undocumented-message finding in
		// both engines — proof the directive's 0 above is central
		// suppression + documentation working together, not a silent config.
		assertBatchParity(t, cfg, undirectedFiles, 2)
	})

	t.Run("detached_comment_documents", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, detachedFiles, 0)
	})

	t.Run("trailing_only_undocumented", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, trailingFiles, 3)
	})

	t.Run("shallow_shapes", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, shallowFiles, 2)
	})

	t.Run("proto2_group_field", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, groupFiles, 0)
	})
}
