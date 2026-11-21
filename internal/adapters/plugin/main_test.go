package plugin

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/tetratelabs/wazero"
)

// TestMain primes wazero v1.9.0's internal/version lazy global before any
// test goroutine constructs a runtime. GetWazeroVersion() writes that package
// global on first use and is not safe under concurrent NewRuntimeWithConfig
// calls. Production runtime construction is serialized by the default
// runtimeFactory (newRuntimeSerialized), but the frozen runtimeFactory seam
// doubles in builtin_run_test.go call the raw wazero constructor directly, so
// no easyp-side mutex can order those calls against each other or against the
// serialized default. Constructing one throwaway runtime here — on the
// TestMain goroutine, before m.Run forks the test goroutines — performs the
// single lazy write with a happens-before edge to every test body: all later
// constructions observe a fully initialized global, and the -race detector
// then proves the production serialization for everything easyp constructs.
func TestMain(m *testing.M) {
	ctx := context.Background()
	rt := newRuntimeSerialized(ctx, wazero.NewRuntimeConfig())
	if err := rt.Close(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: Close: %v\n", err)
	}
	os.Exit(m.Run())
}
