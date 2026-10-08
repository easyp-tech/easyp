package modules

import (
	"fmt"
	"slices"
	"strings"
)

// tidyImportBindings retains verified old/current bindings for token-only
// proposals. Reports refer to physical destinations and are sorted for commit.
type tidyImportBindings map[string]tidyImportBinding

func (bindings tidyImportBindings) planRewrites(tx *resolvedFilesTransaction, files []string, view *tidySourceView) (TidyResult, error) {
	var result TidyResult
	seen := make(map[string]bool)
	for _, name := range files {
		before, captured := tx.capturedFile(name)
		if !captured {
			return TidyResult{}, fmt.Errorf("consumer source %q was not captured during planning", name)
		}
		target := before.path
		if seen[target] {
			continue
		}
		seen[target] = true
		updated, mappings, err := rewriteProtoImportLiterals(name, before.data, func(importPath string) (string, error) {
			binding, exists := bindings[importPath]
			if !exists {
				return importPath, nil
			}
			return binding.replacement(name, importPath)
		})
		if err != nil {
			return TidyResult{}, fmt.Errorf("rewriteProtoImportLiterals: %w", err)
		}
		if len(mappings) == 0 {
			continue
		}
		err = tx.plan(name, updated, 0o644)
		if err != nil {
			return TidyResult{}, fmt.Errorf("plan: %w", err)
		}
		view.propose(target, updated)
		for _, mapping := range mappings {
			binding := bindings[mapping[0]]
			result.Imports = append(result.Imports, ImportRewrite{
				File: target, From: mapping[0], To: mapping[1], Module: binding.before.Lock.Source,
				OldVersion: binding.before.Lock.Version, Version: binding.after.Lock.Version,
				OldCommit: binding.before.Lock.Commit, Commit: binding.after.Lock.Commit,
				OldRoots: slices.Clone(binding.before.Module.Roots), Roots: slices.Clone(binding.after.Module.Roots),
			})
		}
	}
	slices.SortFunc(result.Imports, func(before, after ImportRewrite) int {
		if order := strings.Compare(before.File, after.File); order != 0 {
			return order
		}
		if order := strings.Compare(before.From, after.From); order != 0 {
			return order
		}
		return strings.Compare(before.To, after.To)
	})
	return result, nil
}
