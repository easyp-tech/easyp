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
// doubles in builtin_run_test.go invoke the raw wazero constructor directly
// (bypassing the production mutex); TestMain primes the lazy wazero version
// global before the race detector forks test goroutines, so the detector's
// signal comes exclusively from the production paths the mutex protects.
// Constructing one throwaway runtime here — on the
// TestMain goroutine, before m.Run forks the test goroutines — performs the
// single lazy write with a happens-before edge to every test body. That
// priming write removes the race signal for the rest of the process:
// GetWazeroVersion writes the global once, so every later call takes the
// initialized early return and never writes again — the -race suite cannot
// detect loss of runtimeCreationMu. Production runtime construction (which
// has no priming) is protected by the mutex, and its regression protection
// is pinned structurally by the construction probe in
// TestBuiltinRunRuntimeCreationIsSerializedAcrossScopes — a held-mutex
// ordering check (a construction requested while runtimeCreationMu is held
// must not complete until release) with the peak-concurrency overlap counter
// as a secondary detector.
func TestMain(m *testing.M) {
	ctx := context.Background()
	rt := newRuntimeSerialized(ctx, wazero.NewRuntimeConfig())
	if err := rt.Close(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: Close: %v\n", err)
	}
	os.Exit(m.Run())
}
