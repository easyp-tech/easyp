// Command easyp-protoc is the Go implementation of the easyp .proto compiler.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/compiler"
)

func main() {
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(25)
	}
	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(650 * 1024 * 1024)
	}
	opts, err := compiler.ParseArgs(os.Args[1:])
	if err != nil {
		if usageErr, ok := err.(*compiler.UsageError); ok {
			fmt.Fprintf(os.Stderr, "%s\n", usageErr.Error())
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	driver := compiler.NewDriver(opts)
	if err := driver.Compile(); err != nil {
		if !errors.Is(err, compiler.ErrCompilationFailed) {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		os.Exit(1)
	}
}
