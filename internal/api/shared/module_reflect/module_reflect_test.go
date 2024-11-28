package modulereflect

import (
	"testing"

	"github.com/stretchr/testify/suite"

	mocks2 "github.com/easyp-tech/easyp/internal/api/shared/module_reflect/mocks"
)

type codeGenSuite struct {
	suite.Suite

	mod      *mocks2.Mod
	storage  *mocks2.Storage
	lockFile *mocks2.LockFile

	moduleReflect *ModuleReflect
}

func (s *codeGenSuite) SetupSuite() {
	t := s.T()

	s.mod = mocks2.NewMod(t)
	s.storage = mocks2.NewStorage(t)
	s.lockFile = mocks2.NewLockFile(t)

	s.moduleReflect = New(s.mod, s.storage, s.lockFile)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(codeGenSuite))
}
