package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	"github.com/easyp-tech/easyp/internal/core/models"
)

func Test_generateModulePath(t *testing.T) {
	t.Parallel()

	const (
		weatherName    = "github.com/acme/weather"
		weatherVersion = "v1.2"
		otherName      = "github.com/acme/other"
		otherVersion   = "v1.0"
	)

	tests := []struct {
		name     string
		module   models.Module
		replaces func(localDir string) []modfile.Replace
		setup    func(t *testing.T, cacheDir, localDir string, storage *StorageMock, lockFile *LockFileMock)
		want     func(cacheDir, localDir string) string
		wantErr  bool
	}{
		{
			name:   "replace_hit_absolute_path",
			module: models.NewModule(weatherName + "@" + weatherVersion),
			replaces: func(localDir string) []modfile.Replace {
				return []modfile.Replace{{
					Module: models.NewModule(weatherName + "@" + weatherVersion),
					Path:   localDir,
				}}
			},
			setup: func(t *testing.T, cacheDir, localDir string, storage *StorageMock, lockFile *LockFileMock) {
				t.Helper()
			},
			want: func(cacheDir, localDir string) string {
				return filepath.Clean(localDir)
			},
		},
		{
			name:   "replace_hit_relative_path",
			module: models.NewModule(weatherName + "@" + weatherVersion),
			replaces: func(localDir string) []modfile.Replace {
				return []modfile.Replace{{
					Module: models.NewModule(weatherName + "@" + weatherVersion),
					Path:   "local",
				}}
			},
			setup: func(t *testing.T, cacheDir, localDir string, storage *StorageMock, lockFile *LockFileMock) {
				t.Helper()
			},
			want: func(cacheDir, localDir string) string {
				return localDir
			},
		},
		{
			name:   "replace_hit_omitted_version_uses_lock",
			module: models.NewModule(weatherName),
			replaces: func(localDir string) []modfile.Replace {
				return []modfile.Replace{{
					Module: models.NewModule(weatherName + "@" + weatherVersion),
					Path:   localDir,
				}}
			},
			setup: func(t *testing.T, cacheDir, localDir string, storage *StorageMock, lockFile *LockFileMock) {
				t.Helper()
				lockFile.EXPECT().Read(weatherName).Return(models.LockFileInfo{
					Name:    weatherName,
					Version: weatherVersion,
				}, nil).Once()
			},
			want: func(cacheDir, localDir string) string {
				return filepath.Clean(localDir)
			},
		},
		{
			name:   "replace_miss_uses_cache",
			module: models.NewModule(otherName + "@" + otherVersion),
			replaces: func(localDir string) []modfile.Replace {
				return []modfile.Replace{{
					Module: models.NewModule(weatherName + "@" + weatherVersion),
					Path:   localDir,
				}}
			},
			setup: func(t *testing.T, cacheDir, localDir string, storage *StorageMock, lockFile *LockFileMock) {
				t.Helper()
				lockFile.EXPECT().Read(otherName).Return(models.LockFileInfo{
					Name:    otherName,
					Version: otherVersion,
				}, nil).Once()
				storage.EXPECT().GetInstallDir(otherName, otherVersion).Return(cacheDir).Once()
			},
			want: func(cacheDir, localDir string) string {
				return cacheDir
			},
		},
		{
			name: "version_mismatch_uses_cache",
			module: models.NewModuleFromLockFileInfo(models.LockFileInfo{
				Name:    weatherName,
				Version: "v1.3",
			}),
			replaces: func(localDir string) []modfile.Replace {
				return []modfile.Replace{{
					Module: models.NewModule(weatherName + "@" + weatherVersion),
					Path:   localDir,
				}}
			},
			setup: func(t *testing.T, cacheDir, localDir string, storage *StorageMock, lockFile *LockFileMock) {
				t.Helper()
				lockFile.EXPECT().Read(weatherName).Return(models.LockFileInfo{
					Name:    weatherName,
					Version: "v1.3",
				}, nil).Once()
				storage.EXPECT().GetInstallDir(weatherName, "v1.3").Return(cacheDir).Once()
			},
			want: func(cacheDir, localDir string) string {
				return cacheDir
			},
		},
		{
			name:   "replace_path_missing",
			module: models.NewModule(weatherName + "@" + weatherVersion),
			replaces: func(localDir string) []modfile.Replace {
				return []modfile.Replace{{
					Module: models.NewModule(weatherName + "@" + weatherVersion),
					Path:   filepath.Join(filepath.Dir(localDir), "does-not-exist"),
				}}
			},
			setup: func(t *testing.T, cacheDir, localDir string, storage *StorageMock, lockFile *LockFileMock) {
				t.Helper()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			cacheDir := filepath.Join(root, "cache")
			localDir := filepath.Join(root, "local")
			require.NoError(t, os.MkdirAll(cacheDir, 0o755))
			require.NoError(t, os.MkdirAll(localDir, 0o755))

			storage := &StorageMock{}
			lockFile := &LockFileMock{}
			storage.Test(t)
			lockFile.Test(t)

			if tt.setup != nil {
				tt.setup(t, cacheDir, localDir, storage, lockFile)
			}

			c := &Core{
				storage:  storage,
				lockFile: lockFile,
				replaces: tt.replaces(localDir),
			}

			got, err := c.generateModulePath(root, tt.module)
			if tt.wantErr {
				require.Error(t, err)
				storage.AssertNotCalled(t, "GetInstallDir")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want(cacheDir, localDir), got)

			storage.AssertExpectations(t)
			lockFile.AssertExpectations(t)
		})
	}
}

func Test_generateModulePath_ImportList(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	localDir := filepath.Join(root, "local-weather")
	cacheWeather := filepath.Join(root, "cache-weather")
	cacheOther := filepath.Join(root, "cache-other")
	require.NoError(t, os.MkdirAll(localDir, 0o755))
	require.NoError(t, os.MkdirAll(cacheWeather, 0o755))
	require.NoError(t, os.MkdirAll(cacheOther, 0o755))

	weather := models.LockFileInfo{
		Name:    "github.com/acme/weather",
		Version: "v1.2",
	}
	other := models.LockFileInfo{
		Name:    "github.com/acme/other",
		Version: "v1.0",
	}

	storage := &StorageMock{}
	lockFile := &LockFileMock{}
	storage.Test(t)
	lockFile.Test(t)

	lockFile.EXPECT().Read(other.Name).Return(other, nil).Once()
	storage.EXPECT().GetInstallDir(other.Name, other.Version).Return(cacheOther).Once()

	c := &Core{
		storage:  storage,
		lockFile: lockFile,
		replaces: []modfile.Replace{{
			Module: models.NewModule("github.com/acme/weather@v1.2"),
			Path:   localDir,
		}},
	}

	var imports []string
	for _, info := range []models.LockFileInfo{weather, other} {
		modulePath, err := c.generateModulePath(root, models.NewModuleFromLockFileInfo(info))
		require.NoError(t, err)
		imports = append(imports, modulePath)
	}

	require.Equal(t, []string{filepath.Clean(localDir), cacheOther}, imports)
	require.NotContains(t, imports, cacheWeather)
	storage.AssertNotCalled(t, "GetInstallDir", weather.Name, weather.Version)
	storage.AssertExpectations(t)
	lockFile.AssertExpectations(t)
}
