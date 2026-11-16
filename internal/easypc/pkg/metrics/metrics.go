// Package metrics defines the easyp.compiler.metrics/v1 schema for the Go compiler.
package metrics

import (
	"encoding/json"
	"os"
)

const Schema = "easyp.compiler.metrics/v1"

// Canonical phase names, in pipeline order.
const (
	PhaseStartup    = "startup"     // process start → ready to read inputs
	PhaseRead       = "read"        // file I/O
	PhaseLex        = "lex"         // tokenization (if separate from parse)
	PhaseParse      = "parse"       // syntax → AST/IR
	PhaseDescriptor = "descriptor"  // AST/IR → descriptor model
	PhaseLink       = "link"        // symbol table, reference resolution
	PhaseOptions    = "options"     // option interpretation + validation
	PhaseSourceInfo = "source_info" // source_code_info generation
	PhaseSerialize  = "serialize"   // descriptor → wire bytes
	PhaseWrite      = "write"       // output I/O
	PhaseCompile    = "compile"     // used when phases cannot be separated
)

type Report struct {
	Schema  string  `json:"schema"`
	Impl    string  `json:"impl"`
	Version string  `json:"version,omitempty"`
	Mode    string  `json:"mode"`
	Threads int     `json:"threads"`
	Totals  Totals  `json:"totals"`
	Memory  *Memory `json:"memory,omitempty"`
	Files   []File  `json:"files,omitempty"`
}

type Totals struct {
	WallNs      int64            `json:"wall_ns"`
	Phases      map[string]int64 `json:"phases"`
	Files       int              `json:"files"`
	InputBytes  int64            `json:"input_bytes"`
	OutputBytes int64            `json:"output_bytes"`
}

type Memory struct {
	TotalAllocBytes int64 `json:"total_alloc_bytes,omitempty"`
	Mallocs         int64 `json:"mallocs,omitempty"`
	PeakHeapBytes   int64 `json:"peak_heap_bytes,omitempty"`
	GCCycles        int64 `json:"gc_cycles,omitempty"`
	GCPauseNs       int64 `json:"gc_pause_ns,omitempty"`
}

type File struct {
	Path        string           `json:"path"`
	Bytes       int64            `json:"bytes"`
	Lines       int64            `json:"lines,omitempty"`
	Phases      map[string]int64 `json:"phases"`
	TotalNs     int64            `json:"total_ns"`
	OutputBytes int64            `json:"output_bytes,omitempty"`
}

// Write writes the metrics Report to a JSON file.
func Write(path string, r *Report) error {
	r.Schema = Schema
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
