package moduleconfig

import "github.com/easyp-tech/easyp/internal/logger"

type (
	// ModuleConfig implement module config logic such as buf dirs config etc
	ModuleConfig struct {
		logger logger.Logger
	}
)

func New(logger logger.Logger) *ModuleConfig {
	return &ModuleConfig{logger: logger}
}
