package modulereflect

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/easyp-tech/easyp/internal/api/shared/module_reflect/mocks"
)

type codeGenSuite struct {
	suite.Suite

	mod      *mocks.Mod
	storage  *mocks.Storage
	lockFile *mocks.LockFile

	codeGen *ModuleReflect
}

func (s *codeGenSuite) SetupSuite() {
	t := s.T()

	s.mod = mocks.NewMod(t)
	s.storage = mocks.NewStorage(t)
	s.lockFile = mocks.NewLockFile(t)

	s.codeGen = New(s.mod, s.storage, s.lockFile)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(codeGenSuite))
}
