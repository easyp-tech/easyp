// Package bsr supplies explicit Git compatibility snapshots for known Buf modules.
package bsr

import (
	"context"
	"fmt"
	"log/slog"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

// StaticResolver resolves known modules without contacting BSR or guessing repositories.
// Logger is optional and reports that Git snapshots do not verify BSR pins or digests.
type StaticResolver struct {
	Logger logger.Logger
}

var _ modules.BSRResolver = StaticResolver{}

// These versions are deliberate compatibility snapshots, not BSR-to-Git revision mappings.
var knownModules = map[string]v1.Requirement{
	"buf.build/googleapis/googleapis":          {Module: "github.com/googleapis/googleapis", Version: "03a91044136a014466d4293eb1fe91f2b02075d2"},
	"buf.build/grpc-ecosystem/grpc-gateway":    {Module: "github.com/grpc-ecosystem/grpc-gateway", Version: "v2.31.0+incompatible"},
	"buf.build/mercari/grpc-federation":        {Module: "github.com/mercari/grpc-federation", Version: "v1.27.0"},
	"buf.build/envoyproxy/protoc-gen-validate": {Module: "github.com/envoyproxy/protoc-gen-validate", Version: "v1.3.3"},
	"buf.build/bufbuild/protovalidate":         {Module: "github.com/bufbuild/protovalidate", Version: "v1.2.2"},
}

// UnsupportedModuleError identifies a BSR module absent from the explicit mapping.
type UnsupportedModuleError struct {
	Module string
}

// Error explains why no Git source was selected.
func (err *UnsupportedModuleError) Error() string {
	return fmt.Sprintf("unsupported BSR module: %s; no mapping in the BSR resolver", err.Module)
}

// Resolve selects a fixed Git requirement and preserves all supplied BSR metadata.
func (resolver StaticResolver) Resolve(ctx context.Context, dependency v1.BSRDependency) (v1.BSRResolution, error) {
	if err := ctx.Err(); err != nil {
		return v1.BSRResolution{}, fmt.Errorf("Err: %w", err)
	}
	if err := dependency.Validate(); err != nil {
		return v1.BSRResolution{}, fmt.Errorf("Validate: %w", err)
	}
	git, ok := knownModules[dependency.Module]
	if !ok {
		return v1.BSRResolution{}, &UnsupportedModuleError{Module: dependency.Module}
	}
	if resolver.Logger != nil {
		resolver.Logger.Warn(ctx, "using BSR compatibility snapshot; BSR revision equivalence and digest are not verified",
			slog.String("bsr_module", dependency.Module), slog.String("bsr_reference", dependency.Reference), slog.String("bsr_commit", dependency.Commit),
			slog.String("bsr_digest", dependency.Digest), slog.String("config", dependency.Config), slog.String("git_module", git.Module), slog.String("git_version", git.Version))
	}
	return v1.BSRResolution{Dependency: dependency, Git: git, Resolution: v1.BSRCompatibilitySnapshot}, nil
}
