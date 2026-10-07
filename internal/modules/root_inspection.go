package modules

// RootInspection retains logical proto sources from an immutable Git revision.
// It owns its file bytes and remains usable after the temporary checkout is removed.
type RootInspection struct {
	Files    []RootProtoFile
	Problems []RootPathProblem
	// Provisional marks an unresolved metadata-free root selection. Its fetched
	// lock has no content hash and must be finalized before installation.
	Provisional bool
	// LegacyFiles is reserved for verified archive-to-repository logical mappings.
	LegacyFiles map[string]string
}

// RootProtoFile describes one logical proto path and its pinned physical identity.
// Aliases of the same Git file share Identity without sharing mutable Content bytes.
type RootProtoFile struct {
	Path     string
	Identity string
	Content  []byte
}

// RootPathProblem retains a logical path that could not be resolved safely.
// Err preserves the bounded source view's error identity for later root selection.
type RootPathProblem struct {
	Path string
	Err  error
}
