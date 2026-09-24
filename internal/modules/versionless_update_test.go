package modules

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const (
	versionlessCommitA    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	versionlessCommitANew = "dddddddddddddddddddddddddddddddddddddddd"
	versionlessCommitBOld = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	versionlessCommitBNew = "cccccccccccccccccccccccccccccccccccccccc"
	versionlessHash       = "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
)

func TestTransitiveVersionlessRequirementRemainsUpdatable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		get                      bool
		transitiveVersion        string
		changeUpstream           bool
		updatedTransitiveVersion string
		wantManifestB            string
		wantUpdatedManifestB     string
		wantUpdatedB             string
	}{
		{name: "tidy keeps versionless indirect", wantManifestB: "require example.com/B // indirect\n", wantUpdatedB: versionlessCommitBNew},
		{name: "get keeps versionless indirect", get: true, wantManifestB: "require example.com/B // indirect\n", wantUpdatedB: versionlessCommitBNew},
		{name: "explicit transitive commit stays pinned", transitiveVersion: versionlessCommitBOld, wantManifestB: "require example.com/B " + versionlessCommitBOld + " // indirect\n", wantUpdatedB: versionlessCommitBOld},
		{
			name:                     "changed upstream commit replaces derived pin",
			transitiveVersion:        versionlessCommitBOld,
			changeUpstream:           true,
			updatedTransitiveVersion: versionlessCommitBNew,
			wantManifestB:            "require example.com/B " + versionlessCommitBOld + " // indirect\n",
			wantUpdatedManifestB:     "require example.com/B " + versionlessCommitBNew + " // indirect\n",
			wantUpdatedB:             versionlessCommitBNew,
		},
		{
			name:                 "upstream removes commit pin",
			transitiveVersion:    versionlessCommitBOld,
			changeUpstream:       true,
			wantManifestB:        "require example.com/B " + versionlessCommitBOld + " // indirect\n",
			wantUpdatedManifestB: "require example.com/B // indirect\n",
			wantUpdatedB:         versionlessCommitBNew,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			manifest := "module example.com/app\n"
			if !tt.get {
				manifest += "require example.com/A\n"
			}
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), []byte(manifest), 0o600))
			repository := &movingHeadRepository{
				cacheDir: t.TempDir(), bHead: versionlessCommitBOld,
				aModule: v1.Module{Name: "example.com/A", Roots: []string{"."}, Requires: []v1.Requirement{{Module: "example.com/B", Version: tt.transitiveVersion}}},
				bModule: v1.Module{Name: "example.com/B", Roots: []string{"."}},
			}
			for _, name := range []string{"A", "B"} {
				directory := filepath.Join(repository.cacheDir, name)
				require.NoError(t, os.MkdirAll(directory, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(directory, name+".proto"), []byte("syntax = \"proto3\";\n"), 0o644))
			}

			var err error
			if tt.get {
				err = Get(t.Context(), root, v1.Requirement{Module: "example.com/A"}, repository)
			} else {
				err = Tidy(t.Context(), root, repository)
			}
			require.NoError(t, err)
			updatedManifest, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			assert.Contains(t, string(updatedManifest), tt.wantManifestB)

			repository.bHead = versionlessCommitBNew
			if tt.changeUpstream {
				repository.aModule.Requires[0].Version = tt.updatedTransitiveVersion
			}
			require.NoError(t, Update(t.Context(), root, repository))
			updatedManifest, err = os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			wantUpdatedManifestB := tt.wantUpdatedManifestB
			if wantUpdatedManifestB == "" {
				wantUpdatedManifestB = tt.wantManifestB
			}
			assert.Contains(t, string(updatedManifest), wantUpdatedManifestB)
			if tt.changeUpstream {
				assert.NotContains(t, string(updatedManifest), tt.wantManifestB)
			}
			lock, err := ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			for _, entry := range lock.Modules {
				if entry.Source == "example.com/B" {
					assert.Equal(t, tt.wantUpdatedB, entry.Commit)
					return
				}
			}
			t.Fatal("B is missing from protobuf.lock")
		})
	}
}

func TestGetRefreshesDerivedTransitiveRequirements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		updatedRequires []v1.Requirement
		wantB           string
	}{
		{
			name:            "updated upstream pin replaces derived requirement",
			updatedRequires: []v1.Requirement{{Module: "example.com/B", Version: versionlessCommitBNew}},
			wantB:           "require example.com/B " + versionlessCommitBNew + " // indirect\n",
		},
		{
			name: "removed upstream dependency removes derived requirement",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			manifest := "module example.com/app\nrequire example.com/A\n"
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), []byte(manifest), 0o600))
			repository := &movingHeadRepository{
				cacheDir: t.TempDir(),
				bHead:    versionlessCommitBOld,
				aModule: v1.Module{
					Name:     "example.com/A",
					Roots:    []string{"."},
					Requires: []v1.Requirement{{Module: "example.com/B", Version: versionlessCommitBOld}},
				},
				bModule: v1.Module{Name: "example.com/B", Roots: []string{"."}},
			}
			for _, name := range []string{"A", "B"} {
				directory := filepath.Join(repository.cacheDir, name)
				require.NoError(t, os.MkdirAll(directory, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(directory, name+".proto"), []byte("syntax = \"proto3\";\n"), 0o644))
			}

			require.NoError(t, Tidy(t.Context(), root, repository))
			before, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			oldB := "require example.com/B " + versionlessCommitBOld + " // indirect\n"
			assert.Contains(t, string(before), oldB)

			repository.bHead = versionlessCommitBNew
			repository.aModule.Requires = tt.updatedRequires
			require.NoError(t, Get(t.Context(), root, v1.Requirement{Module: "example.com/A", Version: versionlessCommitANew}, repository))

			after, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			assert.NotContains(t, string(after), oldB)
			if tt.wantB != "" {
				assert.Contains(t, string(after), tt.wantB)
			} else {
				assert.NotContains(t, string(after), "require example.com/B")
			}
		})
	}
}

type movingHeadRepository struct {
	cacheDir, bHead string
	aModule         v1.Module
	bModule         v1.Module
}

func (r *movingHeadRepository) Fetch(_ context.Context, source, version string) (Fetched, error) {
	switch source {
	case "example.com/A":
		if version == "" {
			version = versionlessCommitA
		}
		return Fetched{Module: r.aModule, Lock: v1.LockedModule{Source: source, Version: version, Commit: version, Hash: versionlessHash}}, nil
	case "example.com/B":
		if version == "" {
			version = r.bHead
		}
		return Fetched{Module: r.bModule, Lock: v1.LockedModule{Source: source, Version: version, Commit: version, Hash: versionlessHash}}, nil
	default:
		return Fetched{}, os.ErrNotExist
	}
}

func (r *movingHeadRepository) Install(context.Context, v1.Lock) error { return nil }

func (r *movingHeadRepository) Cached(entry v1.LockedModule) (string, v1.Module, error) {
	switch entry.Source {
	case "example.com/A":
		return filepath.Join(r.cacheDir, "A"), r.aModule, nil
	case "example.com/B":
		return filepath.Join(r.cacheDir, "B"), r.bModule, nil
	default:
		return "", v1.Module{}, os.ErrNotExist
	}
}

func (r *movingHeadRepository) Versions(context.Context, string) ([]string, error) {
	return nil, nil
}
