package migration

import (
	"fmt"
	"maps"
	"slices"
)

// localSelectionProof retains the literal inputs and their proved generation
// selectors. The transaction captures the selected bytes and alias topology;
// rechecking this inventory also detects newly added or removed source names.
type localSelectionProof struct {
	module    string
	inputs    []legacyDirectory
	roots     []string
	selection localSourceSelection
}

func newLocalSelectionProof(tx *transaction, cfg legacyConfig, module string) (*localSelectionProof, error) {
	roots, inputs, err := localRoots(cfg)
	if err != nil {
		return nil, fmt.Errorf("localRoots: %w", err)
	}
	if err := checkManagedLocal(cfg, module, len(inputs) > 0); err != nil {
		return nil, fmt.Errorf("checkManagedLocal: %w", err)
	}
	selection, err := proveLocalSelection(tx.requestedRoot, inputs, roots)
	if err != nil {
		return nil, fmt.Errorf("proveLocalSelection: %w", err)
	}
	for _, source := range selection.files {
		if _, err := tx.captureInput(source); err != nil {
			return nil, fmt.Errorf("captureInput: %w", err)
		}
	}
	return &localSelectionProof{module: module, inputs: inputs, roots: roots, selection: selection}, nil
}

func (p *localSelectionProof) recheck(root string) error {
	if p == nil || len(p.inputs) == 0 {
		return nil
	}
	selection, err := proveLocalSelection(root, p.inputs, p.roots)
	if err != nil {
		return fmt.Errorf("proveLocalSelection: %w", err)
	}
	if !maps.Equal(selection.files, p.selection.files) || !slices.Equal(selection.packages, p.selection.packages) || !slices.Equal(selection.paths, p.selection.paths) {
		return fmt.Errorf("local .proto source selection changed since planning; preview again")
	}
	return nil
}
