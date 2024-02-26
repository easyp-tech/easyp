package moduleconfig

type (
	// ModuleConfig implement module config logic such as buf dirs config etc
	ModuleConfig struct {
	}
)

func New() *ModuleConfig {
	return &ModuleConfig{}
}
