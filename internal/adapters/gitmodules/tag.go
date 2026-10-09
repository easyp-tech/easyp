package gitmodules

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/gitcommand"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// ResolveTag resolves a named Git tag to a commit before get writes a requirement.
// Semantic version selection continues to use Versions and Fetch.
func (c *Cache) ResolveTag(ctx context.Context, source, tag string) (resolved string, err error) {
	ctx = c.operationContext(ctx, source, tag)
	finish := gitcommand.Start(ctx, "resolve tag")
	defer func() { finish(err) }()
	if _, err := gitV1(ctx, "", "check-ref-format", "refs/tags/"+tag); err != nil {
		return "", fmt.Errorf("gitV1: %w", err)
	}
	candidates, err := v1GitModuleCandidates(source)
	if err != nil {
		return "", fmt.Errorf("v1GitModuleCandidates: %w", err)
	}
	var firstErr error
	foundRepository := false
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("Err: %w", err)
		}
		ref := "refs/tags/" + candidate.tag(tag)
		candidateContext := gitcommand.WithAttributes(ctx, slog.String("repository", candidate.remote), slog.String("module_directory", candidate.subdir))
		raw, err := gitV1(candidateContext, "", "ls-remote", "--tags", "--", candidate.remote, ref, ref+"^{}")
		if err != nil {
			if gitcommand.Interrupted(err) {
				return "", fmt.Errorf("gitV1: %w", err)
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		foundRepository = true
		var commit string
		for _, line := range strings.Split(raw, "\n") {
			hash, name, ok := strings.Cut(line, "\t")
			if !ok || !v1.IsCommitRef(hash) {
				continue
			}
			if name == ref+"^{}" {
				return hash, nil
			}
			if name == ref {
				commit = hash
			}
		}
		if commit != "" {
			return commit, nil
		}
	}
	if !foundRepository && firstErr != nil {
		return "", fmt.Errorf("could not query Git tags for %s@%s; check repository access and credentials: %w", source, tag, firstErr)
	}
	return "", fmt.Errorf("git tag %s@%s was not found", source, tag)
}
