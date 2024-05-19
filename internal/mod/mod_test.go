package mod

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/easyp-tech/easyp/internal/mod/mocks"
)

type modSuite struct {
	suite.Suite

	storage      *mocks.Storage
	moduleConfig *mocks.ModuleConfig
	lockFile     *mocks.LockFile

	mod *Mod
}

func (s *modSuite) SetupTest() {
	t := s.T()

	s.storage = mocks.NewStorage(t)
	s.moduleConfig = mocks.NewModuleConfig(t)
	s.lockFile = mocks.NewLockFile(t)

	s.mod = New(s.storage, s.moduleConfig, s.lockFile)
}

func TestRunModSuite(t *testing.T) {
	suite.Run(t, new(modSuite))
}

func Test_filterOnlyProtoDirs(t *testing.T) {
	tests := map[string]struct {
		files    []string
		expected []string
	}{
		"one dir": {
			files: []string{
				"collect/file.proto",
				"collect/nested/file.proto",
			},
			expected: []string{
				"collect",
			},
		},
		"several dir": {
			files: []string{
				"collect/file.proto",
				"collect/nested/file.proto",
				"storage/file.proto",
				"storage/111/file.proto",
				"storage/222/file.proto",
				"wo_proto/file.txt",
			},
			expected: []string{
				"collect",
				"storage",
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			result := filterOnlyProtoDirs(tc.files)
			require.ElementsMatch(t, tc.expected, result)
		})
	}
}
