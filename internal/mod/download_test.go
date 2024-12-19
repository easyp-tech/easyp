package mod

import (
	"github.com/brianvoe/gofakeit/v6"

	"github.com/easyp-tech/easyp/internal/mod/models"
)

func getFakeModule() models.Module {
	module := models.Module{}
	_ = gofakeit.Struct(&module)

	return module
}
