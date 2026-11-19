package rules_test

// Batch-6 (UNARY_RPC) parity cells: RPC_NO_CLIENT_STREAMING and
// RPC_NO_SERVER_STREAMING join the thirty-nine already-ported rules under the
// cumulative full-registry config. The two ports are verbatim-from-lab
// (research inventory #40/#41): the streaming predicate reads
// rpc.Request.Stream / rpc.Response.Stream (the etalon's IsStream field
// renamed by main's protoast model) and the finding position is the
// declaration keyword (rpc.Meta.Pos ≡ rpc.Pos — main and etalon carried tests
// assert the identical 300/13/3 literal on the shared auth fixture, the
// alias-only test diff). Every cell asserts its expected old-side finding
// count before the multiset compare (T-04-02); fixture identifiers are
// rule-clean for the other 39 rules so the UNARY_RPC verdicts are the only
// findings.
//
// With the registry complete, every cell runs cfg.Use = rules.AllRuleNames()
// (the 04-05 precedent: cells double as full-registry regressions); the
// cumulative cell additionally packs all three streaming shapes into one
// corpus (1 client-only + 1 server-only + 1 both-streaming rpc = 4 findings).
//
// Old-side counts below are probe-confirmed against the live old engine
// before the RED run (see the probe evidence note in 04-07-SUMMARY).

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/rules"
)

// unaryRPCUse is every cell's config: the completed 41-name registry
// (MINIMAL + BASIC + DEFAULT + COMMENTS + UNARY_RPC), so each cell
// re-verifies the whole registry cumulatively.
var unaryRPCUse = rules.AllRuleNames()

// clientStreamingFiles pin the client-streaming predicate: the request-side
// stream keyword issues RPC_NO_CLIENT_STREAMING at the rpc keyword position,
// while the non-streaming counterpart rpc in the same file stays clean.
//
// Probe-confirmed old-side count: 1.
var clientStreamingFiles = map[string]string{
	"str/v1/client_streaming.proto": `
syntax = "proto3";

package str.v1;

// documented request
message FooRequest {}
// documented response
message FooResponse {}
// documented request
message CleanRequest {}
// documented response
message CleanResponse {}

// documented service
service TestService {
  // documented rpc
  rpc Foo(stream FooRequest) returns (FooResponse);

  // documented clean rpc
  rpc Clean(CleanRequest) returns (CleanResponse);
}
`,
}

// serverStreamingFiles pin the server-streaming predicate: the response-side
// stream keyword issues RPC_NO_SERVER_STREAMING at the rpc keyword position,
// while the non-streaming counterpart rpc in the same file stays clean.
//
// Probe-confirmed old-side count: 1.
var serverStreamingFiles = map[string]string{
	"srv/v1/server_streaming.proto": `
syntax = "proto3";

package srv.v1;

// documented request
message BarRequest {}
// documented response
message BarResponse {}
// documented request
message CleanRequest {}
// documented response
message CleanResponse {}

// documented service
service TestService {
  // documented rpc
  rpc Bar(BarRequest) returns (stream BarResponse);

  // documented clean rpc
  rpc Clean(CleanRequest) returns (CleanResponse);
}
`,
}

// bothStreamingFiles pin the both-streaming cell: one rpc with stream on
// BOTH the request and the response issues both rules, each at the same rpc
// keyword position (two findings, same spot, different rule names).
//
// Probe-confirmed old-side count: 2.
var bothStreamingFiles = map[string]string{
	"both/v1/both_streaming.proto": `
syntax = "proto3";

package both.v1;

// documented request
message BothRequest {}
// documented response
message BothResponse {}

// documented service
service TestService {
  // documented rpc
  rpc Both(stream BothRequest) returns (stream BothResponse);
}
`,
}

// cumulativeStreamingFiles pack all three streaming shapes into one corpus
// for the full-registry cumulative re-check: Get is client-streaming only,
// Put is server-streaming only, Sync is both — and the clean rpc proves the
// stream keyword is the trigger, not the rpc itself.
//
// Probe-confirmed old-side count: 4 (2 CLIENT + 2 SERVER).
var cumulativeStreamingFiles = map[string]string{
	"cum/v1/cumulative_streaming.proto": `
syntax = "proto3";

package cum.v1;

// documented request
message GetRequest {}
// documented response
message GetResponse {}
// documented request
message PutRequest {}
// documented response
message PutResponse {}
// documented request
message SyncRequest {}
// documented response
message SyncResponse {}
// documented request
message CleanRequest {}
// documented response
message CleanResponse {}

// documented service
service TestService {
  // documented rpc
  rpc Get(stream GetRequest) returns (GetResponse);

  // documented rpc
  rpc Put(PutRequest) returns (stream PutResponse);

  // documented rpc
  rpc Sync(stream SyncRequest) returns (stream SyncResponse);

  // documented clean rpc
  rpc Clean(CleanRequest) returns (CleanResponse);
}
`,
}

// TestLinkedParityUnaryRPC pins old-vs-new parity for the UNARY_RPC batch. The
// requirePorted guard is the RED assertion: it fails listing the unported
// rules while the package still compiles, and goes green as the batch's ports
// land.
func TestLinkedParityUnaryRPC(t *testing.T) {
	// RED guard — must stay the first statement of the test body.
	requirePorted(t,
		"RPC_NO_CLIENT_STREAMING",
		"RPC_NO_SERVER_STREAMING",
	)

	t.Parallel()

	cfg := config.LintConfig{Use: unaryRPCUse}

	t.Run("client_streaming", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, clientStreamingFiles, 1)
	})

	t.Run("server_streaming", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, serverStreamingFiles, 1)
	})

	t.Run("both_streaming", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, bothStreamingFiles, 2)
	})

	t.Run("cumulative", func(t *testing.T) {
		t.Parallel()

		assertBatchParity(t, cfg, cumulativeStreamingFiles, 4)
	})
}
